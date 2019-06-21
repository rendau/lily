package tmp

import (
	"errors"
	"github.com/rendau/lily"
	lilyHttp "github.com/rendau/lily/http"
	"github.com/rendau/lily/zip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	_dirPath         string
	_dirName         string
	_dirFullPath     string
	_timeLimit       time.Duration
	_cleanupInterval time.Duration
)

func Init(dirPath, dirName string, timeLimit time.Duration, cleanupInterval time.Duration) {
	if dirPath == "" || dirName == "" || timeLimit == 0 || cleanupInterval == 0 {
		log.Panicln("Bad initial params")
	}
	_dirPath = dirPath
	_dirName = dirName
	_dirFullPath = filepath.Join(_dirPath, _dirName)
	_timeLimit = timeLimit
	_cleanupInterval = cleanupInterval

	err := os.MkdirAll(_dirFullPath, os.ModePerm)
	lily.ErrPanic(err)

	go cleaner()
}

func UploadFileFromHttpRequestForm(r *http.Request, key, fnSuffix string,
	requireExt, extractZip bool) (error, string, string, string, string) {
	if _dirPath == "" || _dirName == "" {
		log.Panicln("Tmp module used befor inited")
	}

	fn := generateFilename(fnSuffix)

	err, newFileName := lilyHttp.UploadFileFromRequestForm(
		r,
		key,
		_dirFullPath,
		fn+"_*",
		requireExt,
	)
	if err != nil {
		return err, "", "", "", ""
	}

	fPath := filepath.Join(_dirFullPath, newFileName)
	rPath := filepath.Join(_dirName, newFileName)
	eFPath := ""
	eRPath := ""

	if extractZip && strings.ToLower(filepath.Ext(fPath)) == ".zip" {
		eFPath, err = ioutil.TempDir(_dirFullPath, fn+"_")
		if err == nil {
			err = zip.ExtractFromFile(fPath, eFPath)
			if err == nil {
				eRPath, _ = filepath.Rel(_dirPath, eFPath)
			}
		}
	}

	return err, fPath, rPath, eFPath, eRPath
}

func Copy(urlStr string, dirPath, dir string, filename string, requireExt bool) (string, error) {
	notFoundError := errors.New("bad_url")

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	urlPathSlice := strings.SplitN(u.Path, _dirName+"/", 2)
	if len(urlPathSlice) != 2 {
		return "", notFoundError
	}

	filePath := filepath.Join(append([]string{_dirFullPath}, strings.Split(urlPathSlice[1], "/")...)...)

	fileExt := filepath.Ext(filePath)
	if requireExt && fileExt == "" {
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

	dstFile, err := ioutil.TempFile(finalDstDirPath, filename+"_*"+fileExt)
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

func generateFilename(suffix string) string {
	res := time.Now().UTC().Format("2006_01_02_15_04_05")
	if suffix != "" {
		res += "_" + suffix
	}
	return res
}

func parseFilename(src string) *time.Time {
	if len(src) > 19 {
		t, err := time.Parse("2006_01_02_15_04_05", src[:19])
		if err != nil {
			return nil
		}
		return &t
	}
	return nil
}

func cleaner() {
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
			_dirFullPath,
			func(path string, f os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if f == nil {
					return nil
				}
				//fmt.Println(path, rpath, f.Name())
				if f.IsDir() {
					rpath, err = filepath.Rel(_dirPath, path)
					if err != nil {
						return nil
					}
					if rpath == _dirName {
						return nil
					}
				}
				ftime = parseFilename(f.Name())
				if ftime == nil || ftime.Add(_timeLimit).Before(now) {
					deletePaths = append(deletePaths, path)
				}
				if f.IsDir() {
					return filepath.SkipDir
				}
				return nil
			},
		)
		lily.ErrPanic(err)

		// delete old files
		for _, x := range deletePaths {
			_ = os.RemoveAll(x)
		}

		//fmt.Printf("  deleted %d paths\n", len(deletePaths))

		time.Sleep(_cleanupInterval)
	}
}
