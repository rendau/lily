package lily

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	netUrl "net/url"
	"os"
	"path/filepath"
	"strings"
)

func HTTPMwCORSAllowAll(h http.Handler, maxAge string) http.Handler {
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

func HTTPStatusCodeIsOk(statusCode int) bool {
	return statusCode > 199 && statusCode < 300
}

func HTTPSetContentTypeJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func HTTPSetContentTypeHTML(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
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
	HTTPRespond400(w, "bad_json", "Fail to parse JSON")
}

func HTTPSendRequest(client *http.Client, method, url string, urlParams map[string]string,
	data []byte, headers ...string) (*http.Response, error) {
	var err error
	var req *http.Request

	if client == nil {
		ErrPanic(errors.New("client is nil"))
	}

	if data != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
		ErrPanic(err)
	} else {
		req, err = http.NewRequest(method, url, nil)
		ErrPanic(err)
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

func HTTPSendRequestReceiveBytes(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	data []byte, headers ...string) (int, []byte, error) {
	var res []byte

	resp, err := HTTPSendRequest(client, method, url, urlParams, data, headers...)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	res, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	if !HTTPStatusCodeIsOk(resp.StatusCode) {
		if errSCode {
			return resp.StatusCode, res, errors.New(fmt.Sprintf("bad_http_status_code - %d\nbody: %s", resp.StatusCode, string(res)))
		}
		return resp.StatusCode, res, nil
	}

	return resp.StatusCode, res, nil
}

func HTTPSendRequestReceiveString(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	data []byte, headers ...string) (int, string, error) {
	sCode, resBytes, err := HTTPSendRequestReceiveBytes(client, errSCode, method, url, urlParams, data, headers...)

	return sCode, string(resBytes), err
}

func HTTPSendRequestReceiveJSONObj(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	data []byte, rObj interface{}, headers ...string) (int, []byte, error) {
	sCode, rBytes, err := HTTPSendRequestReceiveBytes(
		client, errSCode, method, url, urlParams, data, headers...)
	if err != nil || !HTTPStatusCodeIsOk(sCode) {
		return sCode, rBytes, err
	}

	err = json.Unmarshal(rBytes, rObj)
	if err != nil {
		return sCode, rBytes, errors.New(fmt.Sprintf("fail_to_parse_json - %s\nbody: %s", err.Error(), string(rBytes)))
	}

	return sCode, rBytes, nil
}

func HTTPSendJSONObjRequest(client *http.Client, method, url string, urlParams map[string]string,
	sObj interface{}, headers ...string) (*http.Response, error) {
	sBytes, err := json.Marshal(sObj)
	if err != nil {
		return nil, err
	}

	return HTTPSendRequest(client, method, url, urlParams, sBytes, headers...)
}

func HTTPSendJSONObjRequestReceiveBytes(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	sObj interface{}, headers ...string) (int, []byte, error) {
	sBytes, err := json.Marshal(sObj)
	if err != nil {
		return 0, nil, err
	}

	return HTTPSendRequestReceiveBytes(client, errSCode, method, url, urlParams, sBytes, headers...)
}

func HTTPSendJSONObjRequestReceiveString(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	sObj interface{}, headers ...string) (int, string, error) {
	sBytes, err := json.Marshal(sObj)
	if err != nil {
		return 0, "", err
	}

	return HTTPSendRequestReceiveString(client, errSCode, method, url, urlParams, sBytes, headers...)
}

func HTTPSendJSONObjRequestReceiveJSONObj(client *http.Client, errSCode bool, method, url string, urlParams map[string]string,
	sObj interface{}, rObj interface{}, headers ...string) (int, []byte, error) {
	sBytes, err := json.Marshal(sObj)
	if err != nil {
		return 0, nil, err
	}

	return HTTPSendRequestReceiveJSONObj(client, errSCode, method, url, urlParams, sBytes, rObj, headers...)
}

func HTTPRetrieveRequestHostURL(r *http.Request) string {
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

func HTTPRetrieveRemoteIP(r *http.Request) (result string) {
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

func HTTPUploadFileFromRequestForm(r *http.Request, key, dirPath, dir string, filename string) (string, error) {
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

	dstFile, err := TempFile(finalDirPath, filename+"_*"+fileExt)
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

func HTTPRespondError(w http.ResponseWriter, code int, err string, detail string, extras ...interface{}) {
	obj := map[string]interface{}{}
	obj["error"] = err
	obj["error_dsc"] = detail
	for i := 0; (i + 1) < len(extras); i += 2 {
		obj[extras[i].(string)] = extras[i+1]
	}
	HTTPRespondJSONObj(w, code, obj)
}

func HTTPRespond400(w http.ResponseWriter, err, detail string, extras ...interface{}) {
	HTTPRespondError(w, 400, err, detail, extras...)
}

func HTTPRespond401(w http.ResponseWriter, detail string) {
	HTTPRespondError(w, 401, "unauthorized", detail)
}

func HTTPRespond403(w http.ResponseWriter, detail string) {
	HTTPRespondError(w, 403, "permission_denied", detail)
}

func HTTPRespond404(w http.ResponseWriter, detail string) {
	HTTPRespondError(w, 404, "not_found", detail)
}
