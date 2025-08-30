package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/gdme1320/zip/internal"
	"github.com/gdme1320/zip/internal/utils"
	zip "github.com/gdme1320/zip/pkg"
)

// ZipFile 表示一个zip文件
type ZipFile struct {
	Path     string
	Password string
}

// UnzipConfig 解压配置
type UnzipConfig struct {
	ZipPath          string // zip文件路径
	OutputPath       string // 输出路径
	FileEncoding     string // 文件编码 (gbk, utf8, windows)
	Password         string // 密码
	PasswordEncoding string // 密码编码 (gbk, utf8, windows)
	Workers          int    // 并发工作线程数
	Verbose          bool   // 详细输出
	Quiet            bool   // 静默输出
}

// usage prints the application's usage information.
func usage() {
	fmt.Printf("用法: %s <命令> [选项] [文件]\n", os.Args[0])
	fmt.Println("\n命令:")
	fmt.Println("  x        从归档中解压文件")
	fmt.Println("  t        列出归档中的内容")
	fmt.Println("\n示例:")
	fmt.Printf("  %s x archive.zip -C ./extracted -p 123456\n", os.Args[0])
	fmt.Printf("  %s t archive.zip -e gbk\n", os.Args[0])
}

// parseAndValidateFlags parses command-line flags and validates the configuration.
func parseAndValidateFlags() (*UnzipConfig, string, error) {
	if len(os.Args) < 2 {
		return nil, "", fmt.Errorf("需要一个命令 (x 或 t)")
	}

	command := os.Args[1]
	args := os.Args[2:]
	config := &UnzipConfig{}

	fs := flag.NewFlagSet(command, flag.ExitOnError)

	fs.StringVar(&config.OutputPath, "C", ".", "解压输出路径")
	fs.StringVar(&config.FileEncoding, "e", "", "文件名编码 (gbk, utf8)")
	fs.StringVar(&config.Password, "p", "", "解压密码")
	fs.StringVar(&config.PasswordEncoding, "pwd-encoding", "utf8", "密码编码 (gbk, utf8)")
	fs.IntVar(&config.Workers, "workers", 4, "并发工作线程数")
	fs.BoolVar(&config.Verbose, "v", false, "详细输出模式")
	fs.BoolVar(&config.Quiet, "q", false, "静默模式，只输出错误")

	fs.Usage = func() {
		fmt.Printf("用法: %s %s [选项] <zip文件>\n", os.Args[0], command)
		fmt.Println("\n选项:")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	if fs.NArg() != 1 {
		fs.Usage()
		return nil, "", fmt.Errorf("需要指定一个zip文件")
	}
	config.ZipPath = fs.Arg(0)

	if config.Workers < 1 {
		return nil, "", fmt.Errorf("工作线程数必须大于0")
	}

	if config.Verbose && config.Quiet {
		return nil, "", fmt.Errorf("verbose 和 quiet 选项不能同时使用")
	}

	return config, command, nil
}

func getPassword(config *UnzipConfig) ([]byte, error) {
	if config.Password != "" {
		if config.PasswordEncoding != "" {
			return internal.GetBytes(config.Password, config.PasswordEncoding)
		}
		return []byte(config.Password), nil
	}
	return nil, nil
}

func processFile(file *zip.File, outputPath string, encoding string, password []byte, wg *sync.WaitGroup, semaphore chan struct{}) {
	defer wg.Done()
	defer func() { <-semaphore }()
	if file.IsEncrypted() {
		if password == nil {
			name, err := internal.ListFile(file, encoding)
			if err != nil {
				name = file.Name
			}
			utils.Errorf("File %s is encrypted but no password provided\n", name)
		}
		file.SetPassword(password)
	}
	outFile, err := internal.ProcessSingleFile(file, outputPath, encoding, password)
	if err != nil {
		utils.Error("解压文件 %s 失败: %v", file.Name, err)
		return
	}
	fileName := path.Base(outFile)
	ok, err := internal.ValidateZip(file, outFile)
	if err != nil {
		utils.Error("校验文件 %s 失败: %v", fileName, err)
		return
	}
	if !ok {
		utils.Error("文件校验不通过 %s", fileName)
	}
	utils.Info("解压完成: %s, validate ok\n", fileName)
}

// 主解压函数
func unzip(config *UnzipConfig) error {
	// 打开zip文件
	reader, err := zip.OpenReader(config.ZipPath)
	if err != nil {
		return utils.Errorf("打开zip文件失败: %v", err)
	}
	defer reader.Close()

	// 创建输出目录
	if err := os.MkdirAll(config.OutputPath, 0755); err != nil {
		return utils.Errorf("创建输出目录失败: %v", err)
	}

	password, err := getPassword(config)
	if err != nil {
		return utils.Errorf("获取密码失败: %v", err)
	}

	// 统计文件数量
	totalFiles := len(reader.File)
	utils.Info("开始解压 %s，共 %d 个文件", config.ZipPath, totalFiles)

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, config.Workers)
	var wg sync.WaitGroup

	// 处理每个文件
	for _, file := range reader.File {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量
		go processFile(file, config.OutputPath, config.FileEncoding, password, &wg, semaphore)
	}

	// 等待所有文件处理完成
	wg.Wait()

	utils.Info("解压完成！输出路径: %s", config.OutputPath)
	return nil
}

func listFiles(config *UnzipConfig) error {
	reader, err := zip.OpenReader(config.ZipPath)
	if err != nil {
		return utils.Errorf("打开zip文件失败: %v", err)
	}
	defer reader.Close()
	for _, file := range reader.File {
		fileName, err := internal.ListFile(file, config.FileEncoding)
		if err != nil {
			utils.Error("列出文件 %s 失败: %v", file.Name, err)
		}
		fmt.Println(fileName)
	}
	return nil
}

func main() {
	// 解析和验证命令行参数
	config, command, err := parseAndValidateFlags()
	if err != nil {
		fmt.Errorf("参数错误: %v\n", err)
		if len(os.Args) < 2 {
			usage()
		}
		os.Exit(1)
	}

	// 初始化日志系统
	var logLevel utils.LogLevel
	if config.Quiet {
		logLevel = utils.Quiet
	} else if config.Verbose {
		logLevel = utils.Verbose
	} else {
		logLevel = utils.Normal
	}
	utils.InitLogger(logLevel)

	// 检查zip文件是否存在
	if _, err := os.Stat(config.ZipPath); os.IsNotExist(err) {
		utils.Error("错误: zip文件不存在: %s", config.ZipPath)
		os.Exit(1)
	}

	switch command {
	case "x":
		// 开始解压
		if err := unzip(config); err != nil {
			utils.Error("解压失败: %v", err)
			os.Exit(1)
		}
	case "t":
		// 列出文件
		if err := listFiles(config); err != nil {
			utils.Error("列出文件失败: %v", err)
			os.Exit(1)
		}
	default:
		utils.Error("未知命令: %s", command)
		usage()
		os.Exit(1)
	}
}
