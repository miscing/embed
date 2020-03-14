package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
)

// for testing puposes

func main() {
	r := tar.NewReader(bytes.NewBuffer(bindata))
	for {
		h, err := r.Next()
		if err != nil {
			fmt.Println(err)
			return
		}
		// if _, err := io.Copy(os.Stdout, r); err != nil {
		// 	log.Fatal(err)
		// }
		if err == io.EOF {
			fmt.Println("done")
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(h.Name)
	}
}
