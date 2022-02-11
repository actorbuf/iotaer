package skeleton

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

var (
	scaffold = map[string]string{
		"/go.mod":                                 templateModule,
		"/.builderc":                              TemplateBuilderc,
		"/.gitignore":                             templateGitignore,
		"/module_update.sh":                       templateModuleUpdate,
		"/cmd/api.go":                             templateCmdNewApiServer,
		"/cmd/exec.go":                            templateCmdExec,
		"/common/log.go":                          templateLog,
		"/common/body.go":                         templateBody,
		"/common/code.go":                         templateCode,
		"/common/response.go":                     templateResponse,
		"/common/random.go":                       TemplateRandom,
		"/config_dev.yaml":                        templateDevYaml,
		"/config_prod.yaml":                       templateProdYaml,
		"/config_local.yaml":                      templateLocalYaml,
		"/config/config.go":                       templateConfig,
		"/model/model.proto":                      templateProto,
		"/internal/logic/demo.go":                 templateLogic,
		"/internal/router/router.go":              templateRouter,
		"/infra/middleware/common.go":             templateMiddleware,
		"/internal/services/demo_service.go":      templateServices,
		"/internal/controller/demo_controller.go": templateController,
	}
)

type Builder struct {
	Name string
	Path string
	SSH  string
}

func (b *Builder) Build() error {
	if err := os.MkdirAll(b.Path, 0755); err != nil {
		return err
	}
	if err := b.write(b.Path+"/"+b.Name+".go", templateMain); err != nil {
		return err
	}
	for sr, v := range scaffold {
		i := strings.LastIndex(sr, "/")
		if i > 0 {
			dir := sr[:i]
			if err := os.MkdirAll(b.Path+dir, 0755); err != nil {
				return err
			}
		}
		if err := b.write(b.Path+sr, v); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) write(name, tpl string) (err error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Failed")
		}
	}()
	fmt.Printf("create   %s \n", name)
	data, err := b.parse(tpl)
	if err != nil {
		return
	}
	return ioutil.WriteFile(name, data, 0644)
}

func (b *Builder) parse(s string) ([]byte, error) {
	t, err := template.New("").Parse(s)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, b); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
