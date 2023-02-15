package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gorilla/websocket"
)



type Server struct {
	fyne.App
	websocket.Upgrader
	*torrent.Client

	SearchManager

	MainTorrent  string
	MainFile     string
	AppIsClosing bool
}

func main() {
	var server Server
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(pwd)
	w64Dir        := path.Join(pwd,"internal","w64system")
	w64filePrefix := "w64system"
	if err := server.SearchManager.Init(w64Dir, w64filePrefix); err != nil {
		log.Fatalf("initializing search manager: %v", err)
	}

	server.App = app.New()
	server.App.Settings().SetTheme(&myTheme{})
	server.App.SetIcon(resourceAppiconPng)

	mainwin := server.App.NewWindow("123movies")
	mainwin.Resize(fyne.NewSize(400, 710))

	go server.startWebsocket()
	go server.startServer()
	server.AppIsClosing = false

	go server.initmainclient()
	server.LoadSettings()

	tabs := container.NewAppTabs(
		container.NewTabItem("Home", server.homeScreen(mainwin)),
		//container.NewTabItem("Settings",  settingsScreen(myWindow)),
	)

	tabs.SetTabLocation(container.TabLocationTop)
	mainwin.SetContent(tabs)
	mainwin.ShowAndRun() // dwell until exit

	server.AppIsClosing = true
}

func (s *Server) homeScreen(win fyne.Window) fyne.CanvasObject {
	data := binding.BindStringList(
		//&[]string{"Item 1", "Item 2", "Item 3"},
		&[]string{},
	)

	list := widget.NewListWithData(data,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	add := widget.NewButton("Open New Webapp Tab", func() {
		//val := fmt.Sprintf("Item %d", data.Length()+1)
		//data.Append(val)
		s.openNewWebappTab()
	})

	return container.NewBorder(add, nil, nil, nil, list)
}

func (s *Server) openNewWebappTab() {
	u, err := url.Parse("http://localhost:8080/core/core.html")
	if err != nil {
		fmt.Printf("parsing URL: %v", err)
	}

	err = s.App.OpenURL(u)
	if err != nil {
		fmt.Printf("opening URL: %v", err)
	}
}

func (s *Server) startServer() {
	s.openNewWebappTab()

	fs := http.FileServer(http.Dir(path.Join("internal","Webapp")))
	http.Handle("/", http.StripPrefix("/", fs))

	fmt.Println(http.ListenAndServe(":8080", nil))
}

func (s *Server) startWebsocket() {
	http.HandleFunc("/websocket", s.handleWebSocket)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
	conn, err := s.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade failed: ", err)
		return
	}

	defer conn.Close()

	// Continuosly read and write message
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read failed:", err)
			//break
			//App.Quit()
		}
		messagestring := string(message)
		messageArr := strings.Split(messagestring, "*")
		log.Println("got:", messagestring)

		returnmessagestring := s.runCmd(messageArr) //[]byte("return message")
		err = conn.WriteMessage(mt, []byte(returnmessagestring))
		if err != nil {
			log.Println("write failed:", err)
			//break
		}
	}
}

type ServerCommand = string

const (
	GetSearchResult    ServerCommand = "GETSEARCHRESULT"
	SetSearchQuery                   = "SETSEARCHQUERY"
	SetMainTorrent                   = "SETMAINTORRENT"
	SetMainFile                      = "SETMAINFILE"
	AddSavedItem                     = "ADDSAVEDITEM"
	RemoveSavedItem                  = "REMOVESAVEDITEM"
	RequestTorrentInfo               = "REQUESTTORRENTINFO"
	RequestIsSavedItem               = "REQUESTISSAVEDITEM"
)

func (s *Server) runCmd(messageArr []string) string {
	if len(messageArr) == 0 {
		return "Unkown command"
	}

	switch messageArr[0] {
	case GetSearchResult:
		tmpint, cierr := strconv.Atoi(messageArr[1])
		if cierr != nil {
			return "Unkown command"
		}

		if tmpint >= len(SearchResults) {
			go s.MoreSearchResults(s)
			return "SEARCHRESULTNOTFOUND"
		}

		return s.GetSearchResult(tmpint)
	//setSearchQuery
	case SetSearchQuery:
		PreviewingTorrentMagnetArr = PreviewingTorrentMagnetArr[:0]
		EmptySearchResults()
		go s.SetSearchQuery(s, messageArr[1])
	case SetMainTorrent:
		s.SetMainTorrent(messageArr[1])
		if len(messageArr) > 2 {
			s.SetMainFile(messageArr[2])
		}
	case SetMainFile:
		s.SetMainFile(messageArr[1])
	case AddSavedItem:
		s.AddSavedItem(messageArr[1], messageArr[2], messageArr[3], messageArr[4])
	case RemoveSavedItem:
		s.RemoveSavedItem(messageArr[1])
	case RequestTorrentInfo:
		if len(messageArr) > 1 {
			return s.getTorrentInfoResponse(messageArr[1])
		}
	case RequestIsSavedItem:
		return s.getIsSavedItemResponse(messageArr[1])
	default:
		fmt.Println("Unkown command")
	}

	return "return message"
}

