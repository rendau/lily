package zip

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func ExtractFromRequestForm(req *http.Request, field, dest string) error {
	f, _, err := req.FormFile(field)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	fs, err := buf.ReadFrom(f)
	if err != nil {
		return err
	}
	return Extract(buf, fs, dest)
}

func ExtractFromFile(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	fs, err := buf.ReadFrom(f)
	if err != nil {
		return err
	}
	return Extract(buf, fs, dest)
}

func Extract(buffer *bytes.Buffer, fileSize int64, dest string) error {
	err := os.RemoveAll(dest)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	r, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), fileSize)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		err = extractFile(f, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(dest, f.Name)

	if f.FileInfo().IsDir() {
		os.MkdirAll(path, f.Mode())
	} else {
		os.MkdirAll(filepath.Dir(path), f.Mode())
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
	}
	return nil
}
