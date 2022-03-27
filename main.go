package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

var (
	foldersScanned, filesFound int
	mu                         sync.Mutex
	wg                         sync.WaitGroup
	root                       = flag.String("r", "/", "root directory to start search")
	pattern                    = flag.String("p", "", "regex pattern to search for")
	verbose                    = flag.Bool("v", false, "show extra info")
	fuzz                       = flag.Bool("f", false, "fuzzy search for filename")
)

func search(path, filename string, re *regexp.Regexp) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		if *verbose {
			fmt.Printf("\u001b[31m[-] Cannot read %s, skipping...\u001b[0m\n", path)
		}
	}
	for _, file := range files {
		if *fuzz {
			if fuzzy.Match(file.Name(), filename) {
				fmt.Printf("\u001b[32m[+] Found File: %s\u001b[0m\n", filepath.Join(path, file.Name()))
				mu.Lock()
				filesFound++
				mu.Unlock()

			}
		} else {

			if re != nil {
				if re.MatchString(file.Name()) {
					fmt.Printf("\u001b[32m[+] Found File: %s\u001b[0m\n", filepath.Join(path, file.Name()))
					mu.Lock()
					filesFound++
					mu.Unlock()
				}
			} else if file.Name() == filename {
				fmt.Printf("\u001b[32m[+] Found File: %s\u001b[0m\n", filepath.Join(path, file.Name()))
				mu.Lock()
				filesFound++
				mu.Unlock()
			}
		}
		if file.IsDir() {
			p := filepath.Join(path, file.Name())
			wg.Add(1)
			go func() {
				mu.Lock()
				foldersScanned++
				mu.Unlock()
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
	if filename == "" && *pattern == "" {
		fmt.Fprintln(os.Stderr, "\u001b[31m[!] File not specified\u001b[0m")
		os.Exit(1)
	}
	start := time.Now()
	search(*root, filename, re)
	wg.Wait()
	fmt.Printf("\n\u001b[33m[*] %d Matches found, %d folders searched in %v\u001b[0m\n", filesFound, foldersScanned, time.Since(start))
}
