package main

import (
	"os"
	"context"
	"strings"
	"net/http"
	"html/template"

	"github.com/superedge/superedge/examples/Tars/testapp/testapp"
	"github.com/TarsCloud/TarsGo/tars"
	"github.com/tarscloud/gopractice/common/log"
)

func Handler(ctx context.Context, app *testapp.Hello, w http.ResponseWriter, r *http.Request) {
	var Req, Res testapp.Message
	Req.Nodename = os.Getenv("NODE_NAME")
	Req.Podname = os.Getenv("POD_NAME")
	Req.Podip = os.Getenv("POD_IP")
	Req.Set = tars.GetServerConfig().Setdivision
	Req.Gridkey = os.Getenv("GRID_KEY")
	str := strings.Split(Req.Podname, "-")
	Req.Gridvalue = str[len(str) - 3]
	Req.Content = "Hello, I'm client!"
	_, err := app.Echo(&Req, &Res)
	if err != nil {
		log.Error(ctx,"Echo call error")
	}
	Buffer := make(map[string] *testapp.Message)
	Buffer["Server"] = &Res
	Buffer["Client"] = &Req
	t, err := template.ParseFiles("/tars/bin/Html.tmpl")
	if err != nil {
		log.Error(ctx,"Parse html template faile:" + err.Error())
	}
	t.Execute(w,Buffer)
}
