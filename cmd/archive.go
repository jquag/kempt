package cmd

import (
	"fmt"
	"github.com/jquag/kempt/conf"
	"github.com/jquag/kempt/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func runArchive(cfg *conf.Config, dirStr string) error {
	rootStat, err := os.Stat(dirStr)
	if err != nil {
		return err
	}

	count, err := archiveFiles(cfg, dirStr, rootStat)
	if err != nil {
		return err
	}

	fmt.Printf("Archived %d file(s)\n", count)

	if cfg.ZipAgeDays > 0 {
		return zipArchives(cfg, dirStr)
	}
	return nil
}

func zipArchives(cfg *conf.Config, dirStr string) error {
	//subtract a month so we don't zip the month represented by current date - ZipAgeDays
	maxTime := cfg.ZipTime().AddDate(0, -1, 0)
	count, err := cfg.EachMonth(dirStr, maxTime, zipArchive)
	if err == nil {
		fmt.Printf("Zipped %d folder(s)\n", count)
	}
	return err
}

func zipArchive(monthPath string, maxDate time.Time) (int, error) {
    if !strings.HasSuffix(monthPath, ".zip") {
        err := util.ZipFolder(monthPath)
        if err != nil {
            return 0, fmt.Errorf("Error! Failed to zip %s - %s\n", monthPath, err)
        }
        err = os.RemoveAll(monthPath)
        if err != nil {
            fmt.Printf("Warning. Failed to remove month (%s) folder after zipping - %s\n", monthPath, err)
        } else {
            util.Verbosef("Zipped %s\n", monthPath)
        }
        return 1, nil
    } else {
        return 0, nil
    }
}

func archiveFiles(cfg *conf.Config, dirStr string, rootStat os.FileInfo) (int, error) {
	count := 0
	fileList, err := filepath.Glob(filepath.Join(dirStr, cfg.Pattern))
	if err != nil {
		return 0, err
	}
	archiveTime := cfg.ArchiveTime()
	for _, fname := range fileList {
		fs, e := os.Stat(fname)
		if e != nil {
			fmt.Printf("Error! Failed to consider %s for archiving: %s\n", fname, e)
			continue
		}

        if fs.Name() == conf.ConfName || fs.IsDir() {
            continue
        }

		if fs.ModTime().Before(archiveTime) {
			if newname, e := archiveFile(fname, cfg, rootStat, fs); e != nil {
				fmt.Printf("Error! Failed to archive %s - %s\n", fname, e)
			} else {
				util.Verbosef("Moved %s --> %s\n", fname, newname)
				count += 1
			}
		}
	}
	return count, nil
}

func archiveFile(fname string, cfg *conf.Config, rootStat os.FileInfo, fs os.FileInfo) (newname string, err error) {
	year := strconv.Itoa(fs.ModTime().Year())
	month := fs.ModTime().Month().String()

	archivePath := filepath.Join(cfg.ArchivePath(rootStat.Name()), year, month)
	err = os.MkdirAll(archivePath, rootStat.Mode())
	if err != nil && !os.IsExist(err) {
		return
	}

	zipPath := archivePath + ".zip"
	if util.FileExists(zipPath) {
		//unzip so we can archive the new file ... it will be rezipped later
		if err = util.Unzip(zipPath, archivePath); err != nil {
			return
		}
	}

	newname = filepath.Join(archivePath, fs.Name())
	if _, err = os.Stat(newname); err == nil {
		err = fmt.Errorf("A file of the same name already exists in the archive")
		return
	}
	err = os.Rename(fname, newname)

	return
}
