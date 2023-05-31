ES-curator 的 golang 版替代方案

會使用到的設定檔有 setting.yml 和 config.yml。

setting.yml 用來設定基本的環境參數、控制排程開關及測試模式開關。

# setting.yml 參數說明

以下是一個 setting.yml 的範例

```
es:
  url: https://10.99.1.64:9200
  sourceAccount: "elastic"
  sourcePassword: "a12345678"

information:
  test_mode: true
  logdir: "/Users/chen/Documents/gitlab/git-out/product/house_keeping/es-curator/log/custom.log"
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
是否在測試模式下執行，執行後只會在 log 中 print 出每個 action 條件設定下匹配到的 index，不會實際執行action，可接受的值為 false or false。

```logdir``` </br>
log 存放的位置。

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

### max_num _segment
---
**used in forcemerge**

type = number 
ex. 1,2,3，forcemerge 後 index 的 segment 數量。




## Filter types
三個過濾條件，可混用，或單獨使用。

- `age`
- `pattern`
- `space`

### Filter elements
掛在 filtertype 下使用，不同的 filtertype 各有適用的 filter elements，詳細說明請往下翻看。

- `source`
- `direction`
- `unit`
- `unit_count`
- `kind`
- `value`
- `disk_space`


### age
---
以執行程式當下，以 unit_count (5)  unit (days) 前的時間為基準，篩選出前 (older) 或後 (younger) 產生的 index。

例： 當前時間為2023/01/16，unit : days ; unit_count : 5 ; direction : older ，篩選結果為2023/01/11 之前的產生的所有 index。

#### source 

- creation_date

#### direction
- older
- younger

#### unit
- years
- months
- days

#### unit_count
- 任一正整數 ex. 1、2、5、10....

### pattern
---
以 index 名稱做匹配條件，可前匹配 (prefix) 、後匹配 (suffix)、及正則匹配 (regex)，value 中輸入的是匹配字樣，例如 kind : prefix ; value : logstash-asa 會匹配到所有 logstash-asa 開頭的 index。

#### kind
- prefix
- suffix
- regex

#### value
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

#### disk_space
- 任一正整數，單位 GB，例如 10 代表 10 GB 。



# config sample
會用到兩個 config ，setting.yml 及 config.yml ，setting.yml 控制環境參數及排程執行相關；config.yml 控制要執行的 Actions。

注意縮排不能有誤，不然程式會出錯。

### setting.yml 
execute_cron ，可以是 true or flase ， if true，程式進入排程，在指定的時間執行；if false，手動執行所有 actions。

```
es:
  url: https://10.99.1.117:9200
  sourceAccount: "elastic"
  sourcePassword: "12345678"

information:
  logdir: "/Users/chen/Downloads/BiMap/程式區/es-curator/log/custom.log"
  execute_cron : true
  period: "59 11 * * *"
```

### config.yml 
config 中可以有一或多個 actions，一個 actions 中現階段最多可加入三個 filter，各個 filter 各有適用的 filter element ，請詳閱上方說明。

```
actions: 
  - action: delete_indices
    description: delete selected indices1
    options: 
      disable_action: false
    filters:
    - filtertype: age
      source: creation_date
      direction: older
      unit: days
      unit_count: 1
    - filtertype: pattern
      kind: prefix
      value: prefixmore
      exclude: 
    - filtertype: space
      disk_space: 2
      use_age: True
      source: creation_date

  - action: close
    description: close selected indices
    options: 
      disable_action: true
    filters:
    - filtertype: age
      source: creation_date
      direction: older
      unit: days
      unit_count: 1
    - filtertype: pattern
      kind: prefix
      value: prefixmore

  - action: open
    description: open selected indices
    options: 
      disable_action: true
    filters:
    - filtertype: age
      source: creation_date
      direction: older
      unit: days
      unit_count: 1
    - filtertype: pattern
      kind: prefix
      value: prefixmore


  - action: allocation
    description:  allocation routing to warm node setup for *-h indices older than 1 days, based on index-name1
    options:
      disable_action: True
      key: _tier_preference
      value: data_hot
      allocation_type: include
      delay: 20
    filters:
    - filtertype: pattern
      kind: prefix
      value: logstash-zs

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


  - action: delete_indices
    description: delete selected indices1
    options: 
      disable_action: true
    filters:
    - filtertype: pattern
      kind: prefix
      value: logstash-zs


```



