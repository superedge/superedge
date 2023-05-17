package manifest

const KinsAgentTemplate = `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    {{ .KinsResourceLabelKey }}: "yes"
    {{ .UnitName }}: {{ .NodeUnitSuperedge }}
  name: {{ .KinsAgentName }}
  namespace: {{ .KinsNamespace }}
spec:
  selector:
    matchLabels:
      site.superedge.io/kins-role: agent
  template:
    metadata:
      labels:
        site.superedge.io/kins-role: agent
      name: k3s-agent
    spec:
      tolerations:
      - key: "{{ .KinsTaintKey }}"
        operator: "Exists"
        effect: "NoSchedule"
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: {{ .KinsRoleLabelKey }}
                operator: NotIn
                values:
                - {{ .KinsRoleLabelServer }}
              - key: {{ .UnitName }}
                operator: In
                values:
                - {{ .NodeUnitSuperedge }}
      initContainers:
      - name: mkcgroup
        image: {{ .K3SAgentImage }}
        imagePullPolicy: IfNotPresent
        command:
          - /bin/sh
          - -c
          - |
            for d in $(ls /sys/fs/cgroup)
            do
              mkdir -p /sys/fs/cgroup/$d/edgek3s
            done
        volumeMounts:
          - name: host-sys
            mountPath: /sys
      containers:
      - name: agent
        image: {{ .K3SAgentImage }}
        securityContext:
          privileged: true
        command: ["/k3s"]
        env:
        - name: K3S_JOIN_TOKEN
          valueFrom:
            secretKeyRef:
              name: {{ .KinsSecretName }}
              key: jointoken
              optional: false
        - name: K3S_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        args: 
        - agent
        - --container-runtime-endpoint=/run/kins.sock
        - --server=https://{{ .KinsServerEndpoint }}
        - --token=$(K3S_JOIN_TOKEN)
        - --kubelet-arg=--cgroup-root=/edgek3s
        - --kubelet-arg=--root-dir=/data/edge/rancher-kubelet
        ports:
        - containerPort: 10250
        volumeMounts:
        - name: host-run
          mountPath: /run
          mountPropagation: "Bidirectional"
        - name: host-dev
          mountPath: /dev
        - name: host-sys
          mountPath: /sys
          mountPropagation: "Bidirectional"
        - name: lib-modules
          mountPath: /lib/modules
          readOnly: true
        - name: host-containerd
          mountPath: /var/lib/containerd
        - name: host-docker
          mountPath: /var/lib/docker
        - name: host-kubelet-log
          mountPath: /data/edge/log/pods
        - name: k3sroot
          mountPath: /data/edge/rancher-kubelet
          mountPropagation: "Bidirectional"
        - name: rancher-root
          mountPath: /var/lib/rancher
        - name: rancher-etc
          mountPath: /etc/rancher
      volumes:
        - hostPath:
            path: /run
          name: host-run
        - hostPath:
            path: /lib/modules
          name: lib-modules
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
        - name: host-kubelet-log
          hostPath:
            path: /data/edge/log/pods
        - name: k3sroot
          hostPath:
            path: /data/edge/rancher-kubelet
            type: DirectoryOrCreate
        - hostPath:
            path: /data/edge/rancher-root
          name: rancher-root
        - hostPath:
            path: /data/edge/rancher-etc
          name: rancher-etc`
