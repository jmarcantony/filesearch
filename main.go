package main

import (
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

var (
	mu             sync.Mutex
	wg             sync.WaitGroup
	foldersScanned int
	root           = flag.String("r", "/", "root directory to start search")
	pattern        = flag.String("p", "", "regex pattern to search for")
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

func search(path, filename string, re *regexp.Regexp) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		if *verbose {
			fmt.Printf("\u001b[31m[-] Cannot read %s, skipping...\u001b[0m\n", path)
		}
	}
	if re == nil {
		if binarySearch(files, filename) {
			fmt.Printf("\u001b[32m[+] Found File: %s\u001b[0m\n", filepath.Join(path, filename))
		}
	}
	for _, file := range files {
		if re != nil {
			name := file.Name()
			if re.MatchString(name) {
				fmt.Printf("\u001b[32m[+] Found File: %s\u001b[0m\n", filepath.Join(path, name))
			}
		}
		if file.IsDir() {
			mu.Lock()
			foldersScanned++
			mu.Unlock()
			p := filepath.Join(path, file.Name())
			wg.Add(1)
			go func() {
				defer wg.Done()
				search(p, filename, re)
			}()
		}
	}
}

func main() {
	flag.Parse()
	filename := flag.Arg(0)
	var re *regexp.Regexp
	if *pattern != "" {
		r, err := regexp.Compile(*pattern)
		if err != nil {
			fmt.Fprintln(os.Stderr, "\u001b[31m[!] Invalid regex\u001b[0m")
			os.Exit(1)
		}
		re = r
	}
	fmt.Println(*pattern)
	if filename == "" && *pattern == "" {
		fmt.Fprintln(os.Stderr, "\u001b[31m[!] File not specified\u001b[0m")
		os.Exit(1)
	}
	start := time.Now()
	search(*root, filename, re)
	wg.Wait()
	fmt.Printf("\n\u001b[33m[*] %d folders searched in %v\u001b[0m\n", foldersScanned, time.Since(start))
}
