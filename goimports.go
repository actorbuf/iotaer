package main

import (
	"fmt"
	"os"
	ose "os/exec"
)

// installGoImports 更新 goimports
func installGoImports() {
	_, _ = fmt.Fprintf(os.Stdout, "-----\n更新 goimports 插件中...")
	tagO, err := ose.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest").CombinedOutput()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"安装 goimports 报错: %+v\n\n你可以手动重装,执行如下命令:\n	go install golang.org/x/tools/cmd/goimports@latest\n", err)
		return
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "%+v\ngoimports更新完成\n", string(tagO))
		return
	}
}
