package manifest

const KinsServerJoinTemplate = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  labels:
    {{ .KinsResourceLabelKey }}: "yes"
    {{ .UnitName }}: {{ .NodeUnitSuperedge }}
  name: {{ .KinsServerName }}-join
  namespace: {{ .KinsNamespace }}
spec:
  podManagementPolicy: Parallel
  replicas: 2
  serviceName: {{ .KinsServerName }}-join
  selector:
    matchLabels:
      site.superedge.io/nodeunit: {{ .UnitName }}
      site.superedge.io/kins-role: server
      site.superedge.io/server-type: join      
  template:
    metadata:
      labels:
        site.superedge.io/nodeunit: {{ .UnitName }}
        site.superedge.io/kins-role: server
        site.superedge.io/server-type: join        
      name: k3s-server-join
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
                operator: In
                values:
                - {{ .KinsRoleLabelServer }}
              - key: {{ .UnitName }}
                operator: In
                values:
                - {{ .NodeUnitSuperedge }}
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: {{ .KinsRoleLabelKey }}
                operator: In
                values:
                - {{ .KinsRoleLabelServer }}
                - {{ .KinsRoleLabelAgent }}               
            topologyKey: kubernetes.io/hostname
      initContainers:
      - name: mkcgroup
        image: {{ .K3SServerImage }}
        imagePullPolicy: IfNotPresent
        command:
          - /bin/sh
          - -cx
          - |
            set -o pipefail
            for d in $(ls /sys/fs/cgroup)
            do
              mkdir -p /sys/fs/cgroup/$d/edgek3s
            done
        volumeMounts:
          - name: host-sys
            mountPath: /sys
          - name: rancher-root
            mountPath: /var/lib/rancher
      containers:
      - name: server
        image: {{ .K3SServerImage }}
        env:
        - name: K3S_TOKEN
          valueFrom:
            secretKeyRef:
              name: {{ .KinsSecretName }}
              key: k3stoken
              optional: false
        - name: K3S_NODE_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: spec.nodeName
        securityContext:
          privileged: true
        command: ["/k3s"]
        args: 
        - server
        - --container-runtime-endpoint=/run/kins.sock
        - --flannel-backend=none
        - --disable-kube-proxy
        - --disable-cloud-controller
        - --cluster-cidr=169.254.0.0/16
        - --service-cidr={{ .ServiceCIDR }}               
        - --service-node-port-range={{ .KinsNodePortRange }}
        - --cluster-dns={{ .KinsCorednsIP }}
        - --kube-apiserver-arg=--token-auth-file=/etc/edge/known_tokens.csv
        - --kubelet-arg=--cgroup-root=/edgek3s
        - --kubelet-arg=--root-dir=/data/edge/rancher-kubelet
        - --server=https://{{ .KinsServerName }}-0.{{ .KinsServerName }}-init.{{ .KinsNamespace }}.svc:6443
        lifecycle: 
          preStop:
            exec:
              command: 
                - /bin/sh
                - -c
                - |
                  set -x
                  set -o pipefail
                  rm -rf /var/lib/rancher/*
                  rm -rf /etc/rancher/*
                  rm -rf /data/edge/rancher-kubelet/*
                  rm -rf /data/edge/log/*
        ports:
        - containerPort: 6443
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
        - mountPath: /etc/edge/
          name: token
          readOnly: true
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
            path: /data/edge/rancher-etc
          name: rancher-etc
        - secret:
            defaultMode: 420
            secretName: {{ .KinsSecretName }}
            items:
              - key: known_tokens.csv
                path: known_tokens.csv
          name: token
  volumeClaimTemplates:
  - metadata:
      name: rancher-root
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "kins-localvolume"
      resources:
        requests:
          storage: 10Gi
`
