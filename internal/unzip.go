package internal

import (
	"errors"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"

	"github.com/gdme1320/zip/internal/utils"
	zip "github.com/gdme1320/zip/pkg"
)

func getFileName(name string, encoding string) (string, error) {
	if encoding == "utf8" || encoding == "utf-8" {
		return name, nil
	}
	if encoding == "" {
		n, e := charDet([]byte(name), "")
		if e != nil {
			return n, nil
		}
		return "", errors.New("Unable to detect encoding")
	}
	d := createDecoder(encoding)
	if d != nil {
		n, err := decodeWithEncoding([]byte(name), d)
		if err != nil {
			return "", err
		}
		return n, nil
	} else {
		return "", errors.New("invalid encoding")
	}
}

func ListFile(zipFile *zip.File, encoding string) (string, error) {
	// 获取文件名
	fileName, err := getFileName(zipFile.Name, encoding)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

// 处理zip中单个文件解压
func ProcessSingleFile(zipFile *zip.File, outputPath string, encoding string, password []byte) (string, error) {
	// 获取文件名
	fileName, err := getFileName(zipFile.Name, encoding)
	if err != nil {
		return "", err
	}

	// 构建完整输出路径
	fullPath := filepath.Join(outputPath, fileName)

	// 如果是目录，则创建目录
	if zipFile.FileInfo().IsDir() {
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			utils.Error("创建目录失败 %s: %v", fullPath, err)
			return "", err
		}
		return fullPath, nil
	}

	// 确保父目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.Error("创建目录失败 %s: %v", dir, err)
		return "", err
	}

	// 打开zip文件
	rc, err := zipFile.Open()
	if err != nil {
		utils.Error("打开zip文件失败 %s: %v", fileName, err)
		return "", err
	}
	defer rc.Close()

	// 创建输出文件
	outFile, err := os.Create(fullPath)
	if err != nil {
		utils.Error("创建文件失败 %s: %v", fullPath, err)
		return "", err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, rc); err != nil {
		utils.Error("复制文件失败 %s: %v", fileName, err)
		return "", err
	}

	// 设置文件权限
	if err := outFile.Chmod(zipFile.Mode()); err != nil {
		utils.Error("设置文件权限失败 %s: %v", fullPath, err)
	}
	return fullPath, nil
}

func ValidateZip(zipFile *zip.File, fullPath string) (bool, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		utils.Error("获取文件信息失败 %s: %v", fullPath, err)
		return false, err
	}
	if info.Size() != int64(zipFile.UncompressedSize64) {
		return false, nil
	}

	hasher := crc32.NewIEEE()
	outFile, _ := os.Open(fullPath)
	if _, err := io.Copy(hasher, outFile); err != nil {
		return false, errors.New("CRC计算失败")
	}

	calculatedCRC32 := hasher.Sum32()
	if calculatedCRC32 != zipFile.CRC32 {
		return false, nil
	}
	return true, nil
}
