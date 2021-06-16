#!/bin/bash
rm -f edgeadm-install
rm -rf ./edgeadm-install
mkdir -p edgeadm-install

tar -zvxf edgeadm-{{.Arch}}-{{.K8sVersion}}.tar.gz -C /root/edgeadm-install/ --strip-components 1
cd edgeadm-install

chmod +x edgeadm

./edgeadm reset -f

./edgeadm join {{.AdvertiseAddress}}:{{.BindPort}} --token {{.AdmToken}} --discovery-token-ca-cert-hash {{.CaHash}} --node-name {{.NodeName}} --install-pkg-path ./kube-linux-{{.Arch}}-{{.K8sVersion}}.tar.gz
