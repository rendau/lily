package lily

import (
	"encoding/json"
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

func HTTPRespondJSONObj(w http.ResponseWriter, obj interface{}) {
	HTTPSetContentTypeJSON(w)
	ErrPanic(json.NewEncoder(w).Encode(obj))
}

func HTTPRespondError(w http.ResponseWriter, err string, desc string) {
	HTTPSetContentTypeJSON(w)
	HTTPRespondStr(w, 400, fmt.Sprintf(`{"error":"%s","detail":"%s"}`, err, desc))
}

func HTTPRespondJSONParseError(w http.ResponseWriter) {
	HTTPRespondError(w, "bad_json", "Fail to parse JSON")
}
