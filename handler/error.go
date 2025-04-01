package handler

import (
	"fmt"
	"log"
	"os"
	"time"
	"log-detect/global"
	"github.com/gin-gonic/gin"
)

func WriteErrorLog(c *gin.Context, msg string) {
	fileName := fmt.Sprintf("%s/apiError-%s.log", global.EnvConfig.Path.Log_record, time.Now().Format("200601"))    

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening file: %v", err)
	}

	defer f.Close()
	log.SetOutput(f)

	logmsg := fmt.Sprintf("Method=\"%s\" URL=\"%s\" msg=\"%s\"", c.Request.Method, c.Request.RequestURI, msg)
	log.Println(logmsg)
}