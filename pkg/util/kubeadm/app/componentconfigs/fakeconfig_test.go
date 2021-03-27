/*
Copyright 2020 The Kubernetes Authors.

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

package componentconfigs

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lithammer/dedent"

	kubeadmapi "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm"
	kubeadmscheme "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1 "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/kubeadm/v1beta2"
	outputapi "github.com/superedge/superedge/pkg/util/kubeadm/app/apis/output"
	"github.com/superedge/superedge/pkg/util/kubeadm/app/constants"
	kubeadmutil "github.com/superedge/superedge/pkg/util/kubeadm/app/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	clientsetfake "k8s.io/client-go/kubernetes/fake"
)

// All tests in this file use an alternative set of `known` component configs.
// In this case it's just one known config and it's kubeadm's very own ClusterConfiguration.
// ClusterConfiguration is normally not managed by this package. It's only used, because of the following:
// - It's a versioned API that is under the control of kubeadm maintainers. This enables us to test
//   the componentconfigs package more thoroughly without having to have full and always up to date
//   knowledge about the config of another component.
// - Other components often introduce new fields in their configs without bumping up the config version.
//   This, often times, requires that the PR that introduces such new fields to touch kubeadm test code.
//   Doing so, requires more work on the part of developers and reviewers. When kubeadm moves out of k/k
//   this would allow for more sporadic breaks in kubeadm tests as PRs that merge in k/k and introduce
//   new fields won't be able to fix the tests in kubeadm.
// - If we implement tests for all common functionality using the config of another component and it gets
//   deprecated and/or we stop supporting it in production, we'll have to focus on a massive test refactoring
//   or just continue importing this config just for test use.
//
// Thus, to reduce maintenance costs without sacrificing test coverage, we introduce this mini-framework
// and set of tests here which replace the normal component configs with a single one (ClusterConfiguration)
// and test the component config independent logic of this package.

// clusterConfigHandler is the handler instance for the latest supported ClusterConfiguration to be used in tests
var clusterConfigHandler = handler{
	GroupVersion: kubeadmapiv1.SchemeGroupVersion,
	AddToScheme:  kubeadmapiv1.AddToScheme,
	CreateEmpty: func() kubeadmapi.ComponentConfig {
		return &clusterConfig{
			configBase: configBase{
				GroupVersion: kubeadmapiv1.SchemeGroupVersion,
			},
		}
	},
	fromCluster: clusterConfigFromCluster,
}

func clusterConfigFromCluster(h *handler, clientset clientset.Interface, _ *kubeadmapi.ClusterConfiguration) (kubeadmapi.ComponentConfig, error) {
	return h.fromConfigMap(clientset, constants.KubeadmConfigConfigMap, constants.ClusterConfigurationConfigMapKey, true)
}

type clusterConfig struct {
	configBase
	config kubeadmapiv1.ClusterConfiguration
}

func (cc *clusterConfig) DeepCopy() kubeadmapi.ComponentConfig {
	result := &clusterConfig{}
	cc.configBase.DeepCopyInto(&result.configBase)
	cc.config.DeepCopyInto(&result.config)
	return result
}

func (cc *clusterConfig) Marshal() ([]byte, error) {
	return cc.configBase.Marshal(&cc.config)
}

func (cc *clusterConfig) Unmarshal(docmap kubeadmapi.DocumentMap) error {
	return cc.configBase.Unmarshal(docmap, &cc.config)
}

func (cc *clusterConfig) Default(_ *kubeadmapi.ClusterConfiguration, _ *kubeadmapi.APIEndpoint, _ *kubeadmapi.NodeRegistrationOptions) {
	cc.config.ClusterName = "foo"
	cc.config.KubernetesVersion = "bar"
}

// fakeKnown replaces temporarily during the execution of each test here known (in configset.go)
var fakeKnown = []*handler{
	&clusterConfigHandler,
}

// fakeKnownContext is the func that houses the fake component config context.
// NOTE: It does not support concurrent test execution!
func fakeKnownContext(f func()) {
	// Save the real values
	realKnown := known
	realScheme := Scheme
	realCodecs := Codecs

	// Replace the context with the fake context
	known = fakeKnown
	Scheme = kubeadmscheme.Scheme
	Codecs = kubeadmscheme.Codecs

	// Upon function exit, restore the real values
	defer func() {
		known = realKnown
		Scheme = realScheme
		Codecs = realCodecs
	}()

	// Call f in the fake context
	f()
}

// testClusterConfigMap is a short hand for creating and possibly signing a test config map.
// This produces config maps that can be loaded by clusterConfigFromCluster
func testClusterConfigMap(yaml string, signIt bool) *v1.ConfigMap {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.KubeadmConfigConfigMap,
			Namespace: metav1.NamespaceSystem,
		},
		Data: map[string]string{
			constants.ClusterConfigurationConfigMapKey: dedent.Dedent(yaml),
		},
	}

	if signIt {
		SignConfigMap(cm)
	}

	return cm
}

// oldClusterConfigVersion is used as an old unsupported version in tests throughout this file
const oldClusterConfigVersion = "v1alpha1"

var (
	// currentClusterConfigVersion represents the current actively supported version of ClusterConfiguration
	currentClusterConfigVersion = kubeadmapiv1.SchemeGroupVersion.Version

	// currentFooClusterConfig is a minimal currently supported ClusterConfiguration
	// with a well known value of clusterName (in this case `foo`)
	currentFooClusterConfig = fmt.Sprintf(`
		apiVersion: %s
		kind: ClusterConfiguration
		clusterName: foo
	`, kubeadmapiv1.SchemeGroupVersion)

	// oldFooClusterConfig is a minimal unsupported ClusterConfiguration
	// with a well known value of clusterName (in this case `foo`)
	oldFooClusterConfig = fmt.Sprintf(`
		apiVersion: %s/%s
		kind: ClusterConfiguration
		clusterName: foo
	`, kubeadmapiv1.GroupName, oldClusterConfigVersion)

	// currentBarClusterConfig is a minimal currently supported ClusterConfiguration
	// with a well known value of clusterName (in this case `bar`)
	currentBarClusterConfig = fmt.Sprintf(`
		apiVersion: %s
		kind: ClusterConfiguration
		clusterName: bar
	`, kubeadmapiv1.SchemeGroupVersion)

	// oldBarClusterConfig is a minimal unsupported ClusterConfiguration
	// with a well known value of clusterName (in this case `bar`)
	oldBarClusterConfig = fmt.Sprintf(`
		apiVersion: %s/%s
		kind: ClusterConfiguration
		clusterName: bar
	`, kubeadmapiv1.GroupName, oldClusterConfigVersion)

	// This is the "minimal" valid config that can be unmarshalled to and from YAML.
	// Due to same static defaulting it's not exactly small in size.
	validUnmarshallableClusterConfig = struct {
		yaml string
		obj  kubeadmapiv1.ClusterConfiguration
	}{
		yaml: dedent.Dedent(`
			apiServer:
			  timeoutForControlPlane: 4m
			apiVersion: kubeadm.k8s.io/v1beta2
			certificatesDir: /etc/kubernetes/pki
			clusterName: LeCluster
			controllerManager: {}
			dns:
			  type: CoreDNS
			etcd:
			  local:
			    dataDir: /var/lib/etcd
			imageRepository: k8s.gcr.io
			kind: ClusterConfiguration
			kubernetesVersion: 1.2.3
			networking:
			  dnsDomain: cluster.local
			  serviceSubnet: 10.96.0.0/12
			scheduler: {}
		`),
		obj: kubeadmapiv1.ClusterConfiguration{
			TypeMeta: metav1.TypeMeta{
				APIVersion: kubeadmapiv1.SchemeGroupVersion.String(),
				Kind:       "ClusterConfiguration",
			},
			ClusterName:       "LeCluster",
			KubernetesVersion: "1.2.3",
			CertificatesDir:   "/etc/kubernetes/pki",
			ImageRepository:   "k8s.gcr.io",
			Networking: kubeadmapiv1.Networking{
				DNSDomain:     "cluster.local",
				ServiceSubnet: "10.96.0.0/12",
			},
			DNS: kubeadmapiv1.DNS{
				Type: kubeadmapiv1.CoreDNS,
			},
			Etcd: kubeadmapiv1.Etcd{
				Local: &kubeadmapiv1.LocalEtcd{
					DataDir: "/var/lib/etcd",
				},
			},
			APIServer: kubeadmapiv1.APIServer{
				TimeoutForControlPlane: &metav1.Duration{
					Duration: 4 * time.Minute,
				},
			},
		},
	}
)

func TestConfigBaseMarshal(t *testing.T) {
	fakeKnownContext(func() {
		cfg := &clusterConfig{
			configBase: configBase{
				GroupVersion: kubeadmapiv1.SchemeGroupVersion,
			},
			config: kubeadmapiv1.ClusterConfiguration{
				TypeMeta: metav1.TypeMeta{
					APIVersion: kubeadmapiv1.SchemeGroupVersion.String(),
					Kind:       "ClusterConfiguration",
				},
				ClusterName:       "LeCluster",
				KubernetesVersion: "1.2.3",
			},
		}

		b, err := cfg.Marshal()
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		got := strings.TrimSpace(string(b))
		expected := strings.TrimSpace(dedent.Dedent(`
			apiServer: {}
			apiVersion: kubeadm.k8s.io/v1beta2
			clusterName: LeCluster
			controllerManager: {}
			dns:
			  type: ""
			etcd: {}
			kind: ClusterConfiguration
			kubernetesVersion: 1.2.3
			networking: {}
			scheduler: {}
		`))

		if expected != got {
			t.Fatalf("Missmatch between expected and got:\nExpected:\n%s\n---\nGot:\n%s", expected, got)
		}
	})
}

func TestConfigBaseUnmarshal(t *testing.T) {
	fakeKnownContext(func() {
		expected := &clusterConfig{
			configBase: configBase{
				GroupVersion: kubeadmapiv1.SchemeGroupVersion,
			},
			config: validUnmarshallableClusterConfig.obj,
		}

		gvkmap, err := kubeadmutil.SplitYAMLDocuments([]byte(validUnmarshallableClusterConfig.yaml))
		if err != nil {
			t.Fatalf("unexpected failure of SplitYAMLDocuments: %v", err)
		}

		got := &clusterConfig{
			configBase: configBase{
				GroupVersion: kubeadmapiv1.SchemeGroupVersion,
			},
		}
		if err = got.Unmarshal(gvkmap); err != nil {
			t.Fatalf("unexpected failure of Unmarshal: %v", err)
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("Missmatch between expected and got:\nExpected:\n%v\n---\nGot:\n%v", expected, got)
		}
	})
}

func TestGeneratedConfigFromCluster(t *testing.T) {
	fakeKnownContext(func() {
		testYAML := dedent.Dedent(`
			apiVersion: kubeadm.k8s.io/v1beta2
			kind: ClusterConfiguration
		`)
		testYAMLHash := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(testYAML)))
		// The SHA256 sum of "The quick brown fox jumps over the lazy dog"
		const mismatchHash = "sha256:d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"
		tests := []struct {
			name         string
			hash         string
			userSupplied bool
		}{
			{
				name: "Matching hash means generated config",
				hash: testYAMLHash,
			},
			{
				name:         "Missmatching hash means user supplied config",
				hash:         mismatchHash,
				userSupplied: true,
			},
			{
				name:         "No hash means user supplied config",
				userSupplied: true,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				configMap := testClusterConfigMap(testYAML, false)
				if test.hash != "" {
					configMap.Annotations = map[string]string{
						constants.ComponentConfigHashAnnotationKey: test.hash,
					}
				}

				client := clientsetfake.NewSimpleClientset(configMap)
				cfg, err := clusterConfigHandler.FromCluster(client, testClusterCfg())
				if err != nil {
					t.Fatalf("unexpected failure of FromCluster: %v", err)
				}

				got := cfg.IsUserSupplied()
				if got != test.userSupplied {
					t.Fatalf("mismatch between expected and got:\n\tExpected: %t\n\tGot: %t", test.userSupplied, got)
				}
			})
		}
	})
}

// runClusterConfigFromTest holds common test case data and evaluation code for handler.From* functions
func runClusterConfigFromTest(t *testing.T, perform func(t *testing.T, in string) (kubeadmapi.ComponentConfig, error)) {
	fakeKnownContext(func() {
		tests := []struct {
			name      string
			in        string
			out       *clusterConfig
			expectErr bool
		}{
			{
				name: "Empty document map should return nothing successfully",
			},
			{
				name: "Non-empty document map without the proper API group returns nothing successfully",
				in: dedent.Dedent(`
					apiVersion: api.example.com/v1
					kind: Configuration
				`),
			},
			{
				name: "Old config version returns an error",
				in: dedent.Dedent(`
					apiVersion: kubeadm.k8s.io/v1alpha1
					kind: ClusterConfiguration
				`),
				expectErr: true,
			},
			{
				name: "Unknown kind returns an error",
				in: dedent.Dedent(`
					apiVersion: kubeadm.k8s.io/v1beta2
					kind: Configuration
				`),
				expectErr: true,
			},
			{
				name: "Valid config gets loaded",
				in:   validUnmarshallableClusterConfig.yaml,
				out: &clusterConfig{
					configBase: configBase{
						GroupVersion: clusterConfigHandler.GroupVersion,
						userSupplied: true,
					},
					config: validUnmarshallableClusterConfig.obj,
				},
			},
			{
				name: "Valid config gets loaded even if coupled with an extra document",
				in:   "apiVersion: api.example.com/v1\nkind: Configuration\n---\n" + validUnmarshallableClusterConfig.yaml,
				out: &clusterConfig{
					configBase: configBase{
						GroupVersion: clusterConfigHandler.GroupVersion,
						userSupplied: true,
					},
					config: validUnmarshallableClusterConfig.obj,
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				componentCfg, err := perform(t, test.in)
				if err != nil {
					if !test.expectErr {
						t.Errorf("unexpected failure: %v", err)
					}
				} else {
					if test.expectErr {
						t.Error("unexpected success")
					} else {
						if componentCfg == nil {
							if test.out != nil {
								t.Error("unexpected nil result")
							}
						} else {
							if got, ok := componentCfg.(*clusterConfig); !ok {
								t.Error("different result type")
							} else {
								if test.out == nil {
									t.Errorf("unexpected result: %v", got)
								} else {
									if !reflect.DeepEqual(test.out, got) {
										t.Errorf("missmatch between expected and got:\nExpected:\n%v\n---\nGot:\n%v", test.out, got)
									}
								}
							}
						}
					}
				}
			})
		}
	})
}

func TestLoadingFromDocumentMap(t *testing.T) {
	runClusterConfigFromTest(t, func(t *testing.T, in string) (kubeadmapi.ComponentConfig, error) {
		gvkmap, err := kubeadmutil.SplitYAMLDocuments([]byte(in))
		if err != nil {
			t.Fatalf("unexpected failure of SplitYAMLDocuments: %v", err)
		}

		return clusterConfigHandler.FromDocumentMap(gvkmap)
	})
}

func TestLoadingFromCluster(t *testing.T) {
	runClusterConfigFromTest(t, func(t *testing.T, in string) (kubeadmapi.ComponentConfig, error) {
		client := clientsetfake.NewSimpleClientset(
			testClusterConfigMap(in, false),
		)

		return clusterConfigHandler.FromCluster(client, testClusterCfg())
	})
}

func TestFetchFromClusterWithLocalOverwrites(t *testing.T) {
	fakeKnownContext(func() {
		cases := []struct {
			desc          string
			obj           runtime.Object
			config        string
			expectedValue string
			isNotLoaded   bool
			expectedErr   bool
		}{
			{
				desc:          "appropriate cluster object without overwrite is used",
				obj:           testClusterConfigMap(currentFooClusterConfig, false),
				expectedValue: "foo",
			},
			{
				desc:          "appropriate cluster object with appropriate overwrite is overwritten",
				obj:           testClusterConfigMap(currentFooClusterConfig, false),
				config:        dedent.Dedent(currentBarClusterConfig),
				expectedValue: "bar",
			},
			{
				desc:        "appropriate cluster object with old overwrite returns an error",
				obj:         testClusterConfigMap(currentFooClusterConfig, false),
				config:      dedent.Dedent(oldBarClusterConfig),
				expectedErr: true,
			},
			{
				desc:        "old config without overwrite returns an error",
				obj:         testClusterConfigMap(oldFooClusterConfig, false),
				expectedErr: true,
			},
			{
				desc:          "old config with appropriate overwrite returns the substitute",
				obj:           testClusterConfigMap(oldFooClusterConfig, false),
				config:        dedent.Dedent(currentBarClusterConfig),
				expectedValue: "bar",
			},
			{
				desc:        "old config with old overwrite returns an error",
				obj:         testClusterConfigMap(oldFooClusterConfig, false),
				config:      dedent.Dedent(oldBarClusterConfig),
				expectedErr: true,
			},
			{
				desc:          "appropriate signed cluster object without overwrite is used",
				obj:           testClusterConfigMap(currentFooClusterConfig, true),
				expectedValue: "foo",
			},
			{
				desc:          "appropriate signed cluster object with appropriate overwrite is overwritten",
				obj:           testClusterConfigMap(currentFooClusterConfig, true),
				config:        dedent.Dedent(currentBarClusterConfig),
				expectedValue: "bar",
			},
			{
				desc:        "appropriate signed cluster object with old overwrite returns an error",
				obj:         testClusterConfigMap(currentFooClusterConfig, true),
				config:      dedent.Dedent(oldBarClusterConfig),
				expectedErr: true,
			},
			{
				desc:        "old signed config without an overwrite is not loaded",
				obj:         testClusterConfigMap(oldFooClusterConfig, true),
				isNotLoaded: true,
			},
			{
				desc:          "old signed config with appropriate overwrite returns the substitute",
				obj:           testClusterConfigMap(oldFooClusterConfig, true),
				config:        dedent.Dedent(currentBarClusterConfig),
				expectedValue: "bar",
			},
			{
				desc:        "old signed config with old overwrite returns an error",
				obj:         testClusterConfigMap(oldFooClusterConfig, true),
				config:      dedent.Dedent(oldBarClusterConfig),
				expectedErr: true,
			},
		}

		for _, test := range cases {
			t.Run(test.desc, func(t *testing.T) {
				client := clientsetfake.NewSimpleClientset(test.obj)

				docmap, err := kubeadmutil.SplitYAMLDocuments([]byte(test.config))
				if err != nil {
					t.Fatalf("unexpected failure of SplitYAMLDocuments: %v", err)
				}

				clusterCfg := testClusterCfg()

				err = FetchFromClusterWithLocalOverwrites(clusterCfg, client, docmap)
				if err != nil {
					if !test.expectedErr {
						t.Errorf("unexpected failure: %v", err)
					}
				} else {
					if test.expectedErr {
						t.Error("unexpected success")
					} else {
						clusterCfg, ok := clusterCfg.ComponentConfigs[kubeadmapiv1.GroupName]
						if !ok {
							if !test.isNotLoaded {
								t.Error("no config was loaded when it should have been")
							}
						} else {
							actualConfig, ok := clusterCfg.(*clusterConfig)
							if !ok {
								t.Error("the config is not of the expected type")
							} else if actualConfig.config.ClusterName != test.expectedValue {
								t.Errorf("unexpected value:\n\tgot: %q\n\texpected: %q", actualConfig.config.ClusterName, test.expectedValue)
							}
						}
					}
				}
			})
		}
	})
}

func TestGetVersionStates(t *testing.T) {
	fakeKnownContext(func() {
		versionStateCurrent := outputapi.ComponentConfigVersionState{
			Group:            kubeadmapiv1.GroupName,
			CurrentVersion:   currentClusterConfigVersion,
			PreferredVersion: currentClusterConfigVersion,
		}
		versionStateOld := outputapi.ComponentConfigVersionState{
			Group:                 kubeadmapiv1.GroupName,
			CurrentVersion:        oldClusterConfigVersion,
			PreferredVersion:      currentClusterConfigVersion,
			ManualUpgradeRequired: true,
		}

		cases := []struct {
			desc     string
			obj      runtime.Object
			config   string
			expected outputapi.ComponentConfigVersionState
		}{
			{
				desc:     "appropriate cluster object without overwrite",
				obj:      testClusterConfigMap(currentFooClusterConfig, false),
				expected: versionStateCurrent,
			},
			{
				desc:     "appropriate cluster object with appropriate overwrite",
				obj:      testClusterConfigMap(currentFooClusterConfig, false),
				config:   dedent.Dedent(currentBarClusterConfig),
				expected: versionStateCurrent,
			},
			{
				desc:     "appropriate cluster object with old overwrite",
				obj:      testClusterConfigMap(currentFooClusterConfig, false),
				config:   dedent.Dedent(oldBarClusterConfig),
				expected: versionStateOld,
			},
			{
				desc:     "old config without overwrite returns an error",
				obj:      testClusterConfigMap(oldFooClusterConfig, false),
				expected: versionStateOld,
			},
			{
				desc:     "old config with appropriate overwrite",
				obj:      testClusterConfigMap(oldFooClusterConfig, false),
				config:   dedent.Dedent(currentBarClusterConfig),
				expected: versionStateCurrent,
			},
			{
				desc:     "old config with old overwrite",
				obj:      testClusterConfigMap(oldFooClusterConfig, false),
				config:   dedent.Dedent(oldBarClusterConfig),
				expected: versionStateOld,
			},
			{
				desc:     "appropriate signed cluster object without overwrite",
				obj:      testClusterConfigMap(currentFooClusterConfig, true),
				expected: versionStateCurrent,
			},
			{
				desc:     "appropriate signed cluster object with appropriate overwrite",
				obj:      testClusterConfigMap(currentFooClusterConfig, true),
				config:   dedent.Dedent(currentBarClusterConfig),
				expected: versionStateCurrent,
			},
			{
				desc:     "appropriate signed cluster object with old overwrit",
				obj:      testClusterConfigMap(currentFooClusterConfig, true),
				config:   dedent.Dedent(oldBarClusterConfig),
				expected: versionStateOld,
			},
			{
				desc: "old signed config without an overwrite",
				obj:  testClusterConfigMap(oldFooClusterConfig, true),
				expected: outputapi.ComponentConfigVersionState{
					Group:            kubeadmapiv1.GroupName,
					CurrentVersion:   "", // The config is treated as if it's missing
					PreferredVersion: currentClusterConfigVersion,
				},
			},
			{
				desc:     "old signed config with appropriate overwrite",
				obj:      testClusterConfigMap(oldFooClusterConfig, true),
				config:   dedent.Dedent(currentBarClusterConfig),
				expected: versionStateCurrent,
			},
			{
				desc:     "old signed config with old overwrite",
				obj:      testClusterConfigMap(oldFooClusterConfig, true),
				config:   dedent.Dedent(oldBarClusterConfig),
				expected: versionStateOld,
			},
		}

		for _, test := range cases {
			t.Run(test.desc, func(t *testing.T) {
				client := clientsetfake.NewSimpleClientset(test.obj)

				docmap, err := kubeadmutil.SplitYAMLDocuments([]byte(test.config))
				if err != nil {
					t.Fatalf("unexpected failure of SplitYAMLDocuments: %v", err)
				}

				clusterCfg := testClusterCfg()

				got, err := GetVersionStates(clusterCfg, client, docmap)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if len(got) != 1 {
					t.Errorf("got %d, but expected only a single result: %v", len(got), got)
				} else if got[0] != test.expected {
					t.Errorf("unexpected result:\n\texpected: %v\n\tgot: %v", test.expected, got[0])
				}
			})
		}
	})
}
