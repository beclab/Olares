# Verifer Service

Refer to [Presentation Exchange](https://identity.foundation/presentation-exchange/)
![verifier](/images/developer/contribute/verifier.png) for the process.

## Presentation Definition

1. The Verifier returns the file to the Holder.
2. The Holder submits a packaged file that meets the format requirements to the Verifier, after filling in the contents according to the specifications.

```json
{
  "name": "Facebook Basic Info Presentation Definition",
  "purpose": "Provide your facebook basic info",
  "inputDescriptors": [
    {
      "id": "name",
      "name": "Name",
      "purpose": "Provide vc name",
      "format": {
        "jwt_vc": {
          "alg": ["EdDSA"]
        }
      },
      "constraints": {
        "fields": [
          {
            "path": ["$.credentialSubject.name"]
          }
        ],
        "subject_is_issuer": "preferred"
      }
    },
    {
      "id": "title",
      "name": "Title",
      "purpose": "Provide vc title",
      "format": {
        "jwt_vc": {
          "alg": ["EdDSA"]
        }
      },
      "constraints": {
        "fields": [
          {
            "path": ["$.credentialSubject.title"]
          }
        ],
        "subject_is_issuer": "preferred"
      }
    },
    {
      "id": "description",
      "name": "description",
      "purpose": "Provide vc description",
      "format": {
        "jwt_vc": {
          "alg": ["EdDSA"]
        }
      },
      "constraints": {
        "fields": [
          {
            "path": ["$.credentialSubject.description"]
          }
        ],
        "subject_is_issuer": "preferred"
      }
    },

    {
      "id": "facebook_name",
      "name": "Provide your facebook name",
      "purpose": "Provide your facebook name",
      "format": {
        "jwt_vc": {
          "alg": ["EdDSA"]
        }
      },
      "constraints": {
        "fields": [
          {
            "path": ["$.credentialSubject.facebook_name"]
          }
        ],
        "subject_is_issuer": "preferred"
      }
    },
    {
      "id": "profile_image",
      "name": "Provide your facebook profile image",
      "purpose": "Provide your facebook profile image",
      "format": {
        "jwt_vc": {
          "alg": ["EdDSA"]
        }
      },
      "constraints": {
        "fields": [
          {
            "path": ["$.credentialSubject.profile_image"]
          }
        ],
        "subject_is_issuer": "preferred"
      }
    },
    {
      "id": "email",
      "name": "Provide your facebook email email info",
      "purpose": "Provide your facebook email info",
      "format": {
        "jwt_vc": {
          "alg": ["EdDSA"]
        }
      },
      "constraints": {
        "fields": [
          {
            "path": ["$.credentialSubject.email"]
          }
        ],
        "subject_is_issuer": "preferred"
      }
    },
    {
      "id": "facebook_id",
      "name": "Provide your facebook id",
      "purpose": "Provide your facebook id",
      "format": {
        "jwt_vc": {
          "alg": ["EdDSA"]
        }
      },
      "constraints": {
        "fields": [
          {
            "path": ["$.credentialSubject.facebook_id"]
          }
        ],
        "subject_is_issuer": "preferred"
      }
    },
    {
      "id": "picture_is_silhouette",
      "name": "Provide your facebook Picture is Silhouette",
      "purpose": "Provide your facebook picture_is_silhouette",
      "format": {
        "jwt_vc": {
          "alg": ["EdDSA"]
        }
      },
      "constraints": {
        "fields": [
          {
            "path": ["$.credentialSubject.picture_is_silhouette"]
          }
        ],
        "subject_is_issuer": "preferred"
      }
    }
  ],
  "author": ""
}
```
