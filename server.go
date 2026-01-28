package main

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	PORT      = ":8000"
	SPA_ENTRY = "/example/index.html"
)

func main() {
	fmt.Printf("starting server http://127.0.0.1%s ...\n", PORT)
	err := http.ListenAndServe(PORT, http.HandlerFunc(serveStatic))
	if err != nil {
		fmt.Printf("Oops %s\n", err.Error())
	}
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		fmt.Printf("%s HTTP %s %s %s\n", start.Format("2006-01-02 15:04:05"), r.Method, r.URL.Path, time.Since(start))
	}()

	fs := http.Dir(".")
	path := r.URL.Path

	// 1. Attempt to find the file or directory
	f, err := fs.Open(path)
	if err == nil {
		stat, statErr := f.Stat()
		if statErr == nil && stat.IsDir() {
			// Redirect if missing trailing slash for directory
			if !strings.HasSuffix(path, "/") {
				f.Close()
				http.Redirect(w, r, path+"/", http.StatusMovedPermanently)
				return
			}
			// Check for index.html in the directory
			indexF, indexErr := fs.Open(path + "index.html")
			if indexErr == nil {
				f.Close()
				f = indexF
				path = path + "index.html"
			} else {
				// No index.html: force fallback
				f.Close()
				err = os.ErrNotExist
			}
		} else if statErr != nil {
			f.Close()
			err = statErr
		}
	}

	// 2. Fallback to SPA entry point if file not found
	if err != nil {
		path = SPA_ENTRY
		f, err = fs.Open(path)
		if err != nil {
			http.Error(w, "Oops", http.StatusNotFound)
			return
		}
	}
	defer f.Close()

	// 3. Prepare metadata (Content-Type, ETag)
	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Determine Content-Type
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		// Sniff content type
		// We only read a bit, then seek back
		sniff := make([]byte, 512)
		n, _ := f.Read(sniff)
		f.Seek(0, 0)
		ctype = http.DetectContentType(sniff[:n])
	}
	w.Header().Set("Content-Type", ctype)

	// ETag generation (simple weak ETag based on ModTime and Size)
	etag := fmt.Sprintf(`"%x-%x"`, stat.ModTime().Unix(), stat.Size())
	w.Header().Set("ETag", etag)

	// Check If-None-Match
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// 4. Check for Compression (Brotli / Gzip)
	// We check if a pre-compressed file exists (path + .br or .gz)
	acceptEncoding := r.Header.Get("Accept-Encoding")

	if strings.Contains(acceptEncoding, "br") {
		if fbr, err := fs.Open(path + ".br"); err == nil {
			defer fbr.Close()
			if statBr, err := fbr.Stat(); err == nil {
				w.Header().Set("Content-Encoding", "br")
				w.Header().Add("Vary", "Accept-Encoding")
				http.ServeContent(w, r, filepath.Base(path), statBr.ModTime(), fbr)
				return
			}
		}
	}

	if strings.Contains(acceptEncoding, "gzip") {
		if fgz, err := fs.Open(path + ".gz"); err == nil {
			defer fgz.Close()
			if statGz, err := fgz.Stat(); err == nil {
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Add("Vary", "Accept-Encoding")
				http.ServeContent(w, r, filepath.Base(path), statGz.ModTime(), fgz)
				return
			}
		}
	}

	// 5. Serve the original file
	http.ServeContent(w, r, filepath.Base(path), stat.ModTime(), f)
}
