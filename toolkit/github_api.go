package toolkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/actorbuf/builder/rename"
	"github.com/elliotchance/pie/pie"
	"github.com/guonaihong/gout"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

var ProtobufRepo = new(RepoInfo) // protobuf 的仓库release信息
var CurlRepo = new(RepoInfo)     // curl 信息更新

func UnmarshalRepoInfo(data []byte) (RepoInfo, error) {
	var r RepoInfo
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *RepoInfo) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type RepoInfo struct {
	URL         string  `json:"url"`
	ReleaseURL  string  `json:"html_url"`
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	CreatedAt   string  `json:"created_at"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

type Asset struct {
	URL                string `json:"url"`
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	DownloadCount      int64  `json:"download_count"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type RepoFileTreeInfo struct {
	Name string       `json:"name"`
	Path string       `json:"path"`
	SHA  string       `json:"sha"`
	Size int64        `json:"size"`
	URL  string       `json:"url"`
	Type RepoFileType `json:"type"`
}

type RepoFileType string

const (
	RepoFileTypeDir  RepoFileType = "dir"
	RepoFileTypeFile RepoFileType = "file"
)

// GetRepoInfo 获取仓库信息
func GetRepoInfo(url string) (*RepoInfo, error) {
	var info *RepoInfo
	err := gout.GET(url).SetTimeout(5 * time.Second).BindJSON(info).Do()
	if err != nil {
		return nil, err
	}
	return info, nil
}

// GetProtobufReleaseURL 获取原始下载链接
func GetProtobufReleaseURL() (string, error) {
	err := gout.GET("https://api.github.com/repos/protocolbuffers/protobuf/releases/latest").
		SetTimeout(5 * time.Second).BindJSON(ProtobufRepo).Do()
	if err != nil {
		return "", err
	}

	assets := ProtobufRepo.Assets
	if len(assets) == 0 {
		return "", errors.New("miss repo info")
	}

	switch GetGOOS() {
	case Windows:
		for _, asset := range assets {
			if strings.HasSuffix(asset.Name, "win64.zip") {
				return asset.BrowserDownloadURL, nil
			}
		}
	case Linux:
		for _, asset := range assets {
			if strings.HasSuffix(asset.Name, "linux-x86_64.zip") {
				return asset.BrowserDownloadURL, nil
			}
		}
	case Darwin:
		for _, asset := range assets {
			if strings.HasSuffix(asset.Name, "osx-x86_64.zip") {
				return asset.BrowserDownloadURL, nil
			}
		}
	}
	return "", fmt.Errorf("os: %s, arch: %s not support", GetGOOS(), GetGOARCH())
}

// DownloadProtocAndMove 下载 protoc 插件 并移动到对应位置
// 文件加速服务： https://github.com/zwc456baby/file-proxy
func DownloadProtocAndMove(releaseProtocPATH, oldProtocPATH string) error {
	src := fmt.Sprintf("https://pd.zwc365.com/cfdownload/%s", releaseProtocPATH)

	dlFilename := "./protobuf.builder.zip"
	unzipDir := "./tmp_protoc"

	defer func() {
		// 移除解压缩的文件夹
		if err := os.RemoveAll(unzipDir); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "remove dir %s error: %+v\n", unzipDir, err)
		}
		// 移除下载的临时文件
		if err := os.Remove(dlFilename); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "remove file %s error: %+v\n", dlFilename, err)
		}
	}()
	var file []byte
	if err := gout.GET(src).BindBody(&file).Do(); err != nil {
		return err
	}
	if err := ioutil.WriteFile(dlFilename, file, 0777); err != nil {
		return err
	}

	// 解压缩文件
	if err := unzip(dlFilename, unzipDir); err != nil {
		return fmt.Errorf("unzip file err: %+v", err)
	}

	// 删除以前的protoc文件
	_ = os.Remove(oldProtocPATH)

	oldProtocDir := GetFileDir(oldProtocPATH)
	oldProtocInc := fmt.Sprintf("%s/include", oldProtocDir)
	if GetGOOS() == Windows {
		oldProtocInc = fmt.Sprintf("%s\\include", oldProtocDir)
	}
	// 删除旧的 include 文件夹
	if err := os.RemoveAll(oldProtocInc); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "remove dir %s error: %+v\n", oldProtocInc, err)
	}

	// 移动
	protocSrc := "./tmp_protoc/bin/protoc"
	protocInc := "./tmp_protoc/include"
	if GetGOOS() == Windows {
		protocSrc = "./tmp_protoc/bin/protoc.exe"
	}
	if err := rename.Atomic(protocSrc, oldProtocPATH); err != nil {
		return fmt.Errorf("remove protoc to %s err: %+v\n", oldProtocPATH, err)
	}

	_, _ = fmt.Fprintf(os.Stdout, "include文件夹将被放置在: %+v\n", oldProtocInc)
	if err := rename.Atomic(protocInc, oldProtocInc); err != nil {
		return fmt.Errorf("remove proto include dir to %s err: %+v\n", oldProtocInc, err)
	}

	return nil
}

