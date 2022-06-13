package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/chonlatee/hclient/htpclient"
)

type rtLog struct {
	next http.RoundTripper
}

func justrtLog() htpclient.RTOption {
	return func(r http.RoundTripper) http.RoundTripper {
		return rtLog{
			next: r,
		}
	}
}

func (r rtLog) RoundTrip(req *http.Request) (*http.Response, error) {

	fmt.Printf("Start: %v\n", time.Now())
	defer fmt.Printf("End: %v\n", time.Now())

	return r.next.RoundTrip(req)
}

type rtCache struct {
	next  http.RoundTripper
	cache *sync.Map
}

func justrtCache() htpclient.RTOption {
	return func(r http.RoundTripper) http.RoundTripper {
		return &rtCache{
			next:  r,
			cache: &sync.Map{},
		}
	}
}

func (r *rtCache) RoundTrip(req *http.Request) (*http.Response, error) {

	v, ok := r.cache.Load(req.URL.String())
	if ok {
		b := bufio.NewReader(bytes.NewBuffer(v.([]byte)))
		resp, err := http.ReadResponse(b, nil)
		if err != nil {
			return nil, err
		}

		fmt.Print("Hit cache get value from cache.")

		return resp, nil
	}

	res, err := r.next.RoundTrip(req)
	if err != nil {
		return res, err
	}

	resp, err := httputil.DumpResponse(res, true)
	if err != nil {
		return res, err
	}

	fmt.Print("Not hit cache get value from new request.\n")
	r.cache.Store(req.URL.String(), resp)
	return r.next.RoundTrip(req)
}

type rtStop struct {
	next http.RoundTripper
	stop bool
}

func justrtStop() htpclient.RTOption {
	return func(r http.RoundTripper) http.RoundTripper {
		return &rtStop{
			next: r,
		}
	}
}

func (r *rtStop) RoundTrip(req *http.Request) (*http.Response, error) {

	if r.stop {
		fmt.Printf("new request need to wait for 2 second.\n")
		time.Sleep(time.Second * 2)
		fmt.Printf("now can go to request.\n")
	}

	res, err := r.next.RoundTrip(req)
	if err != nil || res.StatusCode == http.StatusBadGateway {
		fmt.Printf("error occur stop from new request\n")
		go func() {
			r.stopRequest()
		}()
		r.stop = true
	}

	return res, err
}

func (r *rtStop) stopRequest() {
	fmt.Printf("sleep 10 second\n")
	time.Sleep(time.Second * 10)
	fmt.Printf("Now system can accept new quest.\n")
	r.stop = false
}

func main() {

	c := htpclient.New()

	for i := 0; i < 20; i++ {

		url := "https://httpbin.org/get"

		if i == 4 {
			url = "https://httpbin.org/status/502"
		}

		r, err := c.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		defer r.Body.Close()

		fmt.Printf("request number %v\n", i+1)

		// b, err := io.ReadAll(r.Body)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Printf("%v", string(b))
	}
}
