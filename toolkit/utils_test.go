package toolkit

import (
	"fmt"
	"testing"
)

func TestGoRoot(t *testing.T) {
	fmt.Println(GetGOBIN())
}

func TestGetFileDir(t *testing.T) {
	fmt.Println(GetFileDir("C:\\Program Files\\Go\\bin\\protoc.exe"))
}

func TestGetWindowsDiskDriverName(t *testing.T) {
	fmt.Println(GetWindowsDiskDriverName("C:\\Program Files\\Go\\bin\\protoc.exe"))
}
