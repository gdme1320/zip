package internal

import "testing"

func TestCharDet(t *testing.T) {
	s := "测试"
	gbkData := []byte{0xb2, 0xe2, 0xca, 0xd4}
	// utf8Data := []byte(s)

	// 测试GBK编码检测
	detectedStr, enc := charDet(gbkData, "gbk")
	if detectedStr != s {
		t.Errorf("Expected %s, got %s", s, detectedStr)
	}
	if enc != encodingMap["gbk"] {
		t.Errorf("Expected GBK encoding, got %v", enc)
	}

	// 测试GBK编码检测
	detectedStr, enc = charDet(gbkData, "")
	if detectedStr != s {
		t.Errorf("Expected %s, got %s", s, detectedStr)
	}
	if enc != encodingMap["gbk"] {
		t.Errorf("Expected GBK encoding, got %v", enc)
	}
}
