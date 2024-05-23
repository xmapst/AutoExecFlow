package utils

import (
	"os"
<<<<<<< HEAD
	"strings"
	"time"
=======
	"path/filepath"
	"strings"
>>>>>>> githubB
)

func FileOrPathExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func EnsureDirExist(name string) error {
	if !FileOrPathExist(name) {
		return os.MkdirAll(name, os.ModePerm)
	}
	return nil
}

func SliceToStrMap(s []string) map[string]string {
	m := make(map[string]string)
	for _, v := range s {
		slice := strings.Split(v, "=")
		switch {
		case len(slice) > 2:
			m[slice[0]] = strings.Join(slice[1:], "=")
		case len(slice) == 2:
			m[slice[0]] = slice[1]
		case len(slice) == 1:
			m[v] = ""
		}
	}
	return m
}

<<<<<<< HEAD
func ClearOldScript(path string) {
=======
func ClearDir(path string) {
>>>>>>> githubB
	_ = os.RemoveAll(path)
	_ = EnsureDirExist(path)
}

<<<<<<< HEAD
func TimeStr(nsec int64) string {
	if nsec == 0 {
		return ""
	}
	return time.Unix(0, nsec).Format(time.RFC3339)
=======
func PathEscape(s string) string {
	s = filepath.Clean(strings.TrimPrefix(s, ".."))
	if s == ".." {
		return ""
	}
	if !strings.HasPrefix(s, "..") {
		return s
	}
	return PathEscape(s)
}

// RemoveDuplicate removes duplicate elements from a slice while maintaining the original order.
func RemoveDuplicate[T comparable](slice []T) []T {
	allKeys := make(map[T]bool)
	var list []T
	for _, item := range slice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}

	return list
>>>>>>> githubB
}
