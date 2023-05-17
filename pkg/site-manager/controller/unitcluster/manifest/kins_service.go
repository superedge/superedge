package manifest

const KinsServiceTemplate = `
apiVersion: v1
kind: Service
metadata:
  name: {{ .KinsServerServiceName }}
  namespace: {{ .KinsNamespace }}
  labels:
    {{ .KinsResourceLabelKey }}: "yes"
    {{ .UnitName }}: {{ .NodeUnitSuperedge }}
spec:
  ports:
  - port: 443
    name: https
    targetPort: 6443
    protocol: TCP
  selector:
    site.superedge.io/nodeunit: {{ .UnitName }}
    {{ .KinsRoleLabelKey }}: {{ .KinsRoleLabelServer }}
`
