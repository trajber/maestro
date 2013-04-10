package proxy

import (
	"container/ring"
	"errors"
	"fmt"
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

type MultiHostReverseProxy struct {
	targets  *ring.Ring
	lastUsed *ring.Ring
	mu       sync.RWMutex
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

		log.Println("Sending request to", target)

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

func (r *MultiHostReverseProxy) AddTarget(u *url.URL) {
	r.mu.Lock()
	defer r.mu.Unlock()

	e := &ring.Ring{Value: u}
	r.targets.Link(e)
	log.Println("Added", u)
	dump(r.targets)
}

func dump(r *ring.Ring) {
	if r == nil {
		fmt.Println("empty")
		return
	}

	i, n := 0, r.Len()

	for p := r; i < n; p = p.Next() {
		fmt.Printf("%4d: %p(%s) = {<- %p(%s) | %p(%s) ->}\n",
			i,
			p,
			p.Value,
			p.Prev(),
			p.Prev().Value,
			p.Next(),
			p.Next().Value,
		)
		i++
	}

	fmt.Println()
}

func (r *MultiHostReverseProxy) RemoveTarget(u *url.URL) error {
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

	log.Println("Last used is", r.lastUsed.Value)

	if r.lastUsed.Next().Value.(*url.URL).String() == u.String() {
		r.lastUsed = r.targets.Next()
		log.Println("Last used to", r.targets.Next().Value)
	}

	log.Println("Removed", removed)
	dump(r.targets)

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
