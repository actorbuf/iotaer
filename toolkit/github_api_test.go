package toolkit

import (
	"fmt"
	"testing"
)

func TestGetCurlReleaseURL(t *testing.T) {
	ver, err := GetCurlReleaseVersion()
	if err != nil {
		panic(err)
	}
	if err := DownloadCurlAndMove(ver); err != nil {
		panic(err)
	}
}

func TestGetGoReleaseList(t *testing.T) {
	tags, err := GetGoReleaseList()
	if err != nil {
		panic(err)
	}

	fmt.Println(tags)
}
