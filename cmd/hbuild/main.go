package main

import (
	"bufio"
	"code.google.com/p/go-netrc/netrc"
	"flag"
	"fmt"
	"github.com/naaman/hbuild"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"bytes"
	"time"
	"log"
)

var (
	workDir, _ = os.Getwd()
	fDir       = flag.String("source", workDir, "-source=/path/to/src")
	fAppName   = flag.String("app", appName(), "-app=exampleapp")
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

	fmt.Println("Building...")
	build, err = hbuild.NewBuild(*fApiKey, *fAppName, source)
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

func appName() string {
	gitConfigCmd := exec.Command("git", "config", "--list")
	gitConfig, err := gitConfigCmd.CombinedOutput()
	if err != nil {
		return ""
	}

	gitConfigScanner := bufio.NewScanner(bytes.NewBuffer(gitConfig))
	gitConfigScanner.Split(bufio.ScanLines)

	for gitConfigScanner.Scan() {
		gitConfigLine := gitConfigScanner.Text()
		if strings.HasPrefix(gitConfigLine, "remote.heroku.url") {
			l := strings.TrimSuffix(gitConfigLine, ".git")
			i := strings.LastIndex(l, ":") + 1

			return l[i:]
		}
	}

	return ""
}
