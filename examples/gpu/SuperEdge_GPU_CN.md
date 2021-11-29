# 如何在SuperEdge中使用GPU

# edge-gpu

SuperEdge集群和常规Kubernetes集群一样，可以无痛使用各种GPU或GPU Share的方案。

## nvidia-device-plugin

nvidia-device-plugin是英伟达官方提供的K8S支持GPU方案，此方案中每个pod可以使用一张或多张GPU卡，每张卡会被Pod独占。

在使用nvidia-device-plugin之前，请确保：

- 节点上安装了GPU驱动，可以通过命令 nvidia-smi 来检查
    
    ```bash
    $ nvidia-smi -L
    Mon Nov 29 21:17:49 2021
    +-----------------------------------------------------------------------------+
    | NVIDIA-SMI 450.102.04   Driver Version: 450.102.04   CUDA Version: 11.0     |
    |-------------------------------+----------------------+----------------------+
    | GPU  Name        Persistence-M| Bus-Id        Disp.A | Volatile Uncorr. ECC |
    | Fan  Temp  Perf  Pwr:Usage/Cap|         Memory-Usage | GPU-Util  Compute M. |
    |                               |                      |               MIG M. |
    |===============================+======================+======================|
    |   0  Tesla T4            On   | 00000000:00:08.0 Off |                    0 |
    | N/A   35C    P8    11W /  70W |      0MiB / 15109MiB |      0%      Default |
    |                               |                      |                  N/A |
    +-------------------------------+----------------------+----------------------+
    
    +-----------------------------------------------------------------------------+
    | Processes:                                                                  |
    |  GPU   GI   CI        PID   Type   Process name                  GPU Memory |
    |        ID   ID                                                   Usage      |
    |=============================================================================|
    |  No running processes found                                                 |
    +-----------------------------------------------------------------------------+
    ```
    
- 节点上安装了nvidia-docker2
    
    ```bash
    # ubuntu
    distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
    curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
    curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | sudo tee /etc/apt/sources.list.d/nvidia-docker.list
    sudo apt-get update && sudo apt-get install -y nvidia-docker2
    sudo systemctl restart docker
    
    # centos
    
    # install docker
    yum-config-manager --add-repo=https://download.docker.com/linux/centos/docker-ce.repo
    yum install docker-ce -y && systemctl restart docker
    
    # install nvidia-docker
    distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
    curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.repo | sudo tee /etc/yum.repos.d/nvidia-docker.repo
    yum install -y nvidia-docker2
    ```
    
    可以通过命令 `docker run --rm --gpus all nvidia/cuda:11.0-base nvidia-smi` 来检验安装是否成功，结果同nvidia-smi检查GPU驱动
    

nvidia-device-plugin 的方案中仅需要通过 daemonset 在每个GPU节点上部署 device-plugin，因此我们可以通过以下命令把daemonset添加到集群中。

```docker
$ kubectl create -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/v0.10.0/nvidia-device-plugin.yml
```

接下来可以通过在Pod中指定使用gpu数量来添加创建gpu pod：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: gpu-pod
spec:
  containers:
    - name: cuda-container
      image: nvcr.io/nvidia/cuda:9.0-devel
      resources:
        limits:
          nvidia.com/gpu: 1
```

## nano-gpu-framework

nano-gpu-framework 是腾讯云TKE团队开源的GPU解决方案，nano-gpu的方案中多个 pod 可以**共享**同一张GPU卡从而提高资源利用率。

使用nano-gpu的节点和之前一样也需要安装GPU驱动以及nvidia-docker。nano-gpu的方案中需要部署 extender-scheduler (nano-gpu-scheduler) 和 节点上的 device-plugin(nano-gpu-agent)。

### nano-gpu-scheduler

kube-scheduler 需要在调度时调用 nano-gpu-scheduler 的接口，nano-gpu-scheduler 为了感知当前集群的状态需要调用api server，因此建议将 nano-gpu-scheduler 部署在云端，以保证必要的访问不会因为网络环境等原因被中断。

可以通过以下命令来创建nano-gpu-scheduler的workload和service。

```docker
kubectl apply -f https://raw.githubusercontent.com/nano-gpu/nano-gpu-scheduler/master/deploy/nano-gpu-scheduler.yaml
```

接下来需要在kube-scheduler的配置文件中 extenders 字段里添加extender-scheduler的配置：

```json
{
  "urlPrefix": "http://<kube-apiserver-svc>/api/v1/namespaces/kube-system/services/nano-gpu-scheduler/proxy/scheduler",
  "filterVerb": "filter",
  "prioritizeVerb": "priorities",
  "bindVerb": "bind",
  "weight": 1,
  "enableHttps": false,
  "nodeCacheCapable": true,
  "managedResources": [
    {
      "name": "nano-gpu/gpu-percent"
    }
  ]
}
```

### nano-gpu-agent

接下来可以通过以下命令来创建nano-gpu-agent的daemonset：

```docker
$ kubectl apply -f https://raw.githubusercontent.com/nano-gpu/nano-gpu-agent/master/deploy/nano-gpu-agent.yaml
```

可以通过创建使用nano-gpu的pod来验证配置的正确性：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: gpu-pod
spec:
  containers:
    - name: cuda-container
      image: nvcr.io/nvidia/cuda:9.0-devel
      resources:
        limits:
          nano-gpu/gpu-percent: "20"
```

参考：

1. [https://github.com/nano-gpu](https://github.com/nano-gpu)
2. [https://github.com/NVIDIA/k8s-device-plugin](https://github.com/NVIDIA/k8s-device-plugin)