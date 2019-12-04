package main

import (
	"../logger"
	"fmt"
	"log"
	"time"
)

//"logs":{"handle":2, "dir":"./log/", "filename":"TheLogin", "level":0, "iowriter":0, "numlen": 3}
//handle 0 控制台输出 1 文件输出 2控制台和文件输出
//dir  取最后的/或\\为前路径 创建log文件
//filename  设置文件名
//level  只会输入大于这个等级的log
//iowriter  0 调用字符串拼接，利用numlen进行判断  1调用系统log逐条输出
//numlen  设置多少条数据写入到文件，低于numlen值是运行程序将不会写入数据丢失，只有iowriter设置了此条才会进行判断

func main() {
	err := logger.NewLogger("default")
	if err != nil {
		log.Fatal(err)
	}

	defer logger.Close()

	fmt.Println(time.Now().Unix())
	for i := 0; i < 3; i++ {
		logger.Debug("something1DebugDebugDebug", "debug")
		//logger.Debugf("something1DebugDebugDebug", "debug")
		logger.Info("somethingInfoInfoInfoInfoInfo:", i)
		//logger.Infof("This is info:%s-%d", "go", 11)
		logger.Warn("somethingWarnWarnWarnWarnWarn")
		logger.Error("somethingErrorErrorErrorError")
		//logger.Panic("PanicPanicPanicPanicPanicPanic")
		time.Sleep(1 * time.Microsecond)
	/*	if 200 == i {
			logger.Fatal("fatalFatalFatalFatalFatalFatal")
		}*/
	}
	fmt.Println(time.Now().Unix())
}
