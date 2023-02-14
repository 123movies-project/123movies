package main

import (
	"encoding/binary"
	//"encoding/json"
    	//"bufio"
    	"fmt"
	"log"
    	"io"
    	"os"
	"bytes"
	"strings"
	"path"
)


/*
type Item struct {
	Name string `json:"name"`
	Channelname string `json:"channelname"`
	Description string `json:"description"`
	Magnet string `json:"magnet"`
}


var item Item
*/
//
var w64storage * ChunkStorage
var MainSearchQuery string
var MainSearchIndex int

func InitSearchManager(){
	//tmpstoragepath:="./db"
	//if canIHaveAProxy() {
		tmpstoragepath:=path.Join("w64system","w64system")//"./w64system/w64system"
	//}
	w64storage =OpenChunkStorage(tmpstoragepath)
	MainSearchQuery=""
	MainSearchIndex=w64storage.NbChunks()-1
	fmt.Println("****",w64storage.NbChunks())
}


func SetSearchQuery(searchquery string){
	//searchquery:="bikini"
	if searchquery=="" {
		searchquery="sexy"
		return
	}
	if searchquery!=MainSearchQuery {
		fmt.Println("New searchquery**",w64storage.NbChunks())
		MainSearchQuery=searchquery
		MainSearchIndex=w64storage.NbChunks()-1
		EmptySearchResults()
	}
	go MoreSearchResults()
}

func MoreSearchResults(){
	//fmt.Println("****",w64storage.NbChunks())
	//for i := w64storage.NbChunks()-1; i >=0 ; i-- {
	if MainSearchQuery=="" {
		//searchquery="sexy"
		return
	}
	for (MainSearchIndex>=0){
		tmpbytes:=w64storage.GetChunkById(MainSearchIndex)
		MainSearchIndex--
		//fmt.Println("++++",MainSearchIndex)//len(tmpbytes))
		name,description,magnet:=ReadItemBytes(tmpbytes)
		//fmt.Println("***", name, description, magnet)
		if name==""{
			//fmt.Println("***")
			continue
		}
		if SearchResultsFull() {
			return
		}
		if searchStringFull(strings.ToLower(name),strings.ToLower(MainSearchQuery)) {
			_=name
			_=description
			_=magnet
			fmt.Println("!!! found :", name)
			addtorrent(name,description,magnet)
		}
	}


}

func searchStringFull(text string,searchquery string) bool{

		if strings.Contains(text, searchquery) {
			return true
		}
		return false
}
/*

func main(){
	w64storage:=OpenChunkStorage("./w64system/w64system")
	fmt.Println("**** NbChunks()",w64storage.NbChunks())
	//w64storage.AddChunk(0)
	//w64storage.GetChunkById(0)
	for i := 0; i < 101004; i++ {
		tmpcontent:=GetItemBytesFromFile(fmt.Sprintf("Files/Items/Item%d.json",i))//("Files/Items/Item0.json")
		_=tmpcontent
		//fmt.Println("****",tmpcontent)
		if len(tmpcontent)>0 {
			w64storage.AddChunk(tmpcontent)
			//ReadItemBytes(bwcontent)
		}
	}
}
*/
/*
func GetItemBytesFromFile(tmppath string) ([]byte){
    dat, err := os.ReadFile(tmppath)
	if err!=nil {
		fmt.Println("error",err)
		return nil
	}
    fmt.Println(string(dat))
	nerr := json.Unmarshal(dat, &item)
	if nerr != nil {
		fmt.Println("error",nerr)
		return nil
	}
	fmt.Println("Item name:",item.Name)
	fmt.Println("Item channelname:",item.Channelname)
	fmt.Println("Item description:",item.Description)
	fmt.Println("Item magnet:",item.Magnet)
	
	var bwcontent []byte

	itemnamebytes:=[]byte(item.Name)
	namelen := len(itemnamebytes)
	tmpbufferuintnamelen := make([]byte, 1)
	tmpbufferuintnamelen[0] = byte(namelen)
	bwcontent = append(bwcontent, tmpbufferuintnamelen...)
	bwcontent = append(bwcontent, itemnamebytes...)

	itemdescriptionbytes:=[]byte(item.Description)
	descriptionlen := len(itemdescriptionbytes)
	tmpbufferuintdescriptionlen := make([]byte, 2)
	binary.LittleEndian.PutUint16(tmpbufferuintdescriptionlen, uint16(descriptionlen))
	bwcontent = append(bwcontent, tmpbufferuintdescriptionlen...)
	bwcontent = append(bwcontent, itemdescriptionbytes...)

	itemmagnetbytes:=[]byte(item.Magnet)
	magnetlen := len(itemmagnetbytes)
	tmpbufferuintmagnetlen := make([]byte, 2)
	binary.LittleEndian.PutUint16(tmpbufferuintmagnetlen, uint16(magnetlen))
	bwcontent = append(bwcontent, tmpbufferuintmagnetlen...)
	bwcontent = append(bwcontent, itemmagnetbytes...)
	return bwcontent
}
*/
func ReadItemBytes(brcontent []byte) (string,string,string){
	maxcounter:=len(brcontent)
	counter := 0
	counter += 1
	if counter>maxcounter{
		fmt.Println("++++",counter,maxcounter)
		return "","",""
	}
	namelen := int(brcontent[counter-1])
	counter += namelen
	if counter>maxcounter{
	fmt.Println("++++",counter,maxcounter)
		return "","",""
	}
	namebytes := brcontent[counter-namelen : counter]
	counter += 2
	if counter>maxcounter{
	fmt.Println("++++",counter,maxcounter)
		return "","",""
	}
	descriptionlen := int(binary.LittleEndian.Uint16(brcontent[counter-2 : counter]))
	counter += descriptionlen
	if counter>maxcounter{
	fmt.Println("++++",counter,maxcounter)
		return "","",""
	}
	descriptionbytes := brcontent[counter-descriptionlen : counter]
	counter += 2
	if counter>maxcounter{
	fmt.Println("++++",counter,maxcounter)
		return "","",""
	}
	magnetlen := int(binary.LittleEndian.Uint16(brcontent[counter-2 : counter]))
	counter += magnetlen
	if counter>maxcounter{
	fmt.Println("++++",counter,maxcounter)
		return "","",""
	}
	magnetbytes := brcontent[counter-int(magnetlen) : counter]
	//fmt.Println("***", string(namebytes), string(descriptionbytes), string(magnetbytes))
	return string(namebytes), string(descriptionbytes), string(magnetbytes)
}
/*
package utility

import (
	//"github.com/globaldce/globaldce-gateway/applog"
	"fmt"
	"encoding/binary"
	"bytes"
	"log"
	"os"
	"io"
)
*/
const ChunkFileMaxSize=20*1024*1024//100*1024*1024

// ChunkStorage is
type ChunkStorage struct {
	Path string
	file [] *os.File
	//
	Chunkposition []int64
	Chunksize []int64
	Chunkfileid []int
}
func (cs *ChunkStorage) NbChunks() int{
	return int(len(cs.Chunkposition))
}
// OpenChunkStorage is
func OpenChunkStorage(storagepath string) *ChunkStorage {
	cs := new(ChunkStorage)
	cs.Path = storagepath

	var activechunkfileid int=0
for {
	
	filepath:=fmt.Sprintf("%s%03d",storagepath,activechunkfileid)
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	cs.file =append(cs.file,f)
	
	//--------------------------

	var position int64 =-4
	var chunksize uint32 =uint32 (0)
	var i int =0

	for  {
		position+=int64(chunksize+4)
		_, seekerr := cs.file[activechunkfileid].Seek(position, 0)
		if seekerr != nil {
			cs.file[activechunkfileid].Close() // ignore error; Write error takes precedence
			log.Fatal(seekerr)
		}
		//applog.Trace("position %d", position)

		bufferchunksize := make([]byte, 4)
		_, readerr := cs.file[activechunkfileid].Read(bufferchunksize)

		if readerr == io.EOF {
			break
		}

		if readerr != nil {
			cs.file[activechunkfileid].Close() // ignore error; Write error takes precedence
			log.Fatal(readerr)
		}

		
		readerchunksize := bytes.NewReader(bufferchunksize)


		binary.Read(readerchunksize, binary.LittleEndian, &chunksize)

		cs.Chunksize=append(cs.Chunksize ,int64 (chunksize))
		cs.Chunkposition=append(cs.Chunkposition , int64 (position))
		cs.Chunkfileid=append(cs.Chunkfileid,activechunkfileid)

		i++
	}
	//---------------------------
	if _, err := os.Stat(fmt.Sprintf("%s%03d",storagepath,activechunkfileid+1)); os.IsNotExist(err) {
		break
	}else{
		activechunkfileid++
	}
	//---------------------------
}

	//--------------------------
	return cs
}

