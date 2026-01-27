package utils

import (
	"housekeeping/internal/global"
	"housekeeping/internal/structs"
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"github.com/mitchellh/mapstructure"
)

func LoadEnvironment() error {
	if err := loadSettingFile(); err != nil {
		return fmt.Errorf("載入 setting.yml 失敗: %w", err)
	}
	
	if err := viperSettingToModel(); err != nil {
		return fmt.Errorf("解析 setting.yml 配置失敗: %w", err)
	}
	
	if err := loadConfigFile(); err != nil {
		return fmt.Errorf("載入 config.yml 失敗: %w", err)
	}
	
	return nil
}

func loadConfigFile() error {
	// 創建獨立的 Viper 實例避免與 setting.yml 的配置衝突
	configViper := viper.New()
	configViper.SetConfigName("config")
	configViper.SetConfigType("yml")
	configViper.AddConfigPath(".")
	configViper.AddConfigPath("./configs")
	configViper.AddConfigPath("/etc/bimap-housekeeping")

	//讀取配置文件内容
	if err := configViper.ReadInConfig(); err != nil {
		fmt.Println("讀取 config.yml 錯誤，請檢查 yml 檔案", err.Error())
		if global.Logger != nil {
			global.Logger.Error(err.Error())
		}
		return fmt.Errorf("讀取 config.yml 失敗: %w", err)
	}

	// var c structs.ActionStruct
	// if err := configViper.Unmarshal(&c); err != nil {
	// 	if global.Logger != nil {
	// 		global.Logger.Error(err.Error())
	// 	}
	// 	return fmt.Errorf("解析 config.yml 失敗: %w", err)
	// }

    var c structs.ActionStruct
    if err := configViper.Unmarshal(&c, func(dc *mapstructure.DecoderConfig) {
        dc.TagName = "yaml"            // 讓 Viper 解碼時看 yaml 標籤
        dc.WeaklyTypedInput = true     // 選擇性：允許 "20" -> int(20)
    }); err != nil {
        if global.Logger != nil {
            global.Logger.Error(err.Error())
        }
        return fmt.Errorf("解析 config.yml 失敗: %w", err)
    }

	// 添加內容驗證：確保至少有一個 Action 配置
	if len(c.Actions) == 0 {
		return fmt.Errorf("嚴重錯誤: config.yml 中沒有找到任何 action 配置，請檢查配置檔案")
	}

	global.ActionStruct = &c
	
	// 調試信息：顯示讀取到的 Action 數量
	fmt.Printf("⚠️  配置驗證 - Actions loaded: %d\n", len(c.Actions))
	
	return nil
}

func loadSettingFile() error {
	viper.SetConfigName("setting")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/bimap-housekeeping")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("沒有發現 setting.yml，改抓取環境變數")
			viper.AutomaticEnv()
			viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		} else {
			// 有找到 setting.yml 但是發生了其他未知的錯誤
			if global.Logger != nil {
				global.Logger.Error(err.Error())
			}
			return fmt.Errorf("讀取 setting.yml 發生致命錯誤: %w", err)
		}
	}
	return nil
}

func viperSettingToModel() error {
	var config structs.EnviromentModel
	
	// 驗證配置檔案是否正確讀取 - 檢查關鍵配置項
	urls := viper.GetStringSlice("es.url")
	if len(urls) == 0 {
		return fmt.Errorf("嚴重錯誤: 找不到 es.url 配置 - setting.yml 可能讀取失敗")
	}
	
	// 驗證測試模式配置是否讀取到
	testMode := viper.GetBool("information.test_mode")
	logPath := viper.GetString("log.path")

	
	// 如果關鍵配置都是預設值，很可能配置檔案讀取失敗
	if logPath == "" {
		return fmt.Errorf("嚴重錯誤: log.path 為空值 - 這表示 setting.yml 解析失敗")
	}
	
	fmt.Printf("⚠️  配置驗證 - Test mode: %v, Log path: %s\n", testMode, logPath)
	
	//// ES
	config.ES.URL = urls
	config.ES.SourceAccount = viper.GetString("es.sourceAccount")
	config.ES.SourcePassword = viper.GetString("es.sourcePassword")
	
	// INFORMATION
	config.INFORMATION.Period = viper.GetString("information.period")
	config.INFORMATION.CaPath = viper.GetString("information.caPath")
	config.INFORMATION.ExecuteCron = viper.GetBool("information.execute_cron")
	config.INFORMATION.TestMode = viper.GetBool("information.test_mode")

	config.Log.Path = viper.GetString("log.path")
	config.Log.MaxSize = viper.GetInt("log.maxSize")
	config.Log.MaxBackups = viper.GetInt("log.maxBackups")
	config.Log.MaxAge = viper.GetInt("log.maxAge")
	config.Log.Debug = viper.GetBool("log.debug")
	config.Log.ToES = viper.GetBool("log.toes")
	config.Log.HealthCheckInterval = viper.GetInt("log.health_check_interval")
	
	global.EnvConfig = &config
	return nil
}
