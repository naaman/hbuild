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

func NewBuild(token, app string, source Source) (build Build, err error) {
	buildJson := BuildRequestJSON{}
	buildJson.SourceBlob.Url = source.Get.String()

	client := newHerokuClient(token)
	err = client.request(buildRequest(app, buildJson), &build)
	if err != nil {
		return
	}

	stream, err := http.DefaultClient.Get(build.OutputStreamURL)
	if err != nil {
		return
	}

	build.Output = stream.Body
	return
}

func buildRequest(app string, build BuildRequestJSON) herokuRequest {
	return herokuRequest{"POST", "/apps/" + app + "/builds", build}
}
