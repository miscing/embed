package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

const (
	usage       string = "embed [path(0)]... [path(i)]// embed path dir or file/s into current pwd package"
	tarReminder string = "//variable contains a tar archive"
	begTemplate string = "package %s\n//autogenerated by embed\n%s\n\nvar %s = []byte{ "
	endTemplate string = "}"
)

var (
	packageName  string
	variableName string
	fileName     string
	isTar        bool
)

func findPackageName() error {
	var packageNames []string
	absLoc, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	if err := filepath.Walk(absLoc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == absLoc { //skip root directory
			return nil
		}
		if info.IsDir() {
			return filepath.SkipDir
		}
		if filepath.Ext(path) == ".go" {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			s := bufio.NewScanner(f)
			re := regexp.MustCompile("^package ([[:alnum:]]*$)")
			for i := 0; s.Scan(); i++ {
				if i >= 10 {
					break
				}
				if re.MatchString(s.Text()) {
					matches := re.FindStringSubmatch(s.Text())
					for i, m := range matches {
						if i == 1 {
							packageNames = append(packageNames, m)
							break
						}
					}
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if len(packageNames) == 0 {
		return errors.New("couldn't find package name")
	}
	for i := 0; i < len(packageNames); i++ {
		for _, nameB := range packageNames[i+1:] {
			if nameB != packageNames[i] {
				return errors.New("multiple package names in current directory")
			}
		}
	}
	packageName = packageNames[0]
	return nil
}

func makeTar(paths []string) *bytes.Buffer {
	var files []*os.File
	buf := new(bytes.Buffer)
	for _, p := range paths {
		files = append(files, parsePath(p)...)
	}
	if len(files) == 1 {
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
		log.Panic(err)
	}
	return files
}

func makeSource(rawBuf *bytes.Buffer) *bytes.Buffer {
	buf := new(bytes.Buffer)
	isTarStr := ""
	if isTar {
		isTarStr = tarReminder
	}
	_, err := fmt.Fprintf(buf, begTemplate, packageName, isTarStr, variableName)
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
	_, err = buf.WriteString(endTemplate)
	if err != nil {
		log.Panic(err)
	}
	return buf
}

func main() {
	// set flags:
	flag.Bool("0", true, "embed [path0] ... [pathi] // generates a file {fname} with a variable []byte named {vname} for current directory go project\nfor single files it imply makes a bit copy , for multiple files embed first packs them into a tar archive.\nFor a http.FileSystem implementation look at other software\nused to package binary, image or config files into application binary")
	flag.StringVar(&variableName, "vname", "bindata", "sets generated source files data holding variable name, def bindata")
	flag.StringVar(&packageName, "pname", "", "sets generated source files package name instead of parsing from current directory")
	flag.StringVar(&fileName, "fname", "bindata.go", "sets generated source files name, default is bindata.go, use this to avoid overwritting")
	flag.Parse()
	paths := flag.Args()

	if packageName == "" {
		if err := findPackageName(); err != nil {
			log.Panic(err)
		}
	}

	tarBuf := makeTar(paths)
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
	fmt.Printf("created %s for package %s containing a tar archieve of:\n", fileName, packageName)
	fmt.Println(paths)

}
