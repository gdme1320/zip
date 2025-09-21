package utils

import "fmt"

type OutputTopic int

const (
	ZipValidator OutputTopic = iota
)

func Stdout(topic OutputTopic, format string, v ...interface{}) {
	fmt.Printf(format, v...)
}
