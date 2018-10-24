package mst

import (
	"context"
	"errors"
	"github.com/nicksnyder/go-i18n/i18n"
	lilyHttp "github.com/rendau/lily/http"
	"net/http"
	"time"
)

func RequestRetrieveUsrId(r *http.Request, authUrl string) (error, bool, string) {
	qPars := map[string]string{}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			qPars[k] = v[0]
		}
	}

	var headers []string
	for k, v := range r.Header {
		if len(v) > 0 {
			headers = append(headers, k, v[0])
		}
	}

	var repObj requestRetrieveUsrIdSRepSt
	sCode, _, err := lilyHttp.SendRequestReceiveJSONObj(
		&http.Client{Timeout: 20 * time.Second},
		false,
		"GET",
		authUrl,
		qPars,
		nil,
		&repObj,
		headers...,
	)
	if err != nil {
		return err, false, ""
	}
	if !lilyHttp.StatusCodeIsOk(sCode) {
		if sCode == 401 {
			return nil, false, ""
		}
		return errors.New("fail to auth request"), false, ""
	}
	if repObj.UsrId == "" {
		return nil, true, ""
	}

	return nil, true, repObj.UsrId
}

func MwUserCtx(hf http.HandlerFunc, authUrl string, strict bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		ctx := &RequestUserCTXSt{}

		ctx.Language = r.Header.Get("ch-language")
		if len(ctx.Language) > 10 {
			ctx.Language = ""
		}
		ctx.T, err = i18n.Tfunc(ctx.Language, "en")
		//lily.ErrPanicSilent(err)
		if err != nil {
			ctx.T = i18n.IdentityTfunc()
		}

		var ok bool
		err, ok, ctx.ID = RequestRetrieveUsrId(r, authUrl)
		if err != nil {
			if strict {
				RespondServiceNA(w, ctx)
				return
			}
		} else if !ok {
			lilyHttp.Respond401(w, ctx.T("Not authorized"))
			return
		}

		if ctx.ID == "" && strict {
			lilyHttp.Respond401(w, ctx.T("Not authorized"))
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), interface{}("ctx"), ctx))
		hf(w, r)
	}
}

func RespondServiceNA(w http.ResponseWriter, ctx *RequestUserCTXSt) {
	if ctx == nil {
		lilyHttp.Respond400(w, "service_na", "Sorry, service is temporarily unavailable")
	} else {
		lilyHttp.Respond400(w, "service_na", ctx.T("Sorry, service is temporarily unavailable"))
	}
}
