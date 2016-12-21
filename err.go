package lily

import "log"

func ErrPanic(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func ErrFatal(err error, msg string) {
	if err != nil {
		log.Fatalln(err)
	}
}
