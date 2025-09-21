package internal

import zip "github.com/gdme1320/zip/pkg"

type ZipFileProcessArgs interface {
	GetZipFile() *zip.File
	GetOutputPath() string
	GetEncoding() string
	GetPassword() []byte

	// Called after the file name is decoded using the correct encoding.
	// Return false to skip processing this file.
	OnFileName(name string) bool
}
