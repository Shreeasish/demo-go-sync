package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func dirlist(dir string) []DirPath {
	var list []DirPath
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Failure accessing path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			list = append(list, Dir{strings.SplitAfterN(path, "/", 2)[1]})
		} else {
			list = append(list, File{strings.SplitAfterN(path, "/", 2)[1]})
		}

		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", dir)
	}
	return list
}

func local(list []DirPath, dirpaths chan<- DirPath) {
	for i := range list {
		dirpaths <- list[i]
	}
	close(dirpaths)
}

func createDir(prefix string, d Dir) {
	fmt.Printf("Created a remote dir at %v/%v\n", prefix, d.path)
}

func createFile(prefix string, f File) {
	fmt.Printf("Created a remote file at %v/%v\n", prefix, f.path)
}

func remote(remotepath string, channel <-chan DirPath) {
	for dirpath := range channel {
		switch dp := dirpath.(type) {
		case Dir:
			createDir(remotepath, dp)
		case File:
			createFile(remotepath, dp)
		default:
			fmt.Printf("Unable to determine type")
		}
	}
}

func main() {
	destination := "./destination/"
	source := "./source/"

	list := dirlist(source)
	// Use an unbuffered channel for an order guarantee
	c := make(chan DirPath)
	go local(list, c)
	remote(destination, c)
}
