package toolkit

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	ose "os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"
)

type GOOS string

func (g GOOS) String() string {
	return string(g)
}

const (
	Windows GOOS = "windows"
	Linux   GOOS = "linux"
	Darwin  GOOS = "darwin"
	FreeBSD GOOS = "freebsd"
)

// GetGOOS 获取系统信息
func GetGOOS() GOOS {
	if runtime.GOOS == "windows" {
		return Windows
	}
	if runtime.GOOS == "darwin" {
		return Darwin
	}
	if runtime.GOOS == "freebsd" {
		return FreeBSD
	}
	return Linux
}

// GetGOARCH 获取架构信息
func GetGOARCH() string {
	return runtime.GOARCH
}

// GetGOBIN 获取go相关的二进制目录 没有 GOBIN 则找 GOROOT
func GetGOBIN() string {
	out, err := ose.Command("go", "env", "GOBIN").CombinedOutput()
	if err != nil {
		if GetGOOS() == Windows {
			return fmt.Sprintf("%s\\bin", runtime.GOROOT())
		}
		return fmt.Sprintf("%s/bin", runtime.GOROOT())
	}
	outP := strings.Trim(string(out), " \n")
	if outP == "" {
		if GetGOOS() == Windows {
			return fmt.Sprintf("%s\\bin", runtime.GOROOT())
		}
		return fmt.Sprintf("%s/bin", runtime.GOROOT())
	}
	return outP
}

func GetDefaultProtocPATH() string {
	if GetGOOS() == Windows {
		return fmt.Sprintf("%s\\protoc.exe", GetGOBIN())
	}
	return fmt.Sprintf("%s/protoc", GetGOBIN())
}

func WindowsSplit(src string) string {
	return strings.ReplaceAll(src, "/", "\\")
}

func GetFileDir(src string) string {
	return filepath.Dir(src)
}

func Calm2Case(src string) string {
	buffer := new(bytes.Buffer)
	for i, r := range src {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.WriteRune('_')
			}
			buffer.WriteRune(unicode.ToLower(r))
		} else {
			buffer.WriteRune(r)
		}
	}
	return buffer.String()
}

// GetWindowsDiskDriverName 获取目录盘符
func GetWindowsDiskDriverName(dir string) string {
	return string([]rune(dir)[0])
}

// FirstUpper 字符串首字母大写
func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// FirstLower 字符串首字母小写
func FirstLower(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func RFC3339ToTime(value string) (time.Time, error) {
	ts, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Now(), err
	}

	return ts.In(time.Local), nil
}

func GetDirectorySeparator() string {
	if GetGOOS() == Windows {
		return "\\"
	}

	return "/"
}

// 校正文件路径分隔符
func CorrectingDirSeparator(filepath string) string {
	separator := GetDirectorySeparator()
	if GetGOOS() == Windows {
		return strings.ReplaceAll(filepath, "/", separator)
	}
	return strings.ReplaceAll(filepath, "\\", separator)
}

func IsCurrentDirHasModfile() bool {
	modFile := "go.mod"
	exist := IsExist(modFile)

	if !exist {
		_, _ = fmt.Fprintf(os.Stderr, "go.mod 不存在！ 请使用 'go mod init' ")
	}

	return exist
}

func GetCurrentModuleName() (modName string, err error) {
	modName = ""

	text, err := GetFileLineOne("go.mod")
	if err != nil {
		return
	}

	if len(text) < 8 {
		err = errors.New("go.mod文件格式不正确！")
		return
	}

	modName = text[7 : len(text)-1]
	modName = strings.Trim(modName, "\r")
	modName = strings.Trim(modName, "\n")

	return
}

func ReadAll(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

// ListSpecialSuffixFile 获取当前目录下的所有指定后缀文件
func ListSpecialSuffixFile(path, suffix string) []string {
	path = strings.TrimSuffix(path, "/")
	var result []string
	fi, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, file := range fi {
		if file.IsDir() {
			result = append(result, ListSpecialSuffixFile(path+"/"+file.Name(), suffix)...)
		}
		if strings.HasSuffix(file.Name(), suffix) {
			result = append(result, path+"/"+file.Name())
		}
	}
	return result
}

func RewriteFile(path, str string) error {
	// 打开一个存在的文件，将原来的内容覆盖掉
	// O_WRONLY: 只写, O_TRUNC: 清空文件
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer file.Close() // 关闭文件
	// 带缓冲区的*Writer
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(str)
	if err != nil {
		return err
	}

	// 将缓冲区中的内容写入到文件里
	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func ExecCommand(name string, arg ...string) error {
	_, err := ose.Command(name, arg...).CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

// IsDir checks whether the path is a dir,
// it returns true when it's a directory or does not exist.
func IsDir(f string) bool {

	fi, err := os.Stat(f)

	if err != nil {

		return false

	}

	return fi.IsDir()

}

func GetFileLineOne(filePath string) (lineOneText string, err error) {
	lineOneText = ""
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		err = errors.New(filePath + " open file error: " + err.Error())
		return
	}
	//建立缓冲区，把文件内容放到缓冲区中
	buf := bufio.NewReader(f)
	//遇到\n结束读取
	b, err := buf.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			err = errors.New(filePath + " is empty! ")
			return
		}
		err = errors.New(filePath + " read bytes error: " + err.Error())
		return
	}
	lineOneText = string(b)
	err = nil
	return
}

func GetLineWithchars(filePath, chars string) (lineText string, err error)  {
	lineText = ""
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		err = errors.New(filePath + " open file error: " + err.Error())
		return
	}
	//建立缓冲区，把文件内容放到缓冲区中
	buf := bufio.NewReader(f)
	for {
		//遇到\n结束读取
		var b []byte
		b, err = buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		text := string(b)
		if strings.Contains(text, chars) {
			lineText = text
		}
	}

	return
}