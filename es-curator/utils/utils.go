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
	viperconfigToModel()
}

func loadConfigFile() {
	// var configViperConfig = viper.New()
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	//读取配置文件内容
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	var c structs.ActionStruct
	if err := viper.Unmarshal(&c); err != nil {
		panic(err)
	}
	
	// fmt.Println("0",c.Actions)
	// for i := range c.Actions {
	// 	// c.Actions[i].Description = viper.GetString("actions.description")
	// 	// action := viper.GetString("actions.action")
	// 	for j := range c.Actions[i].Filters {
	// 		int1 := viper.GetInt("actions.action.filters.range_from")
	// 		fmt.Println("1",c.Actions[i].Filters[j].Range_From)
	// 		fmt.Println("int1:",int1)
	// 	}

	// 	fmt.Println("2",c.Actions[i].Description)
	// }
	// a := configViperConfig.GetStringSlice("actions.action")
	// fmt.Println("a:",a)
	// action := viper.GetInt("actions.action.filters.range_from")
	// fmt.Println("4",action)

	global.ActionStruct = &c
}


func viperconfigToModel() {
	// var c structs.ActionStruct
	// c.Actions = viper.GetStringSlice("actions")
}

func loadSettingFile() {
	viper.SetConfigName("setting")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("沒有發現 setting.yml，改抓取環境變數")
			viper.AutomaticEnv()
			viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		} else {
			// 有找到 config.yml 但是發生了其他未知的錯誤
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
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
	config.INFORMATION.Logdir = viper.GetString("information.logdir")
	config.INFORMATION.Period = viper.GetString("information.period")
	config.INFORMATION.CaPath = viper.GetString("information.caPath")
	config.INFORMATION.Execute_cron = viper.GetBool("information.execute_cron")

	global.EnvConfig = &config
	// global.Action = &action
}
