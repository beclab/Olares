# 推荐算法配置指导

为 `recommend` 创建应用程序图表时，主要需要配置位于 `templates/` 文件夹中的四个文件：`embedding.yaml` 、`prerank.yaml` 、`rank.yaml` 和 `train.yaml`。


## embedding.yaml

::: details embedding.yaml 示例

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

### 字段介绍

| 选项名称                   | 描述                                                                                       |
| -------------------------- | ------------------------------------------------------------------------------------------ |
| apiVersion                 | 使用的API版本。                                                                             |
| kind                       | 定义了一个CronWorkflow对象。                                                                 |
| metadata.name              | CronWorkflow的名称。                                                                        |
| metadata.namespace         | CronWorkflow所属的命名空间。                                                                 |
| spec.schedule              | Cron表达式，定义了CronWorkflow的调度时间。                                                   |
| spec.startingDeadlineSeconds | CronWorkflow的启动截止时间，表示从调度时间开始的最大延迟时间。                               |
| spec.concurrencyPolicy      | 并发策略，指定了当CronWorkflow下一次调度时间到来时，如何处理当前正在运行的作业。               |
| spec.successfulJobsHistoryLimit | 成功作业的历史记录限制数。                                                                   |
| spec.failedJobsHistoryLimit | 失败作业的历史记录限制数。                                                                   |
| spec.suspend               | 指示是否暂停CronWorkflow的运行。                                                           |
| spec.ttlStrategy.secondsAfterSuccess | 成功作业完成后的存活时间，以秒为单位。                                                 |
| spec.ttlStrategy.secondsAfterCompletion | 作业完成后的存活时间，以秒为单位。                                             |
| spec.ttlStrategy.secondsAfterFailure | 失败作业完成后的存活时间，以秒为单位。                                               |
| spec.workflowSpec.entrypoint | Workflow的入口点。                                                                          |
| spec.workflowSpec.volumes[0].name | 卷的定义，名称为huggingface。                                                          |
| spec.workflowSpec.volumes[0].hostPath.type | 宿主机路径类型，指定为目录或创建目录。                                     |
| spec.workflowSpec.volumes[0].hostPath.path | 宿主机路径。                                                                               |
| spec.workflowSpec.templates[0].name | Workflow模板的名称。                                                                       |
| spec.workflowSpec.templates[0].steps[0][0].name | 步骤的定义，名称。                                                                   |
| spec.workflowSpec.templates[0].steps[0][0].template | 引用的模板名称。                                                                         |
| spec.workflowSpec.templates[1].name | 模板的名称。                                                                               |
| spec.workflowSpec.templates[1].container.image | 容器的镜像名称。                                                                           |
| spec.workflowSpec.templates[1].container.imagePullPolicy | 镜像拉取策略。                                                          |
| spec.workflowSpec.templates[1].container.env[0].name | 环境变量的定义，名称。                                                                 |
| spec.workflowSpec.templates[1].container.env[0].value | 环境变量的值。                                                                            |
| spec.workflowSpec.templates[1].container.env[1].name | 环境变量的定义，名称。                                                                 |
| spec.workflowSpec.templates[1].container.env[1].value | 环境变量的值。                                                                            |
| spec.workflowSpec.templates[1].container.volumeMounts[0].mountPath | 挂载路径的定义。                                                                   |
| spec.workflowSpec.templates[1].container.volumeMounts[0].name | 挂载的卷名称。                                                                         |

## prerank.yaml

::: details prerank.yaml 示例

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

::: details rank.yaml 示例

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

::: details train.yaml 示例

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