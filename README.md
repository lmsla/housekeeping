# housekeeping

This is an alternative to ES-curator.

## Actions

Open indices</br>
Close indices</br>
Delete indices</br>

## Description
描述執行的動作，可將相關說明寫上，方便日後在 log 中查閱相關紀錄。

ex. description: delete selected indices1

## Filter types
三個過濾條件，可混用，或單獨使用。

- age
- pattern
- space

## age 
以執行程式當下，以 unit_count (5)  unit (days) 前的時間為基準，篩選出前 (older) 或後 (younger) 產生的 index。

例： 當前時間為2023/01/16，unit : days ; unit_count : 5 ; direction : older ，篩選結果為2023/01/11 之前的產生的所有 index。
### source 
- creation_date

### direction
- older
- younger

### unit
- years
- months
- days

### unit_count
- 任一正整數 ex. 1、2、5、10....

## patter
以 index 名稱做匹配條件，可前匹配 (prefix) 、後匹配 (suffix)、及正則匹配 (regex)，value 中輸入的是匹配字樣，例如 kind : prefix ; value : logstash-asa 會匹配到所有 logstash-asa 開頭的 index。

### kind
- prefix
- suffix
- regex

### value
匹配字樣</br>
ex. logstash-ap , ap

## space
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

### disk_space
- 任一正整數，單位 GB，例如 10 代表 10 GB 



## config sample
會用到兩個 config ，setting.yml 及 config.yml ，setting.yml 控制環境參數及排程執行相關；config.yml 控制要執行的 Actions。

注意縮排不能有誤，不然程式會出錯。

### setting.yml 
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
```
actions: 
  - action: delete_indices
    description: delete selected indices1
    filters:
    - filtertype: age
      source: "creation_date"
      direction: older
      unit: days
      unit_count: 1
    - filtertype: pattern
      kind: prefix
      value: "prefixmore"
      exclude: 
    - filtertype: space
      disk_space: 2
      use_age: True
      source: creation_date

  - action: delete_indices
    description: delete selected indices2
    filters:
    - filtertype: age
      source: "creation_date"
      direction: older
      unit: days
      unit_count: 5
    - filtertype: pattern
      kind: prefix
      value: "test_index"
      exclude:
  - action: delete_indices
    description: delete selected indices3
    filters:
    - filtertype: space
      disk_space: 10
      # use_age: True
      # source: creation_date

  - action: close
    description: close selected indices
    execute_period: "*/1 * * * *"
    filters:
    - filtertype: age
      source: "creation_date"
      direction: older
      unit: days
      unit_count: 1
    - filtertype: pattern
      kind: prefix
      value: "prefixmore"
      exclude:

  - action: open
    description: open selected indices
    filters:
    - filtertype: age
      source: "creation_date"
      direction: older
      unit: days
      unit_count: 1
    - filtertype: pattern
      kind: prefix
      value: "prefixmore"
      exclude:
```



