package parser

import (
	"errors"
	"github.com/sqweek/dialog"
	"os"
	"path"
	"slices"
	"strings"
	"time"
)

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetGorgonFolder(possiblePaths []string) (string, error) {
	for _, possiblePath := range possiblePaths {
		if ok, err := Exists(possiblePath); ok && err == nil {
			return possiblePath, nil
		}
	}

	dir, err := dialog.Directory().Title("Project Gorgon Chat Logs Folder").Browse()

	if err != nil {
		return "", err
	}

	return dir, nil
}

func GetSortedLogFiles(p string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(p)

	if err != nil {
		return nil, err
	}

	var files []os.FileInfo

	for _, v := range entries {
		stat, err := os.Stat(path.Join(p, v.Name()))

		if err != nil {
			return nil, err
		}

		if !stat.IsDir() {
			files = append(files, stat)
		}
	}

	slices.SortFunc(files, func(a, b os.FileInfo) int {
		return b.ModTime().Compare(a.ModTime())
	})

	return files, nil
}

func GetLatestLogFile(p string) (string, error) {
	files, err := GetSortedLogFiles(p)

	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", errors.New("no log file found")
	}

	return files[0].Name(), nil
}

func IsFileMostRecent(path string) bool {
	now := time.Now()

	_, rest, found := strings.Cut(path, "Chat-")
	if !found {
		return false
	}

	date, _, found := strings.Cut(rest, ".log")
	if !found {
		return false
	}

	parsedDate, err := time.ParseInLocation(DateFormat, date, now.Location())
	if err != nil {
		return false
	}

	return parsedDate.Before(now) && now.Before(parsedDate.Add(time.Hour*24))
}
