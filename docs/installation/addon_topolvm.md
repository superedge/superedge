

English | [简体中文](./addon_topolvm_CN.md)

# SuperEdge Uses Topolvm To Support Local Persistence

* [SuperEdge Uses Topolvm To Support Local Persistence](#superedge-uses-topolvm-to-support-local-persistence)
   * [1. Background](#1-background)
   * [2. Plan](#2-plan)
   * [3. Deploy Topolvm](#3-deploy-topolvm)
      * [&lt;1&gt;. Preparation conditions](#1-preparation-conditions)
      * [&lt;2&gt;. Install Topolvm](#2-install-topolvm)
   * [4. Verify](#4-verify)
      * [4.1 Local file simulation persistent storage](#41-local-file-simulation-persistent-storage)
      * [4.2 Mount Tencent Cloud Cloud Disk](#42-mount-tencent-cloud-cloud-disk)
   * [5.Future](#5future)

## 1. Background

Although edge nodes are at the edge, there are still many stateful applications that need to be deployed in edge clusters, such as Statefulset-type edge applications. The user has put forward the following demands for us;

-   Hope that the edge can support local persistent storage, and the stateful application data at the edge will not be inaccessible due to the downtime of the edge node;
-   Local persistent storage can automatically configure Local PV, and can add corresponding storage resources and dynamically expand PVC to meet the needs of data expansion;
-   It can monitor the status of local persistent storage resources and read and write IO, so as to visually monitor the status of edge storage;
-   It can perceive the topology of local storage, automatically schedule Pods that need local storage to the appropriate edge nodes, and can configure storage priority scheduling strategies.

## 2. Plan

After a lot of research and comparison, the open source project [Topovlm](https://github.com/topolvm/topolvm) satisfies all of our above needs. In addition, it also supports the use of local Volume as a container's block device. Explore support for locally held storage snapshots... The architecture diagram of Topolvm is shown below:

<img src="../img/topolvm_arch.svg" alt="component diagram" style="zoom: 67%;" />

Topovlm consists of four components：

-   Master Node：
    -   `topolvm-controller`：Topolvm's controller, Watch PVC to create PV and bind PV;
    -   `topolvm-scheduler`：Topolvm's extended scheduler provides a scheduling algorithm using local PV;
-   Edge Node:
    -   `topolvm-node`：CSI node service, call lvmd to allocate PV;
    -   `lvmd`： Virtualization component of Linux storage system, gRPC service used to manage LVM volumes

We have verified the Topovlm function in SuperEdge's edge cluster and integrated the deployment method into the addon subcommand of edgeadm. Users can deploy Topolvm on the edge of SuperEdge with one click through `edgeadm addon topolvm`, and use the local persistence provided by it. Storage function.

## 3. Deploy Topolvm

### <1>. Preparation conditions

Execute the following command to download the edgeadm static installation package, pay attention to modify the "arch=amd64" parameter, currently supports [amd64, arm64], download the corresponding architecture of your own machine, other parameters remain unchanged

```
arch=amd64 version=v0.7.0 kubernetesVersion=1.20.6 && rm -rf edgeadm-linux-* && wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/$version/$arch/edgeadm-linux-$arch-$version-k8s-$kubernetesVersion.tgz && tar -xzvf edgeadm-linux-* && cd edgeadm-linux-$arch-$version-k8s-$kubernetesVersion && ./edgeadm
```

To install an edge cluster, please refer to: [Install an edge independent Kubernetes cluster with one click. ](https://github.com/superedge/superedge/blob/main/docs/installation/install_edge_kubernetes.md)

>   Note: topolvm support starts from SuperEdge v0.6.0, please download edgeadm v0.6.0 and later versions;

### <2>. Install Topolvm

Installation Commands：

```powershell
[root@k8s-master-node ~]# ./edgeadm addon topolvm --kubeconfig ~/.kube/config 
I0926 15:23:46.047574   15465 topolvm.go:80] Start install addon apps to your original cluster
...
I0926 15:23:48.065099   15465 csi_plugin.go:63] Deploy topolvm all module success!
```

>   Currently SuperEdge deploys the default `Topolvm v0.10.0` version, other versions can edit the corresponding yaml by themselves;

Uninstall Command:

```powershell
[root@k8s-master-node ~]# ./edgeadm detach topolvm --kubeconfig ~/.kube/config 
I0926 15:22:01.794347   14364 topolvm.go:85] Start uninstall addon apps from your original cluster
...
I0926 15:22:03.808400   14364 csi_plugin.go:220] Remove topolvm success!
```

## 4. Verify

### 4.1 Local file simulation persistent storage

<1>. Create device and initialize

```powershell
[root@k8s-edge-node ~]# truncate --size=3G /tmp/backing_store
[root@k8s-edge-node ~]# losetup -f /tmp/backing_store
```

<2>. Create VolumeGroup

```powershell
[root@k8s-edge-node ~]# vgcreate -f -y myvg1 $(losetup -j /tmp/backing_store | cut -d: -f1)
```

>   Where myvg1 is the name of VolumeGroup, which can be customized

View the created VolumeGroup

```powershell
[root@k8s-edge-node ~]# vgdisplay myvg1
  --- Volume group ---
  VG Name               myvg1
 ...
  VG Size               <3.00 GiB
  PE Size               4.00 MiB
  VG UUID               Xgeofp-0cbu-tWUX-7eCE-1Wfz-yo97-3NcgrB
```

<3>. Configure lvm VolumeGroup configuration</span>

```powershell
[root@k8s-master-node ~]# kubectl -n topolvm-system edit cm topolvm-lvmd
```

Configure `lvmd.yaml` and save

```powershell
# Source: topolvm/templates/lvmd/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: topolvm-lvmd
  namespace: topolvm-system
  lvmd.yaml: |
    socket-name: /run/topolvm/lvmd.sock
    device-classes: 
      - default: true
        name: ssd            ## device-classes name
        spare-gb: 3          ## Modify to the size of the VolumeGroupName created by yourself
        volume-group: myvg1  ## Modify to VolumeGroupName created by yourself
```

<4>. Add device identification to nodes with devices

```powershell
[root@k8s-master-node ~]# kubectl label node [nodeName] superedge.io/local.pv=topolvm
```

After the execution is successful, the two support components `topolvm-lvmd` and `topolvm-node` of Topovlm's local PV will be automatically scheduled to the corresponding device nodes:

```powershell
[root@k8s-master-node ~]# kubectl -n topolvm-system get pod -o wide
NAME                                  READY   STATUS        NODE 
topolvm-controller-86775b6cf9-gx9db   4/4     Running    k8s-master-node
topolvm-scheduler-fd4gs               1/1     Running    k8s-master-node
...
topolvm-lvmd-gzvzb                    1/1     Running    k8s-edge-node ##Device node
topolvm-node-vkgcf                    3/3     Running    k8s-edge-node ##Device node
```

<5>. Create PVC and Pod for verification

```powershell
[root@k8s-master-node ~]# kubectl apply -f https://raw.githubusercontent.com/topolvm/topolvm/main/example/podpvc.yaml
```

Check the PVC and Pod status after the submission is successful

```powershell
[root@k8s-master-node ~]# kubectl get pvc
NAME          STATUS   VOLUME       CAPACITY   ACCESS   STORAGECLASS         AGE
topolvm-pvc   Bound    pvc-46cdd5f0   1Gi        RWO   topolvm-provisioner    6s
[root@k8s-master-node ~]# kubectl get pod 
NAME     READY   STATUS    RESTARTS   AGE
my-pod   1/1     Running   0          119s
```

You can enter the Pod to write content to its mounted file, and mount the PVC to other Pods, and the written content remains unchanged.

<6>. Expand PVC

```powershell
[root@k8s-master-node ~]# kubectl get pvc topolvm-pvc  -o yaml | sed  "s/1Gi/2Gi/" | kubectl apply -f -
persistentvolumeclaim/topolvm-pvc configured
[root@k8s-master-node ~]# kubectl get pvc
NAME          STATUS   VOLUME       CAPACITY   ACCESS   STORAGECLASS         AGE
topolvm-pvc   Bound    pvc-46cdd5f0   2Gi        RWO   topolvm-provisioner    5s
```

>    Note：
>
>   -   PVC does not support shrinking, the expansion value must be larger than the previous value, and should be an integer as much as possible;
>   -   PVC expansion cannot exceed the total volume of VolumeGroup, which means that the actual physical disk storage size can be expanded first;

<7>. Custom scheduling scheduling strategy

Topolvm's scheduling policy file for Local PV is by default in the Master node `/etc/kubernetes/kube-scheduler/kube-scheduler-policy.cfg` location, and its content is as follows;

```json
[root@k8s-master-node ~]# cat /etc/kubernetes/kube-scheduler/kube-scheduler-policy.cfg
{
  "kind" : "Policy",
  "apiVersion" : "v1",
  "extenders" :
    [{
      "urlPrefix": "http://127.0.0.1:9251",
      "filterVerb": "predicate",
      "prioritizeVerb": "prioritize",
      "nodeCacheCapable": false,
      "weight": 100,
      "managedResources":
      [{
        "name": "topolvm.cybozu.com/capacity",
        "ignoredByScheduler": true
      }]
    }]
}
```

Need to customize the scheduling strategy can be edited, and restart `Kube-scheduler` to take effect, more settings can refer to [Topolvm community expansion scheduling strategy](https://github.com/kubernetes/community/blob/master/contributors/ design-proposals/scheduling/scheduler_extender.md).

### 4.2 Mount Tencent Cloud Cloud Disk

<1>. Apply for Tencent Cloud Cloud Disk

To apply for Tencent Cloud Cloud Disk and mount it on the node, please refer to [Create Tencent Cloud Cloud Disk](https://intl.cloud.tencent.com/document/product/362/31647).

<2>.Format Tencent Cloud Cloud Disk 

-   Execute the following command to view the name of the cloud disk connected

    ```powershell
    [root@k8s-edge-node ~]# fdisk -l
    ...
    Disk /dev/vdb: 21.5 GB, 21474836480 bytes, 41943040 sectors
    Units = sectors of 1 * 512 = 512 bytes
    Sector size (logical/physical): 512 bytes / 512 bytes
    I/O size (minimum/optimal): 512 bytes / 512 bytes
    ```

    The name of the newly mounted Tencent Cloud cloud disk is: `/dev/vdb`

-   Format Disk `/dev/vdb`

    ```powershell
    [root@k8s-edge-node ~]# mkfs.ext4 /dev/vdb
    ```

    >   The format of the format is carried out according to your own needs, here is ext4;

    For the automatic mounting of Tencent Cloud Cloud Disk when booting, please refer to: [Tencent Cloud Cloud Disk Automatic Mounting](https://cloud.tencent.com/document/product/362/32403). For expansion of Tencent Cloud Cloud Disk, please refer to: [Tencent Cloud Cloud Disk Expansion](https://cloud.tencent.com/document/product/362/54728).

<3>. Create VolumeGroup

```powershell
vgcreate test /dev/vdb /dev/sdb
```

>   In addition to the `/dev/sdb` disk added by `/dev/vdb`, multiple cloud disks can be added to VolumeGroup. You can also use the [vgextend](https://man.linuxde.net/vgextend) command to expand an existing VolumeGroup.

The follow-up verification logic is exactly the same as [4.1 <3>, <4>...<7>](#VolumeGroupConfig) of local file simulation persistent storage, so I won’t repeat it here.

## 5.Future

In the future we have two plans:

-   First: We will integrate Topolvm into SuperEdge's commercial version of TKE Edge to support local storage;

-   Second: We will cooperate with the Topolvm community to feed back some of the needs of our users to Topolvm, promote the soundness of Topolvm's functions, and provide our users with simpler and more practical local persistent storage.