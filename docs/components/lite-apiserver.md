# lite-apiserver

`lite-apiserver` is a light-weight version of the kube-apiserver running on the edge nodes. It acts as a proxy for requests from all components and pods on the edge node to the cloud apiserver, and caches the responses to achieve edge autonomy in case of the disconnected cloud -edge network.

`lite-apiserver` has the following functionalities:
- Caches request for all edge components (kubelet, kube-proxy, etc.) and pods running on edge nodes
- Provides various authentication mechanism for edge components and pods, including X509 Client Certs, Bearer Token, etc. Support X509 Client Certs rotation
- Caches all kind of Kubernetes resources, including build-in Kubernetes resources and custom resources
- Support multiple cache storage, including file, kv storage(bolt, badger)

## Architecture
<div align="left">
  <img src="../img/lite-apiserver.png" width=70% title="lite-apiserver Architecture">
</div>

Overall, `lite-apiserver` start a HTTPS Server  to accepting the request of all Client (HTTPS request). According to the `Common Name` of TLS certificate, use corresponding ReverseProxy to forwarding the request to kube-apiserver (if not mtls certificate request, using default). When the cloud-edge network is normal, the corresponding https response is returned to the client, and it is asynchronously stored in the cache on demand; when the cloud-edge is disconnected, the request to kube-apiserver times out, `lite-apiserver` query cache and return the cache data to client, to achieve the purpose of edge autonomy.

## Usage
`lite-apiserver` can be run at the edge as Kubernetes pod or systemd service. See [**Installation Guide**](../installation) to get more detail.

## Demo
1. Installing `lite-apiserver`
2. applying the following yaml to running echoserver
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lite-demo
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: echo
  template:
    metadata:
      labels:
        app: echo
    spec:
      containers:
      - image: superedge/echoserver:2.2
        name: echo
        ports:
        - containerPort: 8080
          protocol: TCP
        env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          - name: POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: POD_IP
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
```
3. Accessing the echoserver，the result is successful
```bash
$ kubectl get pods -owide
NAME                        READY   STATUS    RESTARTS   AGE     IP         NODE           NOMINATED NODE   READINESS GATES
lite-demo-c7b458ddc-6lpnx   1/1     Running   0          4m23s   10.0.6.2   ecm-q5hx6hhd   <none>           <none>

$ curl http://10.0.6.2:8080 | grep pod
	pod name:	lite-demo-c7b458ddc-6lpnx
	pod namespace:	default
	pod IP:	10.0.6.2
```
4. Disconnecting the network between the node which echoserver running on and kube-apiserver, the node is autonomous.
5. Accessing the echoserver，the result is successful
```bash
$ curl http://10.0.6.2:8080 | grep pod
	pod name:	lite-demo-c7b458ddc-6lpnx
	pod namespace:	default
	pod IP:	10.0.6.2
```
6. Rebooting the node, and then accessing echoserver, the result is successful.
```bash
$ curl http://10.0.6.2:8080 | grep pod
	pod name:	lite-demo-c7b458ddc-6lpnx
	pod namespace:	default
	pod IP:	10.0.6.2
```
