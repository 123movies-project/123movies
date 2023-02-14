package main

import (
	"log"
	"os"
	"time"
	"fmt"
	"path/filepath"
	//"github.com/anacrolix/tagflag"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

var builtinAnnounceList = [][]string{
	{"http://p4p.arenabg.com:1337/announce"},
	{"udp://tracker.opentrackr.org:1337/announce"},
	{"udp://tracker.openbittorrent.com:6969/announce"},
}

func main() {
/*
	log.SetFlags(log.Flags() | log.Lshortfile)
	var args struct {
		AnnounceList      []string `name:"a" help:"extra announce-list tier entry"`
		EmptyAnnounceList bool     `name:"n" help:"exclude default announce-list entries"`
		Comment           string   `name:"t" help:"comment"`
		CreatedBy         string   `name:"c" help:"created by"`
		InfoName          *string  `name:"i" help:"override info name (defaults to ROOT)"`
		Url               []string `name:"u" help:"add webseed url"`
		tagflag.StartPos
		Root string
	}
	tagflag.Parse(&args, tagflag.Description("Creates a torrent metainfo for the file system rooted at ROOT, and outputs it to stdout."))
*/
	tmpComment:="Cool torrent description"
	tmpCreatedBy:="coolboys"
	tmpInfoName:="CoolInfoName"
	mi := metainfo.MetaInfo{
		AnnounceList: builtinAnnounceList,
	}
	//if args.EmptyAnnounceList {
	//	mi.AnnounceList = make([][]string, 0)
	//}
	//for _, a := range args.AnnounceList {
	///	mi.AnnounceList = append(mi.AnnounceList, []string{a})
	//}
	mi.SetDefaults()
	//if len(args.Comment) > 0 {
		mi.Comment = tmpComment//args.Comment
	//}
	//if len(args.CreatedBy) > 0 {
		mi.CreatedBy = tmpCreatedBy//args.CreatedBy
	//}
	//mi.UrlList = args.Url//???????????
	filePath:="./TorrentFiles"
	/////////////////////////////////////////////////
		totalLength, err := totalLength(filePath)
			if err != nil {
				 fmt.Printf("calculating total length of %q: %v", filePath, err)
				return
			}
			pieceLength := metainfo.ChoosePieceLength(totalLength)
			info := metainfo.Info{
				PieceLength: pieceLength,
			}
		 fmt.Printf("torrent total length of %q is %d\n", filePath,totalLength)
	//////////////////////////////////////////////////
	//info := metainfo.Info{
	///	PieceLength: 256 * 1024,
	//}
	berr := info.BuildFromFilePath(filePath)//args.Root)
	if berr != nil {
		log.Fatal(berr)
	}
	//if args.InfoName != nil {
		info.Name =tmpInfoName// *args.InfoName
	//}
	mi.InfoBytes, err = bencode.Marshal(info)
	if err != nil {
		log.Fatal(err)
	}
	//err = mi.Write(os.Stdout)
	//if err != nil {
	//	log.Fatal(err)
	//}
	tmpMagnet:=mi.Magnet(nil,nil)
	log.Println("****",tmpMagnet)

	//
	cfg := torrent.NewDefaultClientConfig()
	cfg.Seed = true
	mainclient, ncerr := torrent.NewClient(cfg)
	if ncerr != nil {
		log.Println("new torrent client:", ncerr)
		return
	}
	defer mainclient.Close()	
	//t, _ := mainclient.AddMagnet(tmpMagnet.String())
	//<-t.GotInfo()
	//t.DownloadAll()
	//c.WaitAll()
	/////////////////////////////////////////////////
	/////////////////////////////////////////////////
			pc, err := storage.NewDefaultPieceCompletionForDir(".")
			if err != nil {
				fmt.Printf("new piece completion: %w", err)
				return 
			}
			defer pc.Close()
		ih := mi.HashInfoBytes()
			to, _ := mainclient.AddTorrentOpt(torrent.AddTorrentOpts{
				InfoHash: ih,
				Storage: storage.NewFileOpts(storage.NewFileClientOpts{
					ClientBaseDir: filePath,
					FilePathMaker: func(opts storage.FilePathMakerOpts) string {
						return filepath.Join(opts.File.Path...)
					},
					TorrentDirMaker: nil,
					PieceCompletion: pc,
				}),
			})
			defer to.Drop()
			err = to.MergeSpec(&torrent.TorrentSpec{
				InfoBytes: mi.InfoBytes,
				Trackers: [][]string{{
					`wss://tracker.btorrent.xyz`,
					`wss://tracker.openwebtorrent.com`,
					"http://p4p.arenabg.com:1337/announce",
					"udp://tracker.opentrackr.org:1337/announce",
					"udp://tracker.openbittorrent.com:6969/announce",
				}},
			})
			if err != nil {
				 fmt.Printf("setting trackers: %w", err)
				return
			}
			fmt.Printf("%v: %v\n", to, to.Metainfo().Magnet(&ih, &info))
	/////////////////////////////////////////////////
	/////////////////////////////////////////////////

	for {	
		log.Println("******", to.Seeding())
		time.Sleep(8 * time.Second)

	}
}

func totalLength(path string) (totalLength int64, err error) {
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		totalLength += info.Size()
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("walking path, %w", err)
	}
	return totalLength, nil
}
