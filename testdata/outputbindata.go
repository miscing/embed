package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
)

// for testing puposes

type file struct {
	Name    string
	Content string
}

func main() {

	r := tar.NewReader(bytes.NewBuffer(bindata()))
	enc := json.NewEncoder(os.Stdout)

	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		f := new(file)
		f.Name = h.Name
		byteCont, err := ioutil.ReadAll(r)
		if err != nil {
			panic(err)
		}
		f.Content = string(byteCont)
		err = enc.Encode(f)
		if err != nil {
			panic(err)
		}
	}
}
