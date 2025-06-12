package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnzipSupportBundle is a helper method to make it easier to unzip the contents of the sample
// support bundle for tests
func UnzipSupportBundle(bundleZipFile, destination string) (err error) {
	r, err := zip.OpenReader(bundleZipFile)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		destPath := filepath.Join(destination, f.Name)
		if !strings.HasPrefix(destPath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid dest path %s", destPath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				return err
			}

			destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_CREATE, f.Mode())
			if err != nil {
				return err
			}

			zFile, err := f.Open()
			if err != nil {
				return err
			}

			if _, err = io.Copy(destFile, zFile); err != nil {
				return err
			}
			_ = zFile.Close()
			_ = destFile.Close()
		}

	}
	return nil
}
