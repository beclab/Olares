# Configuration Guideline for Recommend

When creating an application chart for `recommend`, you'll primarily need to configure the four files located in the `templates/` folder: `embedding.yaml`, `prerank.yaml`, `rank.yaml`, `train.yaml`.


## embedding.yaml

::: details embedding.yaml Example

```Yaml
apiVersion: argoproj.io/v1alpha1
kind: CronWorkflow
metadata:
  name: user-embedding-r4sport
  namespace: {{ .Release.Namespace }}
spec:
  schedule: '0 */1 * * *'
  startingDeadlineSeconds: 0
  concurrencyPolicy: Replace
  successfulJobsHistoryLimit: 1
  failedJobsHistoryLimit: 1
  suspend: false
  ttlStrategy:
    secondsAfterSuccess: 3600
    secondsAfterCompletion: 3600
    secondsAfterFailure: 3600
  workflowSpec:
    entrypoint: userEmbeddingFlow
    volumes:
      - name: huggingface
        hostPath:
          type: DirectoryOrCreate
          path: >-
            {{ .Values.userspace.appData }}/rss/model/huggingface
    templates:
      - name: userEmbeddingFlow
        steps:
          - - name: user-embedding
              template: user-embedding-template
      - name: user-embedding-template
        container:
          image: 'beclab/r4userembedding'
          imagePullPolicy: Always
          env:
            - name: KNOWLEDGE_BASE_API_URL
              value: {{ .Values.apiUrl }}
            - name: TERMINUS_RECOMMEND_SOURCE_NAME
              value: r4sport
          volumeMounts:
            - mountPath: /root/.cache/huggingface
              name: huggingface
```

:::

### Field Description
                                                                         |
| Field                      | Description                                                                                 |
| -------------------------- | ------------------------------------------------------------------------------------------ |
| apiVersion                 | The API version in use.                                                                     |
| kind                       | Defines a CronWorkflow object.                                                               |
| metadata.name              | The name of the CronWorkflow.                                                                |
| metadata.namespace         | The namespace that the CronWorkflow belongs to.                                               |
| spec.schedule              | Cron expression, defines the scheduling time of the CronWorkflow.                            |
| spec.startingDeadlineSeconds | The start deadline of the CronWorkflow, represents the maximum delay time from the scheduled time. |
| spec.concurrencyPolicy      | Concurrency policy, specifies how to handle the currently running job when the next schedule time of the CronWorkflow arrives. |
| spec.successfulJobsHistoryLimit | The limit of the successful job history record.                                             |
| spec.failedJobsHistoryLimit | The limit of the failed job history record.                                                  |
| spec.suspend               | Indicates whether to suspend the operation of the CronWorkflow.                            |
| spec.ttlStrategy.secondsAfterSuccess | The time to live after the successful job is completed, in seconds.                   |
| spec.ttlStrategy.secondsAfterCompletion | The time to live after the job is completed, in seconds.                        |
| spec.ttlStrategy.secondsAfterFailure | The time to live after the failed job is completed, in seconds.                    |
| spec.workflowSpec.entrypoint | The entry point of the Workflow.                                                            |
| spec.workflowSpec.volumes[0].name | The definition of the volume, the name is huggingface.                                  |
| spec.workflowSpec.volumes[0].hostPath.type | The host machine path type, specified as a directory or create a directory.    |
| spec.workflowSpec.volumes[0].hostPath.path | The path of the host machine.                                                           |
| spec.workflowSpec.templates[0].name | The name of the Workflow template.                                                        |
| spec.workflowSpec.templates[0].steps[0][0].name | The definition of the step, its name.                                                |
| spec.workflowSpec.templates[0].steps[0][0].template | The name of the referenced template.                                                  |
| spec.workflowSpec.templates[1].name | The name of the template.                                                                |
| spec.workflowSpec.templates[1].container.image | The image name of the container.                                                        |
| spec.workflowSpec.templates[1].container.imagePullPolicy | The image pull policy.                                                 |
| spec.workflowSpec.templates[1].container.env[0].name | The definition of the environment variable, its name.                               |
| spec.workflowSpec.templates[1].container.env[0].value | The value of the environment variable.                                               |
| spec.workflowSpec.templates[1].container.env[1].name | The definition of the environment variable, its name.                               |
| spec.workflowSpec.templates[1].container.env[1].value | The value of the environment variable.                                               |
| spec.workflowSpec.templates[1].container.volumeMounts[0].mountPath | The definition of the mount path.                                                  |
| spec.workflowSpec.templates[1].container.volumeMounts[0].name | The name of the mounted volume.                                                     |

## prerank.yaml

::: details prerank.yaml Example

