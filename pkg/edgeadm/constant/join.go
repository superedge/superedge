package constant

const (
	LiteApiserverConfPath = SystemServiceDir + "lite-apiserver/lite-apiserver.service"
	LiteApiserverBinPath  = InstallBin + "lite-apiserver"
)

const LiteApiserverTemplate = `
[Unit]
Description=lite-apiserver

[Service]
Environment=QCLOUD_NORM_URL=
ExecStart=/usr/local/bin/lite-apiserver \
--ca-file=/etc/kubernetes/pki/ca.crt \
--tls-cert-file=/etc/kubernetes/edge/lite-apiserver.crt \
--tls-private-key-file=/etc/kubernetes/edge/lite-apiserver.key \
--kube-apiserver-url=${MASTER_IP} \
--kube-apiserver-port=6443 \
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
