package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"io/ioutil"
	"os"
	ose "os/exec"
	"path/filepath"
	"strings"

	"github.com/actorbuf/iotaer/toolkit"
	"gopkg.in/yaml.v3"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	format "github.com/actorbuf/proto-format"
	proto "github.com/actorbuf/proto-parser"
)

func main() {
	exec()
}

func exec() {
	if NeedUpdateFlag {
		fmt.Println("稍等, 正在执行更新操作...")
		if err := updateBuilder().Execute(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if len(os.Args) >= 2 && os.Args[1] == "update" {
			return // 是更新操作 自己取消掉 前面已经做了
		}
	}
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(addTask())                         // 新增一个系统定时任务
	rootCmd.AddCommand(versionInfo())                     // 打印builder版本信息
	rootCmd.AddCommand(addRPCCommand())                   // 新增一个RPC
	rootCmd.AddCommand(addAPICommand())                   // 新增一个API
	rootCmd.AddCommand(updateBuilder())                   // 检测并更新builder
	rootCmd.AddCommand(addSvcCommand())                   // 添加一个服务
	rootCmd.AddCommand(addRouteCommand())                 // 添加一个路由组
	rootCmd.AddCommand(buildRunCommand())                 // 快速运行iota项目
	rootCmd.AddCommand(outputMdCommand())                 // 生成一个api的md文档
	rootCmd.AddCommand(buildHTTPCommand())                // 生成一个新项目
	rootCmd.AddCommand(buildProtoCommand())               // proto生成
	rootCmd.AddCommand(formatProtoCommand())              // 格式化一个proto文件
	rootCmd.AddCommand(gitListenerCommand())              // 用于local环境布署的slack监听服务
	rootCmd.AddCommand(toolBuilderCommand())              // 工具生成
	rootCmd.AddCommand(generateK8sIngressYmlCommand())    // 生成k8s ingress文件
	rootCmd.AddCommand(installDependentPackageCommand())  // 更新builder依赖的工具链
	rootCmd.AddCommand(generateK8sDeploymentYmlCommand()) // 生成k8s deployment文件
	rootCmd.AddCommand(addRouteV2Command())               // 添加一个路由组v2 --做了些diy
	rootCmd.AddCommand(addErrorCodeFileCommand())         // 创建错误码proto文件
	rootCmd.AddCommand(buildProtoV2Command())             // proto生成，测试版本
}

var (
	rootCmd = &cobra.Command{
		Use:   "builder",
		Short: "builder 是基于 omega 库的一个提高生产效率的工具链",
	}
)

func addTask() *cobra.Command {
	pbPath, _ := os.Getwd()
	var svc = ""
	var name = ""
	var genTo = ""

	cmd := &cobra.Command{
		Use:     "addtask",
		Short:   "给系统添加一个定时任务",
		Long:    "svc不存在时会自动创建一个service",
		Example: "builder addtask --path model/parser.proto --gen infra/task/crm_task.go --svc CrmTask --name RefreshUserInfo",
		Run: func(cmd *cobra.Command, args []string) {
			if pbPath == "" {
				_, _ = fmt.Fprintln(os.Stderr, "path is required")
				os.Exit(1)
			}
			if svc == "" {
				_, _ = fmt.Fprintln(os.Stderr, "svc is required")
				os.Exit(1)
			}
			if name == "" {
				_, _ = fmt.Fprintln(os.Stderr, "name is required")
				os.Exit(1)
			}
			if err := proto.AddTask(pbPath, svc, name, genTo); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "proto单个文件地址")
	cmd.Flags().StringVar(&name, "name", name, "任务名称")
	cmd.Flags().StringVar(&svc, "svc", svc, "任务service名称")
	cmd.Flags().StringVar(&genTo, "gen", genTo, "代码生成位置")
	return cmd
}

func installDependentPackageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "更新builder依赖的工具链",
		Run: func(cmd *cobra.Command, args []string) {
			// 是否已经安装 protoc
			installProtoc()

			// 安装 goimports
			installGoImports()

			// 更新 protoc-go-inject-tag
			installProtocGoInjectTag()

			// 更新 protoc-gen-go
			installProtocGenGo()

			// 更新 grpc
			installProtocGenGoGrpc()

			// builder 自我更新
			_ = upgradeBuilder()
		},
	}

	return cmd
}

func versionInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "打印builder版本信息",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return cmd
}

func updateBuilder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "检测并更新builder",
		Run: func(cmd *cobra.Command, args []string) {

		},
		DisableFlagParsing: true,
	}

	return cmd
}

func buildHTTPCommand() *cobra.Command {
	var name string
	output, _ := os.Getwd()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建一个新项目",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	cmd.Flags().StringVar(&name, "name", "demo", "项目名称")
	cmd.Flags().StringVar(&output, "path", output, "项目输出路径")
	return cmd
}

