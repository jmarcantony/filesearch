package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
    mu sync.Mutex
    wg sync.WaitGroup
    filesScanned, foldersScanned int
    root = flag.String("r", "/", "root directory to start search")
    verbose = flag.Bool("v", false, "show extra info")
)

func search(path, filename string) {
    files, err := ioutil.ReadDir(path)
    if err != nil {
        if *verbose {
            fmt.Printf("\u001b[31m[-] Cannot read %s, skipping...\u001b[0m\n", path)
        }
    }
    for _, file := range files {
        name := file.Name()
        if name == filename {
            fmt.Printf("\u001b[32m[+] Found File: %s\u001b[0m\n", filepath.Join(path, name))
        }
        if file.IsDir() {
            mu.Lock()
            foldersScanned++
            mu.Unlock()
            wg.Add(1)
            go func() {
                defer wg.Done()
                search(filepath.Join(path, name), filename)
            }()
        } else {
            mu.Lock()
            filesScanned++
            mu.Unlock()
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
    fmt.Printf("\n\u001b[33m[*] %d files and %d folders searched in %v\u001b[0m\n", filesScanned, foldersScanned, time.Since(start))
}
