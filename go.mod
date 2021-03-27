module github.com/superedge/superedge

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/PuerkitoBio/purell v1.1.1
	github.com/armon/consul-api v0.0.0-20180202201655-eb2c6b5be1b6 // indirect
	github.com/caddyserver/caddy v1.0.5
	github.com/coredns/corefile-migration/migration v1.0.1
	github.com/coreos/go-etcd v2.0.0+incompatible // indirect
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/go-openapi/validate v0.19.5 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/grpc-ecosystem/grpc-gateway v1.14.6 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/jinzhu/configor v1.2.1
	github.com/lithammer/dedent v1.1.0
	github.com/moby/term v0.0.0-20200312100748-672ec06f55cd
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/ugorji/go v1.1.4 // indirect
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 // indirect
	github.com/xordataexchange/crypt v0.0.3-0.20170626215501-b2862e3d0a77 // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20201112073958-5cba982894dd
	golang.org/x/tools v0.0.0-20200701221012-f01a4bec33ec
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.5
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.20.5
	k8s.io/apiserver v0.20.1
	k8s.io/cli-runtime v0.20.5
	k8s.io/client-go v0.20.5
	k8s.io/cluster-bootstrap v0.20.5
	k8s.io/code-generator v0.20.1
	k8s.io/component-base v0.20.5
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.5.0
	k8s.io/kube-proxy v0.20.5
	k8s.io/kubelet v0.20.5
	k8s.io/system-validators v1.4.0
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.2
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/moby/term => github.com/moby/term v0.0.0-20200312100748-672ec06f55cd
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200329194405-dd816f0735f8
)
