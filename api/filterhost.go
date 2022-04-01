package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
	"zbx-monitor/config"
	"zbx-monitor/middleware"

	zabbixgo "github.com/canghai908/zabbix-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HistoryFilterRequest struct {
	ServerID string  `json:"serverId"`
	Comp     string  `json:"comp"`
	Cpu      float64 `json"cpu"`
	Memory   float64 `json"memory"`
	Days     int64   `json:"days"`
}

func HistoryFiterApi(c *gin.Context) {
	configMap := config.InitServersConfig()
	statList := make(map[string]interface{})
	var historyFilterRequest HistoryFilterRequest
	err := c.ShouldBindJSON(&historyFilterRequest)
	if err != nil {
		zap.L().Fatal("json 解析错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, statList)
	}
	for _, v := range configMap.Servers {
		if historyFilterRequest.ServerID == v.ID {
			url := fmt.Sprintf("http://%s/api_jsonrpc.php", v.Host)
			api := zabbixgo.NewAPI(url)
			api.Login(v.User, v.Password)
			statList := HistoryFilter(api, historyFilterRequest.ServerID, historyFilterRequest.Comp, historyFilterRequest.Cpu, historyFilterRequest.Cpu, historyFilterRequest.Days)
			c.JSON(http.StatusOK, statList)
		}
	}
	c.JSON(http.StatusBadRequest, statList)
}

func HistoryFilter(api *zabbixgo.API, serverID string, compareWay string, cpuCondition float64, memoryCondition float64, days int64) []map[string]interface{} {
	Hosts := HostsGetByGroupName(api, "")
	hostIds := []string{}
	for _, v := range Hosts {
		hostIds = append(hostIds, v.HostID)
	}
	cpuHostItemMap := SelectItemIds(api, hostIds, "system.cpu.util[,idle]")
	memoryHostItemMap := SelectItemIds(api, hostIds, "vm.memory.utilization")
	var statMap sync.Map
	var wg sync.WaitGroup
	ch := make(chan int, 50)
	for _, v := range hostIds {
		wg.Add(1)
		ch <- 1
		cpuItemId := cpuHostItemMap[v]
		memoryItemId := memoryHostItemMap[v]
		go HistoryStat(api, &statMap, &wg, ch, days, v, cpuItemId, memoryItemId)
	}
	wg.Wait()
	resultSlice := HostHistoryStat(serverID, hostIds, statMap, compareWay, cpuCondition, memoryCondition)
	return resultSlice
}

func HostHistoryStat(serverID string, hostIds []string, statMap sync.Map, compareWay string, cpuCondition float64, memoryCondition float64) (resultSlice []map[string]interface{}) {
	for _, v := range hostIds {
		resultMap := make(map[string]interface{})
		dataInterface, _ := statMap.Load(v)
		dataMap := dataInterface.(map[string]float64)
		data := middleware.RdsClient.HGet("zabbixStat", serverID)
		var statList []map[string]string
		err := json.Unmarshal([]byte(data.Val()), &statList)
		if err != nil {
			zap.L().Fatal("json 解析错误", zap.Error(err))
		}
		statMap := make(map[string]map[string]string)
		for _, v := range statList {
			statMap[v["hostId"]] = v
		}
		if compareWay == "gte" {
			if cpuCondition != 0.0 && memoryCondition != 0.0 {
				if dataMap["cpuUse"] >= cpuCondition && dataMap["memoryUse"] >= memoryCondition {
					resultMap["hostIds"] = v
					resultMap["cpuUse"] = dataMap["cpuUse"]
					resultMap["memoryUse"] = dataMap["memoryUse"]
					resultMap["hostName"] = statMap[v]["hostName"]
					resultMap["cpuNum"] = statMap[v]["cpuNum"]
					resultMap["memoryTotal"] = statMap[v]["memoryTotal"]
					resultMap["ip"] = statMap[v]["ip"]
					resultSlice = append(resultSlice, resultMap)
				}

			} else if cpuCondition != 0.0 && memoryCondition == 0.0 {
				if dataMap["cpuUse"] >= cpuCondition {
					resultMap["hostIds"] = v
					resultMap["cpuUse"] = dataMap["cpuUse"]
					resultMap["memoryUse"] = dataMap["memoryUse"]
					resultMap["hostName"] = statMap[v]["hostName"]
					resultMap["cpuNum"] = statMap[v]["cpuNum"]
					resultMap["memoryTotal"] = statMap[v]["memoryTotal"]
					resultMap["ip"] = statMap[v]["ip"]
					resultSlice = append(resultSlice, resultMap)
				}

			} else if cpuCondition == 0.0 && memoryCondition != 0.0 {
				if dataMap["memoryUse"] >= memoryCondition {
					resultMap["hostIds"] = v
					resultMap["cpuUse"] = dataMap["cpuUse"]
					resultMap["memoryUse"] = dataMap["memoryUse"]
					resultMap["hostName"] = statMap[v]["hostName"]
					resultMap["cpuNum"] = statMap[v]["cpuNum"]
					resultMap["memoryTotal"] = statMap[v]["memoryTotal"]
					resultMap["ip"] = statMap[v]["ip"]
					resultSlice = append(resultSlice, resultMap)
				}

			}

		} else if compareWay == "lte" {
			if cpuCondition != 0.0 && memoryCondition != 0.0 {
				if dataMap["cpuUse"] <= cpuCondition && dataMap["memoryUse"] <= memoryCondition {
					resultMap["hostIds"] = v
					resultMap["cpuUse"] = dataMap["cpuUse"]
					resultMap["memoryUse"] = dataMap["memoryUse"]
					resultMap["hostName"] = statMap[v]["hostName"]
					resultMap["cpuNum"] = statMap[v]["cpuNum"]
					resultMap["memoryTotal"] = statMap[v]["memoryTotal"]
					resultMap["ip"] = statMap[v]["ip"]
					resultSlice = append(resultSlice, resultMap)
				}

			} else if cpuCondition != 0.0 && memoryCondition == 0.0 {
				if dataMap["cpuUse"] <= cpuCondition {
					resultMap["hostIds"] = v
					resultMap["cpuUse"] = dataMap["cpuUse"]
					resultMap["memoryUse"] = dataMap["memoryUse"]
					resultMap["hostName"] = statMap[v]["hostName"]
					resultMap["cpuNum"] = statMap[v]["cpuNum"]
					resultMap["memoryTotal"] = statMap[v]["memoryTotal"]
					resultMap["ip"] = statMap[v]["ip"]
					resultSlice = append(resultSlice, resultMap)
				}

			} else if cpuCondition == 0.0 && memoryCondition != 0.0 {
				if dataMap["memoryUse"] <= memoryCondition {
					resultMap["hostIds"] = v
					resultMap["cpuUse"] = dataMap["cpuUse"]
					resultMap["memoryUse"] = dataMap["memoryUse"]
					resultMap["hostName"] = statMap[v]["hostName"]
					resultMap["cpuNum"] = statMap[v]["cpuNum"]
					resultMap["memoryTotal"] = statMap[v]["memoryTotal"]
					resultMap["ip"] = statMap[v]["ip"]
					resultSlice = append(resultSlice, resultMap)
				}
			}

		}
	}
	return
}

func SelectItemIds(api *zabbixgo.API, hostIds []string, item string) (resultMap map[string]string) {
	params := map[string]interface{}{
		"output":  []string{"hostid", "itemid"},
		"search":  map[string]string{"key_": item}, //监控项
		"hostids": hostIds}
	result, err := api.ItemsGet(params)
	if err != nil {
		zap.L().Fatal("itemId获取错误", zap.Error(err))
	}
	resultMap = make(map[string]string)
	for _, v := range result {
		hostId := v.HostId
		itemId := v.ItemId
		resultMap[hostId] = itemId
	}
	return resultMap
}

func HistoryStat(api *zabbixgo.API, statMap *sync.Map, wg *sync.WaitGroup, ch chan int, days int64, hostId string, cpuItemId string, memoryItemId string) {
	defer wg.Add(-1)
	timeTampEnd := time.Now().Unix()
	timeTampStart := timeTampEnd - days*86400
	cpuUtilAvg := StatItemHistoryAvg(api, cpuItemId, strconv.FormatInt(timeTampStart, 10), strconv.FormatInt(timeTampEnd, 10))
	memoryUtilAvg := StatItemHistoryAvg(api, memoryItemId, strconv.FormatInt(timeTampStart, 10), strconv.FormatInt(timeTampEnd, 10))
	ma := make(map[string]float64)
	ma["cpuUse"] = cpuUtilAvg
	ma["memoryUse"] = memoryUtilAvg
	statMap.Store(hostId, ma)
	<-ch
}

func StatItemHistoryAvg(api *zabbixgo.API, itemId string, time_from string, time_till string) float64 {
	params := map[string]interface{}{
		"output":    "extend", //需求数据，监控项的name 和最新的值
		"history":   0,
		"itemids":   []string{itemId},
		"time_from": time_from,
		"time_till": time_till,
	}
	result, err := api.HistoryGet(params)
	if err != nil {
		zap.L().Fatal("历史数据获取错误", zap.Error(err))
	}
	sum := 0.0
	avg := 0.0
	if len(result) > 0 {
		for _, v := range result {
			value, err := strconv.ParseFloat(v.Value, 64)
			if err != nil {
				zap.L().Fatal("", zap.Error(err))
			}
			sum += value
		}
		avg = sum / float64(len(result))
	}
	return avg
}
