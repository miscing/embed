package main

import (
	"archive/tar"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	usageStr string = "tarer [path] //path is optional, set to current directory if omitted\nWalks through directory and outputs a tar encoded file"
)

func usage() {
	fmt.Println(usageStr)
	flag.PrintDefaults()
}

func main() {
	var buf bytes.Buffer
	var path string

	name := flag.String("-n", "archive", "set output name, .tar is appended to it")
	flag.Parse()
	if flag.NArg() > 1 {
		usage()
		return
	}
	if flag.NArg() == 1 {
		path = flag.Arg(0)
	} else {
		path = "."
	}

	tw := tar.NewWriter(&buf)
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil //skips dir, but will still search recurssively into it
		}
		f, err := os.Open(path)
		if err != nil {
			log.Panic(err)
		}
		fi, err := f.Stat()
		if err != nil {
			log.Panic(err)
		}
		head, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			log.Panic(err) // stop execution
		}
		if err := tw.WriteHeader(head); err != nil {
			log.Panic(err)
		}
		if _, err := io.Copy(tw, f); err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	if err := tw.Close(); err != nil {
		log.Panic(err)
	}

	payload, err := ioutil.ReadAll(&buf)
	if err != nil {
		log.Panic(err)
	}

	file, err := os.OpenFile("./"+*name+".tar", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	defer file.Close()
	if err != nil {
		log.Panic(err)
	}

	if _, err := file.Write(payload); err != nil {
		log.Panic(err)
	}

	fmt.Println("created: ", *name, " tar archive")
}