func (s *Server) initmainclient() (err error) {
	cfg := torrent.NewDefaultClientConfig()
	cfg.Seed = true
	cfg.DataDir = path.Join("internal","Webapp", "core", "torrents") //***************
	cfg.DisableAggressiveUpload = false
	cfg.DisableWebtorrent = false
	cfg.DisableWebseeds = false

	if s.Client, err = torrent.NewClient(cfg); err != nil {
		log.Print("new torrent client: %w", err)
		return //fmt.Errorf("new torrent client: %w", err)
	}

	log.Print("new torrent client INITIATED")

	for !s.AppIsClosing {
		time.Sleep(1 * time.Second)
	}

	log.Print("closing mainclient")
	s.Client.Close()

	return nil
}

func (s *Server) SetMainFile(tmpfilepath string) {
	s.MainFile = tmpfilepath
	s.Prioritize(s.MainTorrent, tmpfilepath)
}

func (s *Server) SetMainTorrent(magnet string) {
	if (!s.IsMainTorrent(magnet)) && (!IsSavedItemWithMagnet(magnet)) {
		s.MainTorrent = magnet

		for {
			if (s.Client != nil) && (!s.AppIsClosing) {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
}
func (s *Server) addtorrent(tmpname string, tmpdescription string, tmpmagneturi string) {
	t, err := s.AddMagnet(tmpmagneturi)
	if err != nil {
		log.Print("new torrent error: %w", err)
	}

	<-t.GotInfo()

	log.Printf("added magnet %s\n", tmpmagneturi)
	files := t.Files()
	totalsize := int64(0)
	tmppreviewfile := ""
	tmppreviewfilesize := int64(0)

	for _, filei := range files {
		if (filei.Length() > tmppreviewfilesize) && (strings.Contains(filei.Path(), ".mp4")) {
			tmppreviewfile = filei.Path()
			totalsize += filei.Length()
		}
	}

	for _, filei := range files {
		if tmppreviewfile == filei.Path() {
			firstprioritizedpiece := int(filei.BeginPieceIndex())
			lastprioritizedpiece := CustomMin(firstprioritizedpiece+20, int(filei.EndPieceIndex()))

			t.DownloadPieces(firstprioritizedpiece, lastprioritizedpiece)
			t.CancelPieces(lastprioritizedpiece, filei.EndPieceIndex())
		} else {
			filei.SetPriority(torrent.PiecePriorityNone)
		}
	}

	AddPreviewingTorrent(tmpmagneturi)
	AddSearchResultItem(tmpname+" "+PrettyBytes(totalsize), tmpdescription, tmpmagneturi, tmppreviewfile)

	for {
		if (!IsSavedItemWithMagnet(tmpmagneturi)) && (!s.IsMainTorrent(tmpmagneturi)) && (!IsPreviewingTorrent(tmpmagneturi)) {
			log.Println("Torrent removed", tmpmagneturi)
			t.Drop()
			return
		}
		time.Sleep(8 * time.Second)
	}
}

func (s *Server) getIsSavedItemResponse(tmpitemmagnet string) string {
	var tmpreturnstring = "ISSAVEDITEM*" + tmpitemmagnet

	if IsSavedItemWithMagnet(tmpitemmagnet) {
		tmpreturnstring += "*TRUE"
	} else {
		tmpreturnstring += "*FALSE"
	}

	return tmpreturnstring

}
func (s *Server) getTorrentInfoResponse(tmpmagneturi string) string {
	fmt.Printf("REQUESTTORRENTINFO %s \n", tmpmagneturi)
	var tmpreturnstring = "TORRENTINFO"
	tmpmagnet, perr := metainfo.ParseMagnetUri(tmpmagneturi)
	_ = perr

	if perr != nil {
		return ""
	}

	t, ok := s.Torrent(tmpmagnet.InfoHash)
	if !ok {
		return ""
	}

	if t == nil {
		return ""
	}

	if t.Info() == nil {
		return ""
	}

	files := t.Files()
	if files == nil {
		return ""
	}

	tmpreturnstring += "*" + tmpmagneturi
	tmpreturnstring += "*" + "TORRENTNAME"
	tmpreturnstring += "*" + fmt.Sprintf("%d", len(t.PeerConns())) //"333"//nbpeers

	for _, filei := range files {
		tmpreturnstring += "*" + fmt.Sprintf("%s*%d", filei.Path(), filei.BytesCompleted()*100/filei.Length())
	}
	fmt.Printf("*** %s\n", tmpreturnstring)
	return tmpreturnstring
}

func (s *Server) Prioritize(tmpmagneturi string, filepath string) {
	tmpmagnet, perr := metainfo.ParseMagnetUri(tmpmagneturi)
	_ = perr
	t, ok := s.Torrent(tmpmagnet.InfoHash)
	_ = ok
	if !ok {
		return
	}

	files := t.Files()
	for _, filei := range files {
		if filepath == filei.Path() {
			filei.SetPriority(torrent.PiecePriorityNormal)
		} else {
			filei.SetPriority(torrent.PiecePriorityNone)
		}
	}
	fmt.Printf("***\n")
}

func (s *Server) IsMainTorrent(magnet string) bool {
	return s.MainTorrent == magnet

}

type SettingsType struct {
	LocalHostPort int
	SavedItems    []ItemType
}

var Settings SettingsType

type ItemType struct {
	//Path string
	Name        string
	Description string
	Magnet      string
	PreviewFile string
}

var PreviewingTorrentMagnetArr []string
var SearchResults []ItemType

func SearchResultsFull() bool {
	if len(SearchResults) > 5 {
		return true
	}
	return false
}
func AddSearchResultItem(tmpname string, tmpdescription string, tmpmagneturi string, tmppreviewfile string) {
	NewItem := new(ItemType)
	NewItem.Name = tmpname
	NewItem.Description = tmpdescription
	NewItem.Magnet = tmpmagneturi
	NewItem.PreviewFile = tmppreviewfile
	SearchResults = append(SearchResults, *NewItem)

}
func (s *Server) GetSearchResult(index int) string {
	var tmpsearchresultstring = "SEARCHRESULT"
	if len(SearchResults) <= index {
		return "NOSEARCHRESULTFOUND"
	}

	tmpsearchresultstring += "*" + SearchResults[index].Name
	tmpsearchresultstring += "*" + SearchResults[index].Description
	tmpsearchresultstring += "*" + SearchResults[index].Magnet
	tmpsearchresultstring += "*" + SearchResults[index].PreviewFile

	return tmpsearchresultstring
}
func EmptySearchResults() {
	SearchResults = SearchResults[:0]
}
func LoadDefaultSettings() {
	Settings.LocalHostPort = 666

}

func (s *Server) LoadSettings() {
	SettingsBytes, err := os.ReadFile("Settings.json") // just pass the file name
	if err != nil {
		fmt.Println("error:", err)
		LoadDefaultSettings()
		return
	}
	NewSettings := new(SettingsType)
	uerr := json.Unmarshal(SettingsBytes, NewSettings)
	if uerr != nil {
		fmt.Println("unmarshal error:", uerr)
		LoadDefaultSettings()
		return
	}
	Settings = *NewSettings

}

func SaveSettings() {
	f, err := os.Create("Settings.json")

	defer f.Close()

	//d2 := []byte{115, 111, 109, 101, 10}
	SettingsBytes, merr := json.Marshal(Settings)
	if merr != nil {
		fmt.Println("marshal error:", err)
		return
	}
	_, werr := f.Write(SettingsBytes)
	if werr != nil {
		fmt.Println("error:", werr)
		return
	}
	fmt.Printf("wrote settings\n")

}

func AddPreviewingTorrent(tmpmagnet string) {
	PreviewingTorrentMagnetArr = append(PreviewingTorrentMagnetArr, tmpmagnet)
}
func IsPreviewingTorrent(magnet string) bool {
	for _, tmpe := range PreviewingTorrentMagnetArr {
		if tmpe == magnet {
			return true
		}
	}
	return false
}

func IsSavedItemWithMagnet(magnet string) bool {
	for _, tmpe := range Settings.SavedItems {
		if tmpe.Magnet == magnet {
			return true
		}
	}
	return false
}

func (s *Server) AddSavedItem(itemname string, itemdescription string, itemmagnet string, itempreviewfile string) {
	var tmpsaveditem ItemType

	tmpsaveditem.Name = itemname
	tmpsaveditem.Description = itemdescription
	tmpsaveditem.Magnet = itemmagnet
	tmpsaveditem.PreviewFile = itempreviewfile

	Settings.SavedItems = append(Settings.SavedItems, tmpsaveditem)
}
func (s *Server) RemoveSavedItem(itemmagnet string) {

	Settings.SavedItems = s.removefromsaveditems(Settings.SavedItems, itemmagnet)
}
func (s *Server) removefromsaveditems(slice []ItemType, itemmagnet string) []ItemType {
	for i, tmpe := range slice {
		if tmpe.Magnet == itemmagnet {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func PrettyBytes(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
func CustomMin(i int, j int) int {
	if i > j {
		return j
	} else {
		return i
	}

}
