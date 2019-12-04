package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Level int

var Buf []byte
var IoWriterkey int
var Newlogfile os.File
var Thenumlen int
var APPendCount int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	PANIC
	FATAL
)

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

/*
===================
 log handlers
===================
*/
type Handler interface {
	SetOutput(w io.Writer)
	Output(calldepth int, s string)
	Outputf(format string, v ...interface{})

	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})

	Flags() int
	SetFlags(flag int)
	SetLevel(level Level)

	close()
}

type LogHandler struct {
	lg    *log.Logger
	mu    sync.Mutex
	level Level
}

type ConsoleHander struct {
	LogHandler
}

type FileHandler struct {
	LogHandler
	dir      string
	filename string
	filetime time.Time
	logfile  *os.File
}

/*
===================
 json config
===================
*/
type configs struct {
	Logs logconfig `json:"logs"`
}

type logconfig struct {
	Handle   int    `json:"handle"`
	Dir      string `json:"dir"`
	Filename string `json:"filename"`
	Level    int    `json:"level"`
	Iowriter int    `json:"iowriter"`
	Numlen   int    `json:"numlen"`
}

var Console, _ = NewConsoleHandler()

func NewConsoleHandler() (*ConsoleHander, error) {
	l := log.New(os.Stderr, "", log.LstdFlags)
	return &ConsoleHander{LogHandler: LogHandler{lg: l}}, nil
}

func NewFileHandler(filepath string, filename string) (*FileHandler, error) {
	i := strings.LastIndex(filepath, "\\")
	if -1 == i {
		i = strings.LastIndex(filepath, "/")
		if -1 == i {
			return nil, fmt.Errorf("Error filepath:%v", filepath)
		}
	}
	dir := filepath[:i]
	err := os.MkdirAll(dir, 0711)
	if err != nil {
		return nil, err
	}
	f := &FileHandler{
		dir:      dir,
		filename: filename,
		filetime: time.Time{},
	}
	f.filetime = time.Now()
	logfile, _ := os.OpenFile(f.generateFileName(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	l := log.New(logfile, "", log.LstdFlags)
	f.LogHandler = LogHandler{lg: l}
	f.logfile = logfile

	return &FileHandler{
		LogHandler: LogHandler{lg: l},
		logfile:    logfile,
	}, nil
}

func newHandler(lg logconfig) (Handler, error) {
	if 0 == lg.Handle {
		return NewConsoleHandler()
	} else if 1 == lg.Handle {
		return NewFileHandler(lg.Dir, lg.Filename)
	}

	return nil, fmt.Errorf("Unknown handle:%v", lg.Handle)
}

func NewLogger(name string) error {
	filename := "./config/logs.config"
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	c := configs{}

	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return err
	}

	lgs := c.Logs
	lgshandler := c.Logs.Handle
	IoWriterkey = c.Logs.Iowriter
	Thenumlen = c.Logs.Numlen
	//对handler进行判断
	switch lgshandler {
	case 0, 1:
		{
			handler, err := newHandler(lgs)
			if err != nil {
				Close()
				return err
			}
			handler.SetLevel(Level(lgs.Level))
			handler.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
			logger.handlers = append(logger.handlers, handler)
		}
	case 2:
		{
			//控制台处理
			consolehandler, _ := NewConsoleHandler()
			consolehandler.SetLevel(Level(lgs.Level))
			consolehandler.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
			logger.handlers = append(logger.handlers, consolehandler)
			//文件处理
			filehandlie, err := NewFileHandler(lgs.Dir, lgs.Filename)
			if err != nil {
				Close()
				return err
			}
			filehandlie.SetLevel(Level(lgs.Level))
			filehandlie.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
			logger.handlers = append(logger.handlers, filehandlie)
			Newlogfile = *filehandlie.logfile

		}
	default:
		Close()
		return fmt.Errorf("Unknown log Handle:%v", lgs.Handle)
	}
	//对level进行判断
	if lgs.Level < 0 || lgs.Level > int(FATAL) {
		Close()
		return fmt.Errorf("Unknown log level:%v", lgs.Level)
	}

	return nil
}

//写入到文件
func Onewrite() {

	Newlogfile.WriteString(string(Buf))

	APPendCount = 0 //清空计数器
	Buf = Buf[:0]   //清空字符串

	return
}

//字符串拼接
func Bufappend(s string, v interface{}) {
	APPendCount++ //计数器自增
	time := time.Now().Format("2006-01-_2 3:04:05.000000")
	_, file, line, _ := runtime.Caller(1)
	Buf = append(Buf, time...)
	Buf = append(Buf, ' ') //添加空格
	Buf = append(Buf, file...)
	Buf = append(Buf, ' ') //添加空格
	strline := strconv.Itoa(line)
	Buf = append(Buf, strline...)
	Buf = append(Buf, ' ') //添加空格
	newstring := fmt.Sprintln(s, v)
	Buf = append(Buf, newstring...)
}

func (l *LogHandler) Flags() int {
	return l.lg.Flags()
}

func (l *LogHandler) SetFlags(flag int) {
	l.lg.SetFlags(flag)
}

func (l *LogHandler) SetLevel(level Level) {
	l.level = level
}

func (l *LogHandler) SetOutput(w io.Writer) {
	l.lg.SetOutput(w)
}

func (l *LogHandler) Output(calldepth int, s string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lg.Output(calldepth, s)
}

func (l *LogHandler) Outputf(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lg.Output(4, fmt.Sprintf(format, v...))
}
func (l *LogHandler) Debug(v ...interface{}) {
	if l.level <= DEBUG {
		if IoWriterkey == 1 {
			l.Output(4, fmt.Sprintln("debug", v))
		} else {
			if APPendCount > Thenumlen {
				Onewrite()
			} else {
				Bufappend("debug", v)
			}
		}
	}
}

func (l *LogHandler) Debugf(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.Outputf("debug ["+format+"]", v...)
	}
}