func buildRunCommand() *cobra.Command {
	const (
		envLocal   = "local" // 本地环境
		envDev     = "dev"   // 开发环境
		envTest    = "test"  // 测试环境
		envRelease = "prod"  // 正式环境

		envLocalConfig   = "config_local.yaml"
		envDevConfig     = "config_dev.yaml"
		envTestConfig    = "config_test.yaml"
		envReleaseConfig = "config_prod.yaml"
	)
	var env = envLocal
	var envConfig = envLocalConfig
	var config = ""
	mainPath, _ := os.Getwd()
	cmd := &cobra.Command{
		Use:   "run",
		Short: "快速运行项目",
		Long:  "通过指定 env,path,config 等参数快速运行一个已有项目",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				logrus.Errorf("run arg empty")
				return
			}

			if env == "" {
				env = os.Getenv("OMEGA_ENV")
			}
			if env == "" {
				env = envLocal
			}
			if env != "" {
				switch env {
				case envLocal:
					envConfig = envLocalConfig
				case envDev:
					envConfig = envDevConfig
				case envTest:
					envConfig = envTestConfig
				case envRelease:
					envConfig = envReleaseConfig
				}
			}

			if config != "" {
				envConfig = config
			}

			var params = []string{"run", mainPath, args[0], "--config", envConfig}

			var protocStdout bytes.Buffer
			var protocStderr bytes.Buffer
			protocCmd := ose.Command("go", params...)
			protocCmd.Stdout = &protocStdout
			protocCmd.Stderr = &protocStderr
			logrus.Infof("run: %s", protocCmd.String())
			if err := protocCmd.Run(); err != nil {
				logrus.Errorf("run err: %+v", err)
				logrus.Errorf("std_out: %s", protocStdout.String())
				logrus.Errorf("std_err: %s", protocStderr.String())
				return
			}
		},
	}
	cmd.Flags().StringVar(&env, "env", "", "指定运行环境, 如果没有指定该参数并且系统中没有指定 `OMEGA_ENV` 环境变量, 将默认指定 `local`")
	cmd.Flags().StringVar(&mainPath, "path", mainPath, "项目 main 函数入口路径")
	cmd.Flags().StringVar(&config, "config", "", "强制指定项目配置文件,不建议使用,这将覆盖env参数")
	return cmd
}

func buildProtoCommand() *cobra.Command {
	pbPath, _ := os.Getwd()
	var goOut = filepath.Dir(pbPath)
	var grpcOut = ""
	var include = []string{"."}
	var needFmt bool
	var noScope bool
	var dbType = "mdbc"
	var isApi bool

	// 解析项目下的builder配置
	var parseConfig = func() Config {
		var c Config
		body, err := ioutil.ReadFile("./.builderc")
		if err != nil {
			return c
		}
		if err = yaml.Unmarshal(body, &c); err != nil {
			return c
		}
		return c
	}

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "解析proto文件, 自动生成开发代码.",
		Long:  "生成代码时请在项目根目录下执行,默认生成路径为当前目录,所以proto依赖请写项目全路径\n",
		Run: func(cmd *cobra.Command, args []string) {
			// 解析项目下的配置项
			c := parseConfig()
			err := proto.CodeGen(&proto.CodeGenConfig{
				PbFilePath:       pbPath,
				OutputPath:       goOut,
				GrpcOutputPath:   grpcOut,
				IncludePbFiles:   include,
				OutputNeedFormat: needFmt,
				NoGetScopeFunc:   noScope,
				DbDriveType:      dbType,
				FreqOutput:       c.FreqTo,
			})
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, err.Error())
				return
			}

			// api
			if isApi {
				pbGoList := toolkit.ListSpecialSuffixFile(pbPath, ".pb.go")

				for _, filePath := range pbGoList {
					if strings.HasSuffix(filePath, "error_code.pb.go") {
						// error_code.pb.go文件跳过
						continue
					}

					fileStopImport, _ := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ImportsOnly)

					b, _ := toolkit.ReadAll(filePath)
					rawStr := string(b)
					importStr := rawStr[0:int(fileStopImport.End())]

					file, err := parser.ParseFile(token.NewFileSet(), "", rawStr, parser.ParseComments)

					if err != nil {
						_, _ = fmt.Fprint(os.Stderr, err)
						os.Exit(1)
						return
					}

					decls := file.Decls

					var structsStr string

					for _, dV := range decls {
						dr, ok := dV.(*ast.GenDecl)
						if ok {
							for _, spec := range dr.Specs {
								typeSpec, ok := spec.(*ast.TypeSpec)
								if ok {
									_, ok := typeSpec.Type.(*ast.StructType)
									if ok {
										//fields := structType.Fields
										structsStr += rawStr[(int(dr.Pos()) - 1):int(dr.End())] //  (int(dr.End() - dr.Pos()))
									}

								}
							}
						}
					}

					allStr := importStr + structsStr
					sli := []string{"state         protoimpl.MessageState", "sizeCache     protoimpl.SizeCache", "unknownFields protoimpl.UnknownFields"}
					for _, old := range sli {
						allStr = strings.ReplaceAll(allStr, old, "")
					}
					err = toolkit.RewriteFile(filePath, allStr)
					if err != nil {
						_, _ = fmt.Fprint(os.Stderr, err)
						os.Exit(1)
						return
					}
				}

				// 格式化
				err = toolkit.ExecCommand("gofmt", "-w", pbPath)
				if err != nil {
					_, _ = fmt.Fprint(os.Stderr, err)
					os.Exit(1)
					return
				}

				// import 格式化
				err = toolkit.ExecCommand("goimports", "-w", pbPath)
				if err != nil {
					_, _ = fmt.Fprint(os.Stderr, err)
					os.Exit(1)
					return
				}
			}
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "proto文件地址,支持传入目录")
	cmd.Flags().StringVar(&goOut, "out", goOut, "proto文件生成位置,是--go_out的别名")
	cmd.Flags().StringVar(&grpcOut, "grpc", grpcOut, "生成grpc,指定grpc生成位置,是--go-grpc_out的别名, 默认不生成")
	cmd.Flags().StringSliceVar(&include, "include", include, "proto文件依赖路径,是-I(-IPATH/--proto_path)的别名")
	cmd.Flags().BoolVar(&needFmt, "fmt", needFmt, "是否需要格式化输出的文件,开启时只需要指定 --fmt 后面不需要任何参数(可以确保代码风格统一)")
	cmd.Flags().BoolVar(&noScope, "no-scope", noScope, "是否忽略数据库驱动的GetScope()代码生成, 不建议开启")
	cmd.Flags().StringVar(&dbType, "db", dbType, "生成代码的数据库驱动类型,可选[mdbc,gdbc]")
	cmd.Flags().BoolVar(&isApi, "is-api", isApi, "是否生成的是api形式")
	return cmd
}

