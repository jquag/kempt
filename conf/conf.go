package conf

import (
	"encoding/json"
	"fmt"
	"github.com/jquag/kempt/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ArchiveHome    string
	ArchiveAgeDays int
	Pattern        string
	ZipAgeDays     int
	PurgeAgeDays   int
}

func (cfg *Config) ArchiveTime() time.Time {
	return daysToTime(cfg.ArchiveAgeDays)
}

func (cfg *Config) ZipTime() time.Time {
	return daysToTime(cfg.ZipAgeDays)
}

func (cfg *Config) PurgeTime() time.Time {
	return daysToTime(cfg.PurgeAgeDays)
}

func (cfg *Config) ArchivePath(root string) string {
	if cfg.ArchiveHome == "" {
		return root
	} else {
		return filepath.Join(root, cfg.ArchiveHome)
	}
}

func (cfg *Config) EachYear(rootpath string, maxDate time.Time, f func(path string, maxDate time.Time) (int, error)) (int, error) {
	archive, err := os.Open(cfg.ArchivePath(rootpath))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil //return but without an error
		}
		return 0, err
	}
	defer archive.Close()

	years, err := archive.Readdirnames(0)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, yearStr := range years {
		yearPath := filepath.Join(archive.Name(), yearStr)
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			util.Verbosef("Warning. Skipping archive folder - %s\n", yearPath)
			continue
		}
		if year <= maxDate.Year() {
            c, err := f(yearPath, maxDate)
            if err != nil {
                fmt.Println(err)
            } else {
                count += c
            }
		}
	}
    return count, nil
}

func (cfg *Config) EachMonth(rootpath string, maxDate time.Time, f func(path string, maxDate time.Time) (int, error)) (int, error) {
	return cfg.EachYear(rootpath, maxDate, func(yearPath string, maxDate time.Time) (int, error) {
		yearFolder, err := os.Open(yearPath)
		if err != nil {
			return 0, fmt.Errorf("Error! Could not access archive year, %s - %s", yearPath, err)
		}
		defer yearFolder.Close()

		monthUpperBound := 12
		if filepath.Base(yearPath) == strconv.Itoa(maxDate.Year()) {
			monthUpperBound = int(maxDate.Month())
		}

		var months []os.FileInfo
		if months, err = yearFolder.Readdir(0); err != nil {
			return 0, fmt.Errorf("Error! Could not access archive year, %s - %s", yearPath, err)
		}

        count := 0
		for _, month := range months {
            monthName := strings.Replace(month.Name(), filepath.Ext(month.Name()), "", -1)
            monthInt, found := util.MonthInts[monthName]
            if found && monthInt <= monthUpperBound {
                c, err := f(filepath.Join(yearPath, month.Name()), maxDate)
                if err != nil {
                    fmt.Println(err)
                } else {
                    count += c
                }
            }
		}
        return count, nil
	})
}

const ConfName = ".kempt-conf"

var Default = Config{
	"archive",
	30,
	"*",
	90,
	365,
}

func daysToTime(days int) time.Time {
	dur, err := time.ParseDuration(fmt.Sprintf("%dh", -24*days))
	if err != nil {
		panic(err)
	}
	return time.Now().Add(dur)
}

func Help() string {
	example, _ := json.MarshalIndent(Default, "", "  ")
	explanation := `
"ArchiveHome": directory to store the archived files
  e.g. if ArchiveHome = "archive" then your folder structure would be something like this...
  <KemptRoot>/
    recent.file
    archive/
        2014/
        ...

"ArchiveAgeDays": the number of days before a file is moved to the archive
  NOTE: the age of a file is based on its mtime

"Pattern": the file name pattern which defines which files kempt will manage
  e.g. if Pattern = "*.log" then it will manage any files that end in .log

"ZipAgeDays": the number of days before a month folder in the archive is zipped
  NOTE: the age of the month folder is based on the last day of the year/month represented by the folder name

"PurgeAgeDays": the number of days before a file is purged
  NOTE: the age of normal files are based on their mtime,
        the age of zipped months are based on the last day of the year/month represented by the file name

`
	return fmt.Sprintf("\n%s example:\n%s\n\nexplanation of settings:%s", ConfName, example, explanation)
}

func Exists(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, ConfName)); os.IsNotExist(err) {
		return false
	}
	return true
}

func Parse(dir string) (*Config, error) {
	jsonFromFile, err := ioutil.ReadFile(filepath.Join(dir, ConfName))
	if err != nil {
		return &Default, err
	}

	c := new(Config)
	err = json.Unmarshal(jsonFromFile, c)
	if err != nil {
		return &Default, err
	}
	return c, nil
}

func Init(dir string) error {
	file, err := os.Create(filepath.Join(dir, ConfName))
	if err != nil {
		return err
	}
	defaultJson, err := json.MarshalIndent(Default, "", "  ")
	if err != nil {
		return err
	}
	_, err = file.Write(defaultJson)
	return err
}