func (l *LogHandler) Info(v ...interface{}) {
	if l.level <= INFO {
		if IoWriterkey == 1 {
			l.Output(4, fmt.Sprintln("info", v))
		} else {
			if APPendCount > Thenumlen {
				Onewrite()
			} else {
				Bufappend("info", v)
			}
		}
	}
}

func (l *LogHandler) Infof(format string, v ...interface{}) {
	if l.level <= INFO {
		l.Outputf("info ["+format+"]", v...)
	}
}

func (l *LogHandler) Warn(v ...interface{}) {
	if l.level <= WARN {
		if IoWriterkey == 1 {
			l.Output(4, fmt.Sprintln("warn", v))
		} else {
			if APPendCount > Thenumlen {
				Onewrite()
			} else {
				Bufappend("warn", v)
			}
		}
	}
}

func (l *LogHandler) Warnf(format string, v ...interface{}) {
	if l.level <= WARN {
		l.Outputf("warn ["+format+"]", v...)
	}
}

func (l *LogHandler) Error(v ...interface{}) {
	if l.level <= ERROR {
		if IoWriterkey == 1 {
			l.Output(4, fmt.Sprintln("error", v))
		} else {
			if APPendCount > Thenumlen {
				Onewrite()
			} else {
				Bufappend("error", v)
			}
		}
	}
}

func (l *LogHandler) Errorf(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.Outputf("error ["+format+"]", v...)
	}
}

func (l *LogHandler) Panic(v ...interface{}) {
	if l.level <= PANIC {
		if IoWriterkey == 1 {
			l.Output(4, fmt.Sprintln("panic", v))
		} else {
			if APPendCount > Thenumlen {
				Onewrite()
			} else {
				Bufappend("panic", v)
			}
		}
	}
}

func (l *LogHandler) Panicf(format string, v ...interface{}) {
	if l.level <= PANIC {
		l.Outputf("panic ["+format+"]", v...)
	}
}

func (l *LogHandler) Fatal(v ...interface{}) {
	if l.level <= FATAL {
		if IoWriterkey == 1 {
			l.Output(4, fmt.Sprintln("fatal", v))
		} else {
			if APPendCount > Thenumlen {
				Onewrite()
			} else {
				Bufappend("fatal", v)
			}
		}
	}
}

func (l *LogHandler) Fatalf(format string, v ...interface{}) {
	if l.level <= FATAL {
		l.Outputf("fatal ["+format+"]", v...)
	}
}

func (l *LogHandler) close() {

}

func (h *FileHandler) close() {
	if h.logfile != nil {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.logfile.Close()
	}
}

func (f *FileHandler) isMustRename() bool {

	t := time.Now()
	if t.Year() > f.filetime.Year() ||
		t.Year() == f.filetime.Year() && t.Month() > f.filetime.Month() ||
		t.Year() == f.filetime.Year() && t.Month() == f.filetime.Month() && t.Day() > f.filetime.Day() {

		//f.newFileData() 下条可删
		f.filetime = time.Now()
		return true
	}
	return false
}

//生成文件名的方法
func (f *FileHandler) generateFileName() string {
	filetime := f.filetime.Format("20060102")
	return fmt.Sprintf("%s/%s.%s.log", f.dir, f.filename, filetime)
}
func (f *FileHandler) rename() {

	newpath := f.generateFileName()
	if isExist(newpath) {
		os.Remove(newpath)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.logfile != nil {
		f.logfile.Close()
	}

	f.logfile, _ = os.OpenFile(newpath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	f.lg.SetOutput(f.logfile)
}
func (f *FileHandler) fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if f.isMustRename() {
		f.rename()
	}
}

/*
===================
 logger
===================
*/
type _Logger struct {
	handlers []Handler
	mu       sync.Mutex
}

var logger = &_Logger{
	handlers: []Handler{},
}

func Debug(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Debug(v...)
	}
}

func Debugf(format string, v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Debugf(format, v...)
	}
}

func Info(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Info(v...)
	}
}

func Infof(format string, v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Infof(format, v...)
	}
}

func Warn(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Warn(v...)
	}
}

func Warnf(format string, v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Warnf(format, v...)
	}
}

func Error(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Error(v...)
	}
}

func Errorf(format string, v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Errorf(format, v...)
	}
}

func Panic(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Panic(v...)
	}
	panic(fmt.Sprint(v...))
}

func Panicf(format string, v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Panicf(format, v...)
	}
	panic(fmt.Sprintf(format, v...))
}

func Fatal(v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Fatal(v...)
	}
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	for i := range logger.handlers {
		logger.handlers[i].Fatalf(format, v...)
	}
	os.Exit(1)
}

func Close() {
	for i := range logger.handlers {
		logger.handlers[i].close()
	}
	logger.handlers = logger.handlers[0:0]
}
