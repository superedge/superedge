package manifests

const AppTopolvmCRD = "topolvm-crd.yaml"

const AppTopolvmCRDYaml = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.6.0
  creationTimestamp: null
  name: logicalvolumes.topolvm.cybozu.com
spec:
  group: topolvm.cybozu.com
  names:
    kind: LogicalVolume
    listKind: LogicalVolumeList
    plural: logicalvolumes
    singular: logicalvolume
  scope: Cluster
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        description: LogicalVolume is the Schema for the logicalvolumes API
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
            description: LogicalVolumeSpec defines the desired state of LogicalVolume
            properties:
              deviceClass:
                type: string
              name:
                type: string
              nodeName:
                type: string
              size:
                anyOf:
                - type: integer
                - type: string
                pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                x-kubernetes-int-or-string: true
            required:
            - name
            - nodeName
            - size
            type: object
          status:
            description: LogicalVolumeStatus defines the observed state of LogicalVolume
            properties:
              code:
                description: A Code is an unsigned 32-bit error code as defined in
                  the gRPC spec.
                format: int32
                type: integer
              currentSize:
                anyOf:
                - type: integer
                - type: string
                pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                x-kubernetes-int-or-string: true
              message:
                type: string
              volumeID:
                description: 'INSERT ADDITIONAL STATUS FIELD - define observed state
                  of cluster Important: Run "make" to regenerate code after modifying
                  this file'
                type: string
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
