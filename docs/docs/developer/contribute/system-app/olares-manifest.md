# OlaresManifest.yaml

## Permission

If you need to access the interface of `provider`, you can add following content in the `permissions` section of the `OlaresManifest.yaml` file.
```Yaml
permission:
  sysData:
  - dataType: app
    group: service.bfl
    version: v1
    ops:
    - InstallDevApp
```

## Reference variable in env

You can reference the variable in the `env` section of the `deployment.yaml` file.

```Yaml
env:
  - name: OS_APP_KEY
    value: {{ .Values.os.appKey }}   # Please note, you need to replace it with .Values.os.desktop.appKey when submit to the install wizard.
  - name: OS_APP_SECRET
    value: {{ .Values.os.appSecret }} # Please note, you need to replace it with .Values.os.desktop.appSecret when submit to the install wizard.
  - name: OS_SYSTEM_SERVER
    value: system-server.user-system-{{ .Values.bfl.username }}
```

---
:::details Example of a complete `OlaresManifest.yaml` file
```Yaml
olaresManifest.version: 1
olaresManifest.type: app
metadata:
  name: desktop
  icon: https://file.bttcdn.com/appstore/default/defaulticon.webp
  description: app desktop
  appid: desktop
  title: desktop
  version: 0.0.2
  categories:
  - dev
entrances:
- name: desktop-frontend-dev
  host: desktop-svc-dev
  port: 80
  icon: https://file.bttcdn.com/appstore/default/defaulticon.webp
  title: Desktop-dev
  authLevel: private
  openMethod: default
spec:
  versionName: 0.0.1
  requiredMemory: 2Gi
  requiredDisk: 50Mi
  supportArch:
  - amd64
  requiredCpu: 50m
  limitedMemory: 3Gi
  limitedCpu: 1000m
permission:
  appData: true
  appCache: true
  userData: []
  sysData:
  - group: service.bfl
    dataType: app
    version: v1
    ops:
    - UserApps
  - group: service.appstore
    dataType: app
    version: v1
    ops:
    - UninstallDevApp
  - group: service.bfl
    dataType: datastore
    version: v1
    ops:
    - GetKey
    - GetKeyPrefix
    - SetKey
    - DeleteKey
  - group: service.files
    dataType: files
    version: v1
    ops:
    - Query
options:
  resetCookie:
    enabled: false
  dependencies:
  - name: olares
    version: '>=0.1.0'
    type: system
  appScope:
    clusterScoped: false
    appRef: []
```
:::