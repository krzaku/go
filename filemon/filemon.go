package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	logsDir           = "/tmp/fifi"
	exampleLogPattern = "*.txt"

	fileDiscoverySleep = 500 * time.Millisecond
)

type logFiles struct {
	example map[string]logInfo
}

type logInfo struct {
	exists, reopen bool
	fileInfo       os.FileInfo
	handled        bool
	existsMark     bool
}

var logs logFiles

func init() {
	logs.example = make(map[string]logInfo)
}

func discoverLogsLoop() {
	for {
		discoverLogs()
		time.Sleep(fileDiscoverySleep)
	}
}

func discoverLogs() {
	discoverFiles(logs.example, exampleLogPattern)
}

func discoverFiles(data map[string]logInfo, filePattern string) {
	var fileInf os.FileInfo

	// mark all files as nonexistent
	for key := range data {
		tmp := data[key]
		tmp.existsMark = false
		tmp.reopen = false
		data[key] = tmp
	}

	filesFound, err := filepath.Glob(filepath.Join(logsDir, filePattern))
	if err != nil {
		log.Println(err)
	} else {
		for _, file := range filesFound {
			f, _ := os.Open(file)
			fileInf, _ = f.Stat()

			f.Close()
			if _, found := data[file]; found {
				tmp := data[file]
				tmp.existsMark = true
				tmp.exists = true

				if !os.SameFile(tmp.fileInfo, fileInf) {
					tmp.fileInfo = fileInf
					tmp.reopen = true
				}
				data[file] = tmp
			} else {
				data[file] = logInfo{exists: true, fileInfo: fileInf, reopen: true, handled: false, existsMark: true}
			}
		}
	}

	// remove left nonexistent
	for key := range data {
		if data[key].existsMark == false {
			tmp := data[key]
			tmp.exists = false
			data[key] = tmp
		}
	}
}

func processFile(file string, data map[string]logInfo) {
	for {
		fmt.Println("read ", file)
		if !data[file].exists {
			tmp := data[file]
			tmp.handled = false
			data[file] = tmp
			break
		}
		time.Sleep(fileDiscoverySleep)
	}
}

func handleLogs(data map[string]logInfo) {
	for file, _ := range data {
		info := data[file]
		if info.reopen && !info.handled {
			info.reopen = false
			info.handled = true
			go processFile(file, data)
			data[file] = info

		}
	}
}

func handleLogsLoop() {
	for {
		handleLogs(logs.example)
		time.Sleep(fileDiscoverySleep)
	}
}

func main() {
	go discoverLogsLoop()
	handleLogsLoop()
}
