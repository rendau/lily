package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rendau/lily"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	netUrl "net/url"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

func MwCORSAllowAll(h http.Handler, maxAge string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE, TRACE, CONNECT, OPTIONS")
			// headers
			requestHeaders := strings.Split(r.Header.Get("Access-Control-Request-Headers"), ",")
			var allowedHeaders []string
			for _, v := range requestHeaders {
				allowedHeaders = append(allowedHeaders, http.CanonicalHeaderKey(strings.TrimSpace(v)))
			}
			if len(allowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ","))
			}
			w.Header().Set("Access-Control-Max-Age", maxAge)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func MwRecovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				headersStr := ""
				for name, headers := range r.Header {
					name = strings.ToLower(name)
					for _, hdr := range headers {
						headersStr += "   " + name + ": " + hdr + "\n"
					}
				}
				log.Printf(
					"\nFail to:\n   %v %v\nError:\n   %v\nHTTP Headers:\n%vStack:\n%v",
					r.Method, r.URL, err, headersStr, string(debug.Stack()),
				)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func StatusCodeIsOk(statusCode int) bool {
	return statusCode > 199 && statusCode < 300
}

func SetContentTypeJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func SetContentTypeHTML(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
}

func RespondStr(w http.ResponseWriter, code int, body string) {
	if len(body) == 0 {
		panic("body must be not empty")
	}
	w.WriteHeader(code)
	fmt.Fprint(w, body)
}

func RespondJSONObj(w http.ResponseWriter, code int, obj interface{}) {
	SetContentTypeJSON(w)
	w.WriteHeader(code)
	lily.ErrPanic(json.NewEncoder(w).Encode(obj))
}

func RespondJSONParseError(w http.ResponseWriter) {
	Respond400(w, "bad_json", "Fail to parse JSON")
}

func SendRequest(client *http.Client, method, url string, urlParams map[string]string,
	data []byte, headers ...string) (*http.Response, error) {
	var err error
	var req *http.Request

	if client == nil {
		lily.ErrPanic(errors.New("client is nil"))
	}

	if data != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
		lily.ErrPanic(err)
	} else {
		req, err = http.NewRequest(method, url, nil)
		lily.ErrPanic(err)
	}

	if urlParams != nil {
		q := netUrl.Values{}
		for k, v := range urlParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	for i := 0; (i + 1) < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}

	return client.Do(req)
}

func SendRequestReceiveBytes(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	data []byte, headers ...string) (int, []byte, error) {
	var res []byte

	resp, err := SendRequest(client, method, url, urlParams, data, headers...)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	res, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	if !StatusCodeIsOk(resp.StatusCode) {
		if errSCode {
			return resp.StatusCode, res, errors.New(fmt.Sprintf("bad_http_status_code - %d\nbody: %s", resp.StatusCode, string(res)))
		}
		return resp.StatusCode, res, nil
	}

	return resp.StatusCode, res, nil
}

func SendRequestReceiveString(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	data []byte, headers ...string) (int, string, error) {
	sCode, resBytes, err := SendRequestReceiveBytes(client, errSCode, method, url, urlParams, data, headers...)

	return sCode, string(resBytes), err
}

func SendRequestReceiveJSONObj(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	data []byte, rObj interface{}, headers ...string) (int, []byte, error) {
	sCode, rBytes, err := SendRequestReceiveBytes(
		client, errSCode, method, url, urlParams, data, headers...)
	if err != nil || !StatusCodeIsOk(sCode) {
		return sCode, rBytes, err
	}

	err = json.Unmarshal(rBytes, rObj)
	if err != nil {
		return sCode, rBytes, errors.New(fmt.Sprintf("fail_to_parse_json - %s\nbody: %s", err.Error(), string(rBytes)))
	}

	return sCode, rBytes, nil
}

func SendJSONObjRequest(client *http.Client, method, url string, urlParams map[string]string,
	sObj interface{}, headers ...string) (*http.Response, error) {
	sBytes, err := json.Marshal(sObj)
	if err != nil {
		return nil, err
	}

	return SendRequest(client, method, url, urlParams, sBytes, headers...)
}

func SendJSONObjRequestReceiveBytes(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	sObj interface{}, headers ...string) (int, []byte, error) {
	sBytes, err := json.Marshal(sObj)
	if err != nil {
		return 0, nil, err
	}

	return SendRequestReceiveBytes(client, errSCode, method, url, urlParams, sBytes, headers...)
}

func SendJSONObjRequestReceiveString(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	sObj interface{}, headers ...string) (int, string, error) {
	sBytes, err := json.Marshal(sObj)
	if err != nil {
		return 0, "", err
	}

	return SendRequestReceiveString(client, errSCode, method, url, urlParams, sBytes, headers...)
}

func SendJSONObjRequestReceiveJSONObj(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	sObj interface{}, rObj interface{}, headers ...string) (int, []byte, error) {
	sBytes, err := json.Marshal(sObj)
	if err != nil {
		return 0, nil, err
	}

	return SendRequestReceiveJSONObj(client, errSCode, method, url, urlParams, sBytes, rObj, headers...)
}

func RetrieveRequestHostURL(r *http.Request) string {
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

func RetrieveRemoteIP(r *http.Request) (result string) {
	result = ""
	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		result = parts[0]
	}
	// If we have a forwarded-for header, take the address from there
	if xff := strings.Trim(r.Header.Get("X-Forwarded-For"), ","); len(xff) > 0 {
		addrs := strings.Split(xff, ",")
		lastFwd := addrs[len(addrs)-1]
		if ip := net.ParseIP(lastFwd); ip != nil {
			result = ip.String()
		}
		// parse X-Real-Ip header
	} else if xri := r.Header.Get("X-Real-Ip"); len(xri) > 0 {
		if ip := net.ParseIP(xri); ip != nil {
			result = ip.String()
		}
	}
	return
}

func UploadFileFromRequestForm(r *http.Request, key, dirPath, dir string, filename string) (string, error) {
	var err error

	finalDirPath := filepath.Join(dirPath, dir)

	err = os.MkdirAll(finalDirPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	srcFile, header, err := r.FormFile(key)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	fileExt := filepath.Ext(header.Filename)
	if fileExt == "" {
		return "", errors.New("bad_extension")
	}

	dstFile, err := ioutil.TempFile(finalDirPath, filename+"_*"+fileExt)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return "", err
	}

	err = os.Chmod(dstFile.Name(), 0644)
	if err != nil {
		return "", err
	}

	newName, err := filepath.Rel(dirPath, dstFile.Name())
	if err != nil {
		return "", err
	}

	return newName, nil
}

func RespondError(w http.ResponseWriter, code int, err string, detail string, extras ...interface{}) {
	obj := map[string]interface{}{}
	obj["error"] = err
	obj["error_dsc"] = detail
	for i := 0; (i + 1) < len(extras); i += 2 {
		obj[extras[i].(string)] = extras[i+1]
	}
	RespondJSONObj(w, code, obj)
}

func Respond400(w http.ResponseWriter, err, detail string, extras ...interface{}) {
	RespondError(w, 400, err, detail, extras...)
}

func Respond401(w http.ResponseWriter, detail string) {
	RespondError(w, 401, "unauthorized", detail)
}

func Respond403(w http.ResponseWriter, detail string) {
	RespondError(w, 403, "permission_denied", detail)
}

func Respond404(w http.ResponseWriter, detail string) {
	RespondError(w, 404, "not_found", detail)
}
