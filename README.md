Maestro reverse proxy
================================
Maestro is a fast reverse proxy based on [Go's SingleHostReverseProxy] (http://golang.org/pkg/net/http/httputil/#NewSingleHostReverseProxy) that handles mutiple hosts at the same time.

For instance, to proxy incoming request on port 8082 to port 8080 or 8081:
     u1, _ := url.Parse("http://localhost:8080/")
     u2, _ := url.Parse("http://localhost:8081/")
     targets := []*url.URL{u1, u2}
     prxy := proxy.NewMultiHostReverseProxy(targets)
     http.ListenAndServe(":8082", prxy)