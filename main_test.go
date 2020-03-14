package main

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	testDir string = "./testdata/target/"
)

func findTestFiles() []string {
	var files []string
	//returns paths of files in testdata/target
	if err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if testDir == path {
			return nil
		}
		files = append(files, filepath.Base(path))
		return nil
	}); err != nil {
		panic(err)
	}
	return files

}

func TestBinary(t *testing.T) {
	var err error

	testFiles := findTestFiles()

	cmd := exec.Command("go", "run", "..", "./target/")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	cmd = exec.Command("go", "run", ".")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	stdout := bytes.NewBuffer(output)
	scanner := bufio.NewScanner(stdout)

	var failed bool
	for scanner.Scan() {
		var ok bool
		line := scanner.Text()
		for _, f := range testFiles {
			if line == f || line == "EOF" {
				ok = true
			}
		}
		if !ok {
			t.Error("error, output of test program doesn't match testdata files. Following line not in directory", line)
			failed = true
			continue
		}
		t.Log("correctly found file name: ", line)
	}
	if failed {
		for _, f := range testFiles {
			t.Log("looked for: ", f)
		}
	}

}
