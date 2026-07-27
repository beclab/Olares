---
noindex: true
description: 在 Olares 上管理 AI 应用并与之交互的 API 参考，涵盖应用管理、文本生成、对话及使用统计。
outline: [2, 3]
---

# AI

## API 前缀

`agent.{username}.olares.com/api/controllers/console/api`

## 基本应用管理 API
### 获取 App 列表
- **Request**
  - **URL**: `/apps`
  - **Method**: `GET`
  - **URL Parameters**: `/apps?page=1&limit=30&name=Ashia`
  :::tip
    本文档中列出的大多数 API 都需要`app_id`，可以从该 API 的 Response 中获取该`app_id`。
  :::

### 创建应用
- **Request**
  - **URL**: `/apps`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "name": "TEST",
      "icon": "🤖",
      "icon_background": "#FFEAD5",
      "mode": "agent-chat",
      "description": "JUST A TEST"
    }
    ```

### 获取应用细节
- **Request**
  - **URL**: `/apps/{uuid:app_id}`
  - **Method**: `GET`
  - **Body Example**: `null`

### 删除应用
- **Request**
  - **URL**: `/apps/{uuid:app_id}`
  - **Method**: `DELETE`
  - **Body Example**: `null`

### 复制应用
- **Request**
  - **URL**: `/apps/{uuid:app_id}/copy`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "name": "Ashia-2",
      "icon": "🤖",
      "icon_background": "#FFEAD5",
      "mode": "agent-chat"
    }
    ```

### 应用重命名
- **Request**
  - **URL**: `/apps/{uuid:app_id}/name`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "name": "Ashia—34"
    }
    ```

### 修改应用图标
- **Request**
  - **URL**: `/apps/{uuid:app_id}/icon`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "icon": "heavy_check_mark"
    }
    ```

### 应用网页访问控制
> 调整应用是否可网页访问。
- **Request**
  - **URL**: `/apps/{uuid:app_id}/site-enable`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "enable_site": true
    }
    ```

### 应用 API 访问控制
> 调整应用是否可 API 访问
- **Request**
  - **URL**: `/apps/{uuid:app_id}/api-enable`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "enable_api": true
    }
    ```

## 应用 Function API
### 文本生成
> 文本生成型应用的执行接口
- **Request**
  - **URL**: `/apps/{uuid:app_id}/completion-messages`
  - **Method**: `POST`
  - **Body Example**:
    :::details
    ```json
    {
      "inputs": {
        "query": "Hello～"
      },
      "model_config": {
          "pre_prompt": "{{query}}",
          "prompt_type": "simple",
          "chat_prompt_config": {},
          "completion_prompt_config": {},
          "user_input_form": [
              {
                  "paragraph": {
                      "label": "Query",
                      "variable": "query",
                      "required": true,
                      "default": ""
                  }
              }
          ],
          "dataset_query_variable": "",
          "opening_statement": null,
          "suggested_questions_after_answer": {
            "enabled": false
          },
          "speech_to_text": {
            "enabled": false
          },
          "retriever_resource": {
            "enabled": false
          },
          "sensitive_word_avoidance": {
              "enabled": false,
              "type": "",
              "configs": []
          },
          "more_like_this": {
            "enabled": false
          },
          "model": {
              "provider": "openai_api_compatible",
              "name": "nitro",
              "mode": "chat",
              "completion_params": {
                  "temperature": 0.7,
                  "top_p": 1,
                  "frequency_penalty": 0,
                  "presence_penalty": 0,
                  "max_tokens": 512
              }
          },
          "text_to_speech": {
              "enabled": false,
              "voice": "",
              "language": ""
          },
          "agent_mode": {
              "enabled": false,
              "tools": []
          },
          "dataset_configs": {
              "retrieval_model": "single",
              "datasets": {
                "datasets": []
              }
          },
          "file_upload": {
              "image": {
                  "enabled": false,
                  "number_limits": 3,
                  "detail": "high",
                  "transfer_methods": [
                      "remote_url",
                      "local_file"
                  ]
              }
          }
      },
      "response_mode": "streaming"
    }    
    ```
    :::

## 文本生成停止
> 文本生成型应用执行中断任务接口
- **Request**
  - **URL**: `/apps/{uuid:app_id}/completion-messages/{string:task_id}/stop`
  - **Method**: `POST`
  - **Body Example**: `null`
  :::tip
  该 API 中所需的 `task_id` 可以从[文本生成](#文本生成) API 的 Response（流式传输）中获取。
  :::

## 聊天
> 聊天型应用的执行接口
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-messages`
  - **Method**: `POST`
  - **Body Example**:
    :::details
    ```json
    {
      "response_mode": "streaming",
      "conversation_id": "",
      "query": "Hello～",
      "inputs": {},
      "model_config": {
          "pre_prompt": "",
          "prompt_type": "simple",
          "chat_prompt_config": {},
          "completion_prompt_config": {},
          "user_input_form": [],
          "dataset_query_variable": "",
          "opening_statement": "",
          "more_like_this": {
            "enabled": false
          },
          "suggested_questions": [],
          "suggested_questions_after_answer": {
            "enabled": false
          },
          "text_to_speech": {
              "enabled": false,
              "voice": "",
              "language": ""
          },
          "speech_to_text": {
            "enabled": false
          },
          "retriever_resource": {
            "enabled": false
          },
          "sensitive_word_avoidance": {
            "enabled": false
          },
          "agent_mode": {
              "max_iteration": 5,
              "enabled": true,
              "tools": [],
              "strategy": "react"
          },
          "dataset_configs": {
              "retrieval_model": "single",
              "datasets": {
                "datasets": []
              }
          },
          "file_upload": {
              "image": {
                  "enabled": false,
                  "number_limits": 2,
                  "detail": "low",
                  "transfer_methods": [
                    "local_file"
                  ]
              }
          },
          "annotation_reply": {
            "enabled": false
          },
          "supportAnnotation": true,
          "appId": "2c937aae-f4f2-4cf9-b6e2-f2f2756858c0",
          "supportCitationHitInfo": true,
          "model": {
              "provider": "openai_api_compatible",
              "name": "nitro",
              "mode": "chat",
              "completion_params": {
                  "temperature": 2,
                  "top_p": 1,
                  "frequency_penalty": 0,
                  "presence_penalty": 0,
                  "max_tokens": 512,
                  "stop": []
              }
          }
      }
    }
    ```
    :::


## 聊天停止
> 聊天型应用执行中断任务接口
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-messages/{string:task_id}/stop`
  - **Method**: `POST`
  - **Body Example**: `null`
  :::tip
  该 API 中所需的 `task_id` 可以从[聊天](#聊天) API 的 Response（流式传输）中获取。
  :::

### 获取会话列表（文本生成）
- **Request**
  - **URL**: `/apps/{uuid:app_id}/completion-conversations`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/completion-conversations?page=1&limit=30`
  :::tip
  下面列出的会话（文本生成）API 需要 `conversation_id`，可以从该 API 的 Response 中获取。
  :::

### 获取会话细节（文本生成）
- **Request**
  - **URL**: `/apps/{uuid:app_id}/completion-conversations/{uuid:conversation_id}`
  - **Method**: `GET`
  - **Body Example**: `null`
  :::tip
  下面列出的会话（文本生成）API 需要 `message_id`，可以从该 API 的 Response 中获取。
  :::

### 删除会话细节（文本生成）
- **Request**
  - **URL**: `/apps/{uuid:app_id}/completion-conversations/{uuid:conversation_id}`
  - **Method**: `DELETE`
  - **Body Example**: `null`

### 获取会话列表（聊天）
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-conversations`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/chat-conversations?page=1&limit=30`
  :::tip
  下面列出的会话（聊天）API 需要 `conversation_id`，可以从该 API 的 Response 中获取。
  :::

### 获取会话细节（聊天）
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-conversations/{uuid:conversation_id}`
  - **Method**: `GET`
  - **Body Example**: `null`
  :::tip
  下面列出的会话（对话）API 需要 `message_id`，可以从该 API 的 Response 中获取。
  :::

### 删除会话细节（聊天）
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-conversations/{uuid:conversation_id}`
  - **Method**: `DELETE`
  - **Body Example**: `null`

### 推荐问题（聊天）
> 在对话型应用中，获取 AI 给出回复后可以提出的建议问题
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-messages/{uuid:message_id}/suggested-questions`
  - **Method**: `GET`
  - **Body Example**: `null`

### 获取消息列表（聊天）
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-messages`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/chat-messages?conversation_id={conversation_id}`

### 消息反馈
> 对应用消息反馈喜欢或不喜欢
- **Request**
  - **URL**: `/apps/{uuid:app_id}/feedbacks`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "rating": "like"  // "like" | "dislike" | null
    }
    ```

### 消息标注
> 对来自应用的消息进行标注（文本生成）
- **Request**
  - **URL**: `/apps/{uuid:app_id}/annotations`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "message_id": "2b79fdad-e513-45ef-9532-8de5086cb81c",
      "question": "query:How are you?",
      "answer": "some answer messages"
    }
    ```

### 消息标注统计
> 获取应用当前消息的注释条数
- **Request**
  - **URL**: `/apps/{uuid:app_id}/annotations/count`
  - **Method**: `GET`
  - **Body Example**: `null`


### 获取消息细节（聊天）
- **Request**
  - **URL**: `/apps/{uuid:app_id}/messages/{uuid:message_id}`
  - **Method**: `GET`
  - **Body Example**: `null`

## 高级应用管理 API

### 模型配置
- **Request**
  - **URL**: `/apps/{uuid:app_id}/model-config`
  - **Method**: `POST`
  - **Body Example**:
    :::details
    ```json
    {
      "pre_prompt": "",
      "prompt_type": "simple",
      "chat_prompt_config": {},
      "completion_prompt_config": {},
      "user_input_form": [],
      "dataset_query_variable": "",
      "opening_statement": "",
      "suggested_questions": [],
      "more_like_this": {
        "enabled": false
      },
      "suggested_questions_after_answer": {
        "enabled": false
      },
      "speech_to_text": {
        "enabled": false
      },
      "text_to_speech": {
        "enabled": false,
        "language": "",
        "voice": ""
      },
      "retriever_resource": {
        "enabled": false
      },
      "sensitive_word_avoidance": {
        "enabled": false
      },
      "agent_mode": {
        "max_iteration": 5,
        "enabled": true,
        "strategy": "react",
        "tools": []
      },
      "model": {
        "provider": "openai_api_compatible",
        "name": "nitro",
        "mode": "chat",
        "completion_params": {
            "frequency_penalty": 0,
            "max_tokens": 512,
            "presence_penalty": 0,
            "stop": [],
            "temperature": 2,
            "top_p": 1
        }
      },
      "dataset_configs": {
        "retrieval_model": "single",
        "datasets": {
            "datasets": []
        }
      },
      "file_upload": {
        "image": {
            "enabled": false,
            "number_limits": 2,
            "detail": "low",
            "transfer_methods": [
                "local_file"
            ]
        }
      }
    }
    ```
    :::

### 修改应用基本信息
- **Request**
  - **URL**: `/apps/{uuid:app_id}/site`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "title": "Ashias-23",
      "icon": "grin",
      "icon_background": "#000000",
      "description": "How do you do~"
    }
    ```
### 重新生成公开访问的 URL
> 重新生成应用的公共访问 URL
- **Request**
  - **URL**: `/apps/{uuid:app_id}/site/access-token-reset`
  - **Method**: `POST`
  - **Body Example**: `null`

## 应用统计 API
### 全部消息数统计
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/daily-conversations`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/daily-conversations?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### 活跃用户统计
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/daily-end-users`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/daily-end-users?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### 费用消耗统计
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/token-costs`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/token-costs?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### 平均会话互动数统计
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/average-session-interactions`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/average-session-interactions?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### 用户满意度统计
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/user-satisfaction-rate`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/user-satisfaction-rate?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### 平均响应时间统计
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/average-response-time`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/average-response-time?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### Token 输出速度统计
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/tokens-per-second`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/tokens-per-second?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`
