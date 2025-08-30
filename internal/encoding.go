package internal

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var (
	encodingMap = map[string]encoding.Encoding{
		"gbk":     simplifiedchinese.GBK,
		"windows": simplifiedchinese.GB18030,
	}
)

func charDet(data []byte, hint string) (string, encoding.Encoding) {
	if utf8.Valid(data) {
		return string(data), unicode.UTF8
	}
	if hint != "" {
		hint = strings.ToLower(hint)
		if h, ok := encodingMap[hint]; ok {
			if s, err := decodeWithEncoding(data, h.NewDecoder()); err == nil {
				return s, h
			}
		}
	}

	for k, e := range encodingMap {
		if k == hint {
			continue
		}
		if s, err := decodeWithEncoding(data, e.NewDecoder()); err == nil {
			return s, e
		}
	}

	return "", nil
}

func createDecoder(charset string) *encoding.Decoder {
	charset = strings.ToLower(charset)
	if d, ok := encodingMap[charset]; ok {
		return d.NewDecoder()
	}
	return nil
}

// 以另一个编码读取字符串
//
// params:
//   - data: 原始数据
//   - d: 解码器, 例如 simplifiedchinese.GBK.NewDecoder()
//
// return: 解码后的数据, 错误
func decodeWithEncoding(data []byte, d *encoding.Decoder) (string, error) {
	reader := transform.NewReader(bytes.NewReader(data), d)
	b, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// 获取字符串charset编码的字节数组
func GetBytes(s string, charset string) ([]byte, error) {
	charset = strings.ToLower(charset)
	if charset == "utf8" || charset == "utf-8" {
		return []byte(s), nil
	}
	if d, ok := encodingMap[charset]; ok {
		enc := d.NewEncoder()
		return enc.Bytes([]byte(s))
	} else {
		return nil, errors.New("invalid charset")
	}
}
