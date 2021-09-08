/*
Copyright 2020 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

const ServiceGridCRDYaml = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.5
  creationTimestamp: null
  name: servicegrids.superedge.io
spec:
  group: superedge.io
  names:
    kind: ServiceGrid
    listKind: ServiceGridList
    plural: servicegrids
    singular: servicegrid
    shortNames:
    - sg
  scope: Namespaced
  preserveUnknownFields: false
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          properties:
            gridUniqKey:
              type: string
            template:
              description: ServiceSpec describes the attributes that a user creates
                on a service.
              properties:
                allocateLoadBalancerNodePorts:
                  description: allocateLoadBalancerNodePorts defines if NodePorts
                    will be automatically allocated for services with type LoadBalancer.  Default
                    is "true". It may be set to "false" if the cluster load-balancer
                    does not rely on NodePorts. allocateLoadBalancerNodePorts may
                    only be set for services with type LoadBalancer and will be cleared
                    if the type is changed to any other type. This field is alpha-level
                    and is only honored by servers that enable the ServiceLBNodePortControl
                    feature.
                  type: boolean
                clusterIP:
                  description: 'clusterIP is the IP address of the service and is
                    usually assigned randomly. If an address is specified manually,
                    is in-range (as per system configuration), and is not in use,
                    it will be allocated to the service; otherwise creation of the
                    service will fail. This field may not be changed through updates
                    unless the type field is also being changed to ExternalName (which
                    requires this field to be blank) or the type field is being changed
                    from ExternalName (in which case this field may optionally be
                    specified, as describe above).  Valid values are "None", empty
                    string (""), or a valid IP address. Setting this to "None" makes
                    a "headless service" (no virtual IP), which is useful when direct
                    endpoint connections are preferred and proxying is not required.  Only
                    applies to types ClusterIP, NodePort, and LoadBalancer. If this
                    field is specified when creating a Service of type ExternalName,
                    creation will fail. This field will be wiped when updating a Service
                    to type ExternalName. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies'
                  type: string
                clusterIPs:
                  description: "ClusterIPs is a list of IP addresses assigned to this
                    service, and are usually assigned randomly.  If an address is
                    specified manually, is in-range (as per system configuration),
                    and is not in use, it will be allocated to the service; otherwise
                    creation of the service will fail. This field may not be changed
                    through updates unless the type field is also being changed to
                    ExternalName (which requires this field to be empty) or the type
                    field is being changed from ExternalName (in which case this field
                    may optionally be specified, as describe above).  Valid values
                    are \"None\", empty string (\"\"), or a valid IP address.  Setting
                    this to \"None\" makes a \"headless service\" (no virtual IP),
                    which is useful when direct endpoint connections are preferred
                    and proxying is not required.  Only applies to types ClusterIP,
                    NodePort, and LoadBalancer. If this field is specified when creating
                    a Service of type ExternalName, creation will fail. This field
                    will be wiped when updating a Service to type ExternalName.  If
                    this field is not specified, it will be initialized from the clusterIP
                    field.  If this field is specified, clients must ensure that clusterIPs[0]
                    and clusterIP have the same value. \n Unless the \"IPv6DualStack\"
                    feature gate is enabled, this field is limited to one value, which
                    must be the same as the clusterIP field.  If the feature gate
                    is enabled, this field may hold a maximum of two entries (dual-stack
                    IPs, in either order).  These IPs must correspond to the values
                    of the ipFamilies field. Both clusterIPs and ipFamilies are governed
                    by the ipFamilyPolicy field. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies"
                  items:
                    type: string
                  type: array
                  x-kubernetes-list-type: atomic
                externalIPs:
                  description: externalIPs is a list of IP addresses for which nodes
                    in the cluster will also accept traffic for this service.  These
                    IPs are not managed by Kubernetes.  The user is responsible for
                    ensuring that traffic arrives at a node with this IP.  A common
                    example is external load-balancers that are not part of the Kubernetes
                    system.
                  items:
                    type: string
                  type: array
                externalName:
                  description: externalName is the external reference that discovery
                    mechanisms will return as an alias for this service (e.g. a DNS
                    CNAME record). No proxying will be involved.  Must be a lowercase
                    RFC-1123 hostname (https://tools.ietf.org/html/rfc1123) and requires
                    Type to be
                  type: string
                externalTrafficPolicy:
                  description: externalTrafficPolicy denotes if this Service desires
                    to route external traffic to node-local or cluster-wide endpoints.
                    "Local" preserves the client source IP and avoids a second hop
                    for LoadBalancer and Nodeport type services, but risks potentially
                    imbalanced traffic spreading. "Cluster" obscures the client source
                    IP and may cause a second hop to another node, but should have
                    good overall load-spreading.
                  type: string
                healthCheckNodePort:
                  description: healthCheckNodePort specifies the healthcheck nodePort
                    for the service. This only applies when type is set to LoadBalancer
                    and externalTrafficPolicy is set to Local. If a value is specified,
                    is in-range, and is not in use, it will be used.  If not specified,
                    a value will be automatically allocated.  External systems (e.g.
                    load-balancers) can use this port to determine if a given node
                    holds endpoints for this service or not.  If this field is specified
                    when creating a Service which does not need it, creation will
                    fail. This field will be wiped when updating a Service to no longer
                    need it (e.g. changing type).
                  format: int32
                  type: integer
                ipFamilies:
                  description: "IPFamilies is a list of IP families (e.g. IPv4, IPv6)
                    assigned to this service, and is gated by the \"IPv6DualStack\"
                    feature gate.  This field is usually assigned automatically based
                    on cluster configuration and the ipFamilyPolicy field. If this
                    field is specified manually, the requested family is available
                    in the cluster, and ipFamilyPolicy allows it, it will be used;
                    otherwise creation of the service will fail.  This field is conditionally
                    mutable: it allows for adding or removing a secondary IP family,
                    but it does not allow changing the primary IP family of the Service.
                    \ Valid values are \"IPv4\" and \"IPv6\".  This field only applies
                    to Services of types ClusterIP, NodePort, and LoadBalancer, and
                    does apply to \"headless\" services.  This field will be wiped
                    when updating a Service to type ExternalName. \n This field may
                    hold a maximum of two entries (dual-stack families, in either
                    order).  These families must correspond to the values of the clusterIPs
                    field, if specified. Both clusterIPs and ipFamilies are governed
                    by the ipFamilyPolicy field."
                  items:
                    description: IPFamily represents the IP Family (IPv4 or IPv6).
                      This type is used to express the family of an IP expressed by
                      a type (e.g. service.spec.ipFamilies).
                    type: string
                  type: array
                  x-kubernetes-list-type: atomic
                ipFamilyPolicy:
                  description: IPFamilyPolicy represents the dual-stack-ness requested
                    or required by this Service, and is gated by the "IPv6DualStack"
                    feature gate.  If there is no value provided, then this field
                    will be set to SingleStack. Services can be "SingleStack" (a single
                    IP family), "PreferDualStack" (two IP families on dual-stack configured
                    clusters or a single IP family on single-stack clusters), or "RequireDualStack"
                    (two IP families on dual-stack configured clusters, otherwise
                    fail). The ipFamilies and clusterIPs fields depend on the value
                    of this field.  This field will be wiped when updating a service
                    to type ExternalName.
                  type: string
                loadBalancerIP:
                  description: 'Only applies to Service Type: LoadBalancer LoadBalancer
                    will get created with the IP specified in this field. This feature
                    depends on whether the underlying cloud-provider supports specifying
                    the loadBalancerIP when a load balancer is created. This field
                    will be ignored if the cloud-provider does not support the feature.'
                  type: string
                loadBalancerSourceRanges:
                  description: 'If specified and supported by the platform, this will
                    restrict traffic through the cloud-provider load-balancer will
                    be restricted to the specified client IPs. This field will be
                    ignored if the cloud-provider does not support the feature." More
                    info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/'
                  items:
                    type: string
                  type: array
                ports:
                  description: 'The list of ports that are exposed by this service.
                    More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies'
                  items:
                    description: ServicePort contains information on service's port.
                    properties:
                      appProtocol:
                        description: The application protocol for this port. This
                          field follows standard Kubernetes label syntax. Un-prefixed
                          names are reserved for IANA standard service names (as per
                          RFC-6335 and http://www.iana.org/assignments/service-names).
                          Non-standard protocols should use prefixed names such as
                          mycompany.com/my-custom-protocol. This is a beta field that
                          is guarded by the ServiceAppProtocol feature gate and enabled
                          by default.
                        type: string
                      name:
                        description: The name of this port within the service. This
                          must be a DNS_LABEL. All ports within a ServiceSpec must
                          have unique names. When considering the endpoints for a
                          Service, this must match the 'name' field in the EndpointPort.
                          Optional if only one ServicePort is defined on this service.
                        type: string
                      nodePort:
                        description: 'The port on each node on which this service
                          is exposed when type is NodePort or LoadBalancer.  Usually
                          assigned by the system. If a value is specified, in-range,
                          and not in use it will be used, otherwise the operation
                          will fail.  If not specified, a port will be allocated if
                          this Service requires one.  If this field is specified when
                          creating a Service which does not need it, creation will
                          fail. This field will be wiped when updating a Service to
                          no longer need it (e.g. changing type from NodePort to ClusterIP).
                          More info: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport'
                        format: int32
                        type: integer
                      port:
                        description: The port that will be exposed by this service.
                        format: int32
                        type: integer
                      protocol:
                        default: TCP
                        description: The IP protocol for this port. Supports "TCP",
                          "UDP", and "SCTP". Default is TCP.
                        type: string
                      targetPort:
                        anyOf:
                        - type: integer
                        - type: string
                        description: 'Number or name of the port to access on the
                          pods targeted by the service. Number must be in the range
                          1 to 65535. Name must be an IANA_SVC_NAME. If this is a
                          string, it will be looked up as a named port in the target
                          Pod''s container ports. If this is not specified, the value
                          of the ''port'' field is used (an identity map). This field
                          is ignored for services with clusterIP=None, and should
                          be omitted or set equal to the ''port'' field. More info:
                          https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service'
                        x-kubernetes-int-or-string: true
                    required:
                    - port
                    type: object
                  type: array
                  x-kubernetes-list-map-keys:
                  - port
                  - protocol
                  x-kubernetes-list-type: map
                publishNotReadyAddresses:
                  description: publishNotReadyAddresses indicates that any agent which
                    deals with endpoints for this Service should disregard any indications
                    of ready/not-ready. The primary use case for setting this field
                    is for a StatefulSet's Headless Service to propagate SRV DNS records
                    for its Pods for the purpose of peer discovery. The Kubernetes
                    controllers that generate Endpoints and EndpointSlice resources
                    for Services interpret this to mean that all endpoints are considered
                    "ready" even if the Pods themselves are not. Agents which consume
                    only Kubernetes generated endpoints through the Endpoints or EndpointSlice
                    resources can safely assume this behavior.
                  type: boolean
                selector:
                  additionalProperties:
                    type: string
                  description: 'Route service traffic to pods with label keys and
                    values matching this selector. If empty or not present, the service
                    is assumed to have an external process managing its endpoints,
                    which Kubernetes will not modify. Only applies to types ClusterIP,
                    NodePort, and LoadBalancer. Ignored if type is ExternalName. More
                    info: https://kubernetes.io/docs/concepts/services-networking/service/'
                  type: object
                sessionAffinity:
                  description: 'Supports "ClientIP" and "None". Used to maintain session
                    affinity. Enable client IP based session affinity. Must be ClientIP
                    or None. Defaults to None. More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies'
                  type: string
                sessionAffinityConfig:
                  description: sessionAffinityConfig contains the configurations of
                    session affinity.
                  properties:
                    clientIP:
                      description: clientIP contains the configurations of Client
                        IP based session affinity.
                      properties:
                        timeoutSeconds:
                          description: timeoutSeconds specifies the seconds of ClientIP
                            type session sticky time. The value must be >0 && <=86400(for
                            1 day) if ServiceAffinity == "ClientIP". Default value
                            is 10800(for 3 hours).
                          format: int32
                          type: integer
                      type: object
                  type: object
                topologyKeys:
                  description: topologyKeys is a preference-order list of topology
                    keys which implementations of services should use to preferentially
                    sort endpoints when accessing this Service, it can not be used
                    at the same time as externalTrafficPolicy=Local. Topology keys
                    must be valid label keys and at most 16 keys may be specified.
                    Endpoints are chosen based on the first topology key with available
                    backends. If this field is specified and all entries have no backends
                    that match the topology of the client, the service has no backends
                    for that client and connections should fail. The special value
                    "*" may be used to mean "any topology". This catch-all value,
                    if used, only makes sense as the last value in the list. If this
                    is not specified or empty, no topology constraints will be applied.
                    This field is alpha-level and is only honored by servers that
                    enable the ServiceTopology feature.
                  items:
                    type: string
                  type: array
                type:
                  description: 'type determines how the Service is exposed. Defaults
                    to ClusterIP. Valid options are ExternalName, ClusterIP, NodePort,
                    and LoadBalancer. "ClusterIP" allocates a cluster-internal IP
                    address for load-balancing to endpoints. Endpoints are determined
                    by the selector or if that is not specified, by manual construction
                    of an Endpoints object or EndpointSlice objects. If clusterIP
                    is "None", no virtual IP is allocated and the endpoints are published
                    as a set of endpoints rather than a virtual IP. "NodePort" builds
                    on ClusterIP and allocates a port on every node which routes to
                    the same endpoints as the clusterIP. "LoadBalancer" builds on
                    NodePort and creates an external load-balancer (if supported in
                    the current cloud) which routes to the same endpoints as the clusterIP.
                    "ExternalName" aliases this service to the specified externalName.
                    Several other fields do not apply to ExternalName services. More
                    info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types'
                  type: string
              type: object
          type: object
        status:
          description: ServiceStatus represents the current status of a service.
          properties:
            conditions:
              description: Current service state
              items:
                description: "Condition contains details for one aspect of the current
                  state of this API Resource."
                properties:
                  lastTransitionTime:
                    description: lastTransitionTime is the last time the condition
                      transitioned from one status to another. This should be when
                      the underlying condition changed.  If that is not known, then
                      using the time when the API field changed is acceptable.
                    format: date-time
                    type: string
                  message:
                    description: message is a human readable message indicating details
                      about the transition. This may be an empty string.
                    maxLength: 32768
                    type: string
                  observedGeneration:
                    description: observedGeneration represents the .metadata.generation
                      that the condition was set based upon. For instance, if .metadata.generation
                      is currently 12, but the .status.conditions[x].observedGeneration
                      is 9, the condition is out of date with respect to the current
                      state of the instance.
                    format: int64
                    minimum: 0
                    type: integer
                  reason:
                    description: reason contains a programmatic identifier indicating
                      the reason for the condition's last transition. Producers of
                      specific condition types may define expected values and meanings
                      for this field, and whether the values are considered a guaranteed
                      API. The value should be a CamelCase string. This field may
                      not be empty.
                    maxLength: 1024
                    minLength: 1
                    pattern: ^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$
                    type: string
                  status:
                    description: status of the condition, one of True, False, Unknown.
                    enum:
                    - "True"
                    - "False"
                    - Unknown
                    type: string
                  type:
                    description: type of condition in CamelCase or in foo.example.com/CamelCase.
                      --- Many .condition.type values are consistent across resources
                      like Available, but because arbitrary conditions can be useful
                      (see .node.status.conditions), the ability to deconflict is
                      important. The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)
                    maxLength: 316
                    pattern: ^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$
                    type: string
                required:
                - lastTransitionTime
                - message
                - reason
                - status
                - type
                type: object
              type: array
              x-kubernetes-list-map-keys:
              - type
              x-kubernetes-list-type: map
            loadBalancer:
              description: LoadBalancer contains the current status of the load-balancer,
                if one is present.
              properties:
                ingress:
                  description: Ingress is a list containing ingress points for the
                    load-balancer. Traffic intended for the service should be sent
                    to these ingress points.
                  items:
                    description: 'LoadBalancerIngress represents the status of a load-balancer
                      ingress point: traffic intended for the service should be sent
                      to an ingress point.'
                    properties:
                      hostname:
                        description: Hostname is set for load-balancer ingress points
                          that are DNS based (typically AWS load-balancers)
                        type: string
                      ip:
                        description: IP is set for load-balancer ingress points that
                          are IP based (typically GCE or OpenStack load-balancers)
                        type: string
                      ports:
                        description: Ports is a list of records of service ports If
                          used, every port defined in the service should have an entry
                          in it
                        items:
                          properties:
                            error:
                              description: 'Error is to record the problem with the
                                service port The format of the error shall comply
                                with the following rules: - built-in error values
                                shall be specified in this file and those shall use   CamelCase
                                names - cloud provider specific error values must
                                have names that comply with the   format foo.example.com/CamelCase.
                                --- The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)'
                              maxLength: 316
                              pattern: ^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$
                              type: string
                            port:
                              description: Port is the port number of the service
                                port of which status is recorded here
                              format: int32
                              type: integer
                            protocol:
                              default: TCP
                              description: 'Protocol is the protocol of the service
                                port of which status is recorded here The supported
                                values are: "TCP", "UDP", "SCTP"'
                              type: string
                          required:
                          - port
                          - protocol
                          type: object
                        type: array
                        x-kubernetes-list-type: atomic
                    type: object
                  type: array
              type: object
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
