package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	mu             sync.Mutex
	wg             sync.WaitGroup
	foldersScanned int
	root           = flag.String("r", "/", "root directory to start search")
	verbose        = flag.Bool("v", false, "show extra info")
)

func binarySearch(s []fs.FileInfo, t string) bool {
	i, j := 0, len(s)-1
	for i <= j {
		m := (i + j) / 2
                n := s[m].Name()
		if n == t {
			return true
		} else if n > t {
			j = m - 1
		} else {
			i = m + 1
		}
	}
	return false
}

func search(path, filename string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		if *verbose {
			fmt.Printf("\u001b[31m[-] Cannot read %s, skipping...\u001b[0m\n", path)
		}
	}
	if binarySearch(files, filename) {
		fmt.Printf("\u001b[32m[+] Found File: %s\u001b[0m\n", filepath.Join(path, filename))
	}
	for _, file := range files {
		if file.IsDir() {
			mu.Lock()
			foldersScanned++
			mu.Unlock()
			p := filepath.Join(path, file.Name())
			wg.Add(1)
			go func() {
				defer wg.Done()
				search(p, filename)
			}()
		}
	}
}

func main() {
	flag.Parse()
	filename := flag.Arg(0)
	if filename == "" {
		fmt.Fprintln(os.Stderr, "\u001b[31m[!] File not specified\u001b[0m")
		os.Exit(1)
	}
	start := time.Now()
	search(*root, filename)
	wg.Wait()
	fmt.Printf("\n\u001b[33m[*] %d folders searched in %v\u001b[0m\n", foldersScanned, time.Since(start))
}