// GetCurlReleaseVersion 获取Curl最新版本信息
func GetCurlReleaseVersion() (string, error) {
	err := gout.GET("https://api.github.com/repos/curl/curl/releases/latest").
		SetTimeout(5 * time.Second).BindJSON(CurlRepo).Do()
	if err != nil {
		return "", err
	}

	logrus.Infof("repo info: %+v", CurlRepo)

	return CurlRepo.Name, nil
}

// DownloadCurlAndMove 下载 curl 安装到 go-bin 下
func DownloadCurlAndMove(version string) error {
	var dlName = "curl.release.zip"
	var unzipDir = "./tmp_curl"
	var dlUrl string

	dlUrl = fmt.Sprintf("https://curl.se/windows/dl-%s/curl-%s-win64-mingw.zip", version, version)
	switch GetGOOS() {
	case Windows:
		dlUrl = fmt.Sprintf("https://curl.se/windows/dl-%s/curl-%s-win64-mingw.zip", version, version)
	case Linux:
		_, _ = fmt.Fprintf(os.Stdout, "Linux用户请手动安装 curl 包\n")
		return fmt.Errorf("linux用户请手动安装 curl 包")
	default:
		return fmt.Errorf("此系统不支持: %+v", GetGOOS())
	}

	defer func() {
		// 移除解压缩的文件夹
		if err := os.RemoveAll(unzipDir); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "remove dir %s error: %+v\n", unzipDir, err)
		}
		// 移除下载的临时文件
		if err := os.Remove(dlName); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "remove file %s error: %+v\n", dlName, err)
		}
	}()

	var file []byte
	if err := gout.GET(dlUrl).SetTimeout(time.Minute).BindBody(&file).Do(); err != nil {
		return err
	}
	if err := ioutil.WriteFile(dlName, file, 0777); err != nil {
		return err
	}

	// 解压缩文件
	if err := unzip(dlName, unzipDir); err != nil {
		return fmt.Errorf("unzip file err: %+v", err)
	}

	// 移动文件
	var curlExe = fmt.Sprintf("%s/curl-%s-win64-mingw/bin/curl.exe", unzipDir, version)
	var curlCrt = fmt.Sprintf("%s/curl-%s-win64-mingw/bin/curl-ca-bundle.crt", unzipDir, version)

	goBin := GetGOBIN()

	moveToCurlExe := fmt.Sprintf("%s\\curl.exe", goBin)
	if err := rename.Atomic(curlExe, moveToCurlExe); err != nil {
		return fmt.Errorf("remove curl.exe to %s err: %+v\n", moveToCurlExe, err)
	}

	moveToCurlCrt := fmt.Sprintf("%s\\curl-ca-bundle.crt", goBin)
	if err := rename.Atomic(curlCrt, moveToCurlCrt); err != nil {
		return fmt.Errorf("remove curl-ca-bundle.crt to %s err: %+v\n", moveToCurlCrt, err)
	}

	return nil
}

// GetGoReleaseList 获取Go释出列表
func GetGoReleaseList() ([]string, error) {
	var fileTree []*RepoFileTreeInfo

	var api = "https://api.github.com/repos/golang/dl/contents/?ref=master&_=1633673670227"
	err := gout.GET(api).SetHeader(gout.H{
		"Accept": "application/vnd.github.v3+json",
	}).
		SetTimeout(5 * time.Second).BindJSON(&fileTree).Do()
	if err != nil {
		return nil, err
	}

	var reg = regexp.MustCompile("go(\\d\\.\\d{2,}\\.\\d+)")
	var tagList []string
	for _, tag := range fileTree {
		if tag.Name == "gotip" {
			tagList = append(tagList, tag.Name)
			continue
		}
		if tag.Type != RepoFileTypeDir {
			continue
		}
		if reg.MatchString(tag.Name) {
			res := reg.FindStringSubmatch(tag.Name)
			if len(res) == 2 {
				tagList = append(tagList, res[1])
				continue
			}
		}
	}

	tagList = pie.Strings(tagList).Unique().Sort()

	return tagList, nil
}
