module superedge

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/golang/protobuf v1.3.3
	github.com/grpc-ecosystem/grpc-gateway v1.14.6 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/moby/term v0.0.0-20200312100748-672ec06f55cd
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	golang.org/x/tools v0.0.0-20200701221012-f01a4bec33ec // indirect
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/apiserver v0.18.5
	k8s.io/client-go v0.18.5
	k8s.io/code-generator v0.18.5
	k8s.io/component-base v0.18.5
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/moby/term => github.com/moby/term v0.0.0-20200312100748-672ec06f55cd
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20200329194405-dd816f0735f8
)