func (cs *ChunkStorage) AddChunk(data []byte) error {
	var activechunkfileid int
	activechunkfileid=len(cs.file)-1
	var newchunkfile bool =false


	//------------------------------
	fileinfo, staterr := cs.file[activechunkfileid].Stat()
	if staterr != nil {
		log.Fatal(staterr)
	}
	//applog.Trace("size %d", fileinfo.Size())
	if (fileinfo.Size()>ChunkFileMaxSize){
		activechunkfileid++
		filepath:=fmt.Sprintf("%s%03d",cs.Path,activechunkfileid)
		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			log.Fatal(err)
		}
		cs.file =append(cs.file,f)
		newchunkfile=true
	}

	//------------------------------


	bufferchunkfilesize := make([]byte, 4)
	binary.LittleEndian.PutUint32(bufferchunkfilesize, uint32(len(data)))
	_,serrs:=cs.file[activechunkfileid].Seek(0,os.SEEK_END)
	if serrs != nil {
		cs.file[activechunkfileid].Close() // error
		log.Fatal(serrs)
	}
	_, lerr := cs.file[activechunkfileid].Write(bufferchunkfilesize)
	if lerr != nil {
		cs.file[activechunkfileid].Close() // error
		log.Fatal(lerr)
	}
	_,serrd:=cs.file[activechunkfileid].Seek(0,os.SEEK_END)
	if serrd != nil {
		cs.file[activechunkfileid].Close() // error
		log.Fatal(serrd)
	}

	_, err := cs.file[activechunkfileid].Write(data)
	if err != nil {
		cs.file[activechunkfileid].Close() // error
		log.Fatal(err)
	}

	if (!newchunkfile) && (cs.NbChunks()>=1){
			newposition:=int64 (  int (cs.Chunkposition[cs.NbChunks()-1]+cs.Chunksize[cs.NbChunks()-1])+4 )
			cs.Chunkposition=append(cs.Chunkposition , newposition)	
		} else {
			cs.Chunkposition=append(cs.Chunkposition, int64(0))
		}






	cs.Chunksize=append(cs.Chunksize ,int64 (len(data)))

	cs.Chunkfileid=append(cs.Chunkfileid ,activechunkfileid)

	return err
}


func (cs *ChunkStorage) GetChunkById(chunkid int) []byte {


	//applog.Trace("position %d size %d file %d", cs.Chunkposition[chunkid]+4, cs.Chunksize[chunkid],cs.Chunkfileid[chunkid])
	return cs.GetChunk(cs.Chunkposition[chunkid]+4, cs.Chunksize[chunkid],cs.Chunkfileid[chunkid])
}


func (cs *ChunkStorage) GetChunk(position int64, length int64,fileid int) []byte {
	_, seekerr := cs.file[fileid].Seek(position, 0)
	//check(err)
	if seekerr != nil {
		cs.file[fileid].Close() // ignore error; Write error takes precedence
		log.Fatal(seekerr)
	}
	chunk := make([]byte, length)
	_, readerr := cs.file[fileid].Read(chunk)

	if readerr != nil {
		cs.file[fileid].Close() // ignore error; Write error takes precedence
		log.Fatal(readerr)
	}
	return chunk
}


