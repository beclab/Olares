---
outline: [2, 3]
---

# 文件上传

Olares 作为一个云端系统，存在很多将本地文件上传到云端的场景。Olares 应用运行时提供了一个通用的 file-upload 组件。简化应用对文件上传需求的开发。同时，file-upload 组件还提供了断点续传功能。

## 如何安装

只要在应用 chart 的 [OlaresManifest.yaml](../package/manifest.md#file-upload) 中申明
```yaml
upload:
  fileType:
    - pdf
  dest: /appdata
  limitedSize: 3729747942
```

## 前端接口对接

:::info 注意
单次上传大小限制为 10M，大于 10M 需要使用分片断点续传功能。
:::


## 上传

该接口用于上传文件到服务器并获取文件id和状态。
:::details 示例
**Request**
```sh
curl --location 'http://host:40030/upload/' \
--form 'storage_path="./testupload/"' \
--form 'file_relative_path="1.csv"' \
--form 'file_type="csv"' \
--form 'file_size="1937"'
```
**Response**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "b0b76f02bdb8ee3269602c983c4a2aeb",
    "offset": 0,
    "file_relative_path": "1.csv",
    "file_type": "csv",
    "file_size": 1937,
    "storage_path": "./testupload/"
  }
}
```
:::
- **Request**

  - **URL**: `/upload/`

  - **Method**: `POST`

  - **Body**:
  ``` json
    "mode": "formdata",  //请求体为 multipart/form-data 格式
    "body parameters": {
      "storage_path": string,   //必填，文件在服务器上的存储文件夹，该文件夹必须存在，
      "file_relative_path": string, //必填，文件相对于 storage_path 的路径，必须包含文件名。如果为文件夹，以“/”结尾。
      "file_type": string,  //必填，文件类型
      "file_size": integer, //必填，文件大小
    }
  ```
- **Success Response**

  响应体为 JSON 格式，包含以下字段：
  - **状态码** : `200 OK`
    ```json
      "code": integer, // 响应码，0 表示成功，非零表示失败。
      "message": string, // 响应消息，成功时为 "success"，失败时为相应的错误消息。
      "data":{  //响应数据，成功时包含以下字段（上传文件夹时无该字段）：
        "id": string,   // 文件唯一标识符
        "offset": integer,  // 文件上传的偏移量
        "file_relative_path": string, 
        "file_type": string,
        "file_size": integer,
        "storage_path": string
      }
    ```

- **错误情况**
  - **状态码** : `400 Bad Request`
    > 请求参数不合法或缺失。
  - **状态码** : `500 Internal Server Error`
    > 服务器内部错误，例如创建文件夹失败或保存文件信息失败。

### 断点续传

该接口用于继续上传文件的剩余部分。
:::details 示例
**Request**
```sh
curl --location --request PATCH 'http://host:40030/upload/b0b76f02bdb8ee3269602c983c4a2aeb' \
--form 'file=@"/Users/yangtao/Downloads/1.csv"' \
--form 'upload_offset="0"'
```
**Response**
```json
{
  "code": 0,
  "message": "File uploaded successfully",
  "data": {
    "id": "b0b76f02bdb8ee3269602c983c4a2aeb",
    "offset": 1937,
    "file_name": "1.csv",
    "file_type": "csv",
    "file_size": 1937,
    "storage_path": "./testupload"
  }
}
{
  "code": 0,
  "message": "Continue uploading",
  "data": {
    "id": "e3133b0f838124ff3ebcc9cb14774f26",
    "offset": 1048576,
    "file_name": "1.pdf",
    "file_type": ".pdf",
    "file_size": 10296258,
    "storage_path": "./testupload"
  }
}
```
:::

- **Request**

  - **URL**: `http://host:40030/upload/{uid}`

  - **Method**: `PATCH`

  - **Body**:
  ``` json
    "mode": "formdata",  //请求体为 multipart/form-data 格式
    "body parameters": {
      "file": string,   //必填，要上传的文件。请使用 multipart/form-data 格式进行文件上传。
      "upload_offset": integer,   //必填，文件上传的偏移量，之前已上传的文件大小。
    }
    "url parameters": {
      "uid": string,   //必填，文件的唯一标识符。可以从上传 API 的 Response 数据中获取。
    }    
  ```
- **Success Response**

响应体为 JSON 格式，包含以下字段：
  - **状态码** : `200 OK`
    ```json
      "code": integer, // 响应码，0 表示成功，非零表示失败。
      "message": string, // 响应消息，成功时为 "File uploaded successfully"，失败时为相应的错误消息。
      "data":{  // 响应数据，成功时包含以下字段：
        "id": string,   // 文件唯一标识符
        "offset": integer,  // 文件上传的偏移量
        "file_relative_path": string, 
        "file_type": string,
        "file_size": integer,
        "storage_path": string
      }
    ```

- **错误情况**
  - **状态码** : `400 Bad Request`
    > 请求参数不合法或缺失。
    ```json
    { "code": 1, "message": "Invalid upload ID" }
    ```
  - **状态码** : `500 Internal Server Error`
    > 服务器内部错误，例如创建文件夹失败、保存文件信息失败或移动文件失败。