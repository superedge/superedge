# Tunnel
Tunnel acts as the bridge between edge and cloud. It consists of `tunnel-cloud` and `tunnel-edge`, which responsible for maintaining persistent cloud-to-edge network connection. It allows edge nodes without public IP address to be managed by Kubernetes on the cloud for unified operation and maintenance.

## Architecture Diagram
<div align="left">
  <img src="../img/tunnel.png" width=70% title="tunnel Architecture">
</div>

## Implementation
### Node Registration
   - The tunnel-edge on the edge node actively connects to  tunnel-cloud service, and tunnel-cloud service transfers the request to the tunnel-cloud pod according to the load balancing policy.
   - After tunnel-edge establishes a grpc connection with tunnel-cloud, tunnel-cloud will write the mapping of its podIp and nodeName of the node where tunnel-edge is located into DNS。If the grpc connection is disconnected, tunnel-cloud will delete the podIp and node name mapping.

### Proxy Forwarding Of Requests
   - When apiserver or other cloud applications access the kubelet or other applications on the edge node, the tunnel-dns uses DNS hijacking (resolving the node name in the host to the podIp of tunnel-cloud) to forward the request to the pod of the tunnel-cloud.
   - The tunnel-cloud forwards the request information to the grpc connection established with the tunnel-edge according to the node name.
   - The tunnel-edge requests the application on the edge node according to the received request information.
