package main

import (
	"fmt"
	gim "github.com/ozankasikci/go-image-merge"
	"image/jpeg"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		log.Fatalln("too few arguments")
	}
	baseDir := args[0]
	outputDir := args[1]
	stat, err := os.Stat(baseDir)
	if err != nil {
		log.Fatalln(err)
	}
	if !stat.IsDir() {
		log.Fatalf("%s is not a directory\n", baseDir)
	}
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		log.Fatalln(err)
	}
	opt := gim.OptBaseDir(baseDir)
	wg := sync.WaitGroup{}
	var i, j int
	i = findFileFromIndex(files, 0)
	j = findFileFromIndex(files, i+1)
	for i >= 0 && j >= 0 {
		wg.Add(1)
		go stitch(&wg, outputDir, opt, files[i], files[j])
		i = findFileFromIndex(files, j+1)
		j = findFileFromIndex(files, i+1)
	}
	wg.Wait()
}

func findFileFromIndex(files []fs.FileInfo, i int) int {
	for j := i; j < len(files); j++ {
		if strings.HasSuffix(strings.ToUpper(files[j].Name()), ".JPG") {
			return j
		}
	}
	return -1
}

func stitch(w *sync.WaitGroup, output string, opt func(*gim.MergeImage), left, right fs.FileInfo) {
	defer w.Done()
	grids := []*gim.Grid{
		{ImageFilePath: left.Name()},
		{ImageFilePath: right.Name()},
	}
	rgba, err := gim.New(grids, 2, 1, opt).Merge()
	if err != nil {
		log.Printf("failed to merge image for %s and %s\n", left.Name(), right.Name())
	}

	outFile := fmt.Sprintf("%s_%s.JPG", strings.TrimSuffix(strings.ToUpper(left.Name()), ".JPG"), strings.TrimSuffix(strings.ToUpper(right.Name()), ".JPG"))
	file, err := os.Create(filepath.Join(output, outFile))
	err = jpeg.Encode(file, rgba, &jpeg.Options{Quality: 80})
	if err != nil {
		log.Printf("failed to save merged image for %s and %s\n", left.Name(), right.Name())
	}
	log.Printf("Stitched %s\n", outFile)
}
