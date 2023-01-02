package markov

import (
	"archive/zip"
	"io"
	"os"
	"strings"
	"time"
)

func zipTicker() {
	for range time.Tick(6 * time.Hour) {
	writingLoop:
		debugLog("zip ticker went off")
		if writing {
			debugLog("cannot zip because currently writing -- trying again in 30s")
			time.Sleep(30 * time.Second)
			goto writingLoop
		}

		zipping = true
		zipChains()
		zipping = false
	}
}

func zipChains() {
	defer duration(track("zip duration", ""))

	debugLog("creating zip archive...")
	archive, err := os.Create("markov-chains.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	if err := addDirectoryToZip(zipWriter, "./markov-chains/"); err != nil {
		panic(err)
	}

	debugLog("closing zip archive...")
	zipWriter.Close()
}

func addDirectoryToZip(zipWriter *zip.Writer, path string) error {
	if !strings.HasPrefix(path, "./") {
		path = "./" + path
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		filePath := path + file.Name()

		if file.IsDir() {
			if err := addDirectoryToZip(zipWriter, filePath); err != nil {
				return err
			}
			continue
		}

		f2, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f2.Close()

		filePath = strings.TrimPrefix(filePath, "./")
		debugLog("zipping directory", filePath, "to archive...")
		w2, err := zipWriter.Create(filePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w2, f2); err != nil {
			return err
		}
	}

	return nil
}
