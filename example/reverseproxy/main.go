package main

import (
	"log"
	"maestro/proxy"
	"net/http"
	"net/url"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	u1, _ := url.Parse("http://localhost:8080/")
	u2, _ := url.Parse("http://localhost:8081/")

	targets := []*url.URL{u1, u2}

	prxy := proxy.NewMultiHostReverseProxy(targets)

	go func() {
		for {
			time.Sleep(6 * time.Second)
			u, _ := url.Parse("http://localhost:8083/")
			prxy.AddTarget(u)
			time.Sleep(6 * time.Second)
			prxy.RemoveTarget(u)
		}
	}()

	log.Fatal(http.ListenAndServe(":8082", prxy))
}
