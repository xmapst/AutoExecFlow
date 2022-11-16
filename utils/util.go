package utils

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	CodeErrParam = iota + 1000
	CodeRunning
	CodeExecErr
	CodeErrNoData
	CodeSuccess = 0
	CodeErrApp  = 500
)

var MsgFlags = map[int]string{
	CodeSuccess:   "成功",
	CodeErrApp:    "内部错误",
	CodeErrParam:  "参数错误",
	CodeRunning:   "执行中",
	CodeExecErr:   "执行失败",
	CodeErrNoData: "沒有数据",
}

// GetMsg get error information based on Code
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[CodeErrApp]
}

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

func RootDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	exPath := filepath.Dir(ex)
	return exPath, nil
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

func ClearOldScript(path string) {
	_ = os.RemoveAll(path)
	_ = EnsureDirExist(path)
}

func GetCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	fmt.Println(res)
	return res
}

func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}
