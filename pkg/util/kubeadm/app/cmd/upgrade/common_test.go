/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package upgrade

import (
	"bytes"
	"testing"

	kubeadmapi "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm"
)

func TestEnforceRequirements(t *testing.T) {
	tcases := []struct {
		name          string
		newK8sVersion string
		dryRun        bool
		flags         applyPlanFlags
		expectedErr   bool
	}{
		{
			name:        "Fail pre-flight check",
			expectedErr: true,
		},
		{
			name: "Bogus preflight check disabled when also 'all' is specified",
			flags: applyPlanFlags{
				ignorePreflightErrors: []string{"bogusvalue", "all"},
			},
			expectedErr: true,
		},
		{
			name: "Fail to create client",
			flags: applyPlanFlags{
				ignorePreflightErrors: []string{"all"},
			},
			expectedErr: true,
		},
	}
	for _, tt := range tcases {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := enforceRequirements(&tt.flags, nil, tt.dryRun, false)

			if err == nil && tt.expectedErr {
				t.Error("Expected error, but got success")
			}
			if err != nil && !tt.expectedErr {
				t.Errorf("Unexpected error: %+v", err)
			}
		})
	}
}

func TestPrintConfiguration(t *testing.T) {
	var tests = []struct {
		name          string
		cfg           *kubeadmapi.ClusterConfiguration
		buf           *bytes.Buffer
		expectedBytes []byte
	}{
		{
			name:          "config is nil",
			cfg:           nil,
			expectedBytes: []byte(""),
		},
		{
			name: "cluster config with local Etcd",
			cfg: &kubeadmapi.ClusterConfiguration{
				KubernetesVersion: "v1.7.1",
				Etcd: kubeadmapi.Etcd{
					Local: &kubeadmapi.LocalEtcd{
						DataDir: "/some/path",
					},
				},
				DNS: kubeadmapi.DNS{
					Type: kubeadmapi.CoreDNS,
				},
			},
			expectedBytes: []byte(`[upgrade/config] Configuration used:
	apiServer: {}
	apiVersion: kubeadm.k8s.io/v1beta2
	controllerManager: {}
	dns:
	  type: CoreDNS
	etcd:
	  local:
	    dataDir: /some/path
	kind: ClusterConfiguration
	kubernetesVersion: v1.7.1
	networking: {}
	scheduler: {}
`),
		},
		{
			name: "cluster config with ServiceSubnet and external Etcd",
			cfg: &kubeadmapi.ClusterConfiguration{
				KubernetesVersion: "v1.7.1",
				Networking: kubeadmapi.Networking{
					ServiceSubnet: "10.96.0.1/12",
				},
				Etcd: kubeadmapi.Etcd{
					External: &kubeadmapi.ExternalEtcd{
						Endpoints: []string{"https://one-etcd-instance:2379"},
					},
				},
				DNS: kubeadmapi.DNS{
					Type: kubeadmapi.CoreDNS,
				},
			},
			expectedBytes: []byte(`[upgrade/config] Configuration used:
	apiServer: {}
	apiVersion: kubeadm.k8s.io/v1beta2
	controllerManager: {}
	dns:
	  type: CoreDNS
	etcd:
	  external:
	    caFile: ""
	    certFile: ""
	    endpoints:
	    - https://one-etcd-instance:2379
	    keyFile: ""
	kind: ClusterConfiguration
	kubernetesVersion: v1.7.1
	networking:
	  serviceSubnet: 10.96.0.1/12
	scheduler: {}
`),
		},
	}
	for _, rt := range tests {
		t.Run(rt.name, func(t *testing.T) {
			rt.buf = bytes.NewBufferString("")
			printConfiguration(rt.cfg, rt.buf)
			actualBytes := rt.buf.Bytes()
			if !bytes.Equal(actualBytes, rt.expectedBytes) {
				t.Errorf(
					"failed PrintConfiguration:\n\texpected: %q\n\t  actual: %q",
					string(rt.expectedBytes),
					string(actualBytes),
				)
			}
		})
	}
}
