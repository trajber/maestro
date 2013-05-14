Maestro HTTP load balancer
================================
Maestro is a fast HTTP load balancer.It uses [Go's SingleHostReverseProxy] (http://golang.org/pkg/net/http/httputil/#NewSingleHostReverseProxy) to takes an incoming request and sends it to another server.

## Use

For instance, to send incoming requests on port 8080 to hosts 1.1.1.1 and 2.2.2.2:

	package main

	import (
		"maestro/balancer"
		"net/http"
		"net/url"
	)

	func main() {
		u1, _ := url.Parse("http://1.1.1.1:8080/")
		u2, _ := url.Parse("http://2.2.2.2:8080/")
		targets := []*url.URL{u1, u2}
		lb := balancer.NewLoadBalancer(targets)
		log.Fatal(http.ListenAndServe(":8080", lb))
	}

You can also change the target hosts dinamically:

	package main

	import (
		"maestro/balancer"
		"net/http"
		"net/url"
	)

	u1, _ := url.Parse("http://1.1.1.1:8080/")
	u2, _ := url.Parse("http://2.2.2.2:8080/")

	targets := []*url.URL{u1, u2}

	lb := balancer.NewLoadBalancer(targets)

	go func() {
		// Adding and removing a new host dinamically every 6 seconds
		for {
			time.Sleep(6 * time.Second)
			u, _ := url.Parse("http://3.3.3.3:8080/")
			lb.AddTarget(u)
			time.Sleep(6 * time.Second)
			lb.RemoveTarget(u)
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", lb))
