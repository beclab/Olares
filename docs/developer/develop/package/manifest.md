---
outline: [2, 3]
---

# OlaresManifest Specification

Every **Olares Application Chart** should include an `OlaresManifest.yaml` file in the root directory. `OlaresManifest.yaml` provides all the essential information about an Olares App. Both the **Olares Market protocol** and the Olares depend on this information to distribute and install applications.

:::info NOTE
Latest Olares Manifest version: `0.12.0`
- Requires Olares OS version 1.12.6 or later
- **Template rendering is no longer supported**: `OlaresManifest.yaml` must not use  <code v-pre>{{ ... }}</code> or other template rendering functions.
- Modified valid values for the `apiVersion` field: added `v3`; the original `v2` format will no longer be supported in Olares OS 1.12.6.
- Added `options.shared` field to indicate shared applications
- Added `spec.accelerator` field for GPU resource declaration
- Added `workloadReplicas` field to declare all workload replica counts
- Added `overlayGateway` field for L2 overlay LAN discovery support
- Added `LLMGatewaySupported` in options for LLM Gateway support
- Added `appCommon` and `externalData` permissions (both default to `false`)
- Added `templateOnly` field to mark template-type applications
- Removed [deprecated fields](#deprecated-fields-0-12-0)
:::

:::tip Upgrading to Manifest 0.12.0
If you maintain an existing app chart, pay special attention to `apiVersion`, `workloadReplicas`, and `permission.externalData` before upgrading. These fields contain breaking changes or new mandatory requirements that may require updates to your existing configurations.
:::

:::details Changelog
`0.11.0`
- Removed deprecated `sysData` field
- Updated shared app example
- Added `apiVersion` field
- Added `sharedEntrance` section

`0.10.0`
- Modified the `categories` field
- Added the `provider` field in the Permission section
- Added the Provider section, to allow apps to expose specific service interfaces within the cluster
- Removed some deprecated fields from the Spec section
- Removed some deprecated fields from the Option section
- Added the `allowMultipleInstall` field, allowing the app to be installed as multiple independent instances
- Added the Envs section, to define environment variables required by the application

`0.9.0`
- Added a `conflict` field in `options` to declare incompatible applications
- Removed `analytics` field in `options`
- Modified the format of the `tailscale` section
- Added a `allowedOutboundPorts` field to allow non-http protocol external access through the specified port
- Modified the format of the `ports` section

`0.8.3`
- Add a `mandatory` field in the `dependencies` section for dependent applications required for the installation
- Add `tailscaleAcls` section to permit applications to open specified ports via Tailscale

`0.8.2`
- Add a `runAsUser` option to force the app to run under non root user

`0.8.1`
- Add a `ports` section to specify exposed ports for UDP or TCP

`0.7.1`
- Add new `authLevel` value `internal`
- Change `spec`>`language` to `spec`>`locale` and support i18n
  :::

Here's an example of what a `OlaresManifest.yaml` file might look like:

::: details OlaresManifest.yaml Example

```yaml
olaresManifest.version: '0.12.0'
olaresManifest.type: app
apiVersion: 'v3'

workloadReplicas:
  helloworld: 1

metadata:
  name: helloworld
  title: Hello World
  description: app helloworld
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp
  version: 0.0.1
  categories:
  - AI

entrances:
- name: helloworld
  port: 8080
  title: Hello World
  host: helloworld
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp
  authLevel: private
  openMethod: default

sharedEntrances:
- name: helloworld
  host: sharedentrances-api
  port: 0
  title: Hello World API
  invisible: true
  authLevel: internal
  icon: https://app.cdn.olares.com/appstore/default/defaulticon.webp

permission:
  appCache: true
  appData: true
  appCommon: true
  userData:
  - Home/Documents/
  externalData: true
  
spec:
  versionName: '0.0.1'
  featuredImage: https://link.to/featured_image.webp
  promoteImage:
  - https://link.to/promote_image1.webp
  - https://link.to/promote_image2.webp
  fullDescription: |
    A full description of your app.
  upgradeDescription: |
    Describe what is new in this upgraded version.
  developer: Developer's Name
  website: https://link.to.your.website
  sourceCode: https://link.to.sourceCode
  submitter: Submitter's Name
  locale:
  - en-US
  - zh-CN
  doc: https://link.to.documents
  supportArch:
  - amd64
  - arm64
  onlyAdmin: true

  accelerator:
  - mode: nvidia
    requiredCpu: "1"
    limitedCpu: "7"
    requiredMemory: 13Gi
    limitedMemory: 34Gi
    requiredDisk: 36Gi
    limitedDisk: 100Gi
    requiredGPUMemory: 12Gi
    limitedGPUMemory: 24Gi
  - mode: cpu  
    requiredCpu: 50m
    limitedCpu: 1000m
    requiredMemory: 12Mi
    limitedMemory: 1000Mi
    requiredDisk: 50Mi
    limitedDisk: 100Gi

options:
  dependencies:
  - name: olares
    type: system
    version: '>=1.12.6-0'
  shared: true
  apiTimeout: 0
  conflicts:
    - name: conflictapp
      type: application

envs:
  - envName: USERNAME
    required: true
    type: string
    editable: true
    applyOnChange: true
    description: 'default username'
    regex: '^[\w\-!@#$%^&*()+={}\[\]:,.?~]{6,}$'
```
:::

## olaresManifest.version

- Type: `string`

As Olares evolves, the configuration specification of `OlaresManifest.yaml` may change. You can identify whether these changes will affect your application by checking the `olaresManifest.version`. The `olaresManifest.version` consists of three integers separated by periods.

- An increase in the **first digit** indicates the introduction of incompatible configuration items. Applications that haven't updated their `OlaresManifest.yaml` will be unable to distribute or install.
- An increase in the **second digit** signifies changes in the mandatory fields for distribution and installation. However, Olares remains compatible with application distribution and installation of previous configuration versions. We recommend developers promptly update and upgrade the application's `OlaresManifest.yaml` file.
- A change in the **third digit** does not affect the application's distribution and installation.

Please use a 3 digit version numbers to indicate the application's configuration version. Here are some examples of valid versions:
```yaml
olaresManifest.version: 0.1.0
olaresManifest.version: '0.1.2'
olaresManifest.version: "0.12.0"
```

## olaresManifest.type

- Type: `string`
- Accepted Value: `app`, `middleware`

Olares currently supports 2 types of applications, which differ in certain fields. This document uses the `app` type as an example to explain each field.

:::info
The `recommend` type is no longer included in the latest OlaresManifest specification.
:::

## apiVersion
- Type: `string`
- Accepted Value: `v1`, `v3`
- Default: `v1`

Starting from `olaresManifest.version: '0.12.0'`, `apiVersion` is upgraded to `v3`, adopting a new resource declaration and shared application configuration format, and removing OlaresManifest template rendering. If this field is omitted, it defaults to `v1`.

Olares OS 1.12.6 supports installation of both `v1` and `v3` format applications, but no longer supports `v2` format shared application installation. 

:::warning `apiVersion: 'v2'` deprecation
- `apiVersion: 'v2'` is used for shared applications in Olares OS 1.12.5 and earlier, and will be gradually discontinued after 1.12.6.
- After upgrading to Olares OS 1.12.6:
  - Already installed `apiVersion: 'v2'` applications are unaffected and can continue to be used, but cannot be upgraded further. We recommend migrating to shared applications with `apiVersion: 'v3'` as soon as possible.
  - Applications with `apiVersion: 'v2'` are no longer shown in the Market and cannot be installed.
:::

## metadata

Basic information about the app shown in the system and Olares Market.

:::info Example
```yaml
metadata:
  name: nextcloud
  title: Nextcloud
  description: The productivity platform that keeps you in control
  icon: https://app.cdn.olares.com/appstore/nextcloud/icon.png
  version: 0.0.2
  categories:
  - Utilities_v112
  - Productivity_v112
```
:::

### name

- Type: `string`
- Accepted Value: `^[a-z][a-z0-9]{0,29}$`

App's namespace in Olares, lowercase alphanumeric characters only. It can be up to 30 characters, and needs to be consistent with `FolderName` and `name` field in `Chart.yaml`.

### title

- Type: `string`

The title of your app shown in the Olares Market. Must be within `30` characters.

### description

- Type: `string`

A short description appears below app title in the Olares Market.

### icon

- Type: `url`

Your app icon that appears in the Olares Market.

The app's icon must be a `PNG` or `WEBP` format file, up to `512 KB`, with a size of `256x256 px`.

### version

- Type: `string`

The **Chart Version** of the application. It should be incremented each time the content in the **Chart** changes. It should follow the [Semantic Versioning 2.0.0](https://semver.org/) and needs to be consistent with the `version` field in `Chart.yaml`.

### categories

- Type: `list<string>`

Used to display your app on different category page in Olares Market.

Accepted Value for OS 1.12:
- `Creativity`
- `Productivity_v112` (displayed as Productivity)
- `Developer Tools`
- `Fun`
- `Lifestyle`
- `Utilities_v112` (displayed as Utilities)
- `AI`

## entrances

The entrances (up to 10) that users can use to access the app. At least 1 is required.

:::info Example
```yaml
entrances:
- name: a
  host: firefox
  port: 3000
  title: Firefox
  authLevel: public
  invisible: false
- name: b
  host: firefox
  port: 3001
  title: admin
```
:::

### name

- Type: `string`
- Accepted Value: `[a-z]([-a-z0-9]*[a-z0-9])?`

  Name of the Entrance. It can be up to `63` characters, and needs to be unique in an app.

### port

- Type: `int`
- Accepted Value: `0-65535`

### host

- Type: `string`
- Accepted Value: `[a-z]([-a-z0-9]*[a-z0-9])?`

  Ingress name of current entrance, lowercase alphanumeric characters and `-` only. It can be up to `63` characters.

### title

- Type: `string`

Title that appears in the Olares desktop after installed. It can be up to `30` characters.

### icon

- Type: `url`
- Optional

Icon that appears in the Olares desktop after installed. The app's icon must be a `PNG` or `WEBP` format file, up to `512 KB`, with a size of `256x256 px`.

### authLevel

- Type: `string`
- Accepted Value: `public`, `private`, `internal`
- Default: `private`
- Optional

Specify the authentication level of the entrance.
- **Public**: Accessible by anyone on the Internet without restrictions.
- **Private**: Requires authorization for access from both internal and external networks.
- **Internal**: Requires authorization for access from external networks. No authentication is required when accessing from within the internal network (via LAN/VPN).

### invisible

- Type: `boolean`
- Default: `false`
- Optional

When `invisible` is `true`, the entrance will not be displayed on the Olares desktop.

### openMethod

- Type: `string`
- Accepted Value: `default`, `iframe`, `window`
- Default: `default`
- Optional

Explicitly defines how to open this entrance in Desktop.

The `iframe` creates a new window within the desktop window through an iframe. The `window` opens a new tab in the browser. The `default` follows the system setting, which is `iframe` by default.

### windowPushState
- Type: `boolean`
- Default: `false`
- Optional

When embedding the application in an iframe on the desktop, the application's URL may change dynamically. Due to browser's same-origin policy, the desktop (parent window) cannot directly detect these changes in the iframe URL. Consequently, if you reopen the application tab, it will display the initial URL instead of the updated one.

To ensure a seamless user experience, you can enable this option by setting it to true. This action prompts the gateway to automatically inject the following code into the iframe. This code sends an event to the parent window (desktop) whenever the iframe's URL changes. As a result, the desktop can track URL changes and open the correct page.

::: details Code
```Javascript
<script>
  (function () {
    if (window.top == window) {
        return;
    }
    const originalPushState = history.pushState;
    const pushStateEvent = new Event("pushstate");
    history.pushState = function (...args) {
      originalPushState.apply(this, args);
      window.dispatchEvent(pushStateEvent);
    };
    window.addEventListener("pushstate", () => {
      window.parent.postMessage(
        {type: "locationHref", message: location.href},
        "*"
      );
    });
  })();
</script>
```
:::


## sharedEntrances

A shared entrance is an internal address provided by a shared application for other applications within the cluster to access. The field configuration for shared entrances is basically the same as for regular entrances. A typical shared entrance configuration is shown below.

:::info Example
```yaml
sharedEntrances:
  - name: ollamav2
    host: sharedentrances-ollama
    port: 0
    title: Ollama API
    icon: https://app.cdn.olares.com/appstore/ollama/icon.png
    invisible: true
    authLevel: internal
```
:::

## ports

Specify exposed ports

:::info Example
```yaml
ports:
- name: rdp-tcp             # Name of the entrance that provides service
  host: windows-svc         # Ingress name of the entrance that provides service
  port: 3389                # Port of the entrance that provides service
  protocol: udp             # Protocol used by the exposed port
  exposePort: 46879         # The port to be exposed can only be assigned to one application at a time within the cluster.
  addToTailscaleAcl: true   # Automatically added to Tailscale's ACL
```
:::

### exposePort
- Type: `int`
- Optional
- Accepted Value: `0-65535`, except reserved ports `22`, `80`, `81`, `443`, `444`, `2379`, `18088`.

Olares will expose the ports you specify for an application, which are accessible via the application domain name in the local network, for example: `84864c1f.your_olares_id.olares.com:46879`. For each port you expose, Olares configures both TCP and UDP with the same port number.

:::info NOTE
The exposed ports can only be accessed on the local network or through a VPN.
:::

### protocol
- Type: `string`
- Optional
- Accepted Value: `udp`, `tcp`

The protocol used for the exposed port. If specified, Olares exposes only the specified protocol. If omitted, Olares exposes both UDP and TCP by default.

### addToTailscaleAcl
- Type: `boolean`
- Optional
- Default: `false`

When the `addToTailscaleAcl` field is set to `true`, the system will automatically assign a random port and add it to the Tailscale ACLs.

## tailscale
- Type: `map`
- Optional

Allow applications to open specified ports in Tailscale ACL (Access Control Lists).

:::info Example
```yaml
tailscale:
  acls:
  - proto: tcp
    dst:
    - "*:46879"
  - proto: "" # Optional. If not specified, all supported protocols are allowed.
    dst:
    -  "*:4557"
```
:::

## permission

### appCache

- Type: `boolean`
- Default: `false`
- Optional

Whether the app needs to create an application directory in the `Cache` folder. When set to `true`, the app can access the `Cache` directory path via `.Values.userspace.appCache` in the deployment YAML.

### appData

- Type: `boolean`
- Default: `false`
- Optional

Whether the app needs to create an application directory in the `Data` folder. When set to `true`, the app can access the `Data` directory path via `.Values.userspace.appData` in the deployment YAML.

### appCommon

- Type: `boolean`
- Default: `false`
- Optional

Whether the app requires read and write permission to the `App Common` folder. When set to `true`, the app can access the `App Common` directory path via `.Values.userspace.appCommon` in the deployment YAML, enabling shared files across multiple apps or nodes (such as AI models).

### externalData

- Type: `boolean`
- Default: `false`
- Optional

Whether the app requires read and write permission to the `External` directory (typically used for accessing mounted NAS or other external disk data). When set to `true`, the app can access the `External` directory path via `.Values.sharedlib` in the deployment YAML.

:::warning
Starting from Manifest 0.12.0, the `externalData` permission is disabled by default. If your app needs access to mounted NAS or external disks, you must explicitly set `externalData: true`. If omitted, permission will be denied.
:::

### userData

- Type: `list<string>`
- Optional

Whether the app requires read and write permission to specific directories in the user's `Home` folder. List all `Home` subdirectories the app needs to access in this field. In the deployment YAML, use `.Values.userspace.userData` to get the root path of the Home directory. All Home subdirectories that need to be mounted must be explicitly declared in this field before they can be accessed.


## spec
Additional information about the application.

:::info Example
```yaml
spec:
  versionName: '10.8.11'
  # The version of the application that this chart contains. It is recommended to enclose the version number in quotes. This value corresponds to the appVersion field in the `Chart.yaml` file. Note that it is not related to the `version` field.

  featuredImage: https://app.cdn.olares.com/appstore/jellyfin/promote_image_1.jpg
  # Displayed on the My Olares page after the app is installed.

  promoteImage:
  - https://app.cdn.olares.com/appstore/jellyfin/promote_image_1.jpg
  - https://app.cdn.olares.com/appstore/jellyfin/promote_image_2.jpg
  - https://app.cdn.olares.com/appstore/jellyfin/promote_image_3.jpg
  fullDescription: |
    Jellyfin is the volunteer-built media solution that puts you in control of your media. Stream to any device from your own server, with no strings attached. Your media, your server, your way.
  upgradeDescription: |
    upgrade descriptions
  developer: Jellyfin
  website: https://jellyfin.org/
  doc: https://jellyfin.org/docs/
  sourceCode: https://github.com/jellyfin/jellyfin
  submitter: Olares
  locale:
  - en-US
  - zh-CN
  # Languages and regions supported on the app's Market page
  
  accelerator:
  - mode: nvidia
    limitedCpu: 7000m
    requiredCpu: 150m
    requiredDisk: 50Mi
    limitedDisk: 500Gi
    limitedMemory: 40Gi
    requiredMemory: 5Gi
    requiredGPUMemory: 1Gi
    limitedGPUMemory: 24Gi
  - mode: cpu
    limitedCpu: 7000m
    requiredCpu: 150m
    requiredDisk: 50Mi
    limitedDisk: 500Gi
    limitedMemory: 40Gi
    requiredMemory: 5Gi
  # If the app requires GPU or other hardware acceleration, list all supported accelerator modes and resource requirements here. For CPU-only scenarios, use mode: cpu; the fields below follow the same format as previous versions.
  
  license:
  - text: GPL-2.0
    url: https://github.com/jellyfin/jellyfin/blob/master/LICENSE
  supportClient:
  - android: https://play.google.com/store/apps/details?id=org.jellyfin.mobile
  - ios: https://apps.apple.com/us/app/jellyfin-mobile/id1480192618
```
:::

### i18n

To add multi-language support for your app in Olares Market:

1. Create an `i18n` folder in the Olares Application Chart root directory.
2. In the `i18n` folder, create separate subdirectories for each supported locale.
3. In each locale subdirectory, place a localized version of the `OlaresManifest.yaml` file.

Olares Market will automatically display the content of the corresponding "OlaresManifest.yaml" file based on users' locale settings.
:::info Example
```
.
├── Chart.yaml
├── README.md
├── OlaresManifest.yaml
├── i18n
│   ├── en-US
│   │   └── OlaresManifest.yaml
│   └── zh-CN
│       └── OlaresManifest.yaml
├── owners
├── templates
│   └── deployment.yaml
└── values.yaml
```
:::
Currently, you can add i18n content for the following fields:
```yaml
metadata:
  description:
  title:
spec:
  fullDescription:
  upgradeDescription:
```

### supportArch
- Type: `list<string>`
- Accepted Value: `amd64`, `arm64`
- Optional

Specifies the CPU architecture supported by the application. Currently only `amd64` and `arm64` are available.

:::info Example
```yaml
spec:
  supportArch:
  - amd64
  - arm64
```
:::

:::info NOTE
Olares does not support mixed-architecture clusters for now.
:::

### onlyAdmin
- Type: `boolean`
- Default: `false`
- Optional

When set to `true`, only the admin can install this app.

### runAsUser
- Type: `boolean`
- Optional

When set to `true`, Olares forces the application to run under user ID `1000` (as a non-root user).

### accelerator
- Type: `map`
- Optional

Declares accelerator compute resources required by the application (such as GPU or integrated graphics). For accelerator-aware apps such as LLMs, image generation, video processing, or AI model serving, use the `spec.accelerator` field to declare resources; do not include the original root-level fields such as `spec.requiredMemory`.

:::info NOTE
When accelerator is used, all related resource requirements (CPU, memory, disk, and GPU memory) must be declared inside this block, instead of relying on the root-level spec.requiredMemory or spec.requiredCpu fields.
:::

Supported modes

- `nvidia`, `nvidia-gb10`, `apple-m`, `strix-halo`, `intel`, `amd`, `intel-gpu`, `amd-gpu`: For apps that support specific GPU/NPU hardware acceleration.
- `cpu`: For conventional apps, or to explicitly force CPU-only computation in accelerator-aware applications.

:::info Example
```yaml
spec:
  accelerator:
  - mode: nvidia  # Supported modes: nvidia, nvidia-gb10, apple-m, strix-halo, mthreads-m1000, intel, amd, intel-gpu, amd-gpu, cpu
    limitedCpu: 7000m
    requiredCpu: 150m
    requiredDisk: 50Mi
    limitedDisk: 500Gi
    limitedMemory: 40Gi
    requiredMemory: 5Gi
    requiredGPUMemory: 1Gi
    limitedGPUMemory: 24Gi
```
:::

For apps that do not require GPU, it is recommended to declare resources using `cpu` mode:
:::info Example
```yaml
spec:
  requiredMemory: 2Gi
  requiredDisk: 50Mi
  requiredCpu: 0.25
  limitedMemory: 10240Mi
  limitedCpu: '4'
```
:::

## workloadReplicas
- Type: `map`

Declares the replica count for each workload in the chart. Starting from 0.12.0, apps must declare replica counts for all workloads in `workloadReplicas`.

Developers must also define the corresponding variables under `workloads` in `values.yaml`, and ensure the workload names in `workloadReplicas` exactly match those under `workloads` in `values.yaml` for compatibility.
:::info OlaresManifest.yaml
```yaml 
workloadReplicas:
  affine: 1
```
:::
:::info values.yaml
```yaml
workloads:
  affine:
    replicaCount: 1
```
:::

When deploying, be sure to reference this value to set the number of replicas for each workload.
:::info deployment.yaml
```yaml 
spec:
  replicas: {{ .Values.workloads.affine.replicaCount }}
```
:::

## middleware
- Type: `map`
- Optional

Olares provides highly available middleware services. Developers do not need to install middleware repeatedly. Add the required middleware here, then use the corresponding middleware values in the application's deployment YAML file.

Use the `scripts` field to specify scripts that should be executed after the database is created. Additionally, use the `extension` field to add the corresponding extension in the  database.

:::info NOTE
MongoDB, MySQL, MariaDB, MinIO, and RabbitMQ must first be installed by an admin from the Market before they can be used by other applications.
:::

### PostgreSQL

:::info Example
```yaml
middleware:
  postgres:
    username: immich
    databases:
    - name: immich
      extensions:
      - vectors
      - earthdistance
      scripts:
      - BEGIN;
      - ALTER DATABASE $databasename SET search_path TO "$user", public, vectors;
      - ALTER SCHEMA vectors OWNER TO $dbusername;
      - COMMIT;
      # The OS provides two variables, $databasename and $dbusername, which will be replaced by Olares Application Runtime when the command is executed.
```
:::
Use the middleware information in deployment YAML
```yaml
# For PostgreSQL, the corresponding value is as follows
- name: DB_POSTGRESDB_DATABASE # The database name you configured in OlaresManifest, specified in middleware.postgres.databases[i].name
  value: {{ .Values.postgres.databases.<dbname> }}
- name: DB_POSTGRESDB_HOST
  value: {{ .Values.postgres.host }}
- name: DB_POSTGRESDB_PORT
  value: "{{ .Values.postgres.port }}"
- name: DB_POSTGRESDB_USER
  value: {{ .Values.postgres.username }}
- name: DB_POSTGRESDB_PASSWORD
  value: {{ .Values.postgres.password }}
```

### Redis
:::info Example
```yaml
middleware:
  redis:
    password: password
    namespace: db0
```
:::
Use the middleware information in deployment YAML
```yaml
# For Redis, the corresponding value is as follows
host --> {{ .Values.redis.host }}
port --> "{{ .Values.redis.port }}"
password --> "{{ .Values.redis.password }}"
```

### MongoDB
:::info Example
```yaml
middleware:
  mongodb:
    username: chromium
    databases:
    - name: chromium
      script:
      - 'db.getSiblingDB("$databasename").myCollection.insertOne({ x: 111 });'
      # Please make sure each line is a complete query.
```
:::
Use the middleware information in deployment YAML
```yaml
# For MongoDB, the corresponding value is as follows
host --> {{ .Values.mongodb.host }}
port --> "{{ .Values.mongodb.port }}"  # The port and password in the yaml file need to be enclosed in double quotes.
username --> {{ .Values.mongodb.username }}
password --> "{{ .Values.mongodb.password }}" # The port and password in the yaml file need to be enclosed in double quotes.
databases --> "{{ .Values.mongodb.databases }}" # The value type of database is a map. You can get the database using {{ .Values.mongodb.databases.<dbname> }}. The <dbname> is the name you configured in OlaresManifest, specified in middleware.mongodb.databases[i].name
```


### MinIO
:::info Example
```yaml
middleware:
  minio:
    username: miniouser
    buckets:
    - name: mybucket
```
:::
Use the middleware information in deployment YAML
```yaml
# For MinIO, the corresponding value is as follows
- env:
  - name: MINIO_ENDPOINT
    value: '{{ .Values.minio.host }}:{{ .Values.minio.port }}'
  - name: MINIO_PORT
    value: "{{ .Values.minio.port }}"
  - name: MINIO_ACCESS_KEY
    value: {{ .Values.minio.username }}
  - name: MINIO_SECRET_KEY
    value: {{ .Values.minio.password }}
  - name: MINIO_BUCKET
    value: {{ .Values.minio.buckets.mybucket }}
```

### RabbitMQ
:::info Example
```yaml
middleware:
  rabbitmq:
    username: rabbitmquser
    vhosts:
    - name: aaa
```
:::
Use the middleware information in deployment YAML
```yaml
# For RabbitMQ, the corresponding value is as follows
- env:
  - name: RABBITMQ_HOST
    value: '{{ .Values.rabbitmq.host }}'
  - name: RABBITMQ_PORT
    value: "{{ .Values.rabbitmq.port }}"
  - name: RABBITMQ_USER
    value: "{{ .Values.rabbitmq.username }}"
  - name: RABBITMQ_PASSWORD
    value: "{{ .Values.rabbitmq.password }}"
  - name: RABBITMQ_VHOST
    value: "{{ .Values.rabbitmq.vhosts.aaa }}"

user := os.Getenv("RABBITMQ_USER")
password := os.Getenv("RABBITMQ_PASSWORD")
vhost := os.Getenv("RABBITMQ_VHOST")
host := os.Getenv("RABBITMQ_HOST")
portMQ := os.Getenv("RABBITMQ_PORT")
url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, portMQ, vhost)
```
### MariaDB
:::info Example
```yaml
middleware:
  mariadb:
    username: mariadbclient
    databases:
    - name: aaa
```
:::
Use the middleware information in deployment YAML
```yaml
# For MariaDB, the corresponding value is as follows
- env:
  - name: MDB_HOST
    value: '{{ .Values.mariadb.host }}'
  - name: MDB_PORT
    value: "{{ .Values.mariadb.port }}"
  - name: MDB_USER
    value: "{{ .Values.mariadb.username }}"
  - name: MDB_PASSWORD
    value: "{{ .Values.mariadb.password }}"
  - name: MDB_DB
    value: "{{ .Values.mariadb.databases.aaa }}"
```

### MySQL
:::info Example
```yaml
middleware:
  mysql:
    username: mysqlclient
    databases:
    - name: aaa
```
:::
Use the middleware information in deployment YAML
```yaml
# For MySQL, the corresponding value is as follows
- env:
  - name: MDB_HOST
    value: '{{ .Values.mysql.host }}'
  - name: MDB_PORT
    value: "{{ .Values.mysql.port }}"
  - name: MDB_USER
    value: "{{ .Values.mysql.username }}"
  - name: MDB_PASSWORD
    value: "{{ .Values.mysql.password }}"
  - name: MDB_DB
    value: "{{ .Values.mysql.databases.aaa }}"
```

## options

Configure Olares OS related options here.

### policies
- Type: `list<map>`
- Optional

Define detailed access control for subdomains of the app.

:::info Example
```yaml
options:
  policies:
    - uriRegex: /$
      level: two_factor
      oneTime: false
      validDuration: 3600s
      entranceName: gitlab
```
:::

### dependencies
- Type: `list<map>`

Specify the dependencies and requirements for your application. It includes other applications that your app depends on, as well as any specific operating system (OS) version requirements.

If this application requires other dependent applications for proper installation, you should set the `mandatory` field to `true`.

:::info Example
```yaml
options:
  dependencies:
    - name: olares
      version: ">=1.0.0-0"
      type: system
    - name: mongodb
      version: ">=6.0.0-0"
      type: middleware
      mandatory: true # Set this field to true if the dependency needs to be installed first.
```
:::

### conflicts
- Type: `list<map>`
- Optional

List other applications that conflict with this app here. Conflicting apps must be uninstalled before this app can be installed.

:::info Example
```yaml
options:
  conflicts:
  - name: comfyui
    type: application
  - name: comfyuiclient
    type: application
```
:::


### mobileSupported
- Type: `boolean`
- Default: `false`
- Optional

Determine whether the application is compatible with mobile web browsers and can be displayed on the mobile version of Olares Desktop. Enable this option if the app is optimized for mobile web browsers. This will make the app visible and accessible on the mobile version of Olares Desktop.

:::info Example
```yaml
mobileSupported: true
```
:::

### oidc
- Type: `map`
- Optional

The Olares includes a built-in OpenID Connect authentication component to simplify identity verification of users. Enable this option to use OpenID in your app.
```yaml
# OpenID related variables in yaml
{{ .Values.oidc.client.id }}
{{ .Values.oidc.client.secret }}
{{ .Values.oidc.issuer }}
```

:::info Example
```yaml
oidc:
  enabled: true
  redirectUri: /path/to/uri
  entranceName: navidrome
```
:::

### apiTimeout
- Type: `int`
- Optional

Specifies the timeout limit for API providers in seconds. The default value is `15`. Use `0` to allow an unlimited API connection.

:::info Example
```yaml
apiTimeout: 0
```
:::


### allowedOutboundPorts
- Type: `list<int>`
- Optional

​​The specified ports will be opened to allow external access via non-HTTP protocols, such as SMTP.

:::info Example
```yaml
allowedOutboundPorts:
  - 465
  - 587
```
:::

### allowMultipleInstall
- Type: `boolean`
- Default: `false`
- Optional

This application supports deploying multiple independent instances within the same Olares cluster. This setting does not apply to paid applications or clients of shared applications.

### shared
- Type: `boolean`
- Default: `false`
- Optional

When set to `true`, indicates this is a shared application. The app will be installed in the `<appname>-shared` namespace.

When `options.shared: true`:
- `apiVersion` must be set to `'v3'`
- `onlyAdmin` must be set to `true`
- The following `apiVersion: 'v3'` fields are not allowed:
  ```yaml
  spec:
    subCharts:
    - name: ollamaserver
      shared: true
    - name: ollamav2
  options:
    appScope:
      clusterScoped: true
      appRef:
      - ollamav2
  ```

### templateOnly
- Type: `boolean`
- Default: `false`
- Optional

When set to `true`, this application is marked as a template and cannot be installed directly. An instance must be created before use. Because a template app is designed to spawn multiple instances, `allowMultipleInstall` must also be set to `true`.

### LLMGatewaySupported
- Type: `boolean`
- Default: `false`
- Optional

When set to `true`, the application supports LLM Gateway calls. Mainly used for model-related application services.

### overlayGateway
- Type: `map`
- Optional

Declares the app's support for L2 overlay LAN discovery. When enabled, other apps can access this app via an IP address in the local network.
:::info Example
```yaml
overlayGateway:
  enable: true  # Enable L2 overlay LAN discovery, default false
  entrances:
  - port: 8096  # Overlay listening port
    title: Jellyfin  # Overlay entrance name
    workload: jellyfin  # Workload name for OAC validation
    description: "Access Jellyfin using IP address in LAN"
    protocol: tcp  # Supported protocols: tcp/udp; defaults to both tcp and udp if omitted
```
:::

## envs

Declare the environment variables required for your application to run here. You can allow users to manually enter these values or reference existing system environment variables directly.

:::info NOTE
This configuration requires Olares OS version 1.12.2 or later to take effect.
:::

:::info Example
```yaml
envs:
  - envName: ENV_NAME
    # This key will be injected as .Values.olaresEnv.ENV_NAME during deployment.

    required: true
    # Specifies whether a value is required for installation. If set to true and no default is provided, users must input a value, and the value cannot be deleted.

    default: "DEFAULT"
    # The default value of the environment variable; set by the developer and not editable by users

    type: string
    # The data type of the environment variable. Supported types: int, bool, url, ip, domain, email, string, password. If specified, system will validate user input accordingly.

    editable: true
    # Specifies whether the environment variable can be edited after the application is deployed.

    options:
    - title: Windows11
      value: "11"
    - title: Windows10
      value: "10"
    # List of allowed values. Users can only select value from these options.
    # "title" is a user-friendly label, while "value" is what's actually set in the system.

    remoteOptions: https://xxx.xxx/xx
    # URL providing a list of accepted options. The response body should be a JSON-encoded options list.

    multiSelect: true
    splitter: ","
    # When `options` or `remoteOptions` is set, enables multi-select. Selected values are joined with `splitter` and passed to the app as a string.

    regex: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    # The value must match this regular expression.

    valueFrom:
      envName: OLARES_SYSTEM_CLUSTER_DNS_SERVICE
    # Reference the value from a system environment variable. When this is used, manual input is not allowed.
    # All declarable fields  (type, editable, etc.) will be overridden by the referenced variable's attributes; default/value fields are also ignored.

    applyOnChange: true
    # Whether to automatically redeploy the app when this variable changes.
    # If set to false, changes take effect only on upgrade/reinstallation, not on restart.

    description: "DESCRIPTION"
    # Description of this environment variable.
```
:::

To use the values of environment variables in your deployment YAML file, simply use `.Values.olaresEnv.ENV_NAME` in the appropriate place. The system will automatically inject the olaresEnv variables into values during deployment. For example:

:::info Example
```yaml
BACKEND_MAIL_HOST: "{{ .Values.olaresEnv.MAIL_HOST }}"
BACKEND_MAIL_PORT: "{{ .Values.olaresEnv.MAIL_PORT }}"
BACKEND_MAIL_AUTH_USER: "{{ .Values.olaresEnv.MAIL_AUTH_USER }}"
BACKEND_MAIL_AUTH_PASS: "{{ .Values.olaresEnv.MAIL_AUTH_PASS }}"
BACKEND_MAIL_SECURE: "{{ .Values.olaresEnv.MAIL_SECURE }}"
BACKEND_MAIL_SENDER: "{{ .Values.olaresEnv.MAIL_SENDER }}"
```
:::

## Deprecated Fields (0.12.0)

The following fields are deprecated in OlaresManifest 0.12.0. Remove them promptly. Apps containing these fields may be **rejected** during installation.

| Field | Action |
|-------|--------|
| `metadata.appid` | Remove; it is created automatically from `metadata.name` |
| `provider` (top-level) | Remove; use an entrance with `authLevel: internal` for other apps to call instead |
| `permission.provider` | Remove |
| `permission.sysData` | Remove |
| `options.appScope` | Remove; declare shared applications using the `apiVersion: 'v3'` format |
| `spec.subCharts` | Remove; declare shared applications using the `apiVersion: 'v3'` format |
| `spec.requiredMemory`, etc. | Update; declare under `spec.accelerator` with `mode: cpu` |
| OS 1.11 categories: `Blockchain`, `Utilities`, `Social Network`, `Entertainment`, `Productivity` | Use OS 1.12 categories |
