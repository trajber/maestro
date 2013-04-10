package balancer

import (
	"container/ring"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

var (
	ErrNoHostAvailable = errors.New("No host available")
	ErrHostNotFound    = errors.New("Host not found")
)

type LoadBalancer struct {
	targets  *ring.Ring
	lastUsed *ring.Ring
	mu       sync.RWMutex
	*httputil.ReverseProxy
}

func NewLoadBalancer(targets []*url.URL) *LoadBalancer {
	reverse := new(LoadBalancer)
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

func (r *LoadBalancer) chooseNextTarget() (*url.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

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

func (r *LoadBalancer) AddTarget(u *url.URL) {
	r.mu.Lock()
	defer r.mu.Unlock()

	e := &ring.Ring{Value: u}
	r.targets.Link(e)
}

func (r *LoadBalancer) RemoveTarget(u *url.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r == nil {
		return ErrHostNotFound
	}

	var removed *ring.Ring

	i, n := 0, r.targets.Len()

	for p := r.targets; i < n; p = p.Next() {
		url, _ := p.Value.(*url.URL)
		if url.String() == u.String() {
			removed = p.Prev().Unlink(1)
			break
		}
	}

	if removed == nil {
		return ErrHostNotFound
	}

	if r.lastUsed.Next().Value.(*url.URL).String() == u.String() {
		r.lastUsed = r.targets.Next()
	}

	return nil
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
