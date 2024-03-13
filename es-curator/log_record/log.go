package log_record

import (
	"log"
	"os"
	"es-curator/global"
	"time"
    "fmt"
)

func Logrecord(title,msg string) string{


    fileName := fmt.Sprintf("%s/housekeeping_%s.log", global.EnvConfig.INFORMATION.LogPath, time.Now().Format("20060102"))    
    // open file and create if non-existent
    file, err := os.OpenFile( fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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


func ActionDetailrecord(title,msg string) string{


    fileName := fmt.Sprintf("%s/details_%s.log", global.EnvConfig.INFORMATION.LogPath, time.Now().Format("20060102"))    
    // open file and create if non-existent
    file, err := os.OpenFile( fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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