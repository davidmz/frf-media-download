package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const indexFileName = "index.jsonl"

var indexLock = new(sync.RWMutex)

type IndexLine struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	OrigName string `json:"origName"`
	FileName string `json:"fileName"`
}

func readIndex() (map[string]bool, error) {
	indexLock.RLock()
	defer indexLock.RUnlock()

	file, err := os.Open(filepath.Join(config.resultsRoot, indexFileName))
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]bool), nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	d := json.NewDecoder(file)
	m := make(map[string]bool)
	for {
		line := new(IndexLine)
		if err := d.Decode(line); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		m[line.ID] = true
	}

	return m, nil
}

func addIndexLine(task Task) error {
	indexLock.Lock()
	defer indexLock.Unlock()

	indexRow, _ := json.Marshal(IndexLine{
		ID:       task.ID,
		URL:      task.URL,
		OrigName: task.FileName,
		FileName: filepath.Base(task.FilePath),
	})
	indexRow = append(indexRow, '\n')

	file, err := os.OpenFile(
		filepath.Join(config.resultsRoot, indexFileName),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(indexRow)

	return err
}
