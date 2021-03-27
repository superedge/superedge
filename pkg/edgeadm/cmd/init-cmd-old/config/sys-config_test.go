package config

import (
	"fmt"
	"github.com/superedge/superedge/pkg/util"
	"testing"
)

func TestNew(t *testing.T) {
	cfg, err := New("./edge_cluster_sys_config.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(util.ToJson(cfg))
}
