package lily

import (
	"net/http"

	"time"
	"log"
	"os"
	"path/filepath"
)

var (
	dirPath string
	dirName string
)

func TmpInit(pDirPath, pDirName string) {
	if pDirPath == "" || pDirName == "" {
		log.Panicln("Bad initial params")
	}
	dirPath = pDirPath
	dirName = pDirName

	err := os.MkdirAll(filepath.Join(dirPath, dirName), os.ModePerm)
	ErrPanic(err)
}

func TmpSaveFileFromRequestForm(r *http.Request, key, fnSuffix string) (string, error) {
	if dirPath == "" || dirName == "" {
		log.Panicln("Tmp module used befor inited")
	}
	return HTTPUploadFileFromRequestForm(r, key, dirPath, dirName, getFilename(fnSuffix))
}

func getFilename(suffix string) string {
	res := time.Now().UTC().Format("2006_01_02_15_04_05")
	if suffix != "" {
		res += "_" + suffix
	}
	return res
}