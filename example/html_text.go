package main

import (
	"fmt"
	"net/http"
)

func init() {
	respContent := `
    <html>
        <body>
            <p>id: %s</p>
        </body>
    </html>
	`
	path := "/html-text"
	param := []string{"id"}
	logVuln(path, param)
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		logParam(r)
		fmt.Fprintf(w, respContent, r.URL.Query().Get(param[0]))
	})
}
