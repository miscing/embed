package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	testDir string = "./testdata/target/"
)

var (
	testFiles []*file
)

type file struct {
	Name    string
	Content string
}

func findTestFiles() []*file {
	var files []*file
	//returns paths of files in testdata/target
	if err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if testDir == path {
			return nil
		}
		absFilePath, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		content, err := ioutil.ReadFile(absFilePath)
		if err != nil {
			panic(err)
		}
		files = append(files, &file{info.Name(), string(content)})
		return nil
	}); err != nil {
		panic(err)
	}
	return files

}

func checkTestProgAgainst(t *testing.T) {
	var err error
	cmd := exec.Command("go", "run", ".")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(bytes.NewBuffer(output))
	dec.DisallowUnknownFields()
	var hol = make([]*file, 0, len(testFiles))
	for dec.More() {
		resF := new(file)
		err := dec.Decode(resF)
		if err != nil {
			t.Error("failed to decode a file from packaged variable")
			break
		}
		hol = append(hol, resF)
	}

	if len(hol) != len(testFiles) {
		t.Error("number of packed files and test files did not match. Output had: ", len(hol), "test files ", len(testFiles))
		return
	}

Beg:
	for _, o := range hol {
		for _, tf := range testFiles {
			if reflect.DeepEqual(tf, o) {
				continue Beg
			}
		}
		t.Error("failed to find a match in testfiles of: ", o.Name, "\n", o.Content)
		for _, tf := range testFiles {
			t.Log(*tf)
		}
		break
	}

}

func TestSingleArg(t *testing.T) {
	var err error
	os.Remove("./testdata/bindata.go")

	cmd := exec.Command("go", "run", "..", "./target/")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	checkTestProgAgainst(t)
}

func TestMultiArg(t *testing.T) {
	var err error
	os.Remove("./testdata/bindata.go")

	testFiles := findTestFiles()
	allArgs := []string{"run", ".."}
	pathPrefix, err := filepath.Abs("./testdata/target/")
	if err != nil {
		panic(err)
	}
	for _, f := range testFiles {
		allArgs = append(allArgs, filepath.Clean(pathPrefix+"/"+f.Name))
	}
	cmd := exec.Command("go", allArgs...)
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		fmt.Println(cmd.String())
		panic(err)
	}

	checkTestProgAgainst(t)
}

func TestOneTarInput(t *testing.T) {
	var err error

	os.Remove("./testdata/bindata.go")

	cmd := exec.Command("go", "run", ".", "../target/")
	cmd.Dir, err = filepath.Abs("./testdata/readypacked/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	cmd = exec.Command("go", "run", "..", "./readypacked/archive.tar")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	checkTestProgAgainst(t)
}

func TestMain(m *testing.M) {
	os.Remove("./testdata/bindata.go")

	testFiles = findTestFiles()

	r := m.Run()
	if err := os.Remove("./testdata/bindata.go"); err != nil {
		panic(err)
	}
	if err := os.Remove("./testdata/readypacked/archive.tar"); err != nil {
		panic(err)
	}

	os.Exit(r)
}
