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
		absFilePath, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		files = append(files, absFilePath)
		return nil
	}); err != nil {
		panic(err)
	}
	return files

}

func checkTestProgAgainst(expected []string, output []byte, t *testing.T) {
	stdout := bytes.NewBuffer(output)
	scanner := bufio.NewScanner(stdout)

	var failed bool
	for scanner.Scan() {
		var ok bool
		line := scanner.Text()
		for _, f := range expected {
			if line == filepath.Base(f) || line == "EOF" {
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
		for _, f := range expected {
			t.Log("looked for: ", f)
		}
	}
}

func TestSingleArg(t *testing.T) {
	var err error
	os.Remove("./testdata/bindata.go")

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

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	checkTestProgAgainst(testFiles, output, t)
}

func TestMultiArg(t *testing.T) {
	var err error
	os.Remove("./testdata/bindata.go")

	testFiles := findTestFiles()
	allArgs := []string{"run", ".."}
	allArgs = append(allArgs, testFiles...)
	cmd := exec.Command("go", allArgs...)
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

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	checkTestProgAgainst(testFiles, output, t)

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
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	panic(err)
	// }
	// t.Error(string(output))

	cmd = exec.Command("go", "run", ".")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	testFiles := findTestFiles()

	checkTestProgAgainst(testFiles, output, t)

}

func TestMain(m *testing.M) {

	r := m.Run()
	if err := os.Remove("./testdata/bindata.go"); err != nil {
		panic(err)
	}
	if err := os.Remove("./testdata/readypacked/archive.tar"); err != nil {
		panic(err)
	}

	os.Exit(r)
}
