package main

/*
* Di Natale Antonino | https://www.dn-a.it
*/

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"gopkg.in/yaml.v3"
)

const (
	REDIRECT      = "REDIRECT"
	REVERSE_PROXY = "REVERSE-PROXY"
	JOLLY         = "**"
)

type customTransport struct {
	http.Transport
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	//log.Println("custom RoundTrip")
	res, err := http.DefaultTransport.RoundTrip(req)
	return res, err
}

type resource struct {
	Name     string `yaml:"name"`
	Matchers string `yaml:"matchers"`
	Type     string `default:"proxy" yaml:"type"`
	Url      string `yaml:"url"`
}

type configurations struct {
	Port            string              `yaml:"port"`
	DefaultUrl      string              `yaml:"default-url"`
	Resources       map[string]resource `yaml:"resources"`
	UpdatedOn       string
	RequestMatchers *map[string]interface{}
}

func (c *configurations) Matchers(path string) (*resource, bool) {
	check := false
	slicedPath := strings.Split(path, "/")[1:]

	//var queryString string

	log.Printf("sliced path: %v", slicedPath)

	var matchers *map[string]any = c.RequestMatchers
	var rsc *resource

	for _, p := range slicedPath {
		//Remove query string
		if strings.Contains(p, "?") {
			p = strings.Split(p, "?")[0]
		}

		m, ok := (*matchers)[p]

		if ok {
			if r, rOk := m.(*resource); rOk {
				rsc = r
				check = true
				break
			} else {
				log.Printf("Ok %v", m)
				matchers = m.(*map[string]interface{})
			}
		} else if m2, ok2 := (*matchers)[JOLLY]; ok2 {
			if r, rOk := m2.(*resource); rOk {
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

var conf *configurations

func getConfig() (*configurations, error) {
	info, er := os.Stat("config.yaml")
	if er != nil {
		log.Printf("FileInfo: %v", er)
	}
	modTime := info.ModTime().String()

	if conf == nil {
		conf = &configurations{UpdatedOn: modTime}
	}

	if len(conf.Resources) == 0 || conf.UpdatedOn != modTime {
		conf.UpdatedOn = modTime
		readFile, e := os.ReadFile("config.yaml")

		if e != nil {
			log.Printf("File err: %v\n", e)
		}

		log.Printf("GetConfing: Unmarshal config.yaml")

		err := yaml.Unmarshal(readFile, conf)

		conf.RequestMatchers = generateRequestMatchers(conf)

		if err != nil {
			log.Printf("err: %v\n", err)
		}
	}

	//log.Printf("%+v\n", modTime)
	//log.Printf("%+v\n", conf.UpdatedOn)

	return conf, nil
}

func generateRequestMatchers(config *configurations) *map[string]interface{} {
	m := make(map[string]interface{})

	for _, resource := range conf.Resources {
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
			//log.Printf("k:%v recursion: %v res:%v", k, m, resource)
		}
	}
	return &m
}

func matchersRecursion(matchers *[]string, resource *resource) *map[string]interface{} {
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
		/* slicedUrl := slicedUrl(url)
		scheme, host := slicedUrl[0], slicedUrl[1]
		req.URL.Scheme = scheme
		req.URL.Host = host
		req.Host = host */
	}

	dial := func(network, addr string) (net.Conn, error) {
		//log.Println("custom Dial")
		return net.Dial(network, addr)
	}

	//transport := &http.Transport{Dial: dial,}
	/* prints:
	   2016/02/18 13:35:34 (func(string, string) (net.Conn, error))(0x401810)
	   2016/02/18 13:35:34 &http.Transport{idleMu:sync.Mutex{state:0, sema:0x0}, wantIdle:false, idleConn:map[http.connectMethodKey][]*http.persistConn(nil), idleConnCh:map[http.connectMethodKey]chan *http.persistConn(nil), reqMu:sync.Mutex{state:0, sema:0x0}, reqCanceler:map[*http.Request]func()(nil), altMu:sync.RWMutex{w:sync.Mutex{state:0, sema:0x0}, writerSem:0x0, readerSem:0x0, readerCount:0, readerWait:0}, altProto:map[string]http.RoundTripper(nil), Proxy:(func(*http.Request) (*url.URL, error))(nil), Dial:(func(string, string) (net.Conn, error))(0x401810), DialTLS:(func(string, string) (net.Conn, error))(nil), TLSClientConfig:(*tls.Config)(nil), TLSHandshakeTimeout:0, DisableKeepAlives:false, DisableCompression:false, MaxIdleConnsPerHost:0, ResponseHeaderTimeout:0, ExpectContinueTimeout:0, TLSNextProto:map[string]func(string, *tls.Conn) http.RoundTripper(nil), nextProtoOnce:sync.Once{m:sync.Mutex{state:0, sema:0x0}, done:0x0}, h2transport:(*http.http2Transport)(nil)}
	   2016/02/18 13:35:36 custom Director
	   2016/02/18 13:35:36 custom Dial
	*/

	transport := &customTransport{http.Transport{Dial: dial}}
	/* prints:
	   2016/02/18 13:36:45 (func(string, string) (net.Conn, error))(0x401950)
	   2016/02/18 13:36:45 &main.customTransport{Transport:http.Transport{idleMu:sync.Mutex{state:0, sema:0x0}, wantIdle:false, idleConn:map[http.connectMethodKey][]*http.persistConn(nil), idleConnCh:map[http.connectMethodKey]chan *http.persistConn(nil), reqMu:sync.Mutex{state:0, sema:0x0}, reqCanceler:map[*http.Request]func()(nil), altMu:sync.RWMutex{w:sync.Mutex{state:0, sema:0x0}, writerSem:0x0, readerSem:0x0, readerCount:0, readerWait:0}, altProto:map[string]http.RoundTripper(nil), Proxy:(func(*http.Request) (*url.URL, error))(nil), Dial:(func(string, string) (net.Conn, error))(0x401950), DialTLS:(func(string, string) (net.Conn, error))(nil), TLSClientConfig:(*tls.Config)(nil), TLSHandshakeTimeout:0, DisableKeepAlives:false, DisableCompression:false, MaxIdleConnsPerHost:0, ResponseHeaderTimeout:0, ExpectContinueTimeout:0, TLSNextProto:map[string]func(string, *tls.Conn) http.RoundTripper(nil), nextProtoOnce:sync.Once{m:sync.Mutex{state:0, sema:0x0}, done:0x0}, h2transport:(*http.http2Transport)(nil)}}
	   2016/02/18 13:36:46 custom Director
	   2016/02/18 13:36:46 custom RoundTrip
	*/

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
			proxy := reverseProxy(url)
			proxy.ServeHTTP(w, req)
		}

	})

}

const DEFAULT_PORT = "4300"
const DEFAULT_URL = "http://localhost"

func main() {

	config, _ := getConfig()

	mux := http.NewServeMux()

	mux.Handle("/", forward(func(path string) (string, string) {
		config, _ := getConfig()

		// E.G:
		// key: flt
		// destination-url: localhost
		//log.Printf("resources: %v", rsc)
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

		// url, type is (redirect || reverse-proxy)
		return url, forwardType
	}))

	handlers.MaxAge(3600)
	origins := handlers.AllowedOrigins([]string{"*"})

	//headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"})
	//methods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	credentials := handlers.AllowCredentials()

	hdl := handlers.CORS(origins /* headers, methods, */, credentials)

	// E.G: 4300
	var port = DEFAULT_PORT
	if config.Port != "" {
		port = config.Port
	}

	log.Printf("Server started. port:%v", port)

	/* client := &http.Client{
	        //Timeout: 20 * time.Second,
	        Transport: &http.Transport{
	                TLSHandshakeTimeout: 700 * time.Millisecond,
	        },
	} */

	err := http.ListenAndServe(":"+port, hdl(mux))
	//err := http.ListenAndServeTLS(":"+port, "server.crt", "server.key", hdl(mux))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
