package lily

import (
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
