package main

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/wetorrent/wetorrent/internal/chunk_storage"
)

type SearchManager struct {
	w64storage      *chunk_storage.ChunkStorage
	MainSearchQuery string
	MainSearchIndex int
}

func (s *SearchManager) Init(storageDir, fileNamePrefix string) (err error) {
	if s.w64storage, err = chunk_storage.New(storageDir, fileNamePrefix); err != nil {
		return fmt.Errorf("creating w64 storage: %v", err)
	}

	s.MainSearchQuery = ""
	s.MainSearchIndex = s.w64storage.NumberOfChunks() - 1

	return nil
}

func (s *SearchManager) SetSearchQuery(server *Server, query string) {
	if query == "" {
		query = "sexy"
		return
	}

	if query != s.MainSearchQuery {
		fmt.Println("New searchquery**", s.w64storage.NumberOfChunks())
		s.MainSearchQuery = query
		s.MainSearchIndex = s.w64storage.NumberOfChunks() - 1
		EmptySearchResults()
	}

	go s.MoreSearchResults(server)
}

func (s *SearchManager) MoreSearchResults(server *Server) {
	if s.MainSearchQuery == "" {
		return
	}

	for s.MainSearchIndex >= 0 {
		chunkBytes := s.w64storage.GetChunkById(s.MainSearchIndex)
		name, description, magnet := s.ReadItemBytes(chunkBytes)

		s.MainSearchIndex--

		if name == "" {
			continue
		}

		if SearchResultsFull() {
			return
		}

		if strings.Contains(strings.ToLower(name), strings.ToLower(s.MainSearchQuery)) {
			_ = name
			_ = description
			_ = magnet

			server.addtorrent(name, description, magnet)
		}
	}
}

func (s *SearchManager) ReadItemBytes(brContent []byte) (string, string, string) {
	maxCounter := len(brContent)
	counter := 1

	if counter > maxCounter {
		return "", "", "" // unexpected end of content
	}

	lenName := int(brContent[counter-1])
	counter += lenName

	if counter > maxCounter {
		return "", "", "" // unexpected end of content
	}

	bytesName := brContent[counter-lenName : counter]
	counter += 2

	if counter > maxCounter {
		return "", "", "" // unexpected end of content
	}

	lenDescription := int(binary.LittleEndian.Uint16(brContent[counter-2 : counter]))
	counter += lenDescription

	if counter > maxCounter {
		return "", "", "" // unexpected end of content
	}

	bytesDecription := brContent[counter-lenDescription : counter]
	counter += 2

	if counter > maxCounter {
		return "", "", "" // unexpected end of content
	}

	lenMagnet := int(binary.LittleEndian.Uint16(brContent[counter-2 : counter]))
	counter += lenMagnet

	if counter > maxCounter {
		return "", "", "" // unexpected end of content
	}

	bytesMagnet := brContent[counter-lenMagnet : counter]

	return string(bytesName), string(bytesDecription), string(bytesMagnet)
}
