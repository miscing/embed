package main

import (
	"archive/tar"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	usage       string = "embed [path(0)]... [path(i)]// embed path dir or file/s into current pwd package"
	tarReminder string = "//variable contains a tar archive"
	preTemplate string = `package %s

//autogenerated by embed

%s
func %s() []byte {
	var bindata = []byte{`

	postTemplate string = `}
	return bindata
}`
)

var (
	packageName string
	funcName    string
	fileName    string
	isTar       bool
)

func findPackageName() error {
	fset := token.NewFileSet()
	fMap, err := parser.ParseDir(fset, ".", nil, parser.PackageClauseOnly)
	if err != nil {
		return err
	}
	if len(fMap) != 1 {
		return errors.New("expected only one package in current directory, found: " + string(len(fMap)))
	}
	var name string
	for k, _ := range fMap {
		if k == "" {
			return errors.New("current pwd package has empty name")
		}
		name = k
	}
	packageName = name
	return nil
}

func openFiles(paths []string) (files []*os.File) {
	for _, p := range paths {
		files = append(files, parsePath(p)...)
	}
	return
}

func makeTar(files []*os.File) *bytes.Buffer {
	buf := new(bytes.Buffer)
	if len(files) == 1 {
		log.Println("only 1 file found, skipping tar archiving")
		// skip tar process if only one file
		_, err := io.Copy(buf, files[0])
		if err != nil {
			log.Panic(err)
		}
		return buf
	}
	isTar = true

	tw := tar.NewWriter(buf)
	for _, f := range files {
		fi, err := f.Stat()
		if err != nil {
			log.Panic(err)
		}
		head, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			log.Panic(err)
		}
		if err := tw.WriteHeader(head); err != nil {
			log.Panic(err)
		}
		if _, err := io.Copy(tw, f); err != nil {
			log.Panic(err)
		}
		f.Close()
	}
	if err := tw.Close(); err != nil {
		log.Panic(err)
	}
	return buf
}

func parsePath(p string) []*os.File {
	var files []*os.File
	if err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if path == p && info.IsDir() { //skip root if dir
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		files = append(files, f)
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return files
}

func makeSource(rawBuf *bytes.Buffer) *bytes.Buffer {
	buf := new(bytes.Buffer)
	isTarStr := ""
	if isTar {
		isTarStr = tarReminder
	}

	_, err := fmt.Fprintf(buf, preTemplate, packageName, isTarStr, funcName)
	if err != nil {
		log.Panic(err)
	}

	raw, err := ioutil.ReadAll(rawBuf)
	if err != nil {
		log.Panic(err)
	}
	for _, b := range raw {
		fmt.Fprintf(buf, "%#v, ", b)
	}

	if _, err = fmt.Fprint(buf, postTemplate); err != nil {
		log.Panic(err)
	}

	return buf
}

func main() {
	// set flags:
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [path0] ... [pathi]\nGenerates a go source file for golang package in current directory containing all files found in given paths. Accessed through 'func bindata() []byte'. If multiple paths or path is a directory files will be packed into a tar archive.\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&funcName, "name", "bindata", "sets generated source files data holding variable name, def bindata. Also sets fname to name + '.go'")
	flag.StringVar(&packageName, "pname", "", "sets generated source files package name instead of parsing from current directory")
	flag.StringVar(&fileName, "fname", "bindata.go", "sets generated source files name, default is bindata.go, use this to avoid overwritting")
	flag.Parse()

	if packageName == "" {
		if err := findPackageName(); err != nil {
			log.Println("embed failed to find a package name to attach data to, quitting")
			log.Fatal(err)
		}
	}
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "name" {
			fileName = f.Value.String() + ".go"
		}
	})

	paths := flag.Args()
	files := openFiles(paths)
	tarBuf := makeTar(files)
	sourceFileBuff := makeSource(tarBuf)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	defer file.Close()
	if err != nil {
		log.Panic(err)
	}
	_, err = sourceFileBuff.WriteTo(file)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("created %s for package %s containing:\n", fileName, packageName)
	fmt.Println(paths)

}
