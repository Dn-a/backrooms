package main

/*
 * Copyright (c) 2023, Antonino Di Natale | https://www.dn-a.it
 *
 */

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gorilla/handlers"
)

const (
	CONFIG_FILE_NANE = "config.yml"
	REDIRECT         = "REDIRECT"
	REVERSE_PROXY    = "REVERSE-PROXY"
	JOLLY            = "**"
)

type customTransport struct {
	http.Transport
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	//log.Println("custom RoundTrip")
	res, err := http.DefaultTransport.RoundTrip(req)
	return res, err
}

func (c *Configurations) Matchers(path string) (*Resource, bool) {
	check := false
	slicedPath := strings.Split(path, "/")[1:]

	//var queryString string

	log.Printf("sliced path: %v", slicedPath)

	var matchers *map[string]any = c.RequestMatchers
	var rsc *Resource

	for _, p := range slicedPath {
		//Remove query string
		if strings.Contains(p, "?") {
			p = strings.Split(p, "?")[0]
		}

		m, ok := (*matchers)[p]

		if ok {
			if r, rOk := m.(*Resource); rOk {
				rsc = r
				check = true
				break
			} else {
				log.Printf("Ok %v", m)
				matchers = m.(*map[string]interface{})
			}
		} else if m2, ok2 := (*matchers)[JOLLY]; ok2 {
			if r, rOk := m2.(*Resource); rOk {
				log.Printf("Jolly %v", m2)
				rsc = r
				check = true
				break
			} else {
				matchers = m2.(*map[string]interface{})
			}
		} else {
			break
		}
	}

	return rsc, check
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

func generateRequestMatchers(config *Configurations) *map[string]interface{} {
	m := make(map[string]interface{})

	for _, resource := range config.Resources {
		slicedMatchers := strings.Split(resource.Matchers, "/")[1:]
		if len(slicedMatchers) > 0 {
			resourceAddress := resource
			mainKey := slicedMatchers[0]
			slice := slicedMatchers[1:]
			if len(slice) == 0 {
				m[mainKey] = &resourceAddress
			} else {
				m[mainKey] = matchersRecursion(&slice, &resourceAddress)
			}
		}
	}
	return &m
}

func matchersRecursion(matchers *[]string, resource *Resource) *map[string]interface{} {
	mp := make(map[string]interface{})
	matcher := (*matchers)[0]

	if len(*matchers) == 1 {
		mp[matcher] = resource
		return &mp
	} else {
		slice := (*matchers)[1:]
		mp[matcher] = matchersRecursion(&slice, resource)
		return &mp
	}
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

		if req.Method == http.MethodPost {
			buf, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(buf))
			log.Printf("REQUEST BODY: %s\n", buf)
		}

		// Here is where the magic happens
		//
		// if resource == auth (E.G: keycloak request) then the request will be redirect to appropiate Keycloak URL
		if forwardType == REDIRECT {
			//w.Header().Set("Content-Type", "application/json")
			http.Redirect(w, req, req.URL.String(), http.StatusTemporaryRedirect)
		} else {
			reverseProxy(url).ServeHTTP(w, req)
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

	handlers.MaxAge(3600)
	origins := handlers.AllowedOrigins([]string{"*"})

	//headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"})
	//methods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	credentials := handlers.AllowCredentials()

	hdl := handlers.CORS(
		origins,
		//headers,
		//methods,
		credentials,
	)

	// E.G: 4300
	var port = DEFAULT_PORT
	if config.Port != "" {
		port = config.Port
	}

	log.Printf("Backrooms started on port %v", port)

	// TODO: TLS
	/* client := &http.Client{
	        //Timeout: 20 * time.Second,
	        Transport: &http.Transport{
	                TLSHandshakeTimeout: 700 * time.Millisecond,
	        },
	} */

	err := http.ListenAndServe(":"+port, hdl(mux))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
