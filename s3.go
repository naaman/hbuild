package hbuild

import (
	"errors"
	"net/http"
	"os"
)

func uploadFile(putUrl string, file *os.File) (err error) {
	stat, err := file.Stat()
	if err != nil {
		return
	}

	openFile, err := os.Open(file.Name())
	if err != nil {
		return
	}
	defer openFile.Close()

	req, err := http.NewRequest("PUT", putUrl, openFile)
	if err != nil {
		return
	}

	req.ContentLength = stat.Size()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		return errors.New("Unable to upload file.")
	}

	return
}
