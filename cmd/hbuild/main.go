package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

        "github.com/heroku/hbuild"
)

var (
	workDir, _ = os.Getwd()
	fDir       = flag.String("source", workDir, "-source=/path/to/src")
	fAppName   = flag.String("app", "", "-app=exampleapp")
	fApiKey    = flag.String("key", "", "-key=123ABC")
)

func init() {
	flag.Parse()
	for _, f := range []string{*fAppName, *fApiKey} {
		if f == "" {
			flag.Usage()
			os.Exit(1)
		}
	}
}

func main() {
	var err error
	var source hbuild.Source
	var build hbuild.Build

	fmt.Print("Creating source...")
	source, err = hbuild.NewSource(*fApiKey, *fAppName, *fDir)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("done.")

	fmt.Print("Compressing source...")
	err = source.Compress()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("done.")

	fmt.Print("Uploading source...")
	err = source.Upload()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("done.")

	fmt.Println("Building...")
	build, err = hbuild.NewBuild(*fApiKey, *fAppName, source, *new(hbuild.BuildOptions))
	if err != nil {
		fmt.Println(err)
		return
	}

	io.Copy(os.Stdout, build.Output)

	for {
		s, err := build.Status()
		if err != nil {
			log.Println(err)
		}

		if s == "pending" {
			fmt.Print(".")
			time.Sleep(time.Second)
		} else {
			fmt.Println("..done.")
			return
		}
	}
}
