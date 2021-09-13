package main

import (
	"os"
	"github.com/TarsCloud/TarsGo/tars"
	"testapp"
	"github.com/tarscloud/gopractice/common/metrics"
)


func init() {
}

func main() {
	cfg := tars.GetServerConfig()
	imp := new(HelloImp)
	err := imp.Init()
	if err != nil {
		os.Exit(-1)
	}
	app := new(testapp.Hello)
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".HelloObj")
	go metrics.Listen()
	metrics.SetPrometheusStat()//用prometheus代替tars stat
	tars.Run()
}
