# 配置tunnel-cloud HPA

## 1. 部署好监控系统

- [部署监控系统](./deploy-monitor_CN.md)

## 2. 确认tunnel-cloud的metrics数据是否采集成功

<details><summary>tunnel-cloud metrics</summary>
<p>

```shell
curl -G  http://<prometheus-server-clusterip>/api/v1/series? --data-urlencode 'match[]=tunnel_cloud_nodes'
{
  "status": "success",
  "data": [
    {
      "__name__": "tunnel_cloud_nodes",
      "instance": "172.31.0.10:6000",
      "job": "tunnel-cloud-metrics",
      "kubernetes_namespace": "edge-system",
      "kubernetes_pod_name": "tunnel-cloud-64ff7d9c9d-4lljh"
    },
    {
      "__name__": "tunnel_cloud_nodes",
      "instance": "172.31.0.13:6000",
      "job": "tunnel-cloud-metrics",
      "kubernetes_namespace": "edge-system",
      "kubernetes_pod_name": "tunnel-cloud-64ff7d9c9d-vmkxh"
    }
  ]
}
```

</p>
</details>

## 3. 部署prometheus-adapter

### 3.1 安装[helm](https://helm.sh/docs/intro/install/)

### 3.2 准备values.yaml

[values.yaml](../../deployment/values.yaml)

```shell
wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/charts/prometheus-adapter-2.15.0.tgz
helm install prometheus-adapter prometheus-adapter-2.15.0.tgz -f values.yaml -n edge-system
```

### 3.3 测试prometheus-adapter是否安装成功

如果安装正确，是可以看到 Custom Metrics API 返回了我们配置的**nodes_per_pod**相关指标:

<details><summary>Custom Metrics API</summary>
<p>

```shell
$ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1
{
  "kind": "APIResourceList",
  "apiVersion": "v1",
  "groupVersion": "custom.metrics.k8s.io/v1beta1",
  "resources": [
    {
      "name": "namespaces/nodes_per_pod",
      "singularName": "",
      "namespaced": false,
      "kind": "MetricValueList",
      "verbs": [
        "get"
      ]
    },
    {
      "name": "pods/nodes_per_pod",
      "singularName": "",
      "namespaced": true,
      "kind": "MetricValueList",
      "verbs": [
        "get"
      ]
    }
  ]
}
```

</p>
</details>

并且可以看到当前所有**tunnel-cloud**的pod，以及各pod上连接的边缘节点个数

<details><summary>nodes_per_pod </summary>
<p>

```shell
$ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/edge-system/pods/*/nodes_per_pod
{
  "kind": "MetricValueList",
  "apiVersion": "custom.metrics.k8s.io/v1beta1",
  "metadata": {
    "selfLink": "/apis/custom.metrics.k8s.io/v1beta1/namespaces/edge-system/pods/%2A/nodes_per_pod"
  },
  "items": [
    {
      "describedObject": {
        "kind": "Pod",
        "namespace": "edge-system",
        "name": "tunnel-cloud-64ff7d9c9d-vmkxh",
        "apiVersion": "/v1"
      },
      "metricName": "nodes_per_pod",
      "timestamp": "2021-07-14T10:19:37Z",
      "value": "1",
      "selector": null
    }
  ]
}
```

</p>
</details>

## 4. 部署tunel-cloud-hpa.yaml

[tunnel-cloud-hpa.yaml](../../deployment/tunel-cloud-hpa.yaml)

通过调整averageValue值和改变边缘节点数量，可以快速观察到 tunnel-cloud 的pod数量变化情况
