package main

import (
	"os/exec"
)

func TestParse(t *testing.T) {
	paths := []string{"./testdata/target/"}
	out := make(chan *bytes.Buffer)
	wg := new(sync.WaitGroup)
	go closer(out, wg)
	for _, p := range paths {
		go parse(p, out, wg)
	}
	// r := tar.NewReader(bytes.NewBuffer(data))
	// for {
	// 	h, err := r.Next()
	// 	if err == io.EOF {
	// 		fmt.Println("done")
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println(h.Name)
	// }
}
