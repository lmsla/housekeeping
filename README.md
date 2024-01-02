ES-curator 的 golang 版替代方案

主程式是與 README.md 同一層資料夾的 main，資料夾 es-curator 中含有原始碼，供開發人員維護使用。  

會使用到的設定檔有 setting.yml 和 config.yml。

setting.yml 用來設定基本的環境參數、控制排程開關及測試模式開關。
config.yml 控制要執行的 Actions。

注意縮排不能有誤，不然程式會出錯。

# setting.yml 參數說明

以下是一個 setting.yml 的範例

```
es:
  url: https://10.99.1.64:9200
  sourceAccount: "elastic"
  sourcePassword: "a12345678"

information:
  test_mode: true
  logPath: "/Users/chen/Documents/gitlab/git-out/product/house_keeping/es-curator/log"
  execute_cron : false
  period: "59 11 * * *"
```

## es
```url ``` </br>
es的網址</br>

```sourceAccount``` </br>
es 的帳號</br>

```sourcePassword ``` </br>
es 的密碼</br>

## information

```test_mode``` </br>
是否在測試模式下執行，執行後只會在 log 中 print 出每個 action 條件設定下匹配到的 index，不會實際執行action，可接受的值為 true or false。

```logPath``` </br>
log 存放的路徑，注意最後不要加／
會按執行的日期產生如 housekeeping_20230605.log 的紀錄檔。

```execute_cron``` </br>
是否使用排程定時執行，可接受的值為 false or false，若為 true 則下一個參數 ```period``` 須加入排程時間，若為 false 則是單次執行。

```period``` </br>
crontab的執行時間，與```execute_cron ```搭配使用。

# config.yml 參數說明
 
## Actions
可使用的actions如下</br>
`open` </br>
`close` </br>
`delete_indices`</br>
`allocation`</br>
`forcemerge`</br>

## Description
描述執行的動作，可將相關說明寫上，方便日後在 log 中查閱相關紀錄。

ex. description: delete selected indices1

## Options
用來設定 action 執行時欲傳入的常用參數

- `disable_action`
- `key`
- `value`
- `allocation_type`
- `delay`
- `max_num_segment`
- `delay`

### disable_action
---
type = bool

接受的值為 true or false，控制 action 執行與否，每個 action 都要加上這個控制選項。
### key
---
**used in allocation**

`_tier_preference`

### value
---
**used in allocation**

`data_hot`,`data_warm`,`data_cold`

### allocation_type
---
**used in allocation**

`include`,`require`,`exclude`

### delay
---
type = number 

action 執行後等待的時間，單位為秒

沒加的話程式默認等待0秒，會直接進入下一個 action

### max_num _segment
---
**used in forcemerge**

type = number 
ex. 1,2,3，forcemerge 後 index 的 segment 數量。




## Filter types
四個過濾條件，可混用，或單獨使用。


與 index 產生時間相關：

- `age`

與 index 名稱相關：

- `pattern`

與 disk 容量相關：

- `space`
- `water_level`

```
*註：時間相關的 filter 不可與 disk 相關的 filter 一起使用 ;  disk 容量相關的兩個 filter 也不可同時使用。
	例：action 中同時引用 age & space 或同時引用 space & water_level

```

### Filter elements
掛在 filtertype 下使用，不同的 filtertype 各有適用的 filter elements，詳細說明請往下翻看。

- `source`
- `direction`
- `unit`
- `unit_count`
- `range_from`
- `range_to`
- `kind`
- `value`
- `disk_space`
- `upper_limit`
- `lower_limit`


### age
---


##### source

- creation_date

##### direction
- older
- younger

以執行程式當下，篩選出 `range_from` 到  `range_to` 之間產生的 index。

例： 當前時間為2023/01/16，`direction` : range ; `unit` : days ; `range_from` : 5 ; `range_to` : 2 ，篩選結果為2023/01/14 ~ 2023/01/15 時間段中的產生的所有 index。

- range

以執行程式當下，以 unit_count (5) unit (days) 前的時間為基準，篩選出前 (older) 或後 (younger) 產生的 index。

例： 當前時間為2023/01/16，unit : days ; unit_count : 5 ; direction : older ，篩選結果為2023/01/11 之前的產生的所有 index。

##### unit
- years
- months
- days

##### unit_count
- 任一正整數 ex. 1、2、5、10....

##### range_from 
- 任一正整數 ex. 1、2、5、10....

