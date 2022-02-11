package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/actorbuf/builder/toolkit"
	"gopkg.in/yaml.v3"
	"io/fs"
	"io/ioutil"
	"os"
	ose "os/exec"
	"strings"
	"time"
)

const builderConfig = ".builderc"
const checkInterval = 900
const updateCheckInterval = 3600

var (
	golangVersion = cobra.Command{
		Use:   "go",
		Short: "golang 版本管理",
	}
)

type BuilderConfig struct {
	Commit           string `yaml:"commit"`              // 当前commit
	UpdatedAt        string `yaml:"updated-at"`          // 更新时间
	UpdatedAtUnix    int64  `yaml:"updated-at-unix"`     // 更新时间戳 小于1小时 则不检测更新
	CheckedAtUnix    int64  `yaml:"checked-at-unix"`     // 检测时间戳 检测间隔小于15分钟 则不检测更新
	Message          string `yaml:"message"`             // 更新信息
	NewCommit        string `yaml:"new-commit"`          // 仓库最新版本
	NewUpdatedAt     string `yaml:"new-updated-at"`      // 仓库最新代码提交时间
	NewUpdatedAtUnix int64  `yaml:"new-updated-at-unix"` // 时间戳
	NewMessage       string `yaml:"new-message"`         // 仓库最新提交信息
	UserID           int64  `yaml:"userid"`              // git 用户ID
	Username         string `yaml:"username"`            // git 用户名
	Email            string `yaml:"email"`               // git 邮箱
}

func getBuilderConfigFilePATH() (string, error) {
	// 检测环境
	var homeDir string
	if toolkit.GetGOOS() == toolkit.Windows {
		homeDir = os.Getenv("USERPROFILE")
		if homeDir == "" {
			driveName := os.Getenv("HOMEDRIVE")
			homePath := os.Getenv("HOMEPATH")
			homeDir = fmt.Sprintf("%s%s\\", driveName, homePath)
		} else {
			homeDir = fmt.Sprintf("%s\\", homeDir)
		}
	} else {
		homeDir = os.Getenv("HOME")
		if homeDir == "" {
			userData, err := ose.Command("sh", "-c", "eval echo $USER").CombinedOutput()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "get user info err: %+v\n", err)
				return "", err
			}
			userName := strings.ReplaceAll(string(userData), "\n", "")
			if userName == "" {
				_, _ = fmt.Fprintf(os.Stderr, "get username empty")
				return "", fmt.Errorf("get username empty")
			}
			if userName == "root" {
				homeDir = fmt.Sprintf("/root/")
			} else {
				homeDir = fmt.Sprintf("/home/%s/", userName)
			}
		} else {
			homeDir = fmt.Sprintf("%s/", homeDir)
		}
	}
	configFile := fmt.Sprintf("%s%s", homeDir, builderConfig)
	return configFile, nil
}

// checkBuilderVersion 判别是否有新版本
func checkBuilderVersion() bool {
	configFile, err := getBuilderConfigFilePATH()
	if err != nil {
		return false
	}
	var config = BuilderConfig{}

	_, err = os.Stat(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			_, _ = fmt.Fprintf(os.Stderr, "check builder config err: %+v\n", err)
			return false
		}
		// 新建配置文件
		data, err := yaml.Marshal(config)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "marshal config file err: %+v\n", err)
			return false
		}
		if err := ioutil.WriteFile(configFile, data, fs.ModePerm); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "write new builder config err: %+v\n", err)
			return false
		}
	}
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "open builder config err: %+v\n", err)
		return false
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unmarshal builder config err: %+v\n", err)
		return false
	}

	// 限制检测频率 1小时
	var tnu = time.Now().Unix()
	if tnu-config.UpdatedAtUnix < updateCheckInterval {
		return false
	}
	// 检测时间间隔
	if tnu-config.CheckedAtUnix < checkInterval {
		return false
	}

	// 获取仓库最新提交
	commit, err := toolkit.GetRepoLatestCommit(566)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "get repo commit info err: %+v\n", err)
		return false
	}

	// 比较是否要更新
	if config.Commit == commit.ShortID {
		config.CheckedAtUnix = tnu
		_ = rewriteConfig(configFile, config)
		return false
	}

	if config.Commit == "" {
		config.Commit = "-"
	}
	if config.UpdatedAt == "" {
		config.UpdatedAt = "-"
	}
	ca, _ := toolkit.RFC3339ToTime(commit.CommittedDate)
	_, _ = fmt.Fprintf(os.Stdout, "builder有新版本: %s || 更新时间: %s\n", commit.ShortID, ca.Format("2006-01-02 15:04:05"))
	_, _ = fmt.Fprintf(os.Stdout, "当前builder版本: %s || 更新时间: %s\n", config.Commit, config.UpdatedAt)
	_, _ = fmt.Fprintf(os.Stdout, "你可以执行 builder update 进行手动更新\n")

	config.NewCommit = commit.ShortID
	config.NewMessage = commit.Title
	config.NewUpdatedAt = ca.Format("2006-01-02 15:04:05")
	config.NewUpdatedAtUnix = ca.Unix()
	config.CheckedAtUnix = ca.Unix()

	_ = rewriteConfig(configFile, config)

	return true
}

func rewriteConfig(path string, config BuilderConfig) error {
	newBody, err := yaml.Marshal(config)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "marshal builder config err: %+v\n", err)
		return err
	}

	if err := ioutil.WriteFile(path, newBody, fs.ModePerm); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "rewrite builder config err: %+v\n", err)
		return err
	}
	return nil
}

func getBuilderConfig() *BuilderConfig {
	cfg, err := getBuilderConfigFilePATH()
	if err != nil {
		return nil
	}

	var config BuilderConfig

	content, err := ioutil.ReadFile(cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "open builder config err: %+v\n", err)
		return nil
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "unmarshal builder config err: %+v\n", err)
		return nil
	}

	return &config
}

func golangVersionCommand() *cobra.Command {
	golangVersion.AddCommand(golangVersionList())   // golang 版本列表
	golangVersion.AddCommand(golangVersionUpdate()) // golang 版本更新\

	return &golangVersion
}

func golangVersionList() *cobra.Command {
	var cmd = cobra.Command{
		Use:   "list",
		Short: "最近的20个版本列表",
		Run: func(cmd *cobra.Command, args []string) {
			tags, err := toolkit.GetGoReleaseList()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "get go tag err: %+v\n", err)
				return
			}
			var template strings.Builder
			for i := 0; i < len(tags); i++ {
				if i%4 == 0 {
					template.WriteString("\n")
				}
				template.WriteString(fmt.Sprintf("%s	", tags[i]))
			}

			template.WriteString("\n\n")

			_, _ = fmt.Fprintf(os.Stdout, template.String())
		},
	}

	return &cmd
}

func golangVersionUpdate() *cobra.Command {
	var ver string
	var cmd = cobra.Command{
		Use:   "update",
		Short: "go 更新到指定版本",
		Run: func(cmd *cobra.Command, args []string) {
			data, err := ose.Command("go", "install",
				fmt.Sprintf("golang.org/dl/go%s@latest", ver)).CombinedOutput()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "download go error: %+v\n", err)
				return
			}
			_, _ = fmt.Fprintln(os.Stdout, string(data))
			_, _ = fmt.Fprintf(os.Stdout, fmt.Sprintf("安装完毕后,请手动配置 go%s 的版本别名\n", ver))
		},
	}

	cmd.Flags().StringVar(&ver, "version", ver, "指定从list命令取得的版本号,gotip体验最新开发分支特性")
	return &cmd
}
