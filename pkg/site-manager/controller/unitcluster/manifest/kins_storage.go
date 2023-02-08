package manifest

const (
	KinsStorageClassTemplate = `
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: kins-localvolume
  labels:
    {{ .KinsResourceLabelKey }}: "yes"
    {{ .UnitName }}: {{ .NodeUnitSuperedge }}
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
`

	KinsPVTemplate = `
{{range $i, $name := .Nodes}}
apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv-{{ $i }}
  labels:
    {{ .KinsResourceLabelKey }}: "yes"
    {{ .UnitName }}: {{ .NodeUnitSuperedge }}
spec:
  capacity:
    storage: 10Gi
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: kins-localvolume
  hostPath:
    path: /date/edge/rancher-root
    type: DirectoryOrCreate
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - {{ .Name }}
---
{{end}}
`
)
