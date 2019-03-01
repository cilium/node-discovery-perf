apiVersion: batch/v1
kind: Job
metadata:
  name: cilium-nodeperf1
  namespace: cilium
spec:
  template:
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: cilium.io/nodeperf
                operator: In
                values:
                - "1"
      containers:
      - args:
        - "--etcd-config=/var/lib/etcd-config/etcd.config"
        - "--initial-count=INIT_COUNT"
        - "--additional-count=1"
        - "--external-count=INIT_COUNT"
        command:
        - "./node-discovery-perf"
        image: nebril/cilium-nodes-perf
        imagePullPolicy: Always
        name: nodeperf
        volumeMounts:
        - mountPath: /var/lib/etcd-config
          name: etcd-config-path
          readOnly: true
        - mountPath: /var/lib/etcd-secrets
          name: etcd-secrets
          readOnly: true
      restartPolicy: OnFailure
      volumes:
      - configMap:
          defaultMode: 420
          items:
          - key: etcd-config
            path: etcd.config
          name: cilium-config
        name: etcd-config-path
      - name: etcd-secrets
        secret:
          defaultMode: 420
          optional: true
          secretName: cilium-etcd-secrets
---
apiVersion: batch/v1
kind: Job
metadata:
  name: cilium-nodeperf2
  namespace: cilium
spec:
  template:
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: cilium.io/nodeperf
                operator: In
                values:
                - "2"
      containers:
      - args:
        - "--etcd-config=/var/lib/etcd-config/etcd.config"
        - "--initial-count=INIT_COUNT"
        - "--additional-count=1"
        - "--external-count=INIT_COUNT"
        command:
        - "./node-discovery-perf"
        image: nebril/cilium-nodes-perf
        imagePullPolicy: Always
        name: nodeperf
        volumeMounts:
        - mountPath: /var/lib/etcd-config
          name: etcd-config-path
          readOnly: true
        - mountPath: /var/lib/etcd-secrets
          name: etcd-secrets
          readOnly: true
      restartPolicy: OnFailure
      volumes:
      - configMap:
          defaultMode: 420
          items:
          - key: etcd-config
            path: etcd.config
          name: cilium-config
        name: etcd-config-path
      - name: etcd-secrets
        secret:
          defaultMode: 420
          optional: true
          secretName: cilium-etcd-secrets
