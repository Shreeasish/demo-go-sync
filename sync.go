package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type DirPath interface {
	isDirPath()
}

type Dir struct {
	path string
}

type File struct {
	path string
}

func (_ Dir) isDirPath()  {}
func (_ File) isDirPath() {}

func (f File) readfile() ([]byte, error) {
	contents, err := ioutil.ReadFile(f.path)
	return contents, err
}

func dirlist(dir string) (error, []DirPath) {
	var list []DirPath
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Failure accessing path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			list = append(list, Dir{path})
		} else {
			list = append(list, File{path})
		}
		return nil
	})

	return err, list
}

func local(list []DirPath, dirpaths chan<- DirPath) {
	for i := range list {
		dirpaths <- list[i]
	}
	close(dirpaths)
}

func createDir(prefix string, d Dir) {
	fmt.Printf("Created a remote dir at %v/%v\n", prefix, d.path)
	// Should use MkdirAll instead creating the directory structure
	err := os.Mkdir(fmt.Sprintf(prefix+d.path), os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create dir %v\n", err)
	}
}

func populate(filechannel chan<- File, f File) {
	filechannel <- f
}

func write(prefix string, files <-chan File) {
	for f := range files {
		contents, err := f.readfile()
		if err != nil {
			continue // to be attempted later
		}
		err = ioutil.WriteFile(fmt.Sprintf(prefix+f.path), contents, os.ModePerm)
		if err != nil {
			continue // same
		}
	}
}

func remote(remotepath string, channel <-chan DirPath) {
	// controls the number of concurrently open files
	filec := make(chan File, 10)
	for dirpath := range channel {
		switch dp := dirpath.(type) {
		case Dir:
			createDir(remotepath, dp)
		case File:
			populate(filec, dp)
			go write(remotepath, filec)
		default:
			fmt.Printf("Unable to determine type")
		}
	}
}

func main() {
	destination := "./destination/"
	source := "./source/"

	err, list := dirlist(source)
	if err != nil {
		fmt.Printf("Unable to get walk directory at %v\n", source)
	}
	// Use an unbuffered channel for an order guarantee
	c := make(chan DirPath)
	go local(list, c)
	remote(destination, c)
}
