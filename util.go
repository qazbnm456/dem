package main

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

func Success(format string, a ...interface{}) {
	color.Green(format, a...)
}

func Debug(format string, a ...interface{}) {
	if !debug {
		return
	}
	color.Cyan(format, a...)
}

func Info(format string, a ...interface{}) {
	color.White(format, a...)
}

func Warn(format string, a ...interface{}) {
	color.Yellow(format, a...)
}

func Error(format string, err error, a ...interface{}) {
	color.Set(color.FgRed)
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	color.Unset()
}

func die(format string, err error, exitCode int, a ...interface{}) {
	Error(format, err, a...)
	os.Exit(exitCode)
}
