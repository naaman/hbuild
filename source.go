package hbuild

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Source struct {
	Dir     *os.File
	Archive *os.File
	Get     *url.URL
	Put     *url.URL
}

type SourceJSON struct {
	SourceBlob struct {
		GetUrl string `json:"get_url"`
		PutUrl string `json:"put_url"`
	} `json:"source_blob"`
}

func NewSource(token, app, dir string) (source Source, err error) {
	client := newHerokuClient(token)

	source.Dir, err = os.Open(dir)
	if err != nil {
		return
	}

	sourceJson := new(SourceJSON)

	err = client.request(sourceRequest(app), &sourceJson)
	if err != nil {
		return
	}

	source.Get, err = url.Parse(sourceJson.SourceBlob.GetUrl)
	if err != nil {
		return
	}
	source.Put, err = url.Parse(sourceJson.SourceBlob.PutUrl)
	if err != nil {
		return
	}

	return
}

func sourceRequest(app string) herokuRequest {
	return herokuRequest{"POST", "/apps/" + app + "/sources", nil, http.Header{}}
}

func (s *Source) Compress() (err error) {
	s.Archive, err = tarGz(s.Dir.Name())
	return
}

func (s *Source) Upload() (err error) {
	return uploadFile(s.Put.String(), s.Archive)
}

func targzWalk(dirPath string, tw *tar.Writer) error {
	var walkfunc filepath.WalkFunc

	walkfunc = func(path string, fi os.FileInfo, err error) error {
		// ignore the .git dir
		if strings.HasPrefix(path, ".git/") {
			return nil
		}

		h, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		h.Name = "./" + path

		if fi.Mode()&os.ModeSymlink != 0 {
			linkPath, err := os.Readlink(path)
			if err != nil {
				return err
			}

			h.Linkname = linkPath
		}

		err = tw.WriteHeader(h)
		if err != nil {
			return err
		}

		if fi.Mode()&os.ModeDir == 0 && fi.Mode()&os.ModeSymlink == 0 {
			fr, err := os.Open(path)
			if err != nil {
				return err
			}

			defer fr.Close()

			_, err = io.Copy(tw, fr)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return filepath.Walk(dirPath, walkfunc)
}

func tarGz(inPath string) (tarArchive *os.File, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	os.Chdir(inPath)
	defer os.Chdir(wd)

	// file write
	tempSource, err := ioutil.TempFile("", "source")
	if err != nil {
		return nil, err
	}

	tarArchive, err = os.Create(tempSource.Name())
	if err != nil {
		return
	}

	// gzip write
	gw := gzip.NewWriter(tarArchive)
	defer gw.Close()

	// tar write
	tw := tar.NewWriter(gw)
	defer tw.Close()

	err = targzWalk(".", tw)

	return
}
