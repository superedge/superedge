package kubeclient

import (
	"testing"
)

func TestK8sVerisonInt(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name       string
		args       args
		k8sversion int
	}{
		{"test01", args{"v1.8.7"}, 10807},
		{"test02", args{"v0.8.0"}, 800},
		{"test03", args{"v1.18.2"}, 11802},
		{"test04", args{"v1.20.02"}, 12002},
		{"test05", args{"v1.16.07"}, 11607},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sVersionInt, err := k8sVersionInt(tt.args.version)
			if err != nil {
				t.Errorf("k8sVerisonInt testName: %s, intPut: %s, error: %v", tt.name, tt.args, err)
			}
			if k8sVersionInt != tt.k8sversion {
				t.Errorf("k8sVerisonInt testname: %s, "+
					"inPut: %s, want: %d, res: %d", tt.name, tt.args, tt.k8sversion, k8sVersionInt)
			}
		})
	}
}
