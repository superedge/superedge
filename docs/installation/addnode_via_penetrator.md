# Use Penetrator to add edge nodes through the cloud

## 1. Use edgeadm to build a SuperEdge Kubernetes edge cluster

How to build:[One-click install of edge Kubernetes cluster](../../README.md)

## 2. Deploy Penetrator

Deploy directly using [penetrator.yaml](../../deployment/penetrator.yaml)

```shell
kubectl apply -f https://raw.githubusercontent.com/superedge/superedge/main/deployment/penetrator.yaml
```

## 3. Preconditions for operating nodes

Use SSH password file passwd to create sshCredential

```shell
kubectl -n edge-system create secret generic login-secret --from-file=passwd=./passwd 
```

Or use SSH private key file sshkey to create sshCredential

```shell
kubectl -n edge-system create secret generic login-secret --from-file=sshkey=./sshkey 
```

## 4.1 Install node

```yaml
apiVersion: nodetask.apps.superedge.io/v1beta1
kind: NodeTask
metadata:
  name: nodes
spec:
  nodeNamePrefix: "edge"
  targetMachines:
    - 172.21.3.194
    ...
  sshCredential: login-secret
  proxyNode: vm-2-117-centos
```

* nodeNamePrefix: node name prefix, the format of node name: nodeNamePrefix-random string (6 digits)
* targetMachines: the ip list of the node to be installed
* sshCredential: Stores the secret of the password (passwd) and private key (sshkey) for SSH login to the node to be
  added. The key value of the password file must be passwd, and the key value of the private key file must be sshkey
* proxyNode: The node name of the node in the cluster that executes the job to add the node. The node and the node to be
  installed are on the same intranet (SSH can log in to the node to be installed)

## 4.2 Reinstall the node

```yaml
apiVersion: nodetask.apps.superedge.io/v1beta1
kind: NodeTask
metadata:
  name: nodes
spec:
  nodeNamesOverride:
    edge-1mokvl: 172.21.3.194
    ...
  sshCredential: login-secret
  proxyNode: vm-2-117-centos
```

* nodeNamesOverride: Reinstall the node name and IP of the node

## 5. Status query

The Status of NodeTask contains the execution status (creating and ready) of the task and the node name and IP of the
node that has not been installed. You can use the command to view:

```shell
kubectl get nt NodeTaskName -o custom-columns='STATUS:status.nodetaskStatus' 
```

The success and error information of the task in the execution process is reported to the apiserver in the form of an
event, and you can use the command to view:

```shell
kubectl -n edge-system get event
```