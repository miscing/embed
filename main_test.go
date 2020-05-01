//
// Copyright 2020 Alexander Saastamoinen
//
//  Licensed under the EUPL, Version 1.2 or – as soon they
// will be approved by the European Commission - subsequent
// versions of the EUPL (the "Licence");
//  You may not use this work except in compliance with the
// Licence.
//  You may obtain a copy of the Licence at:
//
//  https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12
//
//  Unless required by applicable law or agreed to in
// writing, software distributed under the Licence is
// distributed on an "AS IS" basis,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied.
//  See the Licence for the specific language governing
// permissions and limitations under the Licence.
//

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

const (
	testDir string = "./testdata/target/"
)

var (
	testFiles []*file
	ignore    string = "// +build ignore\n"
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
		t.Error(cmd.String())
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
	os.Remove("./testdata/bindata.go")
}

func TestMultiArg(t *testing.T) {
	var err error

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
	os.Remove("./testdata/bindata.go")
}

func TestOneTarInput(t *testing.T) {
	var err error

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
	os.Remove("./testdata/bindata.go")
}

func TestFname(t *testing.T) {
	// test fname option
	var err error
	var outputName string = "fname.go"

	cmd := exec.Command("go", "run", "..", "-fname", outputName, "./target/")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	fi, err := os.Stat("./testdata/" + outputName)
	if os.IsNotExist(err) {
		t.Error("fname test failed to produce output")
		return
	} else if err != nil {
		panic(err)
	}

	if fi.Name() != outputName {
		t.Error("fname file name not correct.\nExpected: ", outputName, "\nGot: ", fi.Name())
	}

	if err = os.Remove("./testdata/" + outputName); err != nil {
		panic(err)
	}
}

func TestName(t *testing.T) {
	// test fname option
	var err error
	var regexString string = "^func %s()"
	var outputName string = "name"

	cmd := exec.Command("go", "run", "..", "-name", outputName, "./target/")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	fi, err := os.Stat("./testdata/" + outputName + ".go")
	if os.IsNotExist(err) {
		t.Error("fname test failed to produce output")
		return
	} else if err != nil {
		panic(err)
	}

	if fi.Name() != outputName+".go" {
		t.Error("name file name not correct.\nExpected: ", outputName, "\nGot: ", fi.Name())
	}

	a := fmt.Sprintf(regexString, outputName)
	r := regexp.MustCompile(a)

	f, err := os.Open("./testdata/" + outputName + ".go")
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(f)

	var found bool
	for s.Scan() {
		if !r.MatchString(s.Text()) {
			found = true
			break
		}
	}
	if !found {
		t.Error("content does not have a func called: ", outputName)
	}

	if err = os.Remove("./testdata/" + outputName + ".go"); err != nil {
		panic(err)
	}

}

func TestDot(t *testing.T) {
}

func TestMain(m *testing.M) {
	os.Remove("./testdata/bindata.go")

	testFiles = findTestFiles()

	r := m.Run()
	os.Remove("./testdata/bindata.go")
	os.Remove("./testdata/readypacked/archive.tar")

	os.Exit(r)
}
