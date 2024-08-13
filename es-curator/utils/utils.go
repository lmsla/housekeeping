package utils

import (
	"es-curator/global"
	"es-curator/structs"
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

func LoadEnvironment() {
	loadSettingFile()
	viperSettingToModel()
	loadConfigFile()
	// viperconfigToModel()
}

func loadConfigFile() {
	// var configViperConfig = viper.New()
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/bimap-housekeeping")
	//读取配置文件内容
	if err := viper.ReadInConfig(); err != nil {
		global.Logger.Error(err.Error())
		panic(err)
		
	}
	var c structs.ActionStruct
	if err := viper.Unmarshal(&c); err != nil {
		global.Logger.Error(err.Error())
		panic(err)
	}


	global.ActionStruct = &c
}


// func viperconfigToModel() {
	// var c structs.ActionStruct
	// c.Actions = viper.GetStringSlice("actions")
// }

func loadSettingFile() {
	viper.SetConfigName("setting")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/bimap-housekeeping")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("沒有發現 setting.yml，改抓取環境變數")
			viper.AutomaticEnv()
			viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		} else {
			// 有找到 config.yml 但是發生了其他未知的錯誤
			global.Logger.Error(err.Error())
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}
}

func viperSettingToModel() {
	var config structs.EnviromentModel
	// var action structs.Action
	//// ES
	config.ES.URL = viper.GetStringSlice("es.url")
	config.ES.SourceAccount = viper.GetString("es.sourceAccount")
	config.ES.SourcePassword = viper.GetString("es.sourcePassword")
	//// INFORMATION
	// config.INFORMATION.LogPath = viper.GetString("information.logPath")
	config.INFORMATION.Period = viper.GetString("information.period")
	config.INFORMATION.CaPath = viper.GetString("information.caPath")
	config.INFORMATION.Execute_cron = viper.GetBool("information.execute_cron")
	config.INFORMATION.Test_mode = viper.GetBool("information.test_mode")


	config.Log.Path = viper.GetString("log.path")
	config.Log.MaxSize = viper.GetInt("log.maxSize")
	config.Log.MaxBackups = viper.GetInt("log.maxBackups")
	config.Log.MaxAge = viper.GetInt("log.maxAge")
	config.Log.Debug = viper.GetBool("log.debug")

	global.EnvConfig = &config
	// global.Action = &action
}
