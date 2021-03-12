package tcp

import (
	"github.com/superedge/superedge/pkg/tunnel/conf"
	"github.com/superedge/superedge/pkg/tunnel/model"
	"github.com/superedge/superedge/pkg/tunnel/proxy/stream"
	"github.com/superedge/superedge/pkg/tunnel/util"
	"os"
	"testing"
)

package tcp

import (
"github.com/superedge/superedge/pkg/tunnel/conf"
"github.com/superedge/superedge/pkg/tunnel/model"
"github.com/superedge/superedge/pkg/tunnel/proxy/stream"
"github.com/superedge/superedge/pkg/tunnel/util"
"os"
"testing"
)

func Test_TcpServer(t *testing.T) {
	err := conf.InitConf(util.CLOUD, "../../../../conf/cloud_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream server configuration file err = %v", err)
		return
	}
	model.InitModules(util.CLOUD)
	InitTcp()
	stream.InitStream(util.CLOUD)
	model.LoadModules(util.CLOUD)
	model.ShutDown()
}

func Test_TcpClient(t *testing.T) {
	os.Setenv(util.NODE_NAME_ENV, "node1")
	err := conf.InitConf(util.EDGE, "../../../../conf/edge_mode.toml")
	if err != nil {
		t.Errorf("failed to initialize stream client configuration file err = %v", err)
		return
	}
	model.InitModules(util.EDGE)
	InitTcp()
	stream.InitStream(util.EDGE)
	model.LoadModules(util.EDGE)
	model.ShutDown()
}