func buildProtoV2Command() *cobra.Command {
	pbPath, _ := os.Getwd()
	var goOut = filepath.Dir(pbPath)
	var grpcOut = ""
	var include = []string{"."}
	var needFmt bool
	var noScope bool
	var dbType = "mdbc"
	var isApi bool

	// 解析项目下的builder配置
	var parseConfig = func() Config {
		var c Config
		body, err := ioutil.ReadFile("./.builderc")
		if err != nil {
			return c
		}
		if err = yaml.Unmarshal(body, &c); err != nil {
			return c
		}
		return c
	}

	cmd := &cobra.Command{
		Use:   "genV2",
		Short: "解析proto文件, 自动生成开发代码.",
		Long:  "生成代码时请在项目根目录下执行,默认生成路径为当前目录,所以proto依赖请写项目全路径",
		Run: func(cmd *cobra.Command, args []string) {
			exist := toolkit.IsCurrentDirHasModfile() // 当前目录是否有mod文件
			if !exist {
				return
			}

			if pbPath == "" {
				_, _ = fmt.Fprintf(os.Stderr, "proto文件地址 -- path 不能为空\n")
				return
			}

			if !toolkit.IsExist(pbPath) {
				_, _ = fmt.Fprintf(os.Stderr, "proto文件地址 -- path 不存在\n")
				return
			}

			// 解析项目下的配置项
			c := parseConfig()
			err := proto.CodeGen(&proto.CodeGenConfig{
				PbFilePath:       pbPath,
				OutputPath:       goOut,
				GrpcOutputPath:   grpcOut,
				IncludePbFiles:   include,
				OutputNeedFormat: needFmt,
				NoGetScopeFunc:   noScope,
				DbDriveType:      dbType,
				FreqOutput:       c.FreqTo,
			})
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, err.Error())
				return
			}

			// api
			if isApi {
				pbGoList := toolkit.ListSpecialSuffixFile(pbPath, ".pb.go")

				for _, filePath := range pbGoList {
					if strings.HasSuffix(filePath, "error_code.pb.go") {
						// error_code.pb.go文件跳过
						continue
					}

					fileStopImport, _ := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ImportsOnly)

					b, _ := toolkit.ReadAll(filePath)
					rawStr := string(b)
					importStr := rawStr[0:int(fileStopImport.End())]

					file, err := parser.ParseFile(token.NewFileSet(), "", rawStr, parser.ParseComments)

					if err != nil {
						_, _ = fmt.Fprint(os.Stderr, err)
						os.Exit(1)
						return
					}

					decls := file.Decls

					var structsStr string

					for _, dV := range decls {
						dr, ok := dV.(*ast.GenDecl)
						if ok {
							for _, spec := range dr.Specs {
								typeSpec, ok := spec.(*ast.TypeSpec)
								if ok {
									_, ok := typeSpec.Type.(*ast.StructType)
									if ok {
										//fields := structType.Fields
										structsStr += rawStr[(int(dr.Pos()) - 1):int(dr.End())] //  (int(dr.End() - dr.Pos()))
									}

								}
							}
						}
					}

					allStr := importStr + structsStr
					sli := []string{"state         protoimpl.MessageState", "sizeCache     protoimpl.SizeCache", "unknownFields protoimpl.UnknownFields"}
					for _, old := range sli {
						allStr = strings.ReplaceAll(allStr, old, "")
					}
					err = toolkit.RewriteFile(filePath, allStr)
					if err != nil {
						_, _ = fmt.Fprint(os.Stderr, err)
						os.Exit(1)
						return
					}
				}

				// 格式化
				err = toolkit.ExecCommand("gofmt", "-w", pbPath)
				if err != nil {
					_, _ = fmt.Fprint(os.Stderr, err)
					os.Exit(1)
					return
				}

				// import 格式化
				err = toolkit.ExecCommand("goimports", "-w", pbPath)
				if err != nil {
					_, _ = fmt.Fprint(os.Stderr, err)
					os.Exit(1)
					return
				}
			}
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "proto文件地址,支持传入目录")
	cmd.Flags().StringVar(&goOut, "out", goOut, "proto文件生成位置,是--go_out的别名")
	cmd.Flags().StringVar(&grpcOut, "grpc", grpcOut, "生成grpc,指定grpc生成位置,是--go-grpc_out的别名, 默认不生成")
	cmd.Flags().StringSliceVar(&include, "include", include, "proto文件依赖路径,是-I(-IPATH/--proto_path)的别名")
	cmd.Flags().BoolVar(&needFmt, "fmt", needFmt, "是否需要格式化输出的文件,开启时只需要指定 --fmt 后面不需要任何参数(可以确保代码风格统一)")
	cmd.Flags().BoolVar(&noScope, "no-scope", noScope, "是否忽略数据库驱动的GetScope()代码生成, 不建议开启")
	cmd.Flags().StringVar(&dbType, "db", dbType, "生成代码的数据库驱动类型,可选[mdbc,gdbc]")
	cmd.Flags().BoolVar(&isApi, "is-api", isApi, "是否生成的是api形式")
	return cmd
}

