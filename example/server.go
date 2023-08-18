package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	addr := "localhost:9090"
	log.Println("listening on " + addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func logParam(r *http.Request) {
	log.Println(r.URL.RawQuery)
}

func logVuln(path string, param []string) {
	fmt.Printf("vuln: path=%s, query_param=%v\n", path, param)
}
