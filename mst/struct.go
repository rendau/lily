package mst

import "github.com/nicksnyder/go-i18n/i18n"

type RequestUserCTXSt struct {
	Language string
	T        i18n.TranslateFunc
	ID       string
}

type requestRetrieveUsrIdSRepSt struct {
	UsrId string `json:"usr_id"`
}
