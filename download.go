package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
)

type Task struct {
	Attachment
	FilePath string
}

var index map[string]bool

func startDownload() error {
	var (
		stop = false
		err  error
	)

	index, err = readIndex()
	if err != nil {
		return err
	}

	stopChan := make(chan struct{})
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go progressMeter(wg, stopChan)

	for page := 1; !stop; page++ {
		stop, err = downloadPage(page)
		if err != nil {
			return err
		}
	}

	close(stopChan)
	wg.Wait()

	return nil
}

func downloadPage(page int) (bool, error) {
	attResponse, err := getAttachmentsList(page)
	if err != nil {
		return false, err
	}

	stop := !attResponse.HasMore

	wg := new(sync.WaitGroup)
	tasks := make(chan Task)
	for i := 0; i < config.Threads; i++ {
		wg.Add(1)
		go worker(wg, tasks)
	}

	existsFiles := make(map[string]map[string]bool)

	for _, att := range attResponse.Attachments {
		dirPath := filepath.Join(config.resultsRoot, att.CreatedAt.Format("2006/01"))
		if err := os.MkdirAll(dirPath, 0777); err != nil {
			return false, err
		}

		if existsFiles[dirPath] == nil {
			// First visit to this directory
			existsFiles[dirPath] = make(map[string]bool)
			existsFiles[dirPath][indexFileName] = true
			entries, err := os.ReadDir(dirPath)
			if err != nil {
				return false, err
			}
			for _, entry := range entries {
				existsFiles[dirPath][entry.Name()] = true
			}
		}

		chPlanned <- struct{}{}

		if index[att.ID] {
			chDone <- struct{}{}
			continue
		} else {
			index[att.ID] = true
		}

		fn := getFileName(att, existsFiles[dirPath])
		existsFiles[dirPath][fn] = true

		tasks <- Task{
			Attachment: att,
			FilePath:   filepath.Join(dirPath, fn),
		}
	}

	close(tasks)
	wg.Wait()

	return stop, nil
}

const maxBaseNameLength = 80

var twitterJpgRe = regexp.MustCompile(`^\.jpg:`)

func getFileName(att Attachment, entries map[string]bool) string {
	ext := filepath.Ext(att.FileName)
	baseName := strings.TrimSuffix(att.FileName, ext)
	if ext == "" {
		ext = filepath.Ext(att.URL)
	}

	if twitterJpgRe.MatchString(ext) {
		ext = ".jpg"
	}

	if baseName == "" {
		baseName = "file"
	}

	if utf8.RuneCountInString(baseName) > maxBaseNameLength {
		baseName = string([]rune(baseName)[:maxBaseNameLength])
	}

	candidate := fileNameSanitizer.Replace(baseName + ext)
	if !entries[candidate] {
		return candidate
	}

	for i := 1; ; i++ {
		candidate := fileNameSanitizer.Replace(fmt.Sprintf("%s (%d)%s", baseName, i, ext))
		if !entries[candidate] {
			return candidate
		}
	}
}

var fileNameSanitizer = strings.NewReplacer(
	`<`, `_`,
	`>`, `_`,
	`:`, `_`,
	`"`, `_`,
	`/`, `_`,
	`\`, `_`,
	`|`, `_`,
	`?`, `_`,
	`*`, `_`,
)

func worker(wg *sync.WaitGroup, tasks <-chan Task) {
	defer wg.Done()
	for task := range tasks {
		func() {
			resp, err := http.Get(task.Attachment.URL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to download %s: %v\n", task.Attachment.URL, err)
				return
			}
			defer resp.Body.Close()

			tmpName := filepath.Join(filepath.Dir(task.FilePath), task.Attachment.ID+".tmp")
			file, err := os.Create(tmpName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create file: %v\n", err)
				return
			}

			_, err = io.Copy(file, resp.Body)
			file.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to copy data: %v\n", err)
				os.Remove(tmpName)
				return
			}

			if err := os.Rename(tmpName, task.FilePath); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to move file: %v\n", err)
				os.Remove(tmpName)
				return
			}

			if err := addIndexLine(task); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write index: %v\n", err)
				return
			}
			chDone <- struct{}{}
		}()
	}
}
