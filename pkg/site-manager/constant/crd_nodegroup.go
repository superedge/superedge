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

const CRDNodegroupDefinitionYamlFileName = "site.superedge.io_nodegroups.yaml"

const CRDNodegroupDefinitionYaml = `
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.4.1
  creationTimestamp: null
  name: nodegroups.site.superedge.io
spec:
  group: site.superedge.io
  names:
    categories:
    - all
    kind: NodeGroup
    listKind: NodeGroupList
    plural: nodegroups
    shortNames:
    - ng
    singular: nodegroup
  scope: Cluster
  versions:
  - additionalPrinterColumns:
    - jsonPath: .status.unitnumber
      name: UNITS
      type: integer
    - jsonPath: .metadata.creationTimestamp
      name: AGE
      type: date
    name: v1
    schema:
      openAPIV3Schema:
        description: NodeGroup is the Schema for the nodegroups API
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
            description: NodeGroupSpec defines the desired state of NodeGroup
            properties:
              nodeunits:
                description: If specified, If nodeUnit exists, join NodeGroup directly
                items:
                  type: string
                type: array
              selector:
                description: If specified, Label selector for nodeUnit.
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
              workload:
                description: If specified, Nodegroup bound workload
                items:
                  properties:
                    name:
                      description: workload name
                      type: string
                    selector:
                      description: If specified, Label selector for workload.
                      properties:
                        annotations:
                          additionalProperties:
                            type: string
                          description: If specified, select node to join nodeUnit
                            according to Annotations
                          type: object
                        matchExpressions:
                          description: matchExpressions is a list of label selector
                            requirements. The requirements are ANDed.
                          items:
                            description: A label selector requirement is a selector
                              that contains values, a key, and an operator that relates
                              the key and values.
                            properties:
                              key:
                                description: key is the label key that the selector
                                  applies to.
                                type: string
                              operator:
                                description: operator represents a key's relationship
                                  to a set of values. Valid operators are In, NotIn,
                                  Exists and DoesNotExist.
                                type: string
                              values:
                                description: values is an array of string values.
                                  If the operator is In or NotIn, the values array
                                  must be non-empty. If the operator is Exists or
                                  DoesNotExist, the values array must be empty. This
                                  array is replaced during a strategic merge patch.
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
                    type:
                      description: workload type, Value can be pod, deploy, ds, service,
                        job, st
                      type: string
                  type: object
                type: array
            type: object
          status:
            description: NodeGroupStatus defines the observed state of NodeGroup
            properties:
              nodeunits:
                description: Nodeunit contained in nodegroup
                items:
                  type: string
                type: array
              unitnumber:
                default: 0
                description: NodeUnit that is number in nodegroup
                type: integer
              workloadstatus:
                description: The status of the workload in the nodegroup in each nodeunit
                items:
                  description: NodeGroupStatus defines the observed state of NodeGroup
                  properties:
                    notreadyunit:
                      description: workload NotReady Units
                      items:
                        type: string
                      type: array
                    readyunit:
                      description: workload Ready Units
                      items:
                        type: string
                      type: array
                    workloadname:
                      description: workload Name
                      type: string
                  type: object
                type: array
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
