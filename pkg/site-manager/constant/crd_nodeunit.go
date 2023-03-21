/*
Copyright 2021 The SuperEdge Authors.

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

package constant

const CRDNodeUnitDefinitionYamlFileName = "site.superedge.io_nodeunits.yaml"

const CRDNodeUnitDefinitionYaml = `

---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.7.0
  creationTimestamp: null
  name: nodeunits.site.superedge.io
spec:
  group: site.superedge.io
  names:
    kind: NodeUnit
    listKind: NodeUnitList
    plural: nodeunits
    shortNames:
    - nu
    singular: nodeunit
  scope: Cluster
  conversion:
    strategy: Webhook
    webhook:
      clientConfig:
        url: {{ .ConvertWebhookServer }}
        caBundle: {{ .CaCrt}}
      conversionReviewVersions: ["v1"]
  versions:
  - additionalPrinterColumns:
    - jsonPath: .spec.type
      name: TYPE
      type: string
    - jsonPath: .status.readyrate
      name: READY
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: AGE
      type: date
    - jsonPath: .metadata.deletionTimestamp
      name: DELETING
      type: date
    name: v1alpha1
    schema:
      openAPIV3Schema:
        description: NodeUnit is the Schema for the nodeunits API
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
            description: NodeUnitSpec defines the desired state of NodeUnit
            properties:
              nodes:
                description: If specified, If node exists, join nodeunit directly
                items:
                  type: string
                type: array
              selector:
                description: If specified, Label selector for nodes.
                properties:
                  annotations:
                    additionalProperties:
                      type: string
                    description: If specified, select node to join nodeUnit according
                      to Annotations
                    type: object
                  matchExpressions:
                    description: matchExpressions is a list of label selector requirements.
                      The requirements are ANDed.
                    items:
                      description: A label selector requirement is a selector that
                        contains values, a key, and an operator that relates the key
                        and values.
                      properties:
                        key:
                          description: key is the label key that the selector applies
                            to.
                          type: string
                        operator:
                          description: operator represents a key's relationship to
                            a set of values. Valid operators are In, NotIn, Exists
                            and DoesNotExist.
                          type: string
                        values:
                          description: values is an array of string values. If the
                            operator is In or NotIn, the values array must be non-empty.
                            If the operator is Exists or DoesNotExist, the values
                            array must be empty. This array is replaced during a strategic
                            merge patch.
                          items:
                            type: string
                          type: array
                      required:
                      - key
                      - operator
                      type: object
                    type: array
                  matchLabels:
                    additionalProperties:
                      type: string
                    description: matchLabels is a map of {key,value} pairs.
                    type: object
                type: object
              setnode:
                description: If specified, set the relevant properties to the node
                  of nodeunit.
                properties:
                  annotations:
                    additionalProperties:
                      type: string
                    description: If specified, set annotations to all nodes of nodeunit
                    type: object
                  labels:
                    additionalProperties:
                      type: string
                    description: If specified, set labels to all nodes of nodeunit
                    type: object
                  taints:
                    description: If specified, set taints to all nodes of nodeunit
                    items:
                      description: The node this Taint is attached to has the "effect"
                        on any pod that does not tolerate the Taint.
                      properties:
                        effect:
                          description: Required. The effect of the taint on pods that
                            do not tolerate the taint. Valid effects are NoSchedule,
                            PreferNoSchedule and NoExecute.
                          type: string
                        key:
                          description: Required. The taint key to be applied to a
                            node.
                          type: string
                        timeAdded:
                          description: TimeAdded represents the time at which the
                            taint was added. It is only written for NoExecute taints.
                          format: date-time
                          type: string
                        value:
                          description: The taint value corresponding to the taint
                            key.
                          type: string
                      required:
                      - effect
                      - key
                      type: object
                    type: array
                type: object
              taints:
                description: If specified, allow to set taints to nodeunit for the
                  scheduler to choose
                items:
                  description: The node this Taint is attached to has the "effect"
                    on any pod that does not tolerate the Taint.
                  properties:
                    effect:
                      description: Required. The effect of the taint on pods that
                        do not tolerate the taint. Valid effects are NoSchedule, PreferNoSchedule
                        and NoExecute.
                      type: string
                    key:
                      description: Required. The taint key to be applied to a node.
                      type: string
                    timeAdded:
                      description: TimeAdded represents the time at which the taint
                        was added. It is only written for NoExecute taints.
                      format: date-time
                      type: string
                    value:
                      description: The taint value corresponding to the taint key.
                      type: string
                  required:
                  - effect
                  - key
                  type: object
                type: array
              type:
                default: edge
                description: 'Type of nodeunit， vaule: Cloud、Edge'
                type: string
              unschedulable:
                default: false
                description: Unschedulable controls nodeUnit schedulability of new
                  workwolads. By default, nodeUnit is schedulable.
                type: boolean
            type: object
          status:
            description: NodeUnitStatus defines the observed state of NodeUnit
            properties:
              notreadynodes:
                description: Node that is not ready in nodeunit
                items:
                  type: string
                type: array
              readynodes:
                description: Node selected by nodeunit
                items:
                  type: string
                type: array
              readyrate:
                default: '''1/1'''
                description: Node that is ready in nodeunit
                type: string
            type: object
        type: object
    served: true
    storage: false
    subresources:
      status: {}
  - additionalPrinterColumns:
    - jsonPath: .spec.type
      name: TYPE
      type: string
    - jsonPath: .status.readyRate
      name: READY
      type: string
    - jsonPath: .metadata.creationTimestamp
      name: AGE
      type: date
    - jsonPath: .metadata.deletionTimestamp
      name: DELETING
      type: date
    name: v1alpha2
    schema:
      openAPIV3Schema:
        description: NodeUnit is the Schema for the nodeunits API
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
            description: NodeUnitSpec defines the desired state of NodeUnit
            properties:
              autonomyLevel:
                default: L3
                description: AutonomyLevel represent the current node unit autonomous
                  capability, L3(default)'s autonomous area is node, L4's autonomous
                  area is unit. If AutonomyLevel larger than L3, it will create a
                  independent control plane in unit.
                type: string
              nodes:
                description: If specified, If node exists, join nodeunit directly
                items:
                  type: string
                type: array
              selector:
                description: If specified, Label selector for nodes.
                properties:
                  annotations:
                    additionalProperties:
                      type: string
                    description: If specified, select node to join nodeUnit according
                      to Annotations
                    type: object
                  matchExpressions:
                    description: matchExpressions is a list of label selector requirements.
                      The requirements are ANDed.
                    items:
                      description: A label selector requirement is a selector that
                        contains values, a key, and an operator that relates the key
                        and values.
                      properties:
                        key:
                          description: key is the label key that the selector applies
                            to.
                          type: string
                        operator:
                          description: operator represents a key's relationship to
                            a set of values. Valid operators are In, NotIn, Exists
                            and DoesNotExist.
                          type: string
                        values:
                          description: values is an array of string values. If the
                            operator is In or NotIn, the values array must be non-empty.
                            If the operator is Exists or DoesNotExist, the values
                            array must be empty. This array is replaced during a strategic
                            merge patch.
                          items:
                            type: string
                          type: array
                      required:
                      - key
                      - operator
                      type: object
                    type: array
                  matchLabels:
                    additionalProperties:
                      type: string
                    description: matchLabels is a map of {key,value} pairs.
                    type: object
                type: object
              setNode:
                description: If specified, set the relevant properties to the node
                  of nodeunit.
                properties:
                  annotations:
                    additionalProperties:
                      type: string
                    description: If specified, set annotations to all nodes of nodeunit
                    type: object
                  labels:
                    additionalProperties:
                      type: string
                    description: If specified, set labels to all nodes of nodeunit
                    type: object
                  taints:
                    description: If specified, set taints to all nodes of nodeunit
                    items:
                      description: The node this Taint is attached to has the "effect"
                        on any pod that does not tolerate the Taint.
                      properties:
                        effect:
                          description: Required. The effect of the taint on pods that
                            do not tolerate the taint. Valid effects are NoSchedule,
                            PreferNoSchedule and NoExecute.
                          type: string
                        key:
                          description: Required. The taint key to be applied to a
                            node.
                          type: string
                        timeAdded:
                          description: TimeAdded represents the time at which the
                            taint was added. It is only written for NoExecute taints.
                          format: date-time
                          type: string
                        value:
                          description: The taint value corresponding to the taint
                            key.
                          type: string
                      required:
                      - effect
                      - key
                      type: object
                    type: array
                type: object
              type:
                default: edge
                description: 'Type of nodeunit, vaule: cloud, edge, master, other'
                type: string
              unitClusterInfo:
                description: UnitClusterInfo holds configuration for unit cluster
                  information.
                properties:
                  parameters:
                    additionalProperties:
                      type: string
                    description: Parameters holds the parameters for the unit cluster
                      create information
                    type: object
                  storageType:
                    description: StorageType support sqlite(one master node) and built-in
                      etcd(three master node) default is etcd
                    type: string
                type: object
              unitCredentialConfigMapRef:
                description: UnitCredentialConfigMapRef for isolate sensitive NodeUnit
                  credential. site-manager will create one after controller-plane
                  ready
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  fieldPath:
                    description: 'If referring to a piece of an object instead of
                      an entire object, this string should contain a valid JSON/Go
                      field access statement, such as desiredState.manifest.containers[2].
                      For example, if the object reference is to a container within
                      a pod, this would take on a value like: "spec.containers{name}"
                      (where "name" refers to the name of the container that triggered
                      the event) or if no container name is specified "spec.containers[2]"
                      (container with index 2 in this pod). This syntax is chosen
                      only to have some well-defined way of referencing a part of
                      an object. TODO: this design is not final and this field is
                      subject to change in the future.'
                    type: string
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names'
                    type: string
                  namespace:
                    description: 'Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/'
                    type: string
                  resourceVersion:
                    description: 'Specific resourceVersion to which this reference
                      is made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids'
                    type: string
                type: object
              unschedulable:
                default: false
                description: Unschedulable controls nodeUnit schedulability of new
                  workwolads. By default, nodeUnit is schedulable.
                type: boolean
            type: object
          status:
            description: NodeUnitStatus defines the observed state of NodeUnit
            properties:
              notReadyNodes:
                description: Node that is not ready in nodeunit
                items:
                  type: string
                type: array
              readyNodes:
                description: Node selected by nodeunit
                items:
                  type: string
                type: array
              readyRate:
                default: '''1/1'''
                description: Node that is ready in nodeunit
                type: string
              unitClusterStatus:
                description: UnitClusterStatus is not nil, when AutonomyLevel is larger
                  than L3
                properties:
                  addresses:
                    items:
                      description: ClusterAddress contains information for the cluster's
                        address.
                      properties:
                        host:
                          description: The cluster address.
                          type: string
                        path:
                          type: string
                        port:
                          format: int32
                          type: integer
                        type:
                          description: Cluster address type, one of Public, ExternalIP
                            or InternalIP.
                          type: string
                      type: object
                    type: array
                  conditions:
                    items:
                      description: ClusterCondition contains details for the current
                        condition of this cluster.
                      properties:
                        lastProbeTime:
                          description: Last time we probed the condition.
                          format: date-time
                          type: string
                        lastTransitionTime:
                          description: Last time the condition transitioned from one
                            status to another.
                          format: date-time
                          type: string
                        message:
                          description: Human-readable message indicating details about
                            last transition.
                          type: string
                        reason:
                          description: Unique, one-word, CamelCase reason for the
                            condition's last transition.
                          type: string
                        status:
                          description: Status is the status of the condition. Can
                            be True, False, Unknown.
                          type: string
                        type:
                          description: Type is the type of the condition.
                          type: string
                      type: object
                    type: array
                  phase:
                    description: ClusterPhase defines the phase of cluster constructor.
                    type: string
                  resource:
                    description: ClusterResource records the current available and
                      maximum resource quota information for the cluster.
                    properties:
                      allocatable:
                        additionalProperties:
                          anyOf:
                          - type: integer
                          - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
                        description: Allocatable represents the resources of a cluster
                          that are available for scheduling. Defaults to Capacity.
                        type: object
                      allocated:
                        additionalProperties:
                          anyOf:
                          - type: integer
                          - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
                        description: ResourceList is a set of (resource name, quantity)
                          pairs.
                        type: object
                      capacity:
                        additionalProperties:
                          anyOf:
                          - type: integer
                          - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
                        description: Capacity represents the total resources of a
                          cluster.
                        type: object
                    type: object
                  serviceCIDR:
                    type: string
                  version:
                    description: If AutonomyLevel larger than L3, it will create a
                      independent control plane in unit,
                    type: string
                type: object
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
