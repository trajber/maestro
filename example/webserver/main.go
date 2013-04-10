package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from backend, %q",
			html.EscapeString(r.URL.Path),
		)
	})

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	select {}
}
