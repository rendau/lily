package lily

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"strconv"
)

type sendReplySt struct {
	ID        uint64 `json:"id"`
	CNT       int    `json:"cnt"`
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

type getBalanceReplySt struct {
	Balance   string `json:"balance"`
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

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
	req, err := http.NewRequest("GET", "https://smsc.kz/sys/send.php", nil)
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
	reply := sendReplySt{}
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
	req, err := http.NewRequest("GET", "https://smsc.kz/sys/jobs.php", nil)
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
	reply := sendReplySt{}
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

func SmscGetBalance(username, password string) (result float64) {
	result = 0
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest("GET", "https://smsc.kz/sys/balance.php", nil)
	ErrPanic(err)
	params := req.URL.Query()
	params.Add("login", username)
	params.Add("psw", password)
	params.Add("fmt", "3")
	req.URL.RawQuery = params.Encode()
	resp, err := client.Do(req)
	if err != nil {
		log.Println("fail to get sms-balance:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("bad status code %d in sms-balance reply\n", resp.StatusCode)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("fail to read sms-balance reply body:", err)
		return
	}
	reply := getBalanceReplySt{}
	err = json.Unmarshal(data, &reply)
	if err != nil {
		log.Println("fail to parse sms-balance reply:", err)
		return
	}
	if (reply.ErrorCode != 0) || (reply.Error != "") {
		log.Printf("sms provider error for getting balance:\n%s\n", string(data))
		return
	}
	result, _ = strconv.ParseFloat(reply.Balance, 64)
	return
}
