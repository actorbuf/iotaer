package main

import (
	"fmt"
	"github.com/actorbuf/builder/toolkit"
	"os"
	ose "os/exec"
	"regexp"
	"strings"
)

// 安装 protoc
func installProtoc() {
	var needReinstallProtoc bool
	var protocUrl string
	_, _ = fmt.Fprintf(os.Stdout, "-----\n检测 protoc 插件中...\n")
	protocPath, err := ose.LookPath("protoc")
	if err != nil {
		needReinstallProtoc = true
		_, _ = fmt.Fprintf(os.Stderr, "检测失败: %+v\n开始重新安装 protoc 插件\n", err)
	}

	// 存在 是否需要更新
	protocUrl, err = toolkit.GetProtobufReleaseURL()
	if err != nil {
		// 获取不到版本信息 不处理重装
		_, _ = fmt.Fprintf(os.Stderr, "获取 protoc 插件信息失败: %+v\n", err)
	}

	// 检测本地版本信息
	var localVersion string
	if !needReinstallProtoc {
		v, err := ose.Command("protoc", "--version").CombinedOutput()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "获取本地 protoc 版本信息失败: %+v\n开始重新安装 protoc 插件\n", err)
			goto Reinstall
		}
		version := strings.Trim(string(v), " \n")
		reg := regexp.MustCompile(`(\d+.\d+.\d+)`)
		realsV := reg.FindAllString(version, 1)
		if len(realsV) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "获取本地 protoc 版本信息失败: %+v\n开始重新安装 protoc 插件\n", err)
			goto Reinstall
		}
		localVersion = fmt.Sprintf("v%s", realsV[0])
		_, _ = fmt.Fprintf(os.Stdout, "protoc local version: %+v\n", localVersion)
	}
	_, _ = fmt.Fprintf(os.Stdout, "protoc latest version: %+v\n", toolkit.ProtobufRepo.TagName)
	if localVersion == toolkit.ProtobufRepo.TagName {
		_, _ = fmt.Fprintf(os.Stdout, "protoc 处理完毕\n")
		return
	}
Reinstall:
	if protocPath == "" {
		protocPath = toolkit.GetDefaultProtocPATH()
	}
	if toolkit.GetGOOS() == toolkit.Windows {
		protocPath = toolkit.WindowsSplit(protocPath)

		// 检测盘符是否正确 rename 操作符无法跨磁盘移动
		nowDir, err := os.Getwd()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "获取当前操作位置失败\n")
			return
		}
		nowDiskDriverName := toolkit.GetWindowsDiskDriverName(nowDir)
		insDiskDriverName := toolkit.GetWindowsDiskDriverName(protocPath)

		if nowDiskDriverName != insDiskDriverName {
			_, _ = fmt.Fprintf(os.Stderr, "当前操作盘符: %s. 由于无法跨盘符安装, 请前往$GOBIN=%s 的盘符: %s 下重入命令进行安装\n",
				nowDiskDriverName, toolkit.GetGOBIN(), insDiskDriverName)
		}
	}

	_, _ = fmt.Fprintf(os.Stdout, "正在下载 protoc...\n插件将被放置到: %+v\n", protocPath)
	if err := toolkit.DownloadProtocAndMove(protocUrl, protocPath); err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "下载操作失败: %+v\n请手动重新安装 protoc 插件\n", err)
		return
	}
	v, err := ose.Command("protoc", "--version").CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "获取本地 protoc 版本信息失败: %+v\n请手动重新安装 protoc 插件\n", err)
		return
	}
	version := strings.Trim(string(v), " \n")
	_, _ = fmt.Fprintf(os.Stdout, "protoc version: %+v\n", version)
	_, _ = fmt.Fprintf(os.Stdout, "安装 protoc完毕\n")
}

// installProtocGoInjectTag 更新 protoc-go-inject-tag
func installProtocGoInjectTag() {
	_, _ = fmt.Fprintf(os.Stdout, "-----\n更新 protoc-go-inject-tag 插件中...")
	tagO, err := ose.Command("go", "install", "github.com/favadi/protoc-go-inject-tag@latest").CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"安装protoc-go-inject-tag报错: %+v\n\n你可以手动重装,执行如下命令:\n	go install github.com/favadi/protoc-go-inject-tag@latest\n", err)
		return
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "%+v\nprotoc-go-inject-tag更新完成\n", string(tagO))
		return
	}
}

// installProtocGenGo 更新 protoc-gen-go
func installProtocGenGo() {
	_, _ = fmt.Fprintf(os.Stdout, "-----\n更新 protoc-gen-go 插件中...")
	genO, err := ose.Command("go", "install", "github.com/golang/protobuf/protoc-gen-go@latest").CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"安装protoc-gen-go报错: %+v\n\n你可以手动重装,执行如下命令:\n	go install github.com/golang/protobuf/protoc-gen-go@latest\n", err)
		return
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "%+v\nprotoc-gen-go更新完成\n", string(genO))
	}
}

// installProtocGenGo 更新 protoc-gen-go-grpc
func installProtocGenGoGrpc() {
	_, _ = fmt.Fprintf(os.Stdout, "-----\n更新 protoc-gen-go-grpc 插件中...\n")
	genO, err := ose.Command("go", "install", "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest").CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"安装 protoc-gen-go-grpc 报错: %+v\n\n你可以手动重装,执行如下命令:\n	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest\n", err)
		return
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "%+v\nprotoc-gen-go-grpc 更新完成\n-----\n", string(genO))
	}
}

// upgradeBuilder builder 的自我更新
func upgradeBuilder() error {
	_, _ = fmt.Fprintf(os.Stdout, "-----\n更新 builder 插件中...\n")
	genO, err := ose.Command("go", "install", "github.com/actorbuf/builder@latest").CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"安装 builder 报错: %+v\n\n你可以手动重装,执行如下命令:\n	github.com/actorbuf/builder@latest\n", err)
		return err
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "%+v\nbuilder 更新完成\n-----\n", string(genO))
	}
	return nil
}