func formatProtoCommand() *cobra.Command {
	pbPath, _ := os.Getwd()
	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "格式化 proto 文件使其看的赏心悦目",
		Run: func(cmd *cobra.Command, args []string) {
			err := format.Format(pbPath)
			if err != nil {
				_, _ = fmt.Fprint(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "proto文件地址,支持传入目录")
	return cmd
}

func gitListenerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitter",
		Short: "拉取git仓库代码并布署",
		Long:  "通过slack监听实现自动发现git仓库tag变更,拉取git代码并实现布署",
		Run: func(cmd *cobra.Command, args []string) {
			//lis := gitter.Gitter{
			//	BotUserToken: "xoxb-2066066646995-2066407547827-Ir3IH5QNw3jhRLPfLS0qOYJq",
			//	AppLeveToken: "xapp-1-A0224A1PNUU-2066398507266-54c4ad8e30801f8dc3b5f80f9b0102aaffd3cdb4722fa1f9df0e69662cd2c3b1",
			//}
			//lis.Listener()
		},
	}
	return cmd
}

func addAPICommand() *cobra.Command {
	pbPath, _ := os.Getwd()
	var routerGroup = ""
	var routerName = ""
	var routerMethod = "POST"
	cmd := &cobra.Command{
		Use:   "addapi",
		Short: "给路由组新增一个api",
		Long:  "仅适用于Register方式注入路由",
		Run: func(cmd *cobra.Command, args []string) {
			if routerName == "" {
				_, _ = fmt.Fprintf(os.Stderr, "路由名称 --name 不能为空")
				return
			}
			routerName = toolkit.FirstUpper(routerName)
			err := proto.AddAPI(pbPath, routerGroup, routerName, routerMethod)
			if err != nil {
				_, _ = fmt.Fprint(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "proto单个文件地址")
	cmd.Flags().StringVar(&routerName, "name", routerName, "该路由名称")
	cmd.Flags().StringVar(&routerGroup, "svc", routerGroup, "需要注入的路由组名称")
	cmd.Flags().StringVar(&routerMethod, "method", routerMethod, "请求方式POST/GET...")
	return cmd
}

func addRPCCommand() *cobra.Command {
	pbPath, _ := os.Getwd()
	var svcName = ""
	var rpcName = ""
	cmd := &cobra.Command{
		Use:   "addrpc",
		Short: "给service新增一个rpc",
		Long:  "给service新增一个rpc",
		Run: func(cmd *cobra.Command, args []string) {
			err := proto.AddRPC(pbPath, svcName, rpcName)
			if err != nil {
				_, _ = fmt.Fprint(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "proto单个文件地址")
	cmd.Flags().StringVar(&rpcName, "name", rpcName, "该RPC名称")
	cmd.Flags().StringVar(&svcName, "svc", svcName, "需要注入到哪个service中")
	return cmd
}

func outputMdCommand() *cobra.Command {
	pbPath, _ := os.Getwd()
	var svcName = ""
	var rpcName = ""
	var include []string
	cmd := &cobra.Command{
		Use:   "md",
		Short: "给路由组的api输出markdown文档",
		Long:  "给路由组的api输出markdown文档",
		Run: func(cmd *cobra.Command, args []string) {
			err := proto.OutputMD(pbPath, svcName, rpcName, include)
			if err != nil {
				_, _ = fmt.Fprint(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "proto单个文件地址")
	cmd.Flags().StringVar(&rpcName, "name", rpcName, "路由/RPC的名称")
	cmd.Flags().StringVar(&svcName, "svc", svcName, "路由/RPC在哪个service中")
	cmd.Flags().StringSliceVar(&include, "include", include, "路由/RPC依赖到的其他proto文件列表,用于字段注释补全(不支持目录)")
	return cmd
}

func generateK8sDeploymentYmlCommand() *cobra.Command {
	namespace := ""
	serviceName := ""
	maxSurge := ""
	maxUnavailable := ""
	replicas := ""
	startCommand := ""
	cpuMax := ""
	memMax := ""
	cpuMin := ""
	memMin := ""
	port := ""
	targetPort := ""
	protocol := ""
	portsName := ""
	version := ""
	appProtocol := ""
	var needVersion string
	cmd := &cobra.Command{
		Use:   "gen-k8s-deployment-yml",
		Short: "生成k8s deployment yml文件",
		Long:  "生成k8s deployment yml文件",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	cmd.Flags().StringVar(&serviceName, "svc", "", "服务名，不能为空")
	cmd.Flags().StringVar(&startCommand, "startCommand", "", "启动命令，不能为空")
	cmd.Flags().StringVar(&port, "port", "", "启动端口，不是网络服务不需要")
	cmd.Flags().StringVar(&namespace, "namespace", "actor", "命名空间")
	cmd.Flags().StringVar(&maxSurge, "maxSurge", "100%", "滚动更新时，可以有多少个额外的 Pod，可以是数字也可以为百分比")
	cmd.Flags().StringVar(&maxUnavailable, "maxUnavailable", "0%", "滚动更新时，我们可以忍受多少个 Pod 无法提供服务，可以是数字也可以为百分比")
	cmd.Flags().StringVar(&replicas, "replicas", "2", "最大pod数")
	cmd.Flags().StringVar(&cpuMax, "cpuMax", "400m", "cpu最大使用资源")
	cmd.Flags().StringVar(&memMax, "memMax", "512Mi", "memory最大使用资源")
	cmd.Flags().StringVar(&cpuMin, "cpuMin", "200m", "cpu预划资源")
	cmd.Flags().StringVar(&memMin, "memMin", "200Mi", "memory预划资源")
	cmd.Flags().StringVar(&targetPort, "targetPort", "", "服务对外端口，不填与启动端口一致")
	cmd.Flags().StringVar(&protocol, "protocol", "TCP", "协议")
	cmd.Flags().StringVar(&portsName, "portsName", "http", "ports name")
	cmd.Flags().StringVar(&version, "version", "", "k8s版本，1.20字段值填写 v1.20")
	cmd.Flags().StringVar(&needVersion, "needVersion", "", "deployment是否需要加上版本号 特殊配置(true/false)")
	cmd.Flags().StringVar(&appProtocol, "appProtocol", "", "service的appProtocol字段，默认不设置，需要用到istio需要用到，可参考：https://istio.io/latest/zh/docs/ops/configuration/traffic-management/protocol-selection/")
	return cmd
}

func generateK8sIngressYmlCommand() *cobra.Command {
	namespace := ""
	serviceName := ""
	host := ""
	port := ""
	secretName := ""
	ingressBasePath := ""
	version := ""
	isWebsocket := ""
	lbMethod := ""

	cmd := &cobra.Command{
		Use:   "gen-k8s-ingress-yml",
		Short: "生成k8s ingress yml文件",
		Long:  "生成k8s ingress yml文件",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	cmd.Flags().StringVar(&serviceName, "svc", "", "服务名，不能为空")
	cmd.Flags().StringVar(&port, "port", "", "服务端口，不能为空")
	cmd.Flags().StringVar(&namespace, "namespace", "actor", "命名空间")
	cmd.Flags().StringVar(&host, "host", "", "host，不能为空")
	cmd.Flags().StringVar(&ingressBasePath, "ingressBasePath", "/", "根路径")
	cmd.Flags().StringVar(&secretName, "secretName", "", "证书名")
	cmd.Flags().StringVar(&version, "version", "", "k8s版本，1.20字段值填写 v1.20")
	cmd.Flags().StringVar(&isWebsocket, "isWebsocket", "0", "是不是websocket服务，是填1")
	cmd.Flags().StringVar(&lbMethod, "lbMethod", "round_robin", "lb方式，可选参数自己看nginx官方文档吧")
	return cmd
}

func addRouteCommand() *cobra.Command {
	var pbPath = ""
	var svcName = ""
	var genTo = ""
	var apiPath = ""
	cmd := &cobra.Command{
		Use:   "addroute",
		Short: "快速添加一个路由组",
		Long:  "快速添加一个路由组",
		Run: func(cmd *cobra.Command, args []string) {
			// windows 兼容 用这几个盘符足够了吧
			if toolkit.GetGOOS() == toolkit.Windows {
				if strings.HasPrefix(apiPath, "C:") || strings.HasPrefix(apiPath, "D:") ||
					strings.HasPrefix(apiPath, "E:") || strings.HasPrefix(apiPath, "F:") ||
					strings.HasPrefix(apiPath, "G:") || strings.HasPrefix(apiPath, "H:") {
					_, _ = fmt.Fprintf(os.Stderr, "windows模式下 --api 参数不要以 / 开头, 程序将自动帮你补全\n")
					return
				}
				if apiPath != "" {
					apiPath = fmt.Sprintf("/%s", apiPath)
				}
			}
			if svcName == "" {
				_, _ = fmt.Fprintf(os.Stderr, "路由组的名称 -- name 不能为空\n")
				return
			}
			cname := toolkit.Calm2Case(svcName)

			if pbPath == "" {
				pbPath, _ = os.Getwd()
				_tempDir := pbPath
				if toolkit.GetGOOS() == toolkit.Windows {
					pbPath = fmt.Sprintf("%s\\%s.proto", pbPath, cname)
				} else {
					pbPath = fmt.Sprintf("%s/%s.proto", pbPath, cname)
				}
				_, _ = fmt.Fprintf(os.Stdout, "填写proto生成路径为空,将生成至缺省路径: %+v\n", pbPath)
				// 创建文件
				_, err := os.Stat(pbPath)
				// 需要创建文件 填充内容
				if err != nil {
					if !os.IsNotExist(err) {
						_, _ = fmt.Fprintf(os.Stderr, "检索文件失败: %+v\n", err)
						return
					}
					pkgName := filepath.Base(_tempDir)
					pkgName = strings.ReplaceAll(pkgName, "-", "_") // 替换- 防止不识别包
					pkgName = toolkit.Calm2Case(pkgName)
					content := fmt.Sprintf("syntax = \"proto3\";\npackage %s;\n", pkgName)
					err := ioutil.WriteFile(pbPath, []byte(content), fs.ModePerm)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "写入%s文件失败: %+v\n", pbPath, err)
						return
					}
				}
			}
			if genTo == "" {
				genTo = fmt.Sprintf("./internal/controller/%s_controller.go", cname)
			}
			if apiPath == "" {
				apiPath = fmt.Sprintf("/api/%s", cname)
			}

			if err := proto.AddRoute(pbPath, svcName, apiPath, genTo); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "add route err: %+v\n", err)
				return
			}
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "要将该路由组生成到哪个proto文件中,填写文件地址")
	cmd.Flags().StringVar(&svcName, "name", svcName, "生成的router路由组的名称,一般为模块名,如User")
	cmd.Flags().StringVar(&genTo, "gento", genTo, "该路由组具体路由的实现方法将被生成到的位置,默认生成到./internal/controller/module_name_controller.go文件下")
	cmd.Flags().StringVar(&apiPath, "api", apiPath, "供访问的路由组api前缀,如: /api/user/info 组路由api前缀为 /api/user")

	return cmd
}

func addErrorCodeFileCommand() *cobra.Command {
	var pbPath = ""
	cmd := &cobra.Command{
		Use:   "addErrorCodeFile",
		Short: "快速添加一个错误码文件",
		Long:  "快速添加一个错误码文件",
		Run: func(cmd *cobra.Command, args []string) {
			dirSeparator := toolkit.GetDirectorySeparator() // 当前文件系统文件夹分隔符
			// 获取 module name
			modName := ""
			exist := toolkit.IsCurrentDirHasModfile() // 当前目录是否有mod文件
			if exist {
				var err error
				modName, err = toolkit.GetCurrentModuleName()
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, err.Error())
					return
				}
			} else {
				return
			}

			fullPbPath := ""
			if pbPath == "" {
				_, _ = fmt.Fprintf(os.Stderr, "-- path 不能为空\n")
				return
			} else {
				// 路径必须是当前项目目录下开始的路径
				if toolkit.GetGOOS() == toolkit.Windows {
					winDirPrefixSlice := []string{"C:", "c:", "D:", "d:", "E:", "e:", "F:", "f:", "G:", "g:", "H:", "h:"}

					for _, winDirPrefix := range winDirPrefixSlice {
						if strings.HasPrefix(pbPath, winDirPrefix) {
							_, _ = fmt.Fprintf(os.Stderr, "-- path 必须是当前项目下的相对路径，Like this .\\proto\\a.proto\n")
							return
						}
					}
				} else {
					if pbPath[0:1] == "/" {
						_, _ = fmt.Fprintf(os.Stderr, "-- path 必须是当前项目下的相对路径，Like this ./proto/ \n")
						return
					}
				}

				// 去除前缀
				if strings.HasPrefix(pbPath, "./") || strings.HasPrefix(pbPath, ".\\") {
					pbPath = pbPath[2:]
				}

				currentDir, _ := os.Getwd()

				// proto文件完整路径
				fullPbPath = currentDir + dirSeparator +
					strings.Trim(toolkit.CorrectingDirSeparator(pbPath), dirSeparator) +
					dirSeparator + "error_code.proto"

				// proto文件是否存在
				fullPbPathIsExist := toolkit.IsExist(fullPbPath)

				if !fullPbPathIsExist {
					// 文件不存在则生成文件先
					pbBasePath := filepath.Dir(fullPbPath)
					pbBasePathExist := toolkit.IsExist(pbBasePath)
					if !pbBasePathExist {
						// 递归创建文件夹
						err := os.MkdirAll(pbBasePath, os.ModePerm)
						if err != nil {
							_, _ = fmt.Fprintf(os.Stderr, "递归创建文件夹失败："+err.Error())
							return
						}
					}

					lastSeparatorIndex := strings.LastIndex(fullPbPath, dirSeparator)

					pkgName := filepath.Base(fullPbPath[:lastSeparatorIndex])
					pkgName = strings.ReplaceAll(pkgName, "-", "_") // 替换- 防止不识别包
					pkgName = toolkit.Calm2Case(pkgName)
					goPackage := modName + "/" + strings.ReplaceAll(pbPath, dirSeparator, "/")

					initData := `enum ErrCode {
    ErrCodeNil         =    0;
}`
					content := fmt.Sprintf("syntax = \"proto3\";\n\npackage %s;\n\noption go_package = \"%s\";\n\n\n%s", pkgName, strings.Trim(goPackage, "/"), initData)
					err := ioutil.WriteFile(fullPbPath, []byte(content), fs.ModePerm)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "写入%s文件失败: %+v\n", pbPath, err)
						return
					}

				} else {
					_, _ = fmt.Fprintf(os.Stderr, "文件已存在："+fullPbPath)
				}
			}

			fmt.Println("Successed ! 请查看文件：" + fullPbPath)
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "要将该error_code.proto生成在哪个文件夹,填写文件夹地址")

	return cmd
}

func addRouteV2Command() *cobra.Command {
	var pbPath = ""
	var svcName = ""
	var genTo = ""
	var apiPath = ""
	cmd := &cobra.Command{
		Use:   "addrouteV2",
		Short: "快速添加一个路由组",
		Long:  "快速添加一个路由组",
		Run: func(cmd *cobra.Command, args []string) {
			dirSeparator := toolkit.GetDirectorySeparator() // 当前文件系统文件夹分隔符
			// 获取 module name
			modName := ""
			exist := toolkit.IsCurrentDirHasModfile() // 当前目录是否有mod文件
			if exist {
				var err error
				modName, err = toolkit.GetCurrentModuleName()
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, err.Error())
					return
				}
			} else {
				return
			}

			if svcName == "" {
				_, _ = fmt.Fprintf(os.Stderr, "路由组的名称 -- name 不能为空\n")
				return
			}
			cname := toolkit.Calm2Case(svcName)

			fullPbPath := ""
			if pbPath == "" {
				_, _ = fmt.Fprintf(os.Stderr, "proto文件 -- path 不能为空\n")
				return
			} else {
				// proto文件路径必须是当前项目目录下开始的路径
				if toolkit.GetGOOS() == toolkit.Windows {
					winDirPrefixSlice := []string{"C:", "c:", "D:", "d:", "E:", "e:", "F:", "f:", "G:", "g:", "H:", "h:"}

					for _, winDirPrefix := range winDirPrefixSlice {
						if strings.HasPrefix(pbPath, winDirPrefix) {
							_, _ = fmt.Fprintf(os.Stderr, "proto文件 -- path 必须是当前项目下的相对路径，Like this .\\proto\\a.proto\n")
							return
						}
					}
				} else {
					if pbPath[0:1] == "/" {
						_, _ = fmt.Fprintf(os.Stderr, "proto文件 -- path 必须是当前项目下的相对路径，Like this ./proto/a.proto\n")
						return
					}
				}

				// 不为空的路径，看看指定的文件是不是 .proto结尾，不是就告错
				protoLength := len(pbPath)
				if protoLength < 6 {
					_, _ = fmt.Fprintf(os.Stderr, "proto文件 -- path 格式不正确\n")
					return
				}
				if pbPath[0:2] == "./" || pbPath[0:2] == ".\\" {
					pbPath = pbPath[2:]
				}

				if !strings.Contains(pbPath, ".proto") {
					_, _ = fmt.Fprintf(os.Stderr, "proto文件 -- path 后缀不正确，必须为 .proto\n")
					return
				}

				currentDir, _ := os.Getwd()

				// proto文件完整路径
				fullPbPath = currentDir + dirSeparator + toolkit.CorrectingDirSeparator(pbPath)

				// proto文件是否存在
				fullPbPathIsExist := toolkit.IsExist(fullPbPath)

				if !fullPbPathIsExist {
					// 文件不存在则生成文件先
					pbBasePath := filepath.Dir(fullPbPath)
					pbBasePathExist := toolkit.IsExist(pbBasePath)
					if !pbBasePathExist {
						// 递归创建文件夹
						err := os.MkdirAll(pbBasePath, os.ModePerm)
						if err != nil {
							_, _ = fmt.Fprintf(os.Stderr, "递归创建文件夹失败："+err.Error())
							return
						}
					}

					lastSeparatorIndex := strings.LastIndex(fullPbPath, dirSeparator)
					pkgName := filepath.Base(fullPbPath[:lastSeparatorIndex])
					pkgName = strings.ReplaceAll(pkgName, "-", "_") // 替换- 防止不识别包
					pkgName = toolkit.Calm2Case(pkgName)
					goPackage := modName + "/" + strings.ReplaceAll(pbPath, dirSeparator, "/")[:]
					goPackageLength := len(goPackage)
					goPackage = goPackage[:goPackageLength-len(fullPbPath[lastSeparatorIndex+1:])-1]

					content := fmt.Sprintf("syntax = \"proto3\";\npackage %s;\noption go_package = \"%s\"", pkgName, goPackage)
					err := ioutil.WriteFile(fullPbPath, []byte(content), fs.ModePerm)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "写入%s文件失败: %+v\n", pbPath, err)
						return
					}

				}
			}

			if genTo == "" {
				genTo = fmt.Sprintf("./internal/controller/%s_controller.go", cname)
			}
			if apiPath == "" {
				apiPath = fmt.Sprintf("/api/%s", cname)
			}

			svcName = toolkit.FirstUpper(svcName)
			if err := proto.AddRoute(pbPath, svcName, apiPath, genTo); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "add route err: %+v\n", err)
				return
			}

			fmt.Println("Successed ! 请查看文件：" + fullPbPath)
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "要将该路由组生成到哪个proto文件中,填写文件地址")
	cmd.Flags().StringVar(&svcName, "name", svcName, "生成的router路由组的名称,一般为模块名,如User")
	cmd.Flags().StringVar(&genTo, "gento", genTo, "该路由组具体路由的实现方法将被生成到的位置,默认生成到./internal/controller/module_name_controller.go文件下")
	cmd.Flags().StringVar(&apiPath, "api", apiPath, "供访问的路由组api前缀,如: /api/user/info 组路由api前缀为 /api/user")

	return cmd
}

func addSvcCommand() *cobra.Command {
	var pbPath = ""
	var svcName = ""
	var genTo = ""
	var apiPath = ""
	cmd := &cobra.Command{
		Use:   "addsvc",
		Short: "快速添加一个rpc服务",
		Long:  "快速添加一个rpc服务",
		Run: func(cmd *cobra.Command, args []string) {
			// windows 兼容 用这几个盘符足够了吧
			if toolkit.GetGOOS() == toolkit.Windows {
				if strings.HasPrefix(apiPath, "C:") || strings.HasPrefix(apiPath, "D:") ||
					strings.HasPrefix(apiPath, "E:") || strings.HasPrefix(apiPath, "F:") ||
					strings.HasPrefix(apiPath, "G:") || strings.HasPrefix(apiPath, "H:") {
					_, _ = fmt.Fprintf(os.Stderr, "windows模式下 --api 参数不要以 / 开头, 程序将自动帮你补全\n")
					return
				}
				if apiPath != "" {
					apiPath = fmt.Sprintf("/%s", apiPath)
				}
			}
			if svcName == "" {
				_, _ = fmt.Fprintf(os.Stderr, "路由组的名称 -- name 不能为空\n")
				return
			}
			cname := toolkit.Calm2Case(svcName)

			if pbPath == "" {
				pbPath, _ = os.Getwd()
				_tempDir := pbPath
				if toolkit.GetGOOS() == toolkit.Windows {
					pbPath = fmt.Sprintf("%s\\%s.proto", pbPath, cname)
				} else {
					pbPath = fmt.Sprintf("%s/%s.proto", pbPath, cname)
				}
				_, _ = fmt.Fprintf(os.Stdout, "填写proto生成路径为空,将生成至缺省路径: %+v\n", pbPath)
				// 创建文件
				_, err := os.Stat(pbPath)
				// 需要创建文件 填充内容
				if err != nil {
					if !os.IsNotExist(err) {
						_, _ = fmt.Fprintf(os.Stderr, "检索文件失败: %+v\n", err)
						return
					}
					pkgName := filepath.Base(_tempDir)
					pkgName = strings.ReplaceAll(pkgName, "-", "_") // 替换- 防止不识别包
					pkgName = toolkit.Calm2Case(pkgName)
					content := fmt.Sprintf("syntax = \"proto3\";\npackage %s;\n", pkgName)
					err := ioutil.WriteFile(pbPath, []byte(content), fs.ModePerm)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "写入%s文件失败: %+v\n", pbPath, err)
						return
					}
				}
			}
			if genTo == "" {
				genTo = fmt.Sprintf("./internal/services/%s_service.go", cname)
			}
			if apiPath == "" {
				apiPath = fmt.Sprintf("/api/%s", cname)
			}

			if err := proto.AddRoute(pbPath, svcName, apiPath, genTo); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "add route err: %+v\n", err)
				return
			}
		},
	}
	cmd.Flags().StringVar(&pbPath, "path", pbPath, "要将该路由组生成到哪个proto文件中,填写文件地址")
	cmd.Flags().StringVar(&svcName, "name", svcName, "生成的服务名称,一般为模块名,如User")
	cmd.Flags().StringVar(&genTo, "gento", genTo, "该服务的实现方法将被生成到的位置,默认生成到./internal/services/module_name_service.go文件下")

	return cmd
}
