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

func SmscSend(username, password string, phones string, msg string) bool {
	if SmscDebug {
		log.Printf("Sent sms: %s - %q\n", phones, msg)
		return true
	}
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest("GET", urlPrefix+"send.php", nil)
	ErrPanic(err)
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
		log.Println("(sms-send) fail to send sms:", err)
		return false
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("(sms-send) fail to read sms-send reply body:", err)
		return false
	}
	if resp.StatusCode != 200 {
		log.Printf("(sms-send) bad status code %d in sms-send reply, data: %s\n", resp.StatusCode, string(data))
		return false
	}
	reply := smscSendReplySt{}
	err = json.Unmarshal(data, &reply)
	if err != nil {
		log.Printf("(sms-send) fail to parse sms-send reply: %s, %s\n", err.Error(), string(data))
		return false
	}
	if (reply.ErrorCode != 0) || (reply.Error != "") {
		if reply.ErrorCode != 7 && reply.ErrorCode != 8 { // 7 - invalid number, 8 - can't to deliver
			log.Printf("(sms-send) sms provider error for (%s, %q):\n%s\n", phones, msg, string(data))
		}
		return false
	}
	return true
}

func SmscSendBcast(username, password string, phones string, msg string) (bool, uint64) {
	if SmscDebug {
		log.Printf("Sent sms-bcast: %s - %q\n", phones, msg)
		return true, 777
	}
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest("GET", urlPrefix+"jobs.php", nil)
	ErrPanic(err)
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
		log.Println("(sms-bcast) fail to send sms:", err)
		return false, 0
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("(sms-bcast) fail to read sms-send reply body:", err)
		return false, 0
	}
	if resp.StatusCode != 200 {
		log.Printf("(sms-bcast) bad status code %d in sms-send reply, data: %s\n", resp.StatusCode, string(data))
		return false, 0
	}
	reply := smscSendReplySt{}
	err = json.Unmarshal(data, &reply)
	if err != nil {
		log.Printf("(sms-bcast) fail to parse sms-send reply: %s, %s\n", err.Error(), string(data))
		return false, 0
	}
	if (reply.ErrorCode != 0) || (reply.Error != "") {
		log.Printf("(sms-bcast) sms provider error for (%s, %q):\n%s\n", phones, msg, string(data))
		return false, 0
	}
	return true, reply.ID
}

func SmscGetBalance(username, password string) (*SmscErrorSt, float64) {
	var result float64

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", urlPrefix+"balance.php", nil)
	if err != nil {
		return &SmscErrorSt{Code: "request_fail", Desc: "Fail to create new request - " + err.Error()}, 0
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

	if resp.StatusCode != 200 {
		return &SmscErrorSt{Code: "bad_status_code", Desc: "Bad status code - " + strconv.Itoa(resp.StatusCode)}, 0
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &SmscErrorSt{Code: "fail_to_read_body", Desc: "Fail to read body - " + err.Error()}, 0
	}

	reply := smscGetBalanceReplySt{}
	err = json.Unmarshal(data, &reply)
	if err != nil {
		return &SmscErrorSt{Code: "fail_to_parse_body", Desc: "Fail to parse body - " + err.Error()}, 0
	}

	if (reply.ErrorCode != 0) || (reply.Error != "") {
		return &SmscErrorSt{Code: "provider_error", Desc: "Provider error - " + string(data)}, 0
	}

	result, _ = strconv.ParseFloat(reply.Balance, 64)

	return nil, result
}
