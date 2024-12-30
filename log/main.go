package logger

import (
	"log"

	"github.com/fatih/color"
)

type LoggerStruct struct {
	red   *color.Color
	blue  *color.Color
	white *color.Color
	green *color.Color
}

func MakeLogger() *LoggerStruct {
	return &LoggerStruct{
		red:   color.New(color.FgRed),
		blue:  color.New(color.FgBlue),
		green: color.New(color.FgGreen),
		white: color.New(color.FgWhite),
	}
}

func (logger *LoggerStruct) Error(message string) {
	log.Printf("[%s] %s", logger.red.Sprintf("Error"), message)
}

func (logger *LoggerStruct) Info(message string) {
	log.Printf("[%s] %s", logger.blue.Sprintf("Info"), message)
}

func (logger *LoggerStruct) Success(message string) {
	log.Printf("[%s] %s", logger.green.Sprintf("Success"), message)
}

func (logger *LoggerStruct) Normal(message string) {
	log.Printf("[%s] %s", logger.white.Sprintf("Log"), message)
}
