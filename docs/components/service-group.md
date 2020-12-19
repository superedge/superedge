# serviceGroup

# Why is it needed
## What's different about the edges
- In edge computing scenarios, multiple edge sites are often managed in the same cluster, and each edge site has one or more computing nodes
- We hope to run a group of services with business logic relationship in each site, and each site will run the same and complete microservices
- Due to network restrictions, the access between microservices is not expected and can not be performed in multiple sites

## What can ServiceGroup solve

ServiceGroup can easily deploy a group of services in different sites or regions, which belonging to the same cluster,
and make the calls between services complete within site or region, avoiding cross regional access of services.

The native k8s needs to plan the affinity of nodes to control the specific location of the pod created by the deployment.
When the number of edge sites and the number of services to be deployed are too large, the management and deployment are extremely complex, and even there is only theoretical possibility

In addition, in order to limit the mutual calls between services to the same site, it is necessary to create a dedicated service for each deployment,
which brings huge management work and is easy to make mistakes

ServiceGroup is designed for this scenario. The deployment-grid and service-grid provided by ServiceGroup can be used to easily deploy and manage edge applications, control service flow,
and ensure the number of services in each region and disaster recovery

# Key concepts
## Architecture
<div align="left">
  <img src="../img/serviceGroup-UseCase.png" width=70% title="service-group">
</div>

## NodeUnit
- NodeUnit is usually one or more computing resource instances located in the same edge site. It is necessary to ensure that the intranet of nodes in the same NodeUnit is connected
- Services in ServiceGroup run within a NodeUnit
- ServiceGroup allows users to set the number of pods for deployment runs in a NodeUnit
- ServiceGroup keeps calls between services in its own NodeUnit

## NodeGroup
- NodeGroup contains one or more NodeUnits
- Ensure that the services in ServiceGroup are deployed on each NodeUnit in NodeGroup
- When nodeunit is added to the cluster, the service in ServiceGroup is automatically deployed to the newly added NodeUnit

## ServiceGroup
ServiceGroup contains one or more business services. The applicable scenarios are as follows:
- Want to package and deploy multiple services
- Or, the service(s) needs to run in multiple NodeUnits and guarantee the number of pods for each NodeUnit
- Or, calls between services need to be controlled in the same NodeUnit, and traffic cannot be forwarded to other NodeUnits.

> Note: ServiceGroup is an abstract resource. Multiple ServiceGroups can be created in a cluster

## Resource types
### DepolymentGrid

The format of the DeploymentGrid is similar to that of deployment. The < deployment template > field is the template field of
the Kubernetes original deployment, and the more special is the gridUniqKey field, which indicates the key value of the label of the node group.
```yaml
apiVersion: superedge.io/v1
kind: DeploymentGrid
metadata:
  name:
  namespace:
spec:
  gridUniqKey: <NodeLabel Key>
  <deployment-template>
```

### ServiceGrid

The format of ServiceGrid is similar to that of service. The < service template > field is the template field of
the Kubernetes original service. The special field is the GridUniqKey field, which indicates the key value of the label of the node group.
```yaml
apiVersion: superedge.io/v1
kind: ServiceGrid
metadata:
  name:
  namespace:
spec:
  gridUniqKey: <NodeLabel Key>
  <service-template>
```


# How to use ServiceGroup

Taking the deployment of nginx as an example, we hope to deploy nginx services in multiple node groups. We need to do the following:

## Determines the unique identity of the ServiceGroup
This step is logical planning, and no actual operation is required. For example, we use the uniqkey for the ServiceGroup logical tag to be created as: zone.

## Group edge nodes
In this step, we need to label the edge nodes with kubectl.

For example, we select node12 and node14 and label them with zone = nodeunit1; node21 and node23 are labeled with zone = nodeunit2.

> Note: in the previous step, the key of the label is consistent with the uniqkey of the servicegroup. Value is the unique key of the NodeUnit. Nodes with the same value belong to the same NodeUnit

If you want to use more than one ServiceGroup, assign each ServiceGroup a different uniqkey.

## deploy deploymentGrid
```yaml
apiVersion: superedge.io/v1
kind: DeploymentGrid
metadata:
  name: deploymentgrid-demo
  namespace: default
spec:
  gridUniqKey: zone
  template:
    selector:
      matchLabels:
        appGrid: nginx
    replicas: 2
    template:
      metadata:
        labels:
          appGrid: nginx
      spec:
        containers:
        - name: nginx
          image: nginx:1.7.9
          ports:
          - containerPort: 80
            protocol: TCP
```

### Deploy serviceGrid
```yaml
apiVersion: superedge.io/v1
kind: ServiceGrid
metadata:
  name: servicegrid-demo
  namespace: default
spec:
  gridUniqKey: zone
  template:
    selector:
      appGrid: nginx
    ports:
    - protocol: TCP
      port: 80
      targetPort: 80
```
Since the gridUniqKey field is set to zone, the key of the label we use when grouping nodes is zone. If there are three NodeUnits,
labeled by zone: zone-0, zone: zone-1, zone: zone-2 for them. at this time, each NodeUnit has the deployment of nginx and the corresponding pod.
If a service is accessed through servicename and clusterip in the node, the request will only be sent to the nodes in this group.
```
[~]# kubectl get deploy
NAME                         READY   UP-TO-DATE   AVAILABLE   AGE
deploymentgrid-demo-zone-0   2/2     2            2           85s
deploymentgrid-demo-zone-1   2/2     2            2           85s
deploymentgrid-demo-zone-2   2/2     2            2           85s

[~]# kubectl get svc
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
kubernetes             ClusterIP   172.19.0.1     <none>        443/TCP   87m
servicegrid-demo-svc   ClusterIP   172.19.0.177   <none>        80/TCP    80s
```

In addition, if a new NodeUnit that are added to the cluster after the DeploymentGrid and ServiceGrid resources are deployed,
the system will automatically create the corresponding deployment and service for the new NodeUnit.
