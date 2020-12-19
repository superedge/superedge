# edge-health

edge-health runs on every edge node by default for accurately determine the actual operation status of edge nodes.

It introduces distributed health check for the first time to achieve independent health check for each node group, which allows each node to detect and monitor nodes within the same group and eliminates the false alerts caused by the unstable cloud-to-edge network.

### Options
#### Detection Related
   - healthcheckperiod: detection period (unit is second, and the default value is 10)
   - healthcheckscoreline: The score line to determine the node is normal (the default value is 100)
   - The following parameters are supported in the plugin:
     - timeout: specify the timeout period for a request
     - retrytime: specify the number of retries for probe failure
     - weight: specify the proportion of the detection method to the detection score, the maximum is 1
     - port: specify the port for detection
   - Two plugins currently supported:
     - Detcet kubelet security authentication port: `--kubeletauthplugin=timeout=5,retrytime=3,weight=1,port=10250`
     - Detcet kubelet non-secure authentication port: `--kubeletplugin=timeout=5,retrytime=3,port=10255,weight=1`

#### Communication And Interaction Related
- communicateperiod: specify the interaction period (in seconds, and the default value is 10)
- communicatetimetout: specify the timeout period for sending interactive information (unit is seconds, and the default value is 3)
- communicateretrytime: specify the number of retries when sending interactive information fails (the default value is 1)
- communicateserverport: the port for receiving and sending interactive information (the default port is 51005)

#### Others
- voteperiod: specify the voting period (unit is seconds, and the default value is 10)
- votetimeout: specify the valid time for voting, invalid voting is longer than this time (unit is second, and the default value is 60)

### Multi-region Detection
- Turn On
  - Label the nodes according to the region with `superedgehealth/topology-zone:<zone>`
  - Create a configmap named `edge-health-zone-config` in the `kube-system` namespace, specify the value of `TaintZoneAdmission` as `true`, you can directly use the following yaml to create
             ```yaml
             apiVersion: v1
             kind: ConfigMap
             metadata:
               name: edge-health-zone-config
               namespace: kube-system
             data:
               TaintZoneAdmission: true
             ```
- Turn Off
 - Change the value of `TaintZoneAdmission` to `false` or delete the configmap of `edge-health-zone-config`

> Note: If multiple regions are enabled but the node is not marked with a region label, the node will only detect itself when detecting
