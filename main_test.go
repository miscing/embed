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
	"archive/tar"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/inspector"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

const (
	testDir string = "./testdata/target/"
)

var (
	ignore string = "// +build ignore\n"
	// testFiles []*testFile
	tFiles testFiles
)

type testFiles []*testFile

type testFile struct {
	Name     string
	Content  string
	IsDir    bool
	IsHidden bool
	IsChild  bool
}

func printTF(tfs []*testFile) {
	for _, f := range tfs {
		fmt.Println(f.Name)
		fmt.Println("\tis dir:", f.IsDir)
		fmt.Println("\tis hidden", f.IsHidden)
		fmt.Println("\tis child", f.IsChild)
	}
}

func (tf *testFiles) Default() []*testFile {
	files := make([]*testFile, 0, 10)
	for _, f := range *tf {
		switch {
		case f.IsHidden:
			continue
		case f.IsDir:
			continue
		case f.IsChild:
			continue
		}
		files = append(files, f)
	}
	return files
}

func (tf *testFiles) NoDirRH() []*testFile {
	files := make([]*testFile, 0, 10)
	for _, f := range *tf {
		switch {
		case f.IsDir:
			continue
		}
		files = append(files, f)
	}
	return files
}

func (tf *testFiles) NoDirR() []*testFile {
	files := make([]*testFile, 0, 10)
	for _, f := range *tf {
		switch {
		case f.IsHidden:
			continue
		case f.IsDir:
			continue
		}
		files = append(files, f)
	}
	return files
}

func (tf *testFiles) NoDir() []*testFile {
	files := make([]*testFile, 0, 10)
	for _, f := range *tf {
		switch {
		case f.IsHidden:
			continue
		case f.IsChild:
			continue
		case f.IsDir:
			continue
		}
		files = append(files, f)
	}
	return files
}

func (tf *testFiles) Recurssive() []*testFile {
	files := make([]*testFile, 0, 10)
	for _, f := range *tf {
		switch {
		case f.IsHidden:
			continue
		case f.IsDir:
		case f.IsChild:
		}
		files = append(files, f)
	}
	return files
}

func (tf *testFiles) Hidden() []*testFile {
	files := make([]*testFile, 0, 10)
	for _, f := range *tf {
		switch {
		case f.IsDir:
			continue
		case f.IsChild:
			continue
		}
		files = append(files, f)
	}
	return files
}

func isHidden(root string, path string) bool {
	for _, n := range strings.Split(strings.TrimPrefix(path, filepath.Clean(root)), string(filepath.Separator)) {
		if r, _ := utf8.DecodeRuneInString(n); string(r) == "." {
			return true
		}
	}
	return false
}

func isChild(root string, path string) bool {
	if filepath.Clean(root) != filepath.Dir(path) {
		return true
	}
	return false
}

func walkTest() []*os.File {
	var files []*os.File
	if err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if testDir == path {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		files = append(files, f)
		return nil
	}); err != nil {
		panic(err)
	}
	return files
}

func findTestFiles() testFiles {
	var files []*testFile
	if err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == testDir && info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		var c []byte
		if !info.IsDir() {
			c, err = ioutil.ReadAll(f)
			if err != nil {
				return nil
			}
		}
		// files = append(files, &testFile{Name: filepath.Base(f.Name()), Content: string(c), IsHidden: isHidden(info.Name()), IsDir: info.IsDir(), IsChild: isChild(testDir, path)})
		files = append(files, &testFile{Name: filepath.Base(f.Name()), Content: string(c), IsHidden: isHidden(testDir, path), IsDir: info.IsDir(), IsChild: isChild(testDir, path)})
		return nil
	}); err != nil {
		return nil
	}
	return files
}

func checkTestProgAgainst(t *testing.T, toTest []*testFile) {
	var err error
	cmd := exec.Command("go", "run", ".")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(cmd.String())
		return
	}

	dec := json.NewDecoder(bytes.NewBuffer(output))
	dec.DisallowUnknownFields()
	var hol = make([]*testFile, 0, len(toTest))
	for dec.More() {
		resF := new(testFile)
		err := dec.Decode(resF)
		if err != nil {
			t.Error("failed to decode a file from packaged variable")
			break
		}
		hol = append(hol, resF)
	}

	if len(hol) != len(toTest) {
		t.Error("number of packed files and test files did not match.\nTest had: ", len(toTest), "\nOutput had: ", len(hol))
		printFiles(t, hol, toTest)
		return
	}

