package main

import (
	"fmt"
	"net/http"
)

func init() {
	respContent := `
    <html>
        <body>
            <div data-%s="xxxxx">
                <p>xxx</p>
            </div>
            <img src="%s">
        </body>
    </html>
	`
	path := "/html-attr"
	param := []string{"id", "file"}
	logVuln(path, param)
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		logParam(r)
		fmt.Fprintf(w, respContent, r.URL.Query().Get(param[0]), r.URL.Query().Get(param[1]))
	})
}
