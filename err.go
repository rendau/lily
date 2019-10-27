package lily

import (
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
)

func ErrPanic(err error, msg ...string) {
	if err != nil {
		if len(msg) > 0 {
			log.Panicln(err, msg[0])
		} else {
			log.Panicln(err)
		}
	}
}

func ErrPanicSilent(err error) {
	if err != nil {
		panic(err)
	}
}

func ErrPanicWS(err error, msg ...string) {
	if err != nil {
		if len(msg) > 0 {
			log.Panicln(err, msg[0], "\r\n", string(debug.Stack()))
		} else {
			log.Panicln(err, "\r\n", string(debug.Stack()))
		}
	}
}

func ErrFatal(err error, msg ...string) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Fatalf("fatal (%s:%d):\n\t%s\n\t%s\n", file, line, msg, err.Error())
	}
}

type VErr struct {
	Err        error
	Code       string
	Detail     string
	DetailPars map[string]string
	Extras     []string // for http-responding
}

func (e *VErr) Error() string {
	return fmt.Sprintf("%s, %s", e.Code, e.Detail)
}

func NewVErr(err error, code, detail string) *VErr {
	return &VErr{
		Err:    err,
		Code:   code,
		Detail: detail,
	}
}
