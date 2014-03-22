package util

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var MonthInts = map[string]int{
    "January":   int(time.January),
    "February":  int(time.February),
    "March":     int(time.March),
    "April":     int(time.April),
    "May":       int(time.May),
    "June":      int(time.June),
    "July":      int(time.July),
    "August":    int(time.August),
    "September": int(time.September),
    "October":   int(time.October),
    "November":  int(time.November),
    "December":  int(time.December),
}

func Confirm(q string) (ans bool, err error) {
	fmt.Printf("%s", q)
	var a string
	if _, err = fmt.Scanln(&a); err == nil {
		a = strings.ToLower(a)
		ans = (a == "y" || a == "yes")
	}
	return
}

func DefaultString(s string, d string) string {
	if s == "" {
		return d
	} else {
		return s
	}
}

func PrintUnderlined(s string) {
	fmt.Println(s)
	for i := 0; i < len(s); i++ {
		fmt.Print("-")
	}
}

func FileExists(path string) bool {
	_, e := os.Stat(path)
	return e == nil
}

func Unzip(zipsrc, dest string) error {
	ziprc, err := zip.OpenReader(zipsrc)
	if err != nil {
		return err
	}
	defer ziprc.Close()

	for _, f := range ziprc.File {
		if err := extractFile(f, dest); err != nil {
			return err
		}
		if err := os.Chtimes(filepath.Join(dest, f.Name), f.ModTime(), f.ModTime()); err != nil {
			return err
		}
	}

	return nil
}

func extractFile(f *zip.File, dest string) error {
	frc, err := f.Open()
	if err != nil {
		return err
	}
	defer frc.Close()

	path := filepath.Join(dest, f.Name)
	if f.FileInfo().IsDir() {
		os.MkdirAll(path, f.Mode())
	} else {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, frc)
		if err != nil {
			return err
		}
	}
	return nil
}

func ZipFolder(srcFolderPath string) (err error) {
    destZipPath := srcFolderPath + ".zip"
    dest, err := os.Create(destZipPath)
    if err != nil {
        return err
    }
    defer dest.Close()

    destZip := zip.NewWriter(dest)
    defer func() {
        zipCloseErr := destZip.Close()
        if err == nil {
            err = zipCloseErr
        }
    }()

    srcFolder, err := os.Open(srcFolderPath)
    if err != nil {
        return err
    }
    defer srcFolder.Close()

    files, err := srcFolder.Readdir(0)
    if err != nil {
        return err
    }

    for _, fi := range files {
        err := addFileToZip(destZip, fi, filepath.Join(srcFolderPath, fi.Name()))
        if err != nil {
            return err
        }
    }

    return nil
}

func addFileToZip(zipWriter *zip.Writer, fi os.FileInfo, srcPath string) error {
	fh, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
    fh.Method = zip.Deflate
    destFile, err := zipWriter.CreateHeader(fh)
    if err != nil {
		return err
	}

    srcFile, err := os.Open(srcPath)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    ba, err := ioutil.ReadAll(srcFile)
    if err != nil {
        return err
    }

    //_, err = io.Copy(destFile, srcFile)
    _, err = destFile.Write(ba)
    return err
}

func Verbosef(template string, args... interface{}) {
    //TODO: if verbose
    fmt.Printf(template, args...)
}
