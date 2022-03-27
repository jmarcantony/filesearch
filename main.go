package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var (
	foldersScanned, filesFound int
	mu                         sync.Mutex
	wg                         sync.WaitGroup
	red                        = color.New(color.FgRed)
	root                       = flag.String("r", "/", "root directory to start search")
	pattern                    = flag.String("p", "", "regex pattern to search for")
	verbose                    = flag.Bool("v", false, "show extra info")
	fuzz                       = flag.Bool("f", false, "fuzzy search for filename")
)

func search(path, filename string, re *regexp.Regexp) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		if *verbose {
			mu.Lock()
			color.Red("[-] Cannot read %s, skipping...", path)
			mu.Unlock()
		}
	}
	for _, file := range files {
		if *fuzz {
			if fuzzy.Match(file.Name(), filename) {
				mu.Lock()
				color.Green("[+] Found File: %s", filepath.Join(path, file.Name()))
				filesFound++
				mu.Unlock()

			}
		} else {

			if re != nil {
				if re.MatchString(file.Name()) {
					mu.Lock()
					color.Green("[+] Found File: %s", filepath.Join(path, file.Name()))
					filesFound++
					mu.Unlock()
				}
			} else if file.Name() == filename {
				mu.Lock()
				color.Green("[+] Found File: %s", filepath.Join(path, file.Name()))
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
			red.Fprintln(os.Stderr, "[!] Invalid regex")
			os.Exit(1)
		}
		re = r
	}
	if filename == "" && *pattern == "" {
		red.Fprintln(os.Stderr, "[!] File not specified")
		os.Exit(1)
	}
	start := time.Now()
	search(*root, filename, re)
	wg.Wait()
	color.Yellow("\n[*] %d Matches found, %d folders searched in %v", filesFound, foldersScanned, time.Since(start))
}
