package log

import (
	"fmt"
	. "github.com/logrusorgru/aurora"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var level int = 3

const (
	TRACE = "trace"
	DEBUG = "debug"
	INFO  = "info"
	WARN  = "warn"
	FATAL = "fatal"
)

var levels = map[string]int{
	"trace": 1,
	"debug": 2,
	"info":  3,
	"warn":  4,
	"fatal": 5,
}

func SetLevel(a string) {
	level = levels[a]
}

func Trace(a ...interface{}) {
	if level > 1 {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		split := strings.SplitN(file, "blueming", 2)
		if len(split) == 2 {
			file = split[1][1:] + ":" + strconv.Itoa(line)
		}
	}
	caller := file
	pr := make([]interface{}, 0)
	pr = []interface{}{"[" + BrightBlack("TRAC").String() + "]", time.Now().Format("01-02 15:04:05"), BrightCyan(caller).String()}
	pr = append(pr, a...)
	fmt.Println(pr...)
}

func Debug(a ...interface{}) {
	if level > 2 {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		split := strings.SplitN(file, "blueming", 2)
		if len(split) == 2 {
			file = split[1][1:] + ":" + strconv.Itoa(line)
		}
	}
	caller := file
	pr := make([]interface{}, 0)
	pr = []interface{}{"[" + BrightMagenta("DBUG").String() + "]", time.Now().Format("01-02 15:04:05"), BrightCyan(caller).String()}
	pr = append(pr, a...)
	fmt.Println(pr...)
}

func Info(a ...interface{}) {
	if level > 3 {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		split := strings.SplitN(file, "blueming", 2)
		if len(split) == 2 {
			file = split[1][1:] + ":" + strconv.Itoa(line)
		}
	}
	caller := file
	pr := make([]interface{}, 0)
	pr = []interface{}{"[" + BrightBlue("INFO").String() + "]", time.Now().Format("01-02 15:04:05"), BrightCyan(caller).String()}
	pr = append(pr, a...)
	fmt.Println(pr...)
}

func Warn(a ...interface{}) {
	if level > 4 {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		split := strings.SplitN(file, "blueming", 2)
		if len(split) == 2 {
			file = split[1][1:] + ":" + strconv.Itoa(line)
		}
	}
	caller := file
	pr := make([]interface{}, 0)
	pr = []interface{}{"[" + BrightYellow("WARN").String() + "]", time.Now().Format("01-02 15:04:05"), BrightCyan(caller).String()}
	pr = append(pr, a...)
	fmt.Println(pr...)
}

func Fatal(a ...interface{}) {
	if level > 5 {
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if ok {
		split := strings.SplitN(file, "blueming", 2)
		if len(split) == 2 {
			file = split[1][1:] + ":" + strconv.Itoa(line)
		}
	}
	caller := file
	pr := make([]interface{}, 0)
	pr = []interface{}{"[" + BrightRed("FATA").String() + "]", time.Now().Format("01-02 15:04:05"), BrightCyan(caller).String()}
	pr = append(pr, a...)
	fmt.Println(pr...)
	os.Exit(1)
}
