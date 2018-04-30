package lily

import (
	"net/http"

	"time"
	"log"
	"os"
	"path/filepath"
	"net/url"
	"strings"
	"errors"
	"io"
)

var (
	dirPath         string
	dirName         string
	dirFullPath     string
	timeLimit       time.Duration
	cleanupInterval time.Duration
)

func TmpInit(pDirPath, pDirName string, pTimeLimit time.Duration, pCleanupInterval time.Duration) {
	if pDirPath == "" || pDirName == "" || pTimeLimit == 0 || pCleanupInterval == 0 {
		log.Panicln("Bad initial params")
	}
	dirPath = pDirPath
	dirName = pDirName
	dirFullPath = filepath.Join(dirPath, dirName)
	timeLimit = pTimeLimit
	cleanupInterval = pCleanupInterval

	err := os.MkdirAll(dirFullPath, os.ModePerm)
	ErrPanic(err)

	go tmpCleaner()
}

func TmpSaveFileFromRequestForm(r *http.Request, key, fnSuffix string) (string, error) {
	if dirPath == "" || dirName == "" {
		log.Panicln("Tmp module used befor inited")
	}
	return HTTPUploadFileFromRequestForm(r, key, dirPath, dirName, tmpGenerateFilename(fnSuffix))
}

func TmpCopyTempTarget(urlStr string, dirPath, dir string, filename string) (string, error) {
	notFoundError := errors.New("bad_url")

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	urlPathSlice := strings.SplitN(u.Path, dirName+"/", 2)
	if len(urlPathSlice) != 2 {
		return "", notFoundError
	}

	filePath := filepath.Join(append([]string{dirFullPath}, strings.Split(urlPathSlice[1], "/")...)...)

	fileExt := filepath.Ext(filePath)
	if fileExt == "" {
		return "", errors.New("bad_extension")
	}

	srcFile, err := os.Open(filePath)
	if err != nil {
		return "", notFoundError
	}
	defer srcFile.Close()

	finalDstDirPath := filepath.Join(dirPath, dir)

	err = os.MkdirAll(finalDstDirPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	dstFile, err := TempFile(finalDstDirPath, filename+"_*"+fileExt)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return "", err
	}

	newName, err := filepath.Rel(dirPath, dstFile.Name())
	if err != nil {
		return "", err
	}

	return newName, nil
}

func tmpGenerateFilename(suffix string) string {
	res := time.Now().UTC().Format("2006_01_02_15_04_05")
	if suffix != "" {
		res += "_" + suffix
	}
	return res
}

func tmpParseFilename(src string) *time.Time {
	if len(src) > 19 {
		t, err := time.Parse("2006_01_02_15_04_05", src[:19])
		if err != nil {
			return nil
		}
		return &t
	}
	return nil
}

func tmpCleaner() {
	var err error
	var rpath string
	var ftime *time.Time
	var now time.Time
	var deletePaths []string

	for {
		//fmt.Println("start cleaning temp files...")

		now = time.Now()

		deletePaths = nil

		err = filepath.Walk(
			dirFullPath,
			func(path string, f os.FileInfo, err error) error {
				if f == nil {
					return nil
				}
				//fmt.Println(path, rpath, f.Name())
				if f.IsDir() {
					rpath, err = filepath.Rel(dirPath, path)
					if err != nil {
						return nil
					}
					if rpath == dirName {
						return nil
					}
				}
				ftime = tmpParseFilename(f.Name())
				if ftime == nil || ftime.Add(timeLimit).Before(now) {
					deletePaths = append(deletePaths, path)
				}
				if f.IsDir() {
					return filepath.SkipDir
				}
				return nil
			},
		)
		ErrPanic(err)

		// delete old files
		for _, x := range deletePaths {
			os.RemoveAll(x)
		}

		//fmt.Printf("  deleted %d paths\n", len(deletePaths))

		time.Sleep(cleanupInterval)
	}
}
