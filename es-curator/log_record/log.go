package log_record

import (
	"es-curator/global"
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"time"
)

// var logger *zap.SugaredLogger

func InitLogger() {
	logPath := global.EnvConfig.Log.Path // 替换为 global.EnvConfig.INFORMATION.LogPath
	// fileName := fmt.Sprintf("%s/housekeeping_%s.log", logPath, time.Now().Format("200601"))

	// 配置 lumberjack for log rotate
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/housekeeping_%s.log", logPath, time.Now().Format("200601")),
		MaxSize:    global.EnvConfig.Log.MaxSize,    // megabytes
		MaxBackups: global.EnvConfig.Log.MaxBackups, // ex.7 最多保留七個備份 log file
		MaxAge:     global.EnvConfig.Log.MaxAge,     // ex.30 最多保留30天
		Compress:   true,                            // 是否壓縮log
	})
	// var core zapcore.Core
	// 配置 zap 編碼器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.MessageKey = "msg"
	encoderConfig.LevelKey = "level"
	encoderConfig.CallerKey = "caller"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// 创建 zap 核心
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		w,
		zap.InfoLevel,
		// zap.ErrorLevel,
	)

	// caller 顯示文件名、行號和zap調用者的函數名
	if global.EnvConfig.Log.Debug {
		global.Logger = zap.New(core, zap.AddCaller()).Sugar()
	} else {
		global.Logger = zap.New(core).Sugar()
	}

}

func InitDetailLogger() {
	logPath := global.EnvConfig.Log.Path // 替换为 global.EnvConfig.INFORMATION.LogPath
	// 配置 lumberjack for log rotate
	w_detail := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/detail_%s.log", logPath, time.Now().Format("200601")),
		MaxSize:    global.EnvConfig.Log.MaxSize,    // megabytes
		MaxBackups: global.EnvConfig.Log.MaxBackups, // ex.7 最多保留七個備份 log file
		MaxAge:     global.EnvConfig.Log.MaxAge,     // ex.30 最多保留30天
		Compress:   true,                            // 是否壓縮log
	})
	// var core zapcore.Core
	// 配置 zap 編碼器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.MessageKey = "msg"
	encoderConfig.LevelKey = "level"
	encoderConfig.CallerKey = "caller"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		w_detail,
		zap.InfoLevel,
		// zap.ErrorLevel,
	)

	// caller 顯示文件名、行號和zap調用者的函數名
	if global.EnvConfig.Log.Debug {
		global.Detail_Logger = zap.New(core, zap.AddCaller()).Sugar()
	} else {
		global.Detail_Logger = zap.New(core).Sugar()
	}

}

func ActionDetailrecords(mode, msg string) string {
	if global.Detail_Logger == nil {
		InitDetailLogger()
	}

	global.Detail_Logger.Infow(msg, "mode", mode)

	return msg
}

func Logrecord(title, msg string) string {

	fileName := fmt.Sprintf("%s/housekeeping_%s.log", global.EnvConfig.Log.Path, time.Now().Format("200601"))
	// open file and create if non-existent
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, title+" ", log.LstdFlags)
	logger.Println(msg)
	return msg

}

func ActionDetailrecord(title, msg string) string {

	fileName := fmt.Sprintf("%s/details_%s.log", global.EnvConfig.Log.Path, time.Now().Format("200601"))
	// open file and create if non-existent
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, title+" ", log.LstdFlags)
	logger.Println(msg)
	return msg
	//time.Sleep(5 * time.Second)
	//logger.Println("A new log, 5 seconds later")
}

func LogTest() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{
		"stdout",
		"./logs.log",
	}

	// 创建日志记录器
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // 确保在程序退出前将缓存的日志刷新到磁盘

	// 使用 logger 记录不同级别的日志
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warn message")
	logger.Error("This is an error message")

	// 记录带有字段的日志
	sugar := logger.Sugar()
	sugar.Infow("This is an info message with fields",
		"key1", "value1",
		"key2", 42,
		"key3", true,
	)

	sugar.Infof("This is an info message with formatted %s", "output")
}
