package main

/*
 * Copyright (c) 2023, Antonino Di Natale | https://www.dn-a.it
 *
 */

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
)

const (
	DEFAULT_FILE_NANE = "config.yml"
	REDIRECT          = "REDIRECT"
	REVERSE_PROXY     = "REVERSE-PROXY"
	JOLLY             = "**"
)

type customTransport struct {
	http.Transport
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	//log.Println("custom RoundTrip")
	res, err := http.DefaultTransport.RoundTrip(req)
	return res, err
}

func slicedUrl(target string) []string {
	slice := strings.Split(target, "://")
	if len(slice) != 2 {
		log.Fatalf("Invalid target URI")
	}
	//log.Printf("Slice: %s", slice)
	return slice
}

func getURIResource(target string) string {
	uriSlice := strings.Split(target, "/")
	resource := ""
	if len(uriSlice) > 1 {
		resource = uriSlice[1]
		//log.Printf("resource: %s", resource)
	}
	return resource
}

func reverseProxy(url string) *httputil.ReverseProxy {

	director := func(req *http.Request) {
		//log.Printf("custom Director")
	}

	dial := func(network, addr string) (net.Conn, error) {
		//log.Println("custom Dial")
		return net.Dial(network, addr)
	}

	transport := &customTransport{http.Transport{Dial: dial}}

	return &httputil.ReverseProxy{Director: director, Transport: transport}
}

func forward(callback func(path string) (string, string)) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// return first resource from url -> http://localhost/resource1/resource2 -> resource1
		resource := getURIResource(req.URL.String())
		url, forwardType := callback(req.URL.String())
		slicedUrl := slicedUrl(url)
		scheme, host := slicedUrl[0], slicedUrl[1]

		req.URL.Scheme = scheme
		req.URL.Host = host
		req.Host = host

		forwardType = strings.ToUpper(forwardType)

		log.Printf("%v", resource)
		log.Printf("%v", forwardType)
		log.Printf("%v: %v", req.Method, req.URL)

		if req.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, PATCH, POST, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", req.Header.Get("access-control-request-headers"))
			fmt.Fprintln(w, "test")
		} else {

			if req.Method == http.MethodPost {
				buf, _ := io.ReadAll(req.Body)
				req.Body = io.NopCloser(bytes.NewBuffer(buf))
				log.Printf("REQUEST BODY: %s\n", buf)
			}

			// Here is where the magic happens
			//
			// if resource == auth (E.G: keycloak request) then the request will be redirect to appropiate Keycloak URL
			if forwardType == REDIRECT {
				http.Redirect(w, req, req.URL.String(), http.StatusTemporaryRedirect)
			} else {
				reverseProxy(url).ServeHTTP(w, req)
			}
		}

	})

}

const DEFAULT_PORT = "4300"
const DEFAULT_URL = "http://localhost"

func main() {

	config, _ := GetConfig()

	mux := http.NewServeMux()

	mux.Handle("/", forward(func(path string) (string, string) {
		config, _ = GetConfig()

		// E.G:
		// key: foo
		// url: localhost
		//rsc, key := config.Resources[resource]
		rsc, ok := config.Matchers(path)

		url := DEFAULT_URL
		forwardType := REVERSE_PROXY

		if ok && len(rsc.Type) > 0 {
			forwardType = rsc.Type
		}

		// if key doesn't exists, default url will be used
		if ok && rsc.Url != "" {
			url = rsc.Url
		} else if config.DefaultUrl != "" {
			url = config.DefaultUrl
		}

		// url, type (redirect || reverse-proxy)
		return url, forwardType
	}))

	//handlers.MaxAge(3600)
	//origins := handlers.AllowedOrigins([]string{"*"})
	//headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"})
	//methods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})

	// E.G: 4300
	var port = DEFAULT_PORT
	if config.Port != "" {
		port = config.Port
	}

	log.Printf("Backrooms started on port %v", port)

	err := http.ListenAndServe(":"+port, mux)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
