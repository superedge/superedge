package manifest

const (
	KinsSecretTemplate = `
apiVersion: v1
data:
  k3stoken: {{ .K3SToken }}
  jointoken: {{ .K3SJoinToken }}
  known_tokens.csv: {{ .KnowToken }}
kind: Secret
metadata:
  labels:
    {{ .KinsResourceLabelKey }}: "yes"
    {{ .UnitName }}: {{ .NodeUnitSuperedge }}
  name: {{ .KinsSecretName }}
  namespace: {{ .KinsNamespace }}
type: Opaque
`
	KinsConfigMapTemplate = `
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    {{ .KinsResourceLabelKey }}: "yes"
    {{ .UnitName }}: {{ .NodeUnitSuperedge }}
  name: {{ .KinsConfigMapName }}
  namespace: {{ .KinsNamespace }}
data:
  kubeconfig.conf: |
    apiVersion: v1
    kind: Config
    clusters:
    - cluster:
        insecure-skip-tls-verify: true
        server: https://{{ .KinsServiceClusterIP }}
      name: default
    contexts:
    - context:
        cluster: default
        namespace: default
        user: default
      name: default
    current-context: default
    users:
    - name: default
      user:
        token: {{ .KnowToken }}
`
)
