package utils

import (
	"fmt"
	"es-curator/global"
	"es-curator/job"
	// "es-curator/log_record"
	"github.com/robfig/cron/v3"
)

func LoadCrontab() {
	//checkLIcense()
	c := cron.New()
	_, err := c.AddFunc(global.EnvConfig.INFORMATION.Period, job.Action_controll)
	//fmt.Print(global.EnvConfig.CRONTAB.Period,global.EnvConfig.INFLUX.URL)
	if err != nil {
		fmt.Println("crontab BiMAP-housekeeping 初始化失敗")
		// log_record.Logrecord("排程 ","ES-curator排程 初始化失敗")
		global.Logger.Error(err.Error(),"BiMAP-housekeeping排程 初始化失敗")
		// fmt.Println(err.Error())
		// log_record.Logrecord("ERROR ",err.Error())
	} else {
		fmt.Println("crontab BiMAP-housekeeping 初始化成功")
		// log_record.Logrecord("排程 ","ES-curator排程 初始化成功")
		global.Logger.Infow("BiMAP-housekeepingr排程 初始化成功","type","排程")
		c.Start()

	}
}



// func LoadCrontab() {
// 	c := cron.New()
// 	ActionList := global.ActionStruct.Actions
// 	for actions := range ActionList {
// 		if ActionList[actions].Execute_Period != ""  {
// 			var err error
// 			if ActionList[actions].Action == "delete_indices" {
// 				_,err = c.AddFunc(ActionList[actions].Execute_Period,job.Job1)
// 			}else if ActionList[actions].Action == "close"{
// 				_,err = c.AddFunc(ActionList[actions].Execute_Period,job.Job2)
// 			}else if ActionList[actions].Action == "delete_indices" {
// 				_,err = c.AddFunc(ActionList[actions].Execute_Period,job.Job2)
// 			}
			
// 			cronSuccessMsg := fmt.Sprintf("ES-curator排程 %s 初始化成功",ActionList[actions].Description)
// 			cronFailedMsg := fmt.Sprintf("ES-curator排程 %s 初始化失敗",ActionList[actions].Description)
// 			if err != nil {
// 				fmt.Println("crontab ES-curator 初始化失敗")
// 				log_record.Logrecord("排程 ",cronSuccessMsg)
// 				log_record.Logrecord("ERROR ",err.Error())
// 			} else {
// 				fmt.Println("crontab ES-curator 初始化成功")
// 				log_record.Logrecord("排程 ",cronFailedMsg)
// 			}

// 		}

// 	}
// 	c.Start()
// }




