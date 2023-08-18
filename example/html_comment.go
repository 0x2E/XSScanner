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
            <!-- %s -->
        </body>
    </html>
	`
	path := "/html-comment"
	param := []string{"id"}
	logVuln(path, param)
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		logParam(r)
		q := r.URL.Query().Get(param[0])
		fmt.Fprintf(w, respContent, q, q)
	})
}
