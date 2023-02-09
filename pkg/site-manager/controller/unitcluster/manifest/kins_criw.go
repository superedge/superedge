package manifest

const CRIWTemplate = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    {{ .KinsResourceLabelKey }}: "yes"
    {{ .UnitName }}: {{ .NodeUnitSuperedge }}
  name: {{ .CRIWName }}
  namespace: {{ .KinsNamespace }}
spec:
  selector:
    matchLabels:
      site.superedge.io/kins-role: cri-wrapper
  template:
    metadata:
      labels:
        site.superedge.io/kins-role: cri-wrapper
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: {{ .UnitName }}
                operator: In
                values:
                - {{ .NodeUnitSuperedge }}
      hostNetwork: true
      tolerations:
      - key: "{{ .KinsTaintKey }}"
        operator: "Exists"
        effect: "NoSchedule"
      containers:
      - name: cri-wrapper
        image: {{ .KinsCRIWImage }}
        securityContext:
          privileged: true
        volumeMounts:
        - name: host-containerd-conf
          mountPath: /etc/containerd/config.toml
        - name: host-run
          mountPath: /run
        - name: host-var-run
          mountPath: /var/run
          mountPropagation: "Bidirectional"
        - name: host-containerd
          mountPath: /var/lib/containerd
        - name: host-docker
          mountPath: /var/lib/docker
        - name: host-dev
          mountPath: /dev
        - name: host-sys
          mountPath: /sys
        - name: host-cni-bin
          mountPath: /opt/cni
        - name: host-cni-conf
          mountPath: /etc/cni
        - name: host-var-lib-cni
          mountPath: /var/lib/cni
        - name: host-kubelet-log
          mountPath: /data/edge/log/pods
      volumes:
        - hostPath:
            path: /run
          name: host-run
        - hostPath:
            path: /opt/cni
          name: host-cni-bin
        - hostPath:
            path: /etc/cni
          name: host-cni-conf
        - hostPath:
            path: /var/lib/cni
          name: host-var-lib-cni
        - hostPath:
            path: /var/run
          name: host-var-run
        - name: host-dev
          hostPath:
            path: /dev
        - name: host-sys
          hostPath:
            path: /sys
        - name: host-containerd
          hostPath:
            path: /var/lib/containerd
        - name: host-docker
          hostPath:
            path: /var/lib/docker
        - hostPath:
            path: /etc/containerd/config.toml
          name: host-containerd-conf
        - name: host-kubelet-log
          hostPath:
            path: /data/edge/log/pods
`