Beg:
	for _, o := range hol {
		t.Log("output had: ", o.Name)
		for _, tf := range toTest {
			if areEqual(tf, o) {
				continue Beg
			}
		}
		t.Error("failed to find a match in testfiles of: ", o.Name)
		printFiles(t, hol, toTest)
		return
	}
}

func areEqual(f0 *testFile, f1 *testFile) bool {
	if filepath.Clean(f0.Name) == filepath.Clean(f1.Name) && f1.Content == f0.Content {
		return true
	}
	return false
}

func printFiles(t *testing.T, hol, toTest []*testFile) {
	t.Log("files inside test set:")
	for _, tf := range toTest {
		t.Log("\t", tf.Name)
	}
	t.Log("files inside output set:")
	for _, tf := range hol {
		t.Log("\t", tf.Name)
	}
}

func TestMakeTar(t *testing.T) {
	files := walkTest()
	m := new(Maker)
	out := m.MakeTar(files)
	r := tar.NewReader(out)
	for {
		h, err := r.Next()
		if err == io.EOF {
			return
		} else if err != nil {
			panic(err)
		}
		var mark bool
		for _, f := range files {
			if filepath.Base(f.Name()) == filepath.Clean(h.Name) {
				mark = true
				break
			}
		}
		if !mark {
			t.Log("File not found", h.Name)
			t.Log("Test set:")
			for _, f := range files {
				t.Log("\t", f.Name())
			}
			t.Fatal("function output had file name not found in testdata/target")
		}
	}
}

func TestMakeSource(t *testing.T) {
	const (
		fName = "bindata"
		pName = "main"
	)
	files := walkTest()
	m := new(Maker)
	out := m.MakeTar(files)
	payload := m.MakeSource(out, pName, fName)
	f, err := parser.ParseFile(token.NewFileSet(), "", payload, 0)
	if err != nil {
		t.Log(err)
		t.Fatal("function output not valid go code")
	}
	inspector.New([]*ast.File{f}).Preorder([]ast.Node{
		new(ast.FuncDecl),
		new(ast.File),
	}, func(n ast.Node) {
		switch k := n.(type) {
		case *ast.FuncDecl:
			if k.Name.Name != fName {
				t.Error("output func name: ", k.Name.Name)
			}
		case *ast.File:
			if k.Name.Name != pName {
				t.Error("output file name: ", k.Name.Name)
			}
		}
	})
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
	checkTestProgAgainst(t, tFiles.Default())
	os.Remove("./testdata/bindata.go")
}

func TestMultiArg(t *testing.T) {
	var err error
	testFiles := tFiles.Default()
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
		t.Skip(cmd.String(), " : ", err)
	}
	checkTestProgAgainst(t, tFiles.Default())
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

	var args = []string{"run", "..", "./readypacked/archive.tar"}
	// args = append(args, "./target/")
	cmd = exec.Command("go", args...)
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	checkTestProgAgainst(t, tFiles.NoDirRH())
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

func TestSkipDir(t *testing.T) {
	var err error
	cmd := exec.Command("go", "run", "..", "-skipdir", "./target/")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	checkTestProgAgainst(t, tFiles.NoDir())
	os.Remove("./testdata/bindata.go")
}

func TestSkipDirR(t *testing.T) {
	var err error
	cmd := exec.Command("go", "run", "..", "-skipdir", "-r", "./target/")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	// printTF(tFiles)
	checkTestProgAgainst(t, tFiles.NoDirR())
	os.Remove("./testdata/bindata.go")
}

func TestParseHidden(t *testing.T) {
	var err error
	cmd := exec.Command("go", "run", "..", "-phidden", "./target/")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	checkTestProgAgainst(t, tFiles.Hidden())
	os.Remove("./testdata/bindata.go")
}

func TestParseHiddenR(t *testing.T) {
	var err error
	cmd := exec.Command("go", "run", "..", "-r", "-phidden", "./target/")
	cmd.Dir, err = filepath.Abs("./testdata/")
	if err != nil {
		panic(err)
	}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	checkTestProgAgainst(t, tFiles)
	os.Remove("./testdata/bindata.go")
}

func TestMain(m *testing.M) {
	tFiles = findTestFiles()

	os.Remove("./testdata/bindata.go")
	r := m.Run()
	os.Remove("./testdata/bindata.go")
	os.Remove("./testdata/readypacked/archive.tar")

	os.Exit(r)
}
