package main

import (
	"log"
	"maestro/proxy"
	"net/http"
	"net/url"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	u1, _ := url.Parse("http://localhost:8080/")
	u2, _ := url.Parse("http://localhost:8081/")

	targets := []*url.URL{u1, u2}

	prxy := proxy.NewMultiHostReverseProxy(targets)

	log.Fatal(http.ListenAndServe(":8082", prxy))
}
