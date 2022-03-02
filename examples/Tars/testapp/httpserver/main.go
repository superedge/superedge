package main

import (
	"os"
	"context"
	"net/http"

	"github.com/superedge/superedge/examples/Tars/testapp/testapp"
	"github.com/TarsCloud/TarsGo/tars"
	"github.com/tarscloud/gopractice/common/metrics"
	"github.com/tarscloud/gopractice/common/log"
)

var comm *tars.Communicator

func init(){
}

func main() {
	cfg := tars.GetServerConfig()
	comm = tars.NewCommunicator()
	locator := os.Getenv("TARS_LOCATOR")
	comm.SetProperty("locator", locator)
	obj := "testapp.testserver.HelloObj"
	app := new(testapp.Hello)
	comm.StringToProxy(obj, app)
	mux := &tars.TarsHttpMux{}
	ctx, _ := context.WithCancel(context.Background())
	mux.HandleFunc("/",func(w http.ResponseWriter, r *http.Request){
		Handler(ctx,app,w,r)
	})
	tars.AddHttpServant(mux, cfg.App+"."+cfg.Server+".HttpObj") 
	go metrics.Listen()
	metrics.SetPrometheusStat() 
	log.Info(ctx,cfg.App+"."+cfg.Server+"run...")
	tars.Run()
}
