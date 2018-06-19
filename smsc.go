package lily

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"strconv"
)

type SmscErrorSt struct {
	Code string
	Desc string
}

func (o *SmscErrorSt) Error() string {
	return o.Code + ", " + o.Desc
}

type smscSendReplySt struct {
	ID        uint64 `json:"id"`
	CNT       int    `json:"cnt"`
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

type smscGetBalanceReplySt struct {
	Balance   string `json:"balance"`
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

const (
	urlPrefix = `https://smsc.kz/sys/`
)

var (
	SmscDebug = false
)

func SmscSend(username, password string, phones string, msg string) *SmscErrorSt {
	if SmscDebug {
		log.Printf("Sent sms: %s - %q\n", phones, msg)
		return nil
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", urlPrefix+"send.php", nil)
	if err != nil {
		return &SmscErrorSt{Code: "request_create_error", Desc: "Fail to create new request - " + err.Error()}
	}

	params := req.URL.Query()
	params.Add("login", username)
	params.Add("psw", password)
	params.Add("phones", phones)
	params.Add("mes", msg)
	params.Add("charset", "utf-8")
	params.Add("fmt", "3")

	req.URL.RawQuery = params.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return &SmscErrorSt{Code: "request_fail", Desc: "Fail to request smsc - " + err.Error()}
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &SmscErrorSt{Code: "fail_to_read_body", Desc: "Fail to read body - " + err.Error()}
	}

	if resp.StatusCode != 200 {
		return &SmscErrorSt{Code: "bad_status_code", Desc: "Bad status code - " + strconv.Itoa(resp.StatusCode) + ", body - " + string(data)}
	}

	reply := smscSendReplySt{}
	err = json.Unmarshal(data, &reply)
	if err != nil {
		return &SmscErrorSt{Code: "fail_to_parse_body", Desc: "Fail to parse body - " + err.Error()}
	}

	if (reply.ErrorCode != 0) || (reply.Error != "") {
		return &SmscErrorSt{Code: "provider_error_" + strconv.Itoa(reply.ErrorCode), Desc: "Provider error for message '" + msg + "' - " + string(data)}
	}

	return nil
}

func SmscSendBcast(username, password string, phones string, msg string) (*SmscErrorSt, uint64) {
	if SmscDebug {
		log.Printf("Sent sms-bcast: %s - %q\n", phones, msg)
		return nil, 777
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", urlPrefix+"jobs.php", nil)
	if err != nil {
		return &SmscErrorSt{Code: "request_create_error", Desc: "Fail to create new request - " + err.Error()}, 0
	}

	params := req.URL.Query()
	params.Add("add", "1")
	params.Add("login", username)
	params.Add("psw", password)
	params.Add("name", "bcast")
	params.Add("phones", phones)
	params.Add("mes", msg)
	params.Add("charset", "utf-8")
	params.Add("fmt", "3")

	req.URL.RawQuery = params.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return &SmscErrorSt{Code: "request_fail", Desc: "Fail to request smsc - " + err.Error()}, 0
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &SmscErrorSt{Code: "fail_to_read_body", Desc: "Fail to read body - " + err.Error()}, 0
	}

	if resp.StatusCode != 200 {
		return &SmscErrorSt{Code: "bad_status_code", Desc: "Bad status code - " + strconv.Itoa(resp.StatusCode) + ", body - " + string(data)}, 0
	}

	reply := smscSendReplySt{}
	err = json.Unmarshal(data, &reply)
	if err != nil {
		return &SmscErrorSt{Code: "fail_to_parse_body", Desc: "Fail to parse body - " + err.Error()}, 0
	}

	if (reply.ErrorCode != 0) || (reply.Error != "") {
		return &SmscErrorSt{Code: "provider_error_" + strconv.Itoa(reply.ErrorCode), Desc: "Provider error for message '" + msg + "' - " + string(data)}, 0
	}

	return nil, reply.ID
}

func SmscGetBalance(username, password string) (*SmscErrorSt, float64) {
	var result float64

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", urlPrefix+"balance.php", nil)
	if err != nil {
		return &SmscErrorSt{Code: "request_create_error", Desc: "Fail to create new request - " + err.Error()}, 0
	}

	params := req.URL.Query()
	params.Add("login", username)
	params.Add("psw", password)
	params.Add("fmt", "3")

	req.URL.RawQuery = params.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return &SmscErrorSt{Code: "request_fail", Desc: "Fail to request smsc - " + err.Error()}, 0
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &SmscErrorSt{Code: "fail_to_read_body", Desc: "Fail to read body - " + err.Error()}, 0
	}

	if resp.StatusCode != 200 {
		return &SmscErrorSt{Code: "bad_status_code", Desc: "Bad status code - " + strconv.Itoa(resp.StatusCode) + ", body - " + string(data)}, 0
	}

	reply := smscGetBalanceReplySt{}
	err = json.Unmarshal(data, &reply)
	if err != nil {
		return &SmscErrorSt{Code: "fail_to_parse_body", Desc: "Fail to parse body - " + err.Error()}, 0
	}

	if (reply.ErrorCode != 0) || (reply.Error != "") {
		return &SmscErrorSt{Code: "provider_error_" + strconv.Itoa(reply.ErrorCode), Desc: "Provider error - " + string(data)}, 0
	}

	result, _ = strconv.ParseFloat(reply.Balance, 64)

	return nil, result
}
