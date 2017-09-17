package lily

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

func HTTPRespondJSONParseError(w http.ResponseWriter) {
	HTTPRespondError(w, 400, "bad_json", "Fail to parse JSON")
}

func HTTPSendRequest(method, url string, data []byte, timeout time.Duration, headers ...string) (*http.Response, error) {
	var err error
	var req *http.Request
	if data != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
		ErrPanic(err)
	} else {
		req, err = http.NewRequest(method, url, nil)
		ErrPanic(err)
	}
	for i := 0; (i + 1) < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	client := &http.Client{
		Timeout: timeout,
	}
	return client.Do(req)
}

func HTTPSendRequestJSON(method, url string, obj interface{}, timeout time.Duration, headers ...string) (*http.Response, error) {
	var err error
	var data []byte
	if obj != nil {
		data, err = json.Marshal(obj)
		ErrPanic(err)
	}
	return HTTPSendRequest(method, url, data, timeout, headers...)
}

func HTTPRetrieveRequestURL(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS == nil {
			scheme = "http"
		} else {
			scheme = "https"
		}
	}
	return scheme + "://" + r.Host
}

func HTTPRespondError(w http.ResponseWriter, code int, err string, detail string, extras ...interface{}) {
	obj := map[string]interface{}{}
	obj["error"] = err
	obj["error_detail"] = detail
	for i := 0; (i + 1) < len(extras); i += 2 {
		obj[extras[i].(string)] = extras[i+1]
	}
	HTTPRespondJSONObj(w, code, obj)
}

func HTTPRespond400(w http.ResponseWriter, err, detail string, extras ...interface{}) {
	HTTPRespondError(w, 400, err, detail, extras...)
}

func HTTPRespond401(w http.ResponseWriter) {
	HTTPRespondError(w, 401, "bad_token", "Bad token")
}

func HTTPRespond403(w http.ResponseWriter) {
	HTTPRespondError(w, 403, "permission_denied", "Permission denied")
}

func HTTPRespond404(w http.ResponseWriter) {
	HTTPRespondError(w, 404, "object_not_found", "Object not found")
}
