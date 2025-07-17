---
outline: [2, 3]
---

# deployment.md

由于系统应用需要安装到 `user-space` 的 namespace 下，所以需要做一些特殊修改。

1. 修改 chart 包中的 `deployment.yaml` 文件。
2. 先将原有的 deployment 和 service 对应的 namespace 改为 `user-space-{\{ .Values.bfl.username }}`。
   ```Yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: desktop-dev
     namespace: user-space-{{ .Values.bfl.username }}
   ```

3. 参照 Olares 中的应用对应 `deployment.yaml` 文件配置，添加 `annotation` 和 `label`。

   ```Yaml
   metadata:
   name: desktop-dev
   namespace: user-space-{{ .Values.bfl.username }}
   labels:
     app: desktop-dev
     applications.app.bytetrade.io/name: desktop-dev
     applications.app.bytetrade.io/owner: {{ .Values.bfl.username }}
     applications.app.bytetrade.io/author: bytetrade.io
   annotations:
     applications.app.bytetrade.io/icon: https://docs-dev.olares.com/icon.png
     applications.app.bytetrade.io/title: Desktop-dev
     applications.app.bytetrade.io/version: '0.0.1'

    # 此处的 entrances 配置要与 OlaresManifest.yaml 中配置保持一致
     applications.app.bytetrade.io/entrances: '[{"name":"desktop-frontend-dev", "host":"desktop-svc-dev", "port":80,"title":"Desktop-dev"}]'
   ```

4. 修改 service。

    ```Yaml
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: desktop-svc-dev
      namespace: user-space-{{ .Values.bfl.username }}
    spec:
      selector:
      app: desktop-dev
      ports:
      - protocol: TCP
          port: 80
          targetPort: 8080  # 注意，现在 nodejs 的 dev container 端口是 8080，要改成这个端口
    ```

5. 修改 `OlaresManifest.yaml` 中 `entrances` 的内容。

    ```Yaml
    entrances:
    - name: desktop-frontend-dev # 与 deployment 上的 annotation 一致
      host: desktop-svc-dev # 与上面的 service 名字一致
      port: 80
      icon: https://file.bttcdn.com/appstore/default/defaulticon.webp
      title: Desktop-dev
      authLevel: private
      openMethod: default
    ```

6. 添加 service 提供 app-service 安装检查。

    ```Yaml
    # 提供 app-service 安装检查
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: desktop-svc-dev  # 必须与原来的 service 同名
      namespace: {{ .Release.Namespace }}
    spec:
      type: ExternalName
      externalName: desktop-svc-dev.user-space-{{ .Values.bfl.username }}.svc.cluster.local
      ports:
        - protocol: TCP
          name: desktop
          port: 80
          targetPort: 80
    ```

7. 如果需要添加本地 cache 或者 juicefs 用户目录的访问，可添加：

    ```Yaml
    volumes:
      - name: appdata
        hostPath:
        type: DirectoryOrCreate
        path: {{ .Values.userspace.appData }}/desktop-dev
    
    - name: userdata
      hostPath:
        type: DirectoryOrCreate
        path: {{ .Values.userspace.userData }}/desktop-dev
    
    - name: appcache
      hostPath:
        type: DirectoryOrCreate
        path: {{ .Values.userspace.appCache }}/desktop-dev
    
    ```

---

