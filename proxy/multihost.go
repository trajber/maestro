package proxy

import (
	"container/ring"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	ErrNoHostAvailable = errors.New("No host available")
)

type MultiHostReverseProxy struct {
	targets  *ring.Ring
	lastUsed *ring.Ring
	*httputil.ReverseProxy
}

func NewMultiHostReverseProxy(targets []*url.URL) *MultiHostReverseProxy {
	reverse := new(MultiHostReverseProxy)
	reverse.targets = slice2ring(targets)

	director := func(req *http.Request) {
		target, err := reverse.chooseNextTarget()
		if err != nil {
			log.Println(err)
			return
		}

		// log.Println("Sending request to", target)

		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}

	reverse.ReverseProxy = &httputil.ReverseProxy{Director: director}

	return reverse
}

func (r *MultiHostReverseProxy) chooseNextTarget() (*url.URL, error) {
	if r.targets.Len() == 0 {
		return nil, ErrNoHostAvailable
	}

	if r.lastUsed == nil {
		r.lastUsed = r.targets
	}

	target, _ := r.lastUsed.Value.(*url.URL)
	r.lastUsed = r.lastUsed.Next()

	return target, nil
}

func singleJoiningSlash(a, b string) string {
	aslash := false
	if a[len(a)-1] == '/' {
		aslash = true
	}

	bslash := false
	if b[0] == '/' {
		bslash = true
	}

	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func slice2ring(values []*url.URL) *ring.Ring {
	r := ring.New(len(values))

	i, n := 0, r.Len()

	for p := r; i < n; p = p.Next() {
		p.Value = values[i]
		i++
	}

	return r
}
