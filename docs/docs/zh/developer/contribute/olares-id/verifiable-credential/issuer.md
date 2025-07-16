# 发行方 (Issuer)

![alt text](/images/developer/contribute/issuer.png)

以下是发行方的流程：

1.  持有方 (Holder) 从发行方 (Issuer) 获取清单 (Manifest)。
2.  持有方签名并提交申请 (Application)。
3.  发行方自动或手动审查该申请。
4.  持有方从发行方接收审查结果，如果获批则得到可验证凭证 (VC)，如果被拒则得到原因。

## 清单 (Manifest)

此文件将返回给持有方。
`outputDescriptors` 用于钱包展示可验证凭证 (VCs)。
`presentationDefinition` 实际上是后续的 `manifest_presentation` 文件，用于确认用户提交的申请 (Application) 的格式。

```json
{
  "name": "Facebook Verifiable Credential Manifest",
  "description": "Facebook Verifiable Credential Manifest",
  "issuerDid": "",
  "issuerName": "",
  "outputDescriptors": [
    {
      "id": "",
      "schema": "",
      "name": "Facebook Verifiable Credential Manifest",
      "description": "Facebook Verifiable Credential Manifest",
      "display": {
        "title": {
          "path": ["$.credentialSubject.name", "$.vc.credentialSubject.name"],
          "schema": { "type": "string" }
        },
        "subtitle": {
          "path": ["$.credentialSubject.title", "$.vc.credentialSubject.title"],
          "schema": { "type": "string" }
        },
        "description": {
          "path": [
            "$.credentialSubject.description",
            "$.vc.credentialSubject.description"
          ],
          "schema": { "type": "string" }
        },
        "properties": [
          {
            "path": ["$.credentialSubject.id", "$.vc.credentialSubject.id"],
            "schema": { "type": "string" },
            "label": "ID"
          },
          {
            "path": [
              "$.credentialSubject.email",
              "$.vc.credentialSubject.email"
            ],
            "schema": { "type": "string" },
            "label": "Email"
          }
        ]
      },
      "styles": {
        "background": {
          "color": "#FFFFFF"
        },
        "text": {
          "color": "#000000"
        }
      }
    }
  ],
  "format": {
    "jwt_vc": {
      "alg": ["EdDSA"]
    }
  },
  "presentationDefinition": {}
}
````

## 清单呈现 (Manifest Presentation)

发行方使用此文件来验证持有方提交的申请 (Application) 的格式。

```json
{
  "name": "Facebook Manifest Presentation Definition",
  "purpose": "Provide your token required to Facebook",
  "inputDescriptors": [
    {
      "id": "token",
      "name": "Access Token",
      "purpose": "Provide your token required to Facebook",
      "format": {
        "jwt_vc": {
          "alg": ["EdDSA"]
        }
      },
      "constraints": {
        "fields": [
          {
            "path": ["$.credentialSubject.token"]
          }
        ],
        "subject_is_issuer": "preferred"
      }
    }
  ],
  "author": ""
}
```

## 申请可验证凭证 (Application Verifiable Credential)

虽然用户在客户端仅需 `manifest_presentation` 来构建 VC 数据，但服务器需要一个 schema 来验证 `manifest_presentation` 是否符合格式要求。
此文件将返回给持有方，以便持有方了解需要提交以供验证的数据格式。
在 Facebook、Twitter 和 Gmail 的场景中，用户在客户端登录并获取访问令牌 (access token)，然后将其提交给发行方服务器，发行方服务器可以凭此访问令牌从 Facebook 等服务器获取用户的基本信息。因此用户只需提交简单的 token 即可。
但在 KYC（了解你的客户）场景中，用户需要提交全名、身份证照片甚至验证视频等，此文件中的字段会变得复杂得多。

```json
{
  "author": "",
  "name": "Facebook Verifiable Credential Request Schema",
  "schema": {
    "$id": "facebook-schema-1.0",
    "$schema": "[https://json-schema.org/draft/2020-12/schema](https://json-schema.org/draft/2020-12/schema)",
    "description": "Facebook Verifiable Credential Schema",
    "type": "object",
    "properties": {
      "token": {
        "type": "string"
      }
    },
    "required": ["token"],
    "additionalProperties": true
  },
  "sign": false
}
```

## 可验证凭证 (Verifiable Credential)

发行方返回给持有方的 VC 格式。

```json
{
  "author": "",
  "name": "Facebook Verifiable Credential Schema",
  "schema": {
    "$id": "facebook-schema-1.0",
    "$schema": "[https://json-schema.org/draft/2020-12/schema](https://json-schema.org/draft/2020-12/schema)",
    "description": "Facebook Verifiable Credential Schema",
    "type": "object",
    "properties": {
      "name": {
        "type": "string"
      },
      "title": {
        "type": "string"
      },
      "description": {
        "type": "string"
      },

      "facebook_name": {
        "type": "string"
      },
      "profile_image": {
        "type": "string"
      },
      "email": {
        "type": "string"
      },
      "facebook_id": {
        "type": "string"
      },
      "picture_is_silhouette": {
        "type": "boolean"
      }
    },
    "required": ["name", "title", "description", "facebook_name"],
    "additionalProperties": true
  },
  "sign": false
}
```

