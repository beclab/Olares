# 账户

如果安装在 Olares 中的应用，需要同步系统的用户信息，以此作为应用中的用户，可以在应用 chart 中申明`SysEventRegistry`来获得系统中用户事件的回调。

- 用户创建回调申明

  ```yaml
  apiVersion: apr.bytetrade.io/v1alpha1
  kind: SysEventRegistry
  metadata:
    name: user-create-cb
    namespace: "{{ .Release.Namespace }}"
  spec:
    type: subscriber
    event: user.create
    callback: http://app-svc.{{ .Release.Namespace }}:8080/callback/create
  ```
  系统回调
  ```http
  POST /callback/create HTTP/1.1
  Content-Type: application/json

  {
    "name": "user1",
    "role": "workspace-manager",
    "email": "user1@xxx.com"
  }
  ```

- 用户删除回调申明

  ```yaml
  apiVersion: apr.bytetrade.io/v1alpha1
  kind: SysEventRegistry
  metadata:
    name: user-delete-cb
    namespace: "{{ .Release.Namespace }}"
  spec:
    type: subscriber
    event: user.delete
    callback: http://app-svc.{{ .Release.Namespace }}:8080/callback/delete
  ```
  系统回调
  ```http
  POST /callback/delete HTTP/1.1
  Content-Type: application/json

  {
      "name": "user1",
      "email": "user1@xxx.com"
  }
  ```

- 用户激活事件回调

  ```yaml
  apiVersion: apr.bytetrade.io/v1alpha1
  kind: SysEventRegistry
  metadata:
    name: user-active-cb
    namespace: "{{ .Release.Namespace }}"
  spec:
    type: subscriber
    event: user.active
    callback: http://app-svc.{{ .Release.Namespace }}:8080/callback/activate
  ```
  系统回调
  ```http
  POST /callback/activate HTTP/1.1
  Content-Type: application/json

  {
      "name": "user1",
      "email": "user1@xxx.com"
  }
  ```

:::tip
为了能获取系统的回调通知，应用需要定义一个 service，并配置到 registry 中。例如，上面配置的 `app-svc`。
:::