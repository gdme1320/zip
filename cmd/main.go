package main

import (
	"flag"
	"fmt"
	"os"
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
	FileEncoding     string // 文件编码 (gbk, utf8)
	Password         string // 密码
	PasswordEncoding string // 密码编码 (gbk, utf8)
	Workers          int    // 并发工作线程数
	Verbose          bool   // 详细输出
	Quiet            bool   // 静默输出
}

// 解析命令行参数
func parseFlags() *UnzipConfig {
	config := &UnzipConfig{}

	flag.StringVar(&config.ZipPath, "zip", "", "zip文件路径 (必需)")
	flag.StringVar(&config.OutputPath, "output", "", "解压输出路径 (必需)")
	flag.StringVar(&config.FileEncoding, "encoding", "utf8", "文件名编码 (gbk, utf8)")
	flag.StringVar(&config.Password, "password", "", "解压密码")
	flag.StringVar(&config.PasswordEncoding, "pwd-encoding", "utf8", "密码编码 (gbk, utf8)")
	flag.IntVar(&config.Workers, "workers", 4, "并发工作线程数")
	flag.BoolVar(&config.Verbose, "verbose", false, "详细输出模式")
	flag.BoolVar(&config.Quiet, "quiet", false, "静默模式，只输出错误")

	flag.Usage = func() {
		fmt.Printf("用法: %s [选项]\n", os.Args[0])
		fmt.Printf("\n选项:\n")
		flag.PrintDefaults()
		fmt.Printf("\n示例:\n")
		fmt.Printf("  %s -zip archive.zip -output ./extracted -password 123456\n", os.Args[0])
		fmt.Printf("  %s -zip archive.zip -output ./extracted -encoding utf8 -password 密码 -pwd-encoding gbk\n", os.Args[0])
		fmt.Printf("  %s -zip archive.zip -output ./extracted -verbose\n", os.Args[0])
		fmt.Printf("  %s -zip archive.zip -output ./extracted -quiet\n", os.Args[0])
	}

	flag.Parse()

	return config
}

// 验证配置参数
func validateConfig(config *UnzipConfig) error {
	if config.ZipPath == "" {
		return utils.Errorf("必须指定zip文件路径 (-zip)")
	}

	if config.OutputPath == "" {
		return utils.Errorf("必须指定输出路径 (-output)")
	}

	if config.FileEncoding != "gbk" && config.FileEncoding != "utf8" {
		return utils.Errorf("文件编码必须是 gbk 或 utf8")
	}

	if config.PasswordEncoding != "gbk" && config.PasswordEncoding != "utf8" {
		return utils.Errorf("密码编码必须是 gbk 或 utf8")
	}

	if config.Workers < 1 {
		return utils.Errorf("工作线程数必须大于0")
	}

	// 检查 verbose 和 quiet 的互斥性
	if config.Verbose && config.Quiet {
		return utils.Errorf("verbose 和 quiet 选项不能同时使用")
	}

	return nil
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
	if password != nil {
		utils.Debug("使用密码解压文件: %s", file.Name)
		file.SetPassword(password)
	}
	internal.ProcessSingleFile(file, outputPath, encoding, password)
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

func main() {
	// 解析命令行参数
	config := parseFlags()

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

	// 验证配置
	if err := validateConfig(config); err != nil {
		utils.Error("配置错误: %v", err)
		flag.Usage()
		os.Exit(1)
	}

	// 检查zip文件是否存在
	if _, err := os.Stat(config.ZipPath); os.IsNotExist(err) {
		utils.Error("错误: zip文件不存在: %s", config.ZipPath)
		os.Exit(1)
	}

	// 开始解压
	if err := unzip(config); err != nil {
		utils.Error("解压失败: %v", err)
		os.Exit(1)
	}
}
