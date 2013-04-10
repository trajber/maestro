package main

import (
	"log"
	"maestro/balancer"
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

	lb := balancer.NewLoadBalancer(targets)

	go func() {
		for {
			time.Sleep(6 * time.Second)
			u, _ := url.Parse("http://localhost:8082/")
			lb.AddTarget(u)
			time.Sleep(6 * time.Second)
			lb.RemoveTarget(u)
		}
	}()

	log.Fatal(http.ListenAndServe(":8083", lb))
}
