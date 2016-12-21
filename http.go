package lily

import (
	"fmt"
	"net/http"
)

func HTTPSetContentTypeJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func HTTPRespondStr(w http.ResponseWriter, code int, body string) {
	if len(body) == 0 {
		panic("body must be not empty")
	}
	w.WriteHeader(code)
	fmt.Fprint(w, body)
}

func HTTPRespond400(w http.ResponseWriter, err string, desc string) {
	HTTPSetContentTypeJSON(w)
	HTTPRespondStr(w, 400, fmt.Sprintf(`{"error":"%s","detail":"%s"}`, err, desc))
}