```Yaml
apiVersion: argoproj.io/v1alpha1
kind: CronWorkflow
metadata:
  name: prerank-r4sport
  namespace: {{ .Release.Namespace }}
spec:
  schedule: '*/5 * * * *'
  startingDeadlineSeconds: 0
  concurrencyPolicy: Replace
  successfulJobsHistoryLimit: 1
  failedJobsHistoryLimit: 1
  suspend: false
  ttlStrategy:
    secondsAfterSuccess: 3600
    secondsAfterCompletion: 3600
    secondsAfterFailure: 3600
  workflowSpec:
    entrypoint: algorithm
    volumes:
      - name: nfs
        hostPath:
          type: DirectoryOrCreate
          path: >-
            {{ .Values.userspace.appData }}/rss/data
      - name: juicefs
        hostPath:
          type: DirectoryOrCreate
          path: >-
            {{ .Values.userspace.appData }}/rss/data
    templates:
      - name: algorithm
        steps:
          - - name: recall
              template: recall-template
          - - name: prerank
              template: prerank-template
      - name: recall-template
        container:
          image: 'beclab/r4recall:v0.0.5'
          imagePullPolicy: Always
          env:
            - name: KNOWLEDGE_BASE_API_URL
              value: {{ .Values.apiUrl }}
            - name: NFS_ROOT_DIRECTORY
              value: /nfs
            - name: JUICEFS_ROOT_DIRECTORY
              value: /juicefs
            - name: ALGORITHM_FILE_CONFIG_PATH
              value: /usr/config/
            - name: TERMINUS_RECOMMEND_SOURCE_NAME
              value: r4sport
            - name: SUPPORT_LANGUAGE
              value: en
            - name: SUPPORT_TIMELINESS
              value: '0'
            - name: SYNC_PROVIDER
              value: bytetrade
            - name: SYNC_FEED_NAME
              value: sport
            - name: SYNC_MODEL_NAME
              value: bert_v2
          volumeMounts:
            - mountPath: /nfs
              name: nfs
            - mountPath: /juicefs
              name: juicefs
      - name: prerank-template
        container:
          image: 'beclab/r4prerank:v0.0.5'
          imagePullPolicy: Always
          env:
            - name: KNOWLEDGE_BASE_API_URL
              value: {{ .Values.apiUrl }}
            - name: NFS_ROOT_DIRECTORY
              value: /nfs
            - name: JUICEFS_ROOT_DIRECTORY
              value: /juicefs
            - name: ALGORITHM_FILE_CONFIG_PATH
              value: /usr/config/
            - name: TERMINUS_RECOMMEND_SOURCE_NAME
              value: r4sport
            - name: SUPPORT_LANGUAGE
              value: en
            - name: SUPPORT_TIMELINESS
              value: '0'
          volumeMounts:
            - mountPath: /nfs
              name: nfs
            - mountPath: /juicefs
              name: juicefs

```

:::


## rank.yaml

::: details rank.yaml Example

```Yaml
apiVersion: argoproj.io/v1alpha1
kind: CronWorkflow
metadata:
  name: rank-r4sport
  namespace: {{ .Release.Namespace }}
spec:
  schedule: '*/5 * * * *'
  startingDeadlineSeconds: 0
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 1
  failedJobsHistoryLimit: 1
  suspend: false
  ttlStrategy:
    secondsAfterSuccess: 3600
    secondsAfterCompletion: 3600
    secondsAfterFailure: 3600
  workflowSpec:
    entrypoint: rankFlow
    volumes:
      - name: model
        hostPath:
          type: DirectoryOrCreate
          path: >-
            {{ .Values.userspace.appData }}/rss/model
    templates:
      - name: rankFlow
        steps:
          - - name: extractor
              template: extractor-template
          - - name: rank
              template: rank-template
      - name: extractor-template
        container:
          image: 'beclab/r4extractor:v0.0.5'
          imagePullPolicy: Always
          env:
            - name: KNOWLEDGE_BASE_API_URL
              value: {{ .Values.apiUrl }}
            - name: TERMINUS_RECOMMEND_SOURCE_NAME
              value: r4sport
          volumeMounts:
            - mountPath: /opt/rank_model
              name: model
      - name: rank-template
        container:
          image: 'beclab/r4rank'
          imagePullPolicy: Always
          env:
            - name: KNOWLEDGE_BASE_API_URL
              value: {{ .Values.apiUrl }}
            - name: TERMINUS_RECOMMEND_SOURCE_NAME
              value: r4sport
          volumeMounts:
            - mountPath: /opt/rank_model
              name: model


```

:::


## train.yaml

::: details train.yaml Example

```Yaml
apiVersion: argoproj.io/v1alpha1
kind: CronWorkflow
metadata:
  name: rank-r4sport
  namespace: {{ .Release.Namespace }}
spec:
  schedule: '*/5 * * * *'
  startingDeadlineSeconds: 0
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 1
  failedJobsHistoryLimit: 1
  suspend: false
  ttlStrategy:
    secondsAfterSuccess: 3600
    secondsAfterCompletion: 3600
    secondsAfterFailure: 3600
  workflowSpec:
    entrypoint: rankFlow
    volumes:
      - name: model
        hostPath:
          type: DirectoryOrCreate
          path: >-
            {{ .Values.userspace.appData }}/rss/model
    templates:
      - name: rankFlow
        steps:
          - - name: extractor
              template: extractor-template
          - - name: rank
              template: rank-template
      - name: extractor-template
        container:
          image: 'beclab/r4extractor:v0.0.5'
          imagePullPolicy: Always
          env:
            - name: KNOWLEDGE_BASE_API_URL
              value: {{ .Values.apiUrl }}
            - name: TERMINUS_RECOMMEND_SOURCE_NAME
              value: r4sport
          volumeMounts:
            - mountPath: /opt/rank_model
              name: model
      - name: rank-template
        container:
          image: 'beclab/r4rank'
          imagePullPolicy: Always
          env:
            - name: KNOWLEDGE_BASE_API_URL
              value: {{ .Values.apiUrl }}
            - name: TERMINUS_RECOMMEND_SOURCE_NAME
              value: r4sport
          volumeMounts:
            - mountPath: /opt/rank_model
              name: model

```
:::