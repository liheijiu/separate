package conf

//结构体
type logconf struct {
	logLevel    string `ini:"logLevel"`
	filePath    string `ini:"filePath"`
	fileName    string `ini:"fileName"`
	maxFileSize int64  `ini:"maxFileSize"`
}

//方法
func LogConf() {
	
}
