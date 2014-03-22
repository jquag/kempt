package cmd

import (
	"fmt"
	"github.com/jquag/kempt/conf"
	"github.com/jquag/kempt/util"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func runPurge(cfg *conf.Config, path string) error {
	if cfg.PurgeAgeDays > 0 {
		rootCount, err := purgeInFolder(path, cfg.PurgeTime(), cfg.Pattern)
		if err != nil {
			fmt.Printf("Error! Failed to purge files from the root - %s\n", err)
		}

		archiveCount, err := cfg.EachMonth(path, cfg.PurgeTime(), purgeMonth)
		if err != nil {
			fmt.Printf("Error! Failed to purge files from archive month folders - %s\n", err)
		}

		fmt.Printf("purged %d file(s)\n", rootCount+archiveCount)
	}
	return nil
}

func purgeMonth(path string, maxDate time.Time) (int, error) {
    if strings.HasSuffix(path, ".zip") {
        err := os.Remove(path)
        if err != nil {
            return 0, err
        }
        util.Verbosef("Purged %s\n", path)
        return 1, nil
    } else {
        return purgeInFolder(path, maxDate, "*")
    }
}

func purgeInFolder(path string, maxDate time.Time, pattern string) (int, error) {
	dir, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("Error! Could not access archive folder %s - %s", path, err)
	}
	defer dir.Close()

	fileList, err := filepath.Glob(filepath.Join(path, pattern))
	if err != nil {
		return 0, fmt.Errorf("Error! Could not access archive folder %s - %s", path, err)
	}

	count := 0
	for _, f := range fileList {
		fi, err := os.Stat(f)
		if err != nil {
			fmt.Printf("Error! Failed to get stats for %s - %s\n", f, err)
		}

        if fi.Name() == conf.ConfName {
            continue
        }

		if fi.ModTime().Before(maxDate) {
			filePath := filepath.Join(path, fi.Name())
			err := os.Remove(filePath)
			if err != nil {
				fmt.Printf("Error! Failed to purge %s - %s\n", filePath, err)
			} else {
				util.Verbosef("Purged %s\n", filePath)
				count += 1
			}
		}
	}

	return count, nil
}