:::details 完整 `deployment.yaml` 文件例子
```YAML
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: desktop-dev
  namespace: user-space-{{ .Values.bfl.username }}
  labels:
    app: desktop-dev
    applications.app.bytetrade.io/name: desktop-dev
    applications.app.bytetrade.io/owner: {{ .Values.bfl.username }}
    applications.app.bytetrade.io/author: bytetrade.io
  annotations:
    applications.app.bytetrade.io/icon: https://docs-dev.jointerminus.com/icon.png
    applications.app.bytetrade.io/title: Desktop-dev
    applications.app.bytetrade.io/version: '0.0.1'
    applications.app.bytetrade.io/entrances: '[{"name":"desktop-frontend-dev", "host":"desktop-svc-dev", "port":80,"title":"Desktop-dev"}]'
spec:
  replicas: 1
  selector:
    matchLabels:
      app: desktop-dev
  template:
    metadata:
      labels:
        app: desktop-dev
    spec:
      volumes:
      - name: terminus-sidecar-config
        configMap:
          name: sidecar-configs
          items:
          - key: envoy.yaml
            path: envoy.yaml
      - name: appdata
        hostPath:
          type: DirectoryOrCreate
          path: {{ .Values.userspace.appData }}/desktop-dev

      - name: userdata
        hostPath:
          type: DirectoryOrCreate
          path: {{ .Values.userspace.userData }}/desktop-dev

      - name: appcache
        hostPath:
          type: DirectoryOrCreate
          path: {{ .Values.userspace.appCache }}/desktop-dev

      initContainers:
        - name: terminus-sidecar-init
          image: openservicemesh/init:v1.2.3
          imagePullPolicy: IfNotPresent
          securityContext:
            privileged: true
            capabilities:
              add:
              - NET_ADMIN
            runAsNonRoot: false
            runAsUser: 0
          command:
          - /bin/sh
          - -c
          - |
            iptables-restore --noflush <<EOF
            # sidecar interception rules
            *nat
            :PROXY_IN_REDIRECT - [0:0]
            :PROXY_INBOUND - [0:0]
            -A PROXY_IN_REDIRECT -p tcp -j REDIRECT --to-port 15003
            -A PROXY_INBOUND -p tcp --dport 15000 -j RETURN
            -A PROXY_INBOUND -p tcp -j PROXY_IN_REDIRECT
            -A PREROUTING -p tcp -j PROXY_INBOUND
            COMMIT
            EOF

          env:
          - name: POD_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
      containers:
        - name: desktop
          image: "aboveos/node-ts-dev"
          imagePullPolicy: IfNotPresent
          ports:
            - name: port
              containerPort: 8080
              protocol: TCP
          resources:
            requests:
              cpu: "50m"
              memory: 100Mi
            limits:
              cpu: "0.5"
              memory: 2Gi
          volumeMounts:
          - name: appdata
            mountPath: /opt/code
          - name: appcache
            mountPath: /root/.config
        - name: terminus-envoy-sidecar
          image: envoyproxy/envoy-distroless:v1.25.2
          imagePullPolicy: IfNotPresent
          securityContext:
            allowPrivilegeEscalation: false
            runAsUser: 1000
          ports:
          - name: proxy-admin
            containerPort: 15000
          - name: proxy-inbound
            containerPort: 15003
          volumeMounts:
          - name: terminus-sidecar-config
            readOnly: true
            mountPath: /etc/envoy/envoy.yaml
            subPath: envoy.yaml
          command:
          - /usr/local/bin/envoy
          - --log-level
          - debug
          - -c
          - /etc/envoy/envoy.yaml
          env:
          - name: POD_UID
            valueFrom:
              fieldRef:
                fieldPath: metadata.uid
          - name: POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: POD_IP
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
          resources:
            requests:
              cpu: "50m"
              memory: 100Mi
            limits:
              cpu: "0.5"
              memory: 500Mi

---
apiVersion: v1
kind: Service
metadata:
  name: desktop-svc-dev
  namespace: user-space-{{ .Values.bfl.username }}
spec:
  selector:
    app: desktop-dev
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080


---
apiVersion: v1
kind: Service
metadata:
  name: desktop-svc-dev
  namespace: {{ .Release.Namespace }}
spec:
  type: ExternalName
  externalName: desktop-svc-dev.user-space-{{ .Values.bfl.username }}.svc.cluster.local
  ports:
    - protocol: TCP
      name: desktop
      port: 80
      targetPort: 80
```
:::