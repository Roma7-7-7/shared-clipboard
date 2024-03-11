package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var port = flag.String("port", "8888", "proxy server port")
var targetWeb = flag.String("target-web", "http://localhost:8080", "target web server")
var apiHost = flag.String("api-host", "api.clipboard-share.home", "api host")
var targetApi = flag.String("target-api", "http://localhost:8080", "target api server")

func main() {
	flag.Parse()

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		url := *targetWeb + r.URL.Path
		if strings.EqualFold(r.Host, *apiHost) {
			url = *targetApi + r.URL.Path
		}

		req, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header = r.Header
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			rw.Header()[k] = v
		}
		all, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(resp.StatusCode)
		_, _ = rw.Write(all)
	})

	fmt.Println("Proxy server is running on port", *port, "target web server", *targetWeb, "target api server", *targetApi)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		panic(err)
	}
}
