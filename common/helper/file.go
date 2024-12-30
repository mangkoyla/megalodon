package helper

import (
	"os"
)

func ReadFileAsString(path string) (string, error) {
	bufByte, err := os.ReadFile(path)
	return string(bufByte), err
}
