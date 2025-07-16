---
outline: [2, 3]
---

# deployment.yaml

The system application need to be installed under the `user-space` namespace. Therefore, certain modifications are required:

1. Modify the `deployment.yaml` file in the Olares Application Chart.
2. Change the original namespace of `deployment` and `service` to `user-space-{\\{ .Values.bfl.username }}`

   ```Yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: desktop-dev
     namespace: user-space-{{ .Values.bfl.username }}
   ```

3. Add `annotations` and `labels` according to the configuration in the `deployment.yaml` file of the app in **Olares**.

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

     # Configuration of entrances here should be consistent with the configuration in OlaresManifest.yaml.
     applications.app.bytetrade.io/entrances: '[{"name":"desktop-frontend-dev", "host":"desktop-svc-dev", "port":80,"title":"Desktop-dev"}]'
   ```

4. Modify service

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
         targetPort: 8080  # Please note, the port of the Node.js dev container is 8080. please switch to this port.
   ```

5. Modify the section of entrances in `OlaresManifest.yaml`

   ```Yaml
   entrances:
   - name: desktop-frontend-dev # Same with annotation in deployment
     host: desktop-svc-dev # Same with the name in service
     port: 80
     icon: https://file.bttcdn.com/appstore/default/defaulticon.webp
     title: Desktop-dev
     authLevel: private
     openMethod: default
   ```

6. Add service to provide `app-service` installation check

   ```Yaml
   # provide `app-service` installation check
   ---
   apiVersion: v1
   kind: Service
   metadata:
     name: desktop-svc-dev  # Same with the name in original service
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

7. If you need to add a `local cache` or require access to the `user directory` in `JuiceFS`, you can add

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

:::details Example of a complete `deployment.yaml` file

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
    applications.app.bytetrade.io/icon: https://docs-dev.olares.com/icon.png
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
      - name: olares-sidecar-config
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
        - name: olares-sidecar-init
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
        - name: olares-envoy-sidecar
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
          - name: olares-sidecar-config
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
