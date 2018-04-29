package lily

import (
	"net/http"

	"time"
	"log"
	"os"
	"path/filepath"
)

var (
	dirPath         string
	dirName         string
	timeLimit       time.Duration
	cleanupInterval time.Duration
)

func TmpInit(pDirPath, pDirName string, pTimeLimit time.Duration, pCleanupInterval time.Duration) {
	if pDirPath == "" || pDirName == "" || pTimeLimit == 0 || pCleanupInterval == 0 {
		log.Panicln("Bad initial params")
	}
	dirPath = pDirPath
	dirName = pDirName
	timeLimit = pTimeLimit
	cleanupInterval = pCleanupInterval

	err := os.MkdirAll(filepath.Join(dirPath, dirName), os.ModePerm)
	ErrPanic(err)

	go tmpCleaner()
}

func TmpSaveFileFromRequestForm(r *http.Request, key, fnSuffix string) (string, error) {
	if dirPath == "" || dirName == "" {
		log.Panicln("Tmp module used befor inited")
	}
	return HTTPUploadFileFromRequestForm(r, key, dirPath, dirName, tmpGenerateFilename(fnSuffix))
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
			filepath.Join(dirPath, dirName),
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
