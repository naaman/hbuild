# hbuild

Deploy an app to Heroku without using git.

## About

`hbuild` is a utility for deploying an app to Heroku using its 
`builds` API instead of `git push heroku master`.

## Install

```sh
$ go get github.com/naaman/hbuild/cmd/hbuild
```

## Command Usage

```sh
$ hbuild
Usage of hbuild:
  -app="": -app=exampleapp
  -key="12345678-1234-5678-1234-567812345678": -key=123ABC
  -source="/home/user/code/exampleapp": -source=/path/to/src
$ hbuild -app=exampleapp
Creating source...done.
Compressing source...done.
Uploading source...done.
Building:

-----> PHP app detected
...
```

## Library Usage

```go
// Setup a Source object
source, _ := hbuild.NewSource(apiKey, appName, sourceDir)

// Create a compressed targz of the source
source.Compress()

// Upload the source
source.Upload()

// Run a build
build, _ := hbuild.NewBuild(apiKey, appName, source)

// Watch the build output
io.Copy(os.Stdout, build.Output)
```

Note: see the [hbuild command source](https://github.com/naaman/hbuild/blob/master/cmd/hbuild/main.go)
for slightly more detail.

## TODO

* Tests
* Pipeline Compress and Upload
