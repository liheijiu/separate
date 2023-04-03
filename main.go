package main

import (
	"Ninesongs/logfile/logger"
)

//测试我们自己写的日志
func main() {

	log := logger.NewFileLogger("info", "./log", "logagent.log", 10*1024*1024)
	for {
		log.Debug("已开启Debug日志")
		log.Trace("已开启Trace日志")
		log.Info("已开启Info日志")
		log.Warning("已开启Warning日志")
		id := 100
		name := "理想"
		log.Error("已开启Error日志 %d  %s", id, name)
		log.Fatal("已开启Fatal日志")
		//time.Sleep(time.Second * 2)
	}
}
