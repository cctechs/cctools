// 自定义日志系统
// 支持以下功能：
// 1.按小时进行日志记录
// 2.按日志级别进行日志记录

package cclog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"strconv"
	"time"
)

// 定义日志相关的级别
const (
	// content of emnu Level ,level of log
	NULL = 1 << iota
	TRACE
	DEBUG
	INFO
	WARN
	NOTICE
	ERROR
	CRIT
)

// 定义日志相关的设备类型
const (
	FACILITITY_APP = 1 << iota
	FACILITITY_DB
	FACILITITY_GAME
	FACILITITY_USER
)

type Outputer int

const (
	STD = iota
	FILE
)

type logger struct {
	logFd         *os.File // 文件描述符
	starLev       int      // 日志记录的等级
	buf           []byte   // 缓冲区
	path          string   // 路径
	baseName      string   // 通用名称
	logName       string   // 日志名称
	debugOutputer Outputer // 调试输出
	debugSwitch   bool     // 调试模式切换
	callDepth     int      // 日志文件记录深度
	fullPath      string   // 文件全路径
	lastHour      int      // 上一次记录的小时
	IsShowConsole bool     // 是否控制台显示
	logChan       chan string
}

var gLogger *logger

func init(){
	gLogger = newLogger("./logs", "", "Log4Golang", DEBUG)
	gLogger.setCallDepth(3)
	gLogger.start()
}

// 创建日志
func newLogger(path, baseName, logName string, level int) *logger {
	var err error
	logger := &logger{path: path, baseName: baseName, logName: logName, starLev: level}

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		panic(err)
	}

	logger.logFd = logger.getLoggerFd()

	logger.debugSwitch = true
	logger.debugOutputer = STD
	logger.callDepth = 3
	logger.logChan = make(chan string, 8096)
	return logger
}

func (this *logger) getLoggerFd() *os.File {
	var err error
	path := strings.TrimSuffix(this.path, "/")
	flag := os.O_WRONLY | os.O_APPEND | os.O_CREATE
	this.fullPath = path + "/" + this.baseName
	now := time.Now()
	this.fullPath += fmt.Sprintf("%04d%02d%02d%02d.log", now.Year(), now.Month(), now.Day(), now.Hour())
	this.logFd, err = os.OpenFile(this.fullPath, flag, 0666)
	if err != nil {
		panic(err)
	}
	return this.logFd
}

func (this *logger) start() {
	go this.autoWrite()
}

// 每次最多写入2000条数据
func (this *logger) autoWrite() {
	for {
		str := <-this.logChan
		this.writeLog(bytes.NewBufferString(str).Bytes())
	}
}

func (this *logger) writeLog(buf []byte) {
	now := time.Now()
	if now.Hour() != this.lastHour {
		//先将当前文件关闭
		err := this.logFd.Close()
		if err != nil {
			str := fmt.Sprintf("关闭日志文件[%v]失败[err:%v]", this.fullPath, err.Error())
			fmt.Println(str)
		}
		//获取下一个索引的文件
		this.logFd = this.getLoggerFd()
	}
	_, err := this.logFd.Write(buf)
	if err != nil {
		fmt.Printf("写入错误")
	}
}

func (this *logger) output(fd io.Writer, level, prefix string, format string, v ...interface{}) (err error) {
	var msg string
	if format == "" {
		msg = fmt.Sprintln(v...)
	} else {
		msg = fmt.Sprintf(format, v...)
	}

	this.buf = this.buf[:0]

	this.buf = append(this.buf, "["+this.logName+"]"...)
	this.buf = append(this.buf, level...)
	this.buf = append(this.buf, prefix...)

	this.buf = append(this.buf, ":"+msg...)
	if len(msg) > 0 && msg[len(msg)-1] != '\n' {
		this.buf = append(this.buf, '\n')
	}

	_, err = fd.Write(this.buf)

	return nil
}

func (l *logger) setCallDepth(d int) {
	l.callDepth = d
}

func (l *logger) openDebug() {
	l.debugSwitch = true
}

func (l *logger) getFileLine() string {
	_, file, line, ok := runtime.Caller(l.callDepth)
	if !ok {
		file = "???"
		line = 0
	}
	return l.getFileName(file) + ":" + itoa(line, -1)
}

func( l * logger) getFileName(path string) string{
	strArr := strings.Split(path, "/")
	nLen := len(strArr)
	if nLen > 0{
		return strArr[nLen - 1]
	}
	return path
}

/**
* Change from Golang's log.go
* Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
* Knows the buffer has capacity.
 */
func itoa(i int, wid int) string {
	var u uint = uint(i)
	if u == 0 && wid <= 1 {
		return "0"
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}
	return string(b[bp:])
}

func (l *logger) getTime() string {
	// Time is yyyy-mm-dd hh:mm:ss.microsec
	var buf []byte
	t := time.Now()
	year, month, day := t.Date()
	buf = append(buf, itoa(int(year), 4)+"-"...)
	buf = append(buf, itoa(int(month), 2)+"-"...)
	buf = append(buf, itoa(int(day), 2)+" "...)

	hour, min, sec := t.Clock()
	buf = append(buf, itoa(hour, 2)+":"...)
	buf = append(buf, itoa(min, 2)+":"...)
	buf = append(buf, itoa(sec, 2)...)

	buf = append(buf, '.')
	buf = append(buf, itoa(t.Nanosecond()/1e3, 6)...)

	return string(buf[:])
}

func (l *logger) closeDebug() {
	l.debugSwitch = false
}

func (l *logger) setDebugOutput(o Outputer) {
	l.debugOutputer = o
}

func LogTrace(format string, v ...interface{}) error  {
	return gLogger.addlog(TRACE, FACILITITY_APP, format, v...)
}

func LogDebug(format string, v ...interface{}) error {
	return gLogger.addlog(DEBUG, FACILITITY_APP, format, v...)
}

func LogInfo(format string, v ...interface{}) error {
	return gLogger.addlog(INFO, FACILITITY_APP, format, v...)
}

func LogWarn(format string, v ...interface{}) error {
	return gLogger.addlog(WARN, FACILITITY_APP, format, v...)
}

func LogNotice(format string, v ...interface{}) error {
	return gLogger.addlog(NOTICE, FACILITITY_APP, format, v...)
}

func LogError(format string, v ...interface{}) error {
	return gLogger.addlog(ERROR, FACILITITY_APP, format, v...)
}

func LogCrit(format string, v ...interface{}) error {
	return gLogger.addlog(CRIT, FACILITITY_APP, format, v...)
}

func (this *logger) getLogLvlStr(logType int) string {
	var str string = ""
	switch logType {
	case TRACE:
		str = "[TRACE]"
	case DEBUG:
		str = "[DEBUG]"
	case INFO:
		str = "[ INFO]"
	case WARN:
		str = "[ WARN]"
	case ERROR:
		str = "[ERROR]"
	case CRIT:
		str = "[ CRIT]"
	default:
		str = "[DEBUG]"
	}
	return str
}

func(this*logger) getFacilityStr(facility int) string{
	var str string = ""
	switch facility {
	case FACILITITY_APP:
		str = "[ APP]"
	case FACILITITY_DB:
		str = "[ DB]"
	case FACILITITY_GAME:
		str = "[GAME]"
	case FACILITITY_USER:
		str = "[USER]"
	default:
		break
	}
	return str
}


// 获取协程ID
func (this* logger) GetGoID() int32{
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return int32(id)
}

func (this *logger) addlog(logLev int, facility int, format string, v ...interface{}) error {
	strLevel := this.getLogLvlStr(logLev)
	strFacility := this.getFacilityStr(facility)
	strGoID := fmt.Sprintf("[%05d]", this.GetGoID())

	// 直接打印
	strTime := this.getTime() + " "
	strFile := "[" + this.getFileLine() + "]"

	var msg string
	if format == "" {
		msg = fmt.Sprint(v...)
	} else {
		msg = fmt.Sprintf(format, v...)
	}

	//[时间][级别][设备][协程ID][文件]日志内容
	strLog := fmt.Sprintf("%s%s%s%s%s%s", strTime, strLevel, strFacility, strGoID, strFile, msg)
	fmt.Println(strLog)

	if logLev < this.starLev {
		return nil
	}
	strLog += "\n"
	this.logChan <- strLog

	return nil
}

func (this *logger) popLog() *string {
	str := <-this.logChan
	return &str
}
