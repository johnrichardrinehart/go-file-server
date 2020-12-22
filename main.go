// FROM: https://gist.github.com/paulmach/7271283
// AND: https://stackoverflow.com/questions/57281010/remove-the-html-extension-from-every-file-in-a-simple-http-server
//
/*
Serve is a very simple static file server in go
Usage:
	-p="8100": port to serve on
	-d=".":    the directory of static files to host
Navigating to http://localhost:8100 will display the index.html or directory
listing file.
*/
package main

import (
	"compress/gzip"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
)

// HTMLDir is used to provide a custom Open method
type HTMLDir struct {
    d http.Dir
}

// Open wraps http.Dir.Open, allowing for filenames without extension
func (d HTMLDir) Open(name string) (http.File, error) {
	log.Printf("trying to serve: %s", name)
    // Try name as supplied
    f, err := d.d.Open(name)
    if os.IsNotExist(err) {
        // Not found, try with .html
        if f, err := d.d.Open(name + ".html"); err == nil {
            return f, nil
        }
	}
	if err != nil {
		log.Printf("encountered error: %s", err)
	}
    return f, err
}


type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Middleware allows for doing custom things to the requests and responses after file serving
func Middleware(f http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "max-age=31536000")
		// https://gist.github.com/the42/1956518
		w.Header().Add("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}

		f.ServeHTTP(gzr,r)
	}
}

func main() {
	port := flag.String("p", "8100", "port to serve on")
	directory := flag.String("d", ".", "the directory of static file to host")
	flag.Parse()

	fs := http.FileServer(HTMLDir{http.Dir(*directory)})
	h := Middleware(fs)

	http.Handle("/", http.StripPrefix("/", h))


	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
