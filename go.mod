module github.com/superedge/superedge

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/TarsCloud/TarsGo v1.1.7-0.20210519070617-e2dc06be2b59
	github.com/caddyserver/caddy v1.0.5 // indirect
	github.com/cockroachdb/pebble v0.0.0-20220218191007-13f8f7cee6ef
	github.com/davecgh/go-spew v1.1.1
	github.com/dgraph-io/badger/v3 v3.2011.1
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.1.2
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.6 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/lithammer/dedent v1.1.0
	github.com/moby/term v0.0.0-20200312100748-672ec06f55cd
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
	github.com/pelletier/go-toml v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/google/flatbuffers v1.12.1
	github.com/tarscloud/gopractice v1.0.1
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20210909193231-528a39cd75f3
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/apiserver v0.20.5
	k8s.io/cli-runtime v0.20.5
	k8s.io/client-go v0.22.2
	k8s.io/cluster-bootstrap v0.0.0
	k8s.io/code-generator v0.20.5
	k8s.io/component-base v0.22.2
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.8.0
	k8s.io/kubernetes v1.19.14
	k8s.io/system-validators v1.4.0 // indirect
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/coredns/corefile-migration => github.com/coredns/corefile-migration v1.0.10
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	github.com/moby/term => github.com/moby/term v0.0.0-20200312100748-672ec06f55cd
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200329194405-dd816f0735f8
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	k8s.io/api => k8s.io/api v0.20.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.5
	k8s.io/apiserver => k8s.io/apiserver v0.20.5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.5
	k8s.io/client-go => k8s.io/client-go v0.20.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.5
	k8s.io/code-generator => k8s.io/code-generator v0.20.5
	k8s.io/component-base => k8s.io/component-base v0.20.5
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.5
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.5
	k8s.io/cri-api => k8s.io/cri-api v0.20.5
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.5
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.5
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.5
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.5
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.5
	k8s.io/kubectl => k8s.io/kubectl v0.20.5
	k8s.io/kubelet => k8s.io/kubelet v0.20.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.5
	k8s.io/metrics => k8s.io/metrics v0.20.5
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.5
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.5
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.5
	k8s.io/sample-controller => k8s.io/sample-controller v0.20.5
)
