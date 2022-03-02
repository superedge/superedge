package main

import (
	"os"
	"context"
	"strings"

	"github.com/superedge/superedge/examples/Tars/testapp/testapp"
	"github.com/TarsCloud/TarsGo/tars"
	"github.com/tarscloud/gopractice/common/log"
)

type HelloImp struct{
}

func (imp *HelloImp) Init() error {
	return nil
}

// Destroy servant destroy
func (imp *HelloImp) Destroy() {
}

func (imp *HelloImp) Echo(ctx context.Context, Req *testapp.Message, Res *testapp.Message) (int32, error) {
	log.Info(ctx, Req.Content)
	Res.Nodename = os.Getenv("NODE_NAME")
	Res.Podname = os.Getenv("POD_NAME")
	Res.Podip = os.Getenv("POD_IP")
	Res.Set = tars.GetServerConfig().Setdivision
	Res.Content = "Hello,I'm Server!"
	Res.Gridkey = os.Getenv("GRID_KEY")
	str := strings.Split(Res.Podname,"-")
	Res.Gridvalue = str[len(str)-3]
	return 0, nil
}
