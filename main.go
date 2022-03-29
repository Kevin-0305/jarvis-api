package main

import (
	"zbx-monitor/api"
	"zbx-monitor/cmd"
)

func main() {
	go api.SetZabbixStatToReddis()
	cmd.Execute()
}
