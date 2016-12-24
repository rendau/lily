package lily

import (
	"bytes"
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

func HTTPRespondJSONObj(w http.ResponseWriter, code int, obj interface{}) {
	HTTPSetContentTypeJSON(w)
	w.WriteHeader(code)
	ErrPanic(json.NewEncoder(w).Encode(obj))
}

func HTTPRespondError(w http.ResponseWriter, code int, err string, detail string, extras ...interface{}) {
	obj := map[string]interface{}{}
	obj["error"] = err
	obj["detail"] = detail
	for i := 0; i < len(extras); i += 2 {
		obj[extras[i].(string)] = extras[i+1]
	}
	HTTPRespondJSONObj(w, code, obj)
}

func HTTPRespondJSONParseError(w http.ResponseWriter) {
	HTTPRespondError(w, 400, "bad_json", "Fail to parse JSON")
}

func HTTPSendRequestJSON(method, url string, obj interface{}) (*http.Response, error) {
	reqData, err := json.Marshal(obj)
	ErrPanic(err)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqData))
	ErrPanic(err)
	client := &http.Client{}
	return client.Do(req)
}
