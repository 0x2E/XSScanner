package main

import (
	"fmt"
	"net/http"
)

func init() {
	respContent := `
    <html>
        <body>
            <%s>hello</%s>
            <%s src=1 />
        </body>
    </html>
	`
	path := "/html-tagname"
	param := []string{"id"}
	logVuln(path, param)
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		logParam(r)
		id := r.URL.Query().Get(param[0])
		fmt.Fprintf(w, respContent, id, id, id)
	})
}
