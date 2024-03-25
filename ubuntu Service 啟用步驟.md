# 以 Service 啟用 housekeeping 程式

## step 1
1. 將主程式 main 搬移到指定路徑下
2. 主程式 main 需與 setting.yml 、 config.yml 在同一層
3. 設定 log 檔存放的資料夾 (setting.yml 中的 logPath 設定)

## step 2

1. 編輯 housekeeping.service
2. 需要修改的參數：
	- WorkingDirectory - 存放主程式的路徑
	- ExecStart - 主程式的路徑

```
vim /etc/systemd/system/housekeeping.service
```

內容範例如下

```
[Unit]
Description=Bimap housekeeping program
After=network.target

[Service]
User=root
Group=root
Restart=always
WorkingDirectory=/home/localadmin/curator
ExecStart=/home/localadmin/curator/main

[Install]
WantedBy=multi-user.target
```


