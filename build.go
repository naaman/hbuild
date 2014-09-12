package hbuild

import (
	"io"
	"net/http"
)

type Build struct {
	Id     UUID
	Output io.Reader
	status string
	token  string
	app    string
}

type BuildResponseJSON struct {
	Id              string `json:"id"`
	OutputStreamURL string `json:"output_stream_url"`
	Status          string `json:"status"`
}

type BuildRequestJSON struct {
	SourceBlob struct {
		Url     string `json:"url"`
		Version string `json:"version"`
	} `json:"source_blob"`
}

type BuildOptions struct {
	SourceVersion     string
	AdditionalHeaders http.Header
}

func NewBuild(token, app string, source Source, opts BuildOptions) (build Build, err error) {
	buildReqJson := BuildRequestJSON{}
	buildReqJson.SourceBlob.Url = source.Get.String()
	buildReqJson.SourceBlob.Version = opts.SourceVersion

	client := newHerokuClient(token)
	buildResJson := BuildResponseJSON{}
	err = client.request(buildRequest(app, buildReqJson, opts.AdditionalHeaders), &buildResJson)
	if err != nil {
		return
	}

	stream, err := http.DefaultClient.Get(buildResJson.OutputStreamURL)
	if err != nil {
		return
	}

	build.Id = UUID(buildResJson.Id)
	build.token = token
	build.app = app
	build.Output = stream.Body
	return
}

func (b *Build) Status() (string, error) {
	buildJson := new(BuildResponseJSON)
	client := newHerokuClient(b.token)

	err := client.request(buildStatusRequest(*b), &buildJson)
	if err != nil {
		return "", err
	}

	return buildJson.Status, nil
}

func buildRequest(app string, build BuildRequestJSON, additionalHeaders http.Header) herokuRequest {
	return herokuRequest{"POST", "/apps/" + app + "/builds", build, additionalHeaders}
}

func buildStatusRequest(build Build) herokuRequest {
	return herokuRequest{"GET", "/apps/" + build.app + "/builds/" + string(build.Id), nil, http.Header{}}
}
