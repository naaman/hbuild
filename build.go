package hbuild

import (
	"io"
	"net/http"
)

type Build struct {
	Id              UUID   `json:"id"`
	OutputStreamURL string `json:"output_stream_url"`
	Output          io.Reader
}

type BuildRequestJSON struct {
	SourceBlob struct {
		Url     string `json:"url"`
		Version string `json:"version"`
	} `json:"source_blob"`
}

func (b *Build) Run(token, app string, source Source) (err error) {
	buildJson := BuildRequestJSON{}
	buildJson.SourceBlob.Url = source.Get.String()

	client := newHerokuClient(token)
	client.request(buildRequest(app, buildJson), &b)

	stream, err := http.DefaultClient.Get(b.OutputStreamURL)
	if err != nil {
		return
	}

	b.Output = stream.Body
	return
}

func buildRequest(app string, build BuildRequestJSON) herokuRequest {
	return herokuRequest{"POST", "/apps/" + app + "/builds", build}
}
