# 部署监控系统

监控系统分为2个部分：prometheus-server、node-exporter

## 1. 部署prometheus-server

使用tunnel-coredns的clusterip替换[prometheus-server.yaml](../../deployment/prometheus-server.yaml)
中的spec.dnsConfig.nameservers变量

```shell
kubectl -n edge-system get svc  tunnel-coredns  -o=jsonpath='{.spec.clusterIP}'
```

```shell
kubectl apply -f prometheus-server.yaml
```

## 2. 部署prometheus-node-exporter

```shell
kubectl apply -f https://raw.githubusercontent.com/superedge/superedge/main/deployment/prometheus-node-exporter.yaml
```

## 3. 验证部署是否成功

<details><summary>是否采集到kubelet metrics</summary>
<p>

```shell
$ curl -G  http://<prometheus-server的clusterip>/api/v1/series? --data-urlencode 'match[]=container_processes{job="node-cadvisor"}'
{
  [
    {
      "__name__": "container_processes",
      "id": "/system.slice/docker.service",
      "instance": "edge-7x94bd",
      "job": "node-cadvisor",
      "unInstanceId": "none"
    },
    {
      "__name__": "container_processes",
      "id": "/system.slice/kubelet.service",
      "instance": "edge-7x94bd",
      "job": "node-cadvisor",
      "unInstanceId": "none"
    }
  ]
}
```

</p>
</details>


<details><summary>是否采集到node metrics</summary>
<p>

```shell
curl -G  http://<prometheus-server的clusterip>/api/v1/series? --data-urlencode 'match[]=node_cpu_guest_seconds_total{job="node-exporter"}'
{
  "status": "success",
  "data": [
    {
      "__name__": "node_cpu_guest_seconds_total",
      "cpu": "0",
      "instance": "edge-7x94bd",
      "job": "node-exporter",
      "mode": "nice",
      "unInstanceId": "none"
    },
    {
      "__name__": "node_cpu_guest_seconds_total",
      "cpu": "0",
      "instance": "edge-7x94bd",
      "job": "node-exporter",
      "mode": "user",
      "unInstanceId": "none"
    },
    {
      "__name__": "node_cpu_guest_seconds_total",
      "cpu": "1",
      "instance": "edge-7x94bd",
      "job": "node-exporter",
      "mode": "nice",
      "unInstanceId": "none"
    },
    {
      "__name__": "node_cpu_guest_seconds_total",
      "cpu": "1",
      "instance": "edge-7x94bd",
      "job": "node-exporter",
      "mode": "user",
      "unInstanceId": "none"
    }
  ]
}
```

</p>
</details>