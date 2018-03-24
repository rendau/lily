package lily

import (
	"bytes"
	"encoding/json"
	"fmt"
	"errors"
	"net/http"
	"time"
	"net/http/cookiejar"
	"strings"
	"net"
	"path/filepath"
	"os"
	"io"
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
	HTTPRespond400(w, "bad_json", "Fail to parse JSON")
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
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: timeout,
		Jar:     jar,
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
	headers = append(headers, "Content-Type", "application/json")
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

func HTTPUploadFormFile(r *http.Request, key, dirPath, dir string, id uint64) (string, error) {
	var err error

	err = os.MkdirAll(filepath.Join(dirPath, dir), os.ModePerm)
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

	newName := fmt.Sprintf("%s%c%d%s", dir, filepath.Separator, id, fileExt)
	suffix := 0
	for {
		_, err = os.Stat(filepath.Join(dirPath, newName))
		if os.IsNotExist(err) {
			break
		} else if err != nil {
			return "", err
		}
		suffix += 1
		newName = fmt.Sprintf("%s%c%d_%d%s", dir, filepath.Separator, id, suffix, fileExt)
	}

	dstFile, err := os.Create(filepath.Join(dirPath, newName))
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
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
