# Account

If an app in Olares wants to use the system user as the app's user, it can obtain the user information by defining a `SysEventRegistry` in application chart to receive system user event callbacks.

- Define `user create` callback

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
  Response data
  ```http
  POST /callback/create HTTP/1.1
  Content-Type: application/json

  {
    "name": "user1",
    "role": "workspace-manager",
    "email": "user1@xxx.com"
  }
  ```

- Define `user delete` callback

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
  Response data
  ```http
  POST /callback/delete HTTP/1.1
  Content-Type: application/json

  {
      "name": "user1",
      "email": "user1@xxx.com"
  }
  ```

- Define `user active` callback

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
  Response data
  ```http
  POST /callback/activate HTTP/1.1
  Content-Type: application/json

  {
      "name": "user1",
      "email": "user1@xxx.com"
  }
  ```

:::tip
To receive system callback notifications, the app must define a service and register it. For instance, the `app-svc` mentioned above.
:::