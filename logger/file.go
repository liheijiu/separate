package logger

import (
	"fmt"
	"os"
	"path"
	"time"
)

//往文件里面写日志
type FlieLogger struct {
	Level       LogLevel
	filePath    string   //保存路径
	fileName    string   //文件名
	errFileObj  *os.File //错误日志
	fileObj     *os.File //创建文件
	maxFileSize int64    //文件大小
	logChan     chan *logMsgChan
}
type logMsgChan struct {
	level     LogLevel //日志级别
	msg       string   //日志信息
	funcName  string   //行号函数名
	fileName  string   //文件名
	line      int      //行号
	timestamp string   //时间
}

const MaxSize = 5000

func NewFileLogger(levelStr, fp, fn string, maxSize int64) *FlieLogger {
	logLevel, err := parseLogLevel(levelStr)
	if err != nil {
		panic(err)
	}
	fl := &FlieLogger{
		Level:       logLevel,
		filePath:    fp,
		fileName:    fn,
		maxFileSize: maxSize,
		logChan:     make(chan *logMsgChan, MaxSize),
	}
	err = fl.initFile() //按照路径和文件名将文件打开
	if err != nil {
		return nil
	}
	return fl
}

//
func (f *FlieLogger) initFile() (err error) {
	fullFileName := path.Join(f.filePath, f.fileName)
	fileObj, err := os.OpenFile(fullFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("open log failed, err:%v\n", err)
		return
	}
	errFileObj, err := os.OpenFile(fullFileName+".err", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("open err log failed, err:%v\n", err)
		return
	}
	//日志文件已打开完成
	f.fileObj = fileObj
	f.errFileObj = errFileObj

	//开启一个后台的G

	for i := 0; i < 2; i++ {
		go f.writeLogBackground()
	}
	return nil
}

//日志开关
func (f *FlieLogger) enable(loglevel LogLevel) bool {
	return loglevel >= f.Level
}

//
func (f *FlieLogger) checkSize(file *os.File) bool {
	FileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info failed,err:%v\n", err)
		return false
	}
	//如果当前文件大小大于等于设置的大小，就切割
	return FileInfo.Size() > f.maxFileSize
}

//
func (f *FlieLogger) splitFile(file *os.File) (*os.File, error) {
	nowStr := time.Now().Format("200601021501050000")
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file info failed, err:%v\n", err)
		return nil, err
	}

	logName := path.Join(f.filePath, fileInfo.Name())
	newLogName := fmt.Sprintf("%s.%s", logName, nowStr)
	//需要切割
	err = file.Close()
	if err != nil {
		return nil, err
	}

	err = os.Rename(logName, newLogName)
	if err != nil {
		return nil, err
	}
	fileObj, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("open new log file failed, err:%v\n", err)
		return nil, err
	}
	return fileObj, nil
}

//后台写日志
func (f *FlieLogger) writeLogBackground() {

	for {

		if f.checkSize(f.fileObj) {
			newFile, err := f.splitFile(f.fileObj)
			if err != nil {
				return
			}
			f.fileObj = newFile
		}
		select {
		case logTmp := <-f.logChan:
			logInfo := fmt.Sprintf("[%s][%s] [%s %s %d]  %s\n", logTmp.timestamp, getLogString(logTmp.level), logTmp.funcName, logTmp.fileName, logTmp.line, logTmp.msg)
			fmt.Fprintf(f.fileObj, logInfo)
			if logTmp.level >= ERROR {
				if f.checkSize(f.errFileObj) {
					newFile, err := f.splitFile(f.errFileObj)
					if err != nil {
						return
					}
					f.errFileObj = newFile
				}
				fmt.Fprintf(f.errFileObj, logInfo)
			}
		default:
			time.Sleep(time.Millisecond * 500)
		}
	}
}

//日志接收
func (f *FlieLogger) log(lv LogLevel, format string, a ...interface{}) {
	if f.enable(lv) {
		msg := fmt.Sprintf(format, a...)
		now := time.Now()
		funcName, fileName, lineNo := getInfo(3)

		//把日志发送到通道中
		logTmp := &logMsgChan{
			level:     lv,
			msg:       msg,
			funcName:  funcName,
			fileName:  fileName,
			timestamp: now.Format("2006-01-02  15:04:06"),
			line:      lineNo,
		}
		select {
		case f.logChan <- logTmp:
		default:

		}
	}
}

//Debug
func (f *FlieLogger) Debug(format string, a ...interface{}) {
	f.log(DEBUG, format, a...)
}

//Trace
func (f *FlieLogger) Trace(format string, a ...interface{}) {
	f.log(TRACE, format, a...)
}

//Info
func (f *FlieLogger) Info(format string, a ...interface{}) {
	f.log(INFO, format, a...)
}

//Warning
func (f *FlieLogger) Warning(format string, a ...interface{}) {
	f.log(WARNING, format, a...)
}

//Error
func (f *FlieLogger) Error(format string, a ...interface{}) {
	f.log(ERROR, format, a...)
}

//Fatal
func (f *FlieLogger) Fatal(format string, a ...interface{}) {
	f.log(FATAL, format, a...)
}

//关闭文件
func (f *FlieLogger) Close() {
	f.fileObj.Close()
	f.errFileObj.Close()
}
