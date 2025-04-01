package log

import (
    "github.com/natefinch/lumberjack"
	"log-detect/global"
	"time"
    "fmt"
	"log"
	"os"

)


func Logrecord(title, msg string) string {
	// setting log rotate
	logfile := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/LogDetect-%s.log", global.EnvConfig.Path.Log_record, time.Now().Format("200601")),
		MaxSize:    50,    // MB
		MaxBackups: 3,    // 保留的歷史檔案數量
		MaxAge:     30,   // 保留的歷史檔案天數
		Compress:   true, // 是否壓縮歷史檔案
	}

	// 使用 lumberjack.Logger 作為日誌的輸出
	log.SetOutput(logfile)

	// 寫入日誌
	logger := log.New(logfile, title+" ", log.LstdFlags)
	logger.Println(msg)

	return msg
}



func Logrecord_no_rotate(title,msg string) string{
    fileName := fmt.Sprintf("%s/LogDetect-%s.log", global.EnvConfig.Path.Log_record, time.Now().Format("200601"))    
    // open file and create if non-existent
    file, err := os.OpenFile( fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    logger := log.New(file, title + " ", log.LstdFlags)
    logger.Println(msg)
	return msg

}