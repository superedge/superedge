package constant

const (
	KubeConfigPath        = "/root/.kube/config"
	LiteAPIServerConfFile = SystemServiceDir + "lite-apiserver.service"
)

const LiteAPIServerTemplate = `
[Unit]
Description=lite-apiserver

[Service]
Environment=QCLOUD_NORM_URL=
ExecStart=/usr/bin/lite-apiserver \
--ca-file=/etc/kubernetes/pki/lite-apiserver-ca.crt \
--tls-cert-file=/etc/kubernetes/edge/lite-apiserver.crt \
--tls-private-key-file=/etc/kubernetes/edge/lite-apiserver.key \
--kube-apiserver-url=${MASTER_IP} \
--kube-apiserver-port=${MASTER_PORT} \
--port=51003 \
--tls-config-file=/etc/kubernetes/edge/tls.json \
--file-cache-path=/data/lite-apiserver/cache \
--sync-duration=120 \
--timeout=3 \
--v=4
Restart=always
RestartSec=10
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
`
