package api

import (
	"net/http"
	"zbx-monitor/middleware"

	"encoding/json"
	"fmt"
	"log"
	"time"
	"zbx-monitor/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	zabbixgo "github.com/canghai908/zabbix-go"
)

type HostGetResult struct {
	Host       string `json:"host"`
	HostID     string `json:"hostid"`
	Interfaces []map[string]string
}

func GetZabbixStat(c *gin.Context) {
	serverId := c.Param("id")
	data := middleware.RdsClient.HGet("zabbixStat", serverId)
	var statList []map[string]string
	err := json.Unmarshal([]byte(data.Val()), &statList)
	if err != nil {
		c.JSON(http.StatusBadRequest, statList)
	} else {
		c.JSON(http.StatusOK, statList)
	}
}

func SetZabbixStatToReddis() {
	configMap := config.InitServersConfig()
	apiMap := make(map[string]*zabbixgo.API)
	for _, v := range configMap.Servers {
		url := fmt.Sprintf("http://%s/api_jsonrpc.php", v.Host)
		api := zabbixgo.NewAPI(url)
		api.Login(v.User, v.Password)
		apiMap[v.ID] = api
	}
	for _, v := range configMap.Servers {
		itemStat, err := ItemStat(apiMap[v.ID], configMap.Items, configMap.Group)
		if err != nil {
			log.Fatal(err)
		}
		middleware.RdsClient.HSet("zabbixStat", v.ID, itemStat)
	}
	time.Sleep(time.Duration(configMap.Interval) * 60 * time.Second)

}

func ItemStat(api *zabbixgo.API, statConfigMap map[string]string, groupName string) (string, error) {
	statMap := make(map[string]zabbixgo.Items)
	for k, v := range statConfigMap {
		statMap[k] = ItemStatGet(api, v, groupName)
	}
	resultMap := make(map[string]map[string]interface{})
	for _, v := range HostsGetByGroupName(api, groupName) {

		resultMap[v.HostID] = make(map[string]interface{})
		resultMap[v.HostID]["hostId"] = v.HostID
		resultMap[v.HostID]["ip"] = v.Interfaces[0]["ip"]
	}
	for k, _ := range statConfigMap {
		for _, v2 := range statMap[k] {
			_, ok := resultMap[v2.HostId]
			if ok {
				resultMap[v2.HostId][k] = v2.Lastvalue
			}
		}
	}
	var resultList []interface{}
	for _, v := range resultMap {
		resultList = append(resultList, v)
	}
	data, err := json.Marshal(resultList)
	if err != nil {
		return "", err
	}
	return string(data), nil

}

func ItemStatGet(api *zabbixgo.API, key string, groupName string) zabbixgo.Items {
	params := map[string]interface{}{
		"output":      []string{"hostid", "name", "lastvalue"}, //需求数据，监控项的name 和最新的值
		"search":      map[string]string{"key_": key},          //监控项
		"selectHosts": []string{groupName}}                     //群组名称
	ret, err := api.ItemsGet(params)
	if err != nil {
		fmt.Println(err)
	}
	return ret
}

func HostsGetByGroupName(api *zabbixgo.API, groupName string) []HostGetResult {
	params := map[string]interface{}{
		"output":           []string{"hostid", "host"}, //需求数据，监控项的name 和最新的值
		"selectInterfaces": []string{"ip"}}             //群组名称
	response, err := api.CallWithError("host.get", params)
	var result []HostGetResult
	ret, err := json.Marshal(response.Result)
	if err != nil {
		zap.L().Fatal("json 解析错误", zap.Error(err))
	}
	err = json.Unmarshal(ret, &result)
	return result
}
