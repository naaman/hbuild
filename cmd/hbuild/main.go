package main

import (
	"code.google.com/p/go-netrc/netrc"
	"flag"
	"fmt"
	"github.com/naaman/hbuild"
	"io"
	"os"
	"os/user"
)

var (
	workDir, _ = os.Getwd()
	fDir       = flag.String("source", workDir, "-source=/path/to/src")
	fAppName   = flag.String("app", "", "-app=exampleapp")
	fApiKey    = flag.String("key", netrcApiKey(), "-key=123ABC")
)

func init() {
	flag.Parse()
	if *fAppName == "" {
		flag.Usage()
		os.Exit(1)
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

	fmt.Println("Building:")
	err = build.Run(*fApiKey, *fAppName, source)
	if err != nil {
		fmt.Println(err)
		return
	}

	io.Copy(os.Stdout, build.Output)
}

func netrcApiKey() string {
	if u, err := user.Current(); err == nil {
		netrcPath := u.HomeDir + "/.netrc"
		if _, err := os.Stat(netrcPath); err == nil {
			key, _ := netrc.FindMachine(netrcPath, "api.heroku.com")
			if key.Password != "" {
				return key.Password
			}
		}
	}
	return ""
}
