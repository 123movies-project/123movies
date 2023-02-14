package chunk_storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

func New(storageDir, fileNamePrefix string) (*ChunkStorage, error) {
	var storage ChunkStorage

	storageDirWithFilePrefix := path.Join(storageDir, fileNamePrefix)

	if err := os.MkdirAll(storageDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("making dir for w64system files: %v", err)
	}

	storage.Path = storageDirWithFilePrefix
	activeChunkFileID := 0

	for {
		filepath := fmt.Sprintf("%s%03d", storageDirWithFilePrefix, activeChunkFileID)

		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, fmt.Errorf("opening file '%s': %v", filepath, err)
		}

		storage.file = append(storage.file, f)

		var position int64 = -4
		var chunkSize uint32 = 0

		for {
			position += int64(chunkSize + 4)
			if _, errSeek := storage.file[activeChunkFileID].Seek(position, 0); errSeek != nil {
				_ = storage.file[activeChunkFileID].Close() // ignore error; Write error takes precedence
				return nil, fmt.Errorf("seeking in file '%s' at chunk ID %d: %v", filepath, activeChunkFileID, errSeek)
			}

			bufferChunkSize := make([]byte, 4)

			if _, errRead := storage.file[activeChunkFileID].Read(bufferChunkSize); errRead == io.EOF {
				break
			} else if errRead != nil {
				storage.file[activeChunkFileID].Close() // ignore error; Write error takes precedence
				return nil, fmt.Errorf("reading file '%s' at chunk ID %d: %v", filepath, activeChunkFileID, errRead)
			}

			readerChunkSize := bytes.NewReader(bufferChunkSize)

			_ = binary.Read(readerChunkSize, binary.LittleEndian, &chunkSize)

			storage.size = append(storage.size, int64(chunkSize))
			storage.position = append(storage.position, int64(position))
			storage.fileID = append(storage.fileID, activeChunkFileID)
		}

		if _, err := os.Stat(fmt.Sprintf("%s%03d", storageDirWithFilePrefix, activeChunkFileID+1)); os.IsNotExist(err) {
			break
		}

		activeChunkFileID++
	}

	return &storage, nil
}

const ChunkFileMaxSize = 20 * 1024 * 1024

// ChunkStorage is
type ChunkStorage struct {
	Path     string
	file     []*os.File
	position []int64
	size     []int64
	fileID   []int
}

func (cs *ChunkStorage) NumberOfChunks() int {
	return len(cs.position)
}

func (cs *ChunkStorage) AddChunk(data []byte) error {
	var activechunkfileid int
	activechunkfileid = len(cs.file) - 1
	var newchunkfile bool = false

	//------------------------------
	fileinfo, staterr := cs.file[activechunkfileid].Stat()
	if staterr != nil {
		log.Fatal(staterr)
	}
	//applog.Trace("size %d", fileinfo.Size())
	if fileinfo.Size() > ChunkFileMaxSize {
		activechunkfileid++
		filepath := fmt.Sprintf("%s%03d", cs.Path, activechunkfileid)
		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			log.Fatal(err)
		}
		cs.file = append(cs.file, f)
		newchunkfile = true
	}

	//------------------------------

	bufferchunkfilesize := make([]byte, 4)
	binary.LittleEndian.PutUint32(bufferchunkfilesize, uint32(len(data)))
	_, serrs := cs.file[activechunkfileid].Seek(0, os.SEEK_END)
	if serrs != nil {
		cs.file[activechunkfileid].Close() // error
		log.Fatal(serrs)
	}
	_, lerr := cs.file[activechunkfileid].Write(bufferchunkfilesize)
	if lerr != nil {
		cs.file[activechunkfileid].Close() // error
		log.Fatal(lerr)
	}
	_, serrd := cs.file[activechunkfileid].Seek(0, os.SEEK_END)
	if serrd != nil {
		cs.file[activechunkfileid].Close() // error
		log.Fatal(serrd)
	}

	_, err := cs.file[activechunkfileid].Write(data)
	if err != nil {
		cs.file[activechunkfileid].Close() // error
		log.Fatal(err)
	}

	if (!newchunkfile) && (cs.NumberOfChunks() >= 1) {
		newposition := int64(int(cs.position[cs.NumberOfChunks()-1]+cs.size[cs.NumberOfChunks()-1]) + 4)
		cs.position = append(cs.position, newposition)
	} else {
		cs.position = append(cs.position, int64(0))
	}

	cs.size = append(cs.size, int64(len(data)))

	cs.fileID = append(cs.fileID, activechunkfileid)

	return err
}

func (cs *ChunkStorage) GetChunkById(chunkid int) []byte {

	//applog.Trace("position %d size %d file %d", cs.position[chunkid]+4, cs.size[chunkid],cs.fileID[chunkid])
	return cs.GetChunk(cs.position[chunkid]+4, cs.size[chunkid], cs.fileID[chunkid])
}

func (cs *ChunkStorage) GetChunk(position int64, length int64, fileid int) []byte {
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
