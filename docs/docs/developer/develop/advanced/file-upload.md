---
outline: [2, 3]
---

# File Upload

There are many situations where you might need to upload local files to your edge when using **Olares**. The `Olares Application Runtime` provides a common file-upload component to simplify this process in app development. Moreover, this file-upload component features **resumable upload**.

## How to install

To use this feature, simply add the following configuration to the [OlaresManifest.yaml](../package/manifest.md#upload) file in the application chart.
```yaml
upload:
  fileType:
    - pdf
  dest: /appdata
  limitedSize: 3729747942
```

## Frontend API

:::info NOTE
The limit for a single upload is 10MB. If the file is larger than 10MB, you must upload using the resumable upload API.
:::


### Upload

This interface is used to upload files to the server and get the file id and status.
:::details Example
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
    "mode": "formdata",  //The request body type must set to multipart/form-data
    "body parameters": {
      "storage_path": string,   //required, the storage path for the file on the server. Ensure that this folder exists.
      "file_relative_path": string, //required, the location of file relative to the storage_path, it must include the filename. If it is a floder, end with '/'
      "file_type": string,  //required, File type
      "file_size": integer, //required, File size
    }
  ```
- **Success Response**
  
  The response body contains the following content, formatted in JSON.
  - **Status Code** : `200 OK`
    ```json
      "code": integer, // Response code, 0 means success, non-zero means failure.
      "message": string, // Response message, return "success" upon success, and the corresponding error message upon failure.
      "data":{  //Response data. Upon success, it includes the following fields (these contents is absent when uploading a folder):
        "id": string,   // Unique identifier of the file
        "offset": integer,  //The offset of the file uploaded
        "file_relative_path": string, 
        "file_type": string,
        "file_size": integer,
        "storage_path": string
      }
    ```

- **Error Response**
  - **Status Code** : `400 Bad Request`
    > The request is invalid due to illegal or missing parameters.
  - **Status Code** : `500 Internal Server Error`
    > An internal server error has occurred, which prevented the request from being fulfilled. This could be due to reasons such as failure to create a folder or save file information.

### Resumable Upload

This interface is used to continue uploading the remaining part of the file.
:::details Example
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
    "mode": "formdata",  //The request body type must set to multipart/form-data
    "body parameters": {
      "file": string,   //required. The file to be uploaded. Please upload the file in multipart/form-data format.
      "upload_offset": integer,   //required. The offset refers to the size of the file that has already been uploaded.
    }
    "url parameters": {
      "uid": string,   //required. This is the unique identifier of the file. You can obtain it from the response data of the Upload API.
    }    
  ```
- **Success Response**
  
  The response body contains the following content, formatted in JSON.
  - **Status Code** : `200 OK`
    ```json
      "code": integer, // Response code, 0 means success, non-zero means failure.
      "message": string, // Response message, return "File uploaded successfully" upon success, and the corresponding error message upon failure.
      "data":{  //Response data. Upon success, it includes the following fields
        "id": string,   // Unique identifier of the file
        "offset": integer,  //The offset of the file uploaded
        "file_relative_path": string, 
        "file_type": string,
        "file_size": integer,
        "storage_path": string
      }
    ```

- **Error Response**
  - **Status Code** : `400 Bad Request`
    > The request is invalid due to illegal or missing parameters.
    ```json
    { "code": 1, "message": "Invalid upload ID" }
    ```
  - **Status Code** : `500 Internal Server Error`
    > An internal server error has occurred, which prevented the request from being fulfilled. This could be due to reasons such as failure to create a folder, save file information, or move file.