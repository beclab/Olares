---
outline: [2, 3]
---

# AI

## API Prefix

`agent.{username}.olares.com/api/controllers/console/api`

## Basic Application Management API
### Get App List
- **Request**
  - **URL**: `/apps`
  - **Method**: `GET`
  - **URL Parameters**: `/apps?page=1&limit=30&name=Ashia`
  :::tip
  Most of the APIs listed in this document require the `app_id`, which can be obtained from the response of this API.
  :::

### Create App
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

### Get App Details
- **Request**
  - **URL**: `/apps/{uuid:app_id}`
  - **Method**: `GET`
  - **Body Example**: `null`

### Delete App
- **Request**
  - **URL**: `/apps/{uuid:app_id}`
  - **Method**: `DELETE`
  - **Body Example**: `null`

### Copy App
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

### Rename App
- **Request**
  - **URL**: `/apps/{uuid:app_id}/name`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "name": "Ashia—34"
    }
    ```

### Change App Icon
- **Request**
  - **URL**: `/apps/{uuid:app_id}/icon`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "icon": "heavy_check_mark"
    }
    ```

### App Web Access Control
> Whether the app can be accessed from the site.
- **Request**
  - **URL**: `/apps/{uuid:app_id}/site-enable`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "enable_site": true
    }
    ```

### App API Access Control
> Whether the app can be accessed via API.
- **Request**
  - **URL**: `/apps/{uuid:app_id}/api-enable`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "enable_api": true
    }
    ```

## Application Function API
### Text Generation
> Execution interface for text generation APP
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

### Stop Text Generation
> Interruption interface for text generation APP
- **Request**
  - **URL**: `/apps/{uuid:app_id}/completion-messages/{string:task_id}/stop`
  - **Method**: `POST`
  - **Body Example**: `null`
  :::tip
  The `task_id` required in this API can be obtained from the response (streaming) of the [Text Generation](#text-generation) API.
  :::

### Chat
> Execution interface for chat APP
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


### Stop Chat
> Interruption interface for chat APP
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-messages/{string:task_id}/stop`
  - **Method**: `POST`
  - **Body Example**: `null`
  :::tip
  The `task_id` required in this API can be obtained from the response (streaming) of the [Chat](#chat) API.
  :::

### Get Conversations List (Text Generation) 
- **Request**
  - **URL**: `/apps/{uuid:app_id}/completion-conversations`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/completion-conversations?page=1&limit=30`
  :::tip
  the Conversations (Text Generation) APIs listed below require the `conversation_id`, which can be obtained from the response of this API.
  :::

### Get Conversations Details (Text Generation) 
- **Request**
  - **URL**: `/apps/{uuid:app_id}/completion-conversations/{uuid:conversation_id}`
  - **Method**: `GET`
  - **Body Example**: `null`
  :::tip
  The Conversations (Text Generation) APIs listed below require the `message_id`, which can be obtained from the response of this API.
  :::

### Delete Conversations Details (Text Generation)
- **Request**
  - **URL**: `/apps/{uuid:app_id}/completion-conversations/{uuid:conversation_id}`
  - **Method**: `DELETE`
  - **Body Example**: `null`

### Get Conversations List (Chat)
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-conversations`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/chat-conversations?page=1&limit=30`
  :::tip
  the Conversations (Chat) APIs listed below require the `conversation_id`, which can be obtained from the response of this API.
  :::

### Get Conversations Details (Chat)
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-conversations/{uuid:conversation_id}`
  - **Method**: `GET`
  - **Body Example**: `null`
  :::tip
  The Conversations (Chat) APIs listed below require the `message_id`, which can be obtained from the response of this API.
  :::

### Delete Conversations Details (Chat)
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-conversations/{uuid:conversation_id}`
  - **Method**: `DELETE`
  - **Body Example**: `null`

### Get Suggested Questions (Chat)
> In a chat APP, get the suggested questions that can be asked after the AI gives a response
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-messages/{uuid:message_id}/suggested-questions`
  - **Method**: `GET`
  - **Body Example**: `null`

### Get Message List (Chat)
- **Request**
  - **URL**: `/apps/{uuid:app_id}/chat-messages`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/chat-messages?conversation_id={conversation_id}`

### Message Feedback
> Give like or dislike feedback to the message from the APP
- **Request**
  - **URL**: `/apps/{uuid:app_id}/feedbacks`
  - **Method**: `POST`
  - **Body Example**:
    ```json
    {
      "rating": "like"  // "like" | "dislike" | null
    }
    ```

### Message Annotation
> Give annotation to the message from the APP (Text Generation)
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

### Count Annotation
> Get the current number of annotations of the APP's message
- **Request**
  - **URL**: `/apps/{uuid:app_id}/annotations/count`
  - **Method**: `GET`
  - **Body Example**: `null`


### Get Message Details (Chat)
- **Request**
  - **URL**: `/apps/{uuid:app_id}/messages/{uuid:message_id}`
  - **Method**: `GET`
  - **Body Example**: `null`

## Advanced Application Management API

### Model Config
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

### Change APP Basic Info
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
### Access Token Reset
> Regenerate the public access URL for the APP
- **Request**
  - **URL**: `/apps/{uuid:app_id}/site/access-token-reset`
  - **Method**: `POST`
  - **Body Example**: `null`

## Application Statistics API    
### All Conversations
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/daily-conversations`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/daily-conversations?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### Active Users
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/daily-end-users`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/daily-end-users?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### Token Costs
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/token-costs`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/token-costs?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### Average Session Interactions
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/average-session-interactions`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/average-session-interactions?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### User Satisfaction
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/user-satisfaction-rate`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/user-satisfaction-rate?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### Average Response Time
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/average-response-time`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/average-response-time?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`

### Token Output Speed
- **Request**
  - **URL**: `/apps/{uuid:app_id}/statistics/tokens-per-second`
  - **Method**: `GET`
  - **URL Parameters**: `/apps/{uuid:app_id}/statistics/tokens-per-second?start=2024-04-19%2016%3A28&end=2024-04-26%2016%3A28`
