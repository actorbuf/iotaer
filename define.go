package main

// ProjectName 项目名称的字符画
const ProjectName = ` _           _ _     _           
| |__  _   _(_) | __| | ___ _ __ 
| '_ \| | | | | |/ _` + "`" + ` |/ _ \ '__|
| |_) | |_| | | | (_| |  __/ |
|_.__/ \__,_|_|_|\__,_|\___|_|
`

// NeedUpdateFlag 是否需要强制更新的标志位
var NeedUpdateFlag bool

// Config 接管项目时 解析项目根下的配置项
type Config struct {
	FreqTo string `yaml:"freq_to" json:"freq_to"`
}
