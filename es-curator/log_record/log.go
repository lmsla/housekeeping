package log_record

import (
	"log"
	"os"
	"es-curator/global"
	//"xdr/utils"
	//"time"
)

func Logrecord(title,msg string) string{

    // open file and create if non-existent
    file, err := os.OpenFile( global.EnvConfig.INFORMATION.Logdir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    logger := log.New(file, title + " ", log.LstdFlags)
    logger.Println(msg)
	return msg
    //time.Sleep(5 * time.Second)
    //logger.Println("A new log, 5 seconds later")
}