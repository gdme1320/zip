package utils

import (
	"os"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	// 测试 Quiet 级别
	t.Run("Quiet Level", func(t *testing.T) {
		// 捕获 stdout 和 stderr
		stdout := captureOutput(os.Stdout, func() {
			stderr := captureOutput(os.Stderr, func() {
				InitLogger(Quiet)

				Info("这条信息不应该显示")
				Debug("这条调试信息不应该显示")
				Error("这条错误信息应该显示")
			})

			// 验证 stderr 包含错误信息
			if !strings.Contains(stderr, "这条错误信息应该显示") {
				t.Errorf("Quiet 级别应该输出错误信息到 stderr")
			}
		})

		// 验证 stdout 不包含信息
		if strings.Contains(stdout, "这条信息不应该显示") {
			t.Errorf("Quiet 级别不应该输出信息到 stdout")
		}
		if strings.Contains(stdout, "这条调试信息不应该显示") {
			t.Errorf("Quiet 级别不应该输出调试信息到 stdout")
		}
	})

	// 测试 Normal 级别
	t.Run("Normal Level", func(t *testing.T) {
		stdout := captureOutput(os.Stdout, func() {
			stderr := captureOutput(os.Stderr, func() {
				InitLogger(Normal)

				Info("这条信息应该显示")
				Debug("这条调试信息不应该显示")
				Error("这条错误信息应该显示")
			})

			// 验证 stderr 包含错误信息
			if !strings.Contains(stderr, "这条错误信息应该显示") {
				t.Errorf("Normal 级别应该输出错误信息到 stderr")
			}
		})

		// 验证 stdout 包含信息但不包含调试信息
		if !strings.Contains(stdout, "这条信息应该显示") {
			t.Errorf("Normal 级别应该输出信息到 stdout")
		}
		if strings.Contains(stdout, "这条调试信息不应该显示") {
			t.Errorf("Normal 级别不应该输出调试信息到 stdout")
		}
	})

	// 测试 Verbose 级别
	t.Run("Verbose Level", func(t *testing.T) {
		stdout := captureOutput(os.Stdout, func() {
			stderr := captureOutput(os.Stderr, func() {
				InitLogger(Verbose)

				Info("这条信息应该显示")
				Debug("这条调试信息应该显示")
				Error("这条错误信息应该显示")
			})

			// 验证 stderr 包含错误信息
			if !strings.Contains(stderr, "这条错误信息应该显示") {
				t.Errorf("Verbose 级别应该输出错误信息到 stderr")
			}
		})

		// 验证 stdout 包含所有信息
		if !strings.Contains(stdout, "这条信息应该显示") {
			t.Errorf("Verbose 级别应该输出信息到 stdout")
		}
		if !strings.Contains(stdout, "这条调试信息应该显示") {
			t.Errorf("Verbose 级别应该输出调试信息到 stdout")
		}
	})
}

func TestErrorf(t *testing.T) {
	InitLogger(Normal)

	err := Errorf("测试错误: %s", "错误详情")
	if err == nil {
		t.Errorf("Errorf 应该返回错误")
	}

	if !strings.Contains(err.Error(), "测试错误: 错误详情") {
		t.Errorf("Errorf 返回的错误消息不正确: %s", err.Error())
	}
}

// captureOutput 捕获输出到指定 writer 的内容
func captureOutput(w *os.File, fn func()) string {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 保存原始的 writer
	original := w
	defer func() { w = original }()

	// 重定向到临时文件
	w = tmpFile

	// 执行函数
	fn()

	// 读取捕获的输出
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		panic(err)
	}

	return string(content)
}
