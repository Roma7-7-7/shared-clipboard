package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
)

var port = flag.String("port", "8888", "proxy server port")
var target = flag.String("target", "http://localhost:8080", "target server")

func main() {
	flag.Parse()

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), r.Method, *target+r.URL.Path, r.Body)
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

	fmt.Println("Proxy server is running on port", *port, "and target server is", *target)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		panic(err)
	}
}