##### range_to
- 任一正整數 ex. 1、2、5、10....

### pattern
---
以 index 名稱做匹配條件，可前匹配 (prefix) 、後匹配 (suffix)、及正則匹配 (regex)，value 中輸入的是匹配字樣，例如 kind : prefix ; value : logstash-asa 會匹配到所有 logstash-asa 開頭的 index。

##### kind
- prefix
- suffix
- regex

##### value
匹配字樣</br>
ex. logstash-ap , ap

### space
---
計算 index 所佔空間，由最新的 index 開始算，超過設定的閥值 disk_space ，篩選出較舊的 index，新舊判定依據為 index 產生時間。

例，有五個 index ，舊到新的順序分別為 01~05 ，disk_space : 20 ，index-05、index-04 加起來共 20 GB，超過的部分 index-03、index-02、index-01，會被篩選出來。

```
index-01 10GB
index-02 10GB
index-03 10GB
index-04 10GB
index-05 10GB
```

閥值為絕對值，只要超過閥值就會被篩出來，如下例，index-05、index-04 加起來共 15 GB，再加上 index-03 的 5.1GB 共 20.1 GB ，即使只超過 0.1 GB，index-03 也會被列入篩選名單中。

```
index-01 10GB
index-02 10GB
index-03 5.1GB
index-04 5GB
index-05 10GB

```

##### disk_space
- 任一正整數，單位 GB，例如 10 代表 10 GB 。

### water_level
---
disk 水位控管，統計目前 cluster 中所有 nodes 的 disk 使用率取平均值，如果超過 `upper_limit` (上限值百分比)，
則觸發篩選機制 - 由舊到新加總 index 所佔容量直到約等於 `upper_limit` 與 `lower_limit` 百分比差值佔 cluster disk 總量。

##### upper_limit

- 任一正整數，單位 %，例如 50 代表上限 50% 。

##### lower_limit

- 任一正整數，單位 %，例如 40 代表下限 40% 。


# config sample
會用到兩個 config ，setting.yml 及 config.yml ，setting.yml 控制環境參數及排程執行相關；config.yml 控制要執行的 Actions。

注意縮排不能有誤，不然程式會出錯。

### config.yml 
config 中可以有一或多個 action，一個 action 可是情況搭配不同的 filter，不同的 filter 有各自適用的 filter element ，請詳閱上方說明。

```
actions: 
### 刪除產生在10天以前且開頭為 logstash- 的所有 index

  - action: delete_indices
    description: delete selected indices1
    options: 
      disable_action: false
      delay: 5
    filters:
    - filtertype: age
      source: creation_date
      direction: older
      unit: days
      unit_count: 10
    - filtertype: pattern
      kind: prefix
      value: logstash-
      
### 開頭為 logstash- 的所有 index 只保留最新的 20G ，超過的刪除

  - action: delete_indices
    description: delete selected indices2
    options: 
      disable_action: false
      delay: 5
    filters:
    - filtertype: pattern
      kind: prefix
      value: logstash-
    - filtertype: space
      disk_space: 20

### 當 cluster 的平均 disk 使用量高於 80% 時，篩選出開頭為 logstash- ，
	  從舊的開始刪，刪到釋放出 5% 的 cluster 的平均 disk 使用量。

  - action: delete_indices
    description: delete selected indices3
    options: 
      disable_action: true
      delay: 5
    filters:
    - filtertype: pattern
      kind: prefix
      value: logstash-
    - filtertype: water_level
      upper_limit: 80
      lower_limit: 75

### 將前1~5天之間產生的 index 搬遷到 warm data node 
	 
  - action: allocation
    description:  allocation selected indices to data_warm
    options:
      disable_action: true
      key: _tier_preference
      value: data_warm
      allocation_type: include
      delay: 20
    filters:
    - filtertype: pattern
      kind: prefix
      value: logstash-zs
    - filtertype: age
      source: creation_date
      direction: range
      unit: days
      range_from: 5
      range_to: 1
      
### 將前1~5天之間產生，開頭為 logstash-zs 的所有 index forcemerge to segments = 1
      
  - action: forcemerge
    description: Perform a forceMerge on selected indices to 'max_num_segments' per shard
    options:
      disable_action: true
      max_num_segment: 1
      delay: 10
    filters:
    - filtertype: pattern
      kind: prefix
      value: logstash-zs
    - filtertype: age
      source: creation_date
      direction: range
      unit: days
      range_from: 5
      range_to: 1    


```



