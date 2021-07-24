package tcp

import (
	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/module"
	"github.com/superedge/superedge/pkg/tunnel/proxy/modules/stream"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"os"
	"testing"
)

func Test_TcpServer(t *testing.T) {
	err := conf.InitConf(util.CLOUD, "../../../../../conf/cloud_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream server configuration file err = %v", err)
		return
	}
	module.InitModules(util.CLOUD)
	InitTcp()
	stream.InitStream(util.CLOUD)
	module.LoadModules(util.CLOUD)
	module.ShutDown()
}

func Test_TcpClient(t *testing.T) {
	os.Setenv(util.NODE_NAME_ENV, "node1")
	err := conf.InitConf(util.EDGE, "../../../../../conf/edge_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream client configuration file err = %v", err)
		return
	}
	module.InitModules(util.EDGE)
	InitTcp()
	stream.InitStream(util.EDGE)
	module.LoadModules(util.EDGE)
	module.ShutDown()
}
