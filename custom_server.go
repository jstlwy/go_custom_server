// Experimenting with custom routing and manual file serving.

package main

import (
	//"encoding/json"
	"fmt"
	"io/fs"
	"log"
	//"net"
	"net/http"
	"os"
	//"path"
	"path/filepath"
	"regexp"
	//"strconv"
	"strings"
	//"time"
)

var usedFileTypes = [...]string{
	".html",
	".css",
	".js",
}
var expectedFileExtensions map[string]bool
var fileCache map[string]string
var fileRegex = regexp.MustCompile(`\.[[:alpha:]]{2,4}$`)

func SetContentType(w http.ResponseWriter, extension string) {
	switch extension {
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "text/javascript")
	}
}

func CustomServeHTTP(w http.ResponseWriter, r *http.Request) {
	extension := fileRegex.FindString(r.URL.Path)
	if len(extension) > 0 {
		// For non-cached files:
		//http.ServeFile(w, r, path[0] + ".html")
		contents, ok := fileCache[r.URL.Path[1:]]
		if ok {
			log.Printf("Requested %s file: %s\n", extension, r.URL.Path)
			SetContentType(w, extension)
			fmt.Fprint(w, contents)
		} else {
			log.Printf("Requested nonexistent %s file: %s\n", extension, r.URL.Path)
			http.NotFound(w, r)
		}
		return
	}

	path := strings.Split(r.URL.Path, "/")[1:]
	n := len(path)
	log.Printf("Requested URL \"%s\"; %d segments: %v\n", r.URL.Path, n, path)
	switch {
	case n == 1 && path[0] == "":
		fmt.Fprintf(w, "<h1>Home: %s</h1>", r.URL.Path)
	case n == 1 && path[0] == "contact":
		fmt.Fprint(w, "<h1>Contact</h1>")
	case path[0] == "calculator" && n == 1:
		http.Redirect(w, r, r.URL.Path + "/index.html", 301)
	default:
		log.Printf("Not found: %s\n", r.URL.Path)
		http.NotFound(w, r)
	}
}

func init() {
	// Only permit caching and transfer of certain file types
	expectedFileExtensions = make(map[string]bool)
	for _, extension := range usedFileTypes {
		expectedFileExtensions[extension] = true
	}

	// Automatically locate and cache all static assets
	fileCache = make(map[string]string)
	log.Println("Searching for static assets...")
	err := filepath.Walk("./static", func(currentPath string, info fs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		extension := fileRegex.FindString(currentPath)
		if len(extension) > 0 {
			_, ok := expectedFileExtensions[extension]
			if ok {
				log.Printf("Caching: %s\n", currentPath)
				data, readErr := os.ReadFile(currentPath)
				if readErr != nil {
					log.Println(readErr)
				} else {
					currentPath = strings.TrimPrefix(currentPath, "static/")
					fileCache[currentPath] = string(data)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("Error walking directory: %v\n", err)
	}

	if len(fileCache) == 0 {
		log.Fatalln("Failed to cache any files.")
	}

	log.Println("Confirming cached files:")
	for filename, _ := range fileCache {
		fmt.Println(filename)
	}
}

func main() {
	addr := ":8000"
	log.Printf("Starting server at address: %s\n", addr)
	handler := http.HandlerFunc(CustomServeHTTP)
	err := http.ListenAndServe(addr, handler)
	log.Fatalln(err)
}

