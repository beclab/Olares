# `jwt` <!-- omit in toc -->

# Table of Contents

- [Table of Contents](#table-of-contents)
- [Usage](#usage)
  - [Signing](#signing)
  - [Verifying](#verifying)
- [Directory Structure](#directory-structure)
    - [Rationale](#rationale)


# Usage 

## Signing

```go
package main

import (
	"fmt"
	"github.com/beclab/Olares/cli/pkg/web5/didjwk"
    "github.com/beclab/Olares/cli/pkg/web5/jwt"
)

func main() {	
	did, err := didjwk.Create()
	if err != nil {
		panic(err)
	}

	claims := jwt.Claims{
		Issuer: did.URI,
		Misc:   map[string]interface{}{"c_nonce": "abcd123"},
	}

	jwt, err := jwt.Sign(claims, did)
	if err != nil {
		panic(err)
	}
}
```


## Verifying
```go
package main

import (
	"fmt"
	"github.com/beclab/Olares/cli/pkg/web5/dids"
    "github.com/beclab/Olares/cli/pkg/web5/jwt"
)

func main() {
    someJWT := "SOME_JWT"
	ok, err := jwt.Verify(signedJWT)
	if err != nil {
		panic(err)
	}

    if (!ok) {
        fmt.Printf("dookie JWT")
    }
}
```

specifying a specific category of key to use relative to the did provided can be done in the same way shown with `jws.Sign`

# Directory Structure

```
jwt
├── jwt.go
└── jwt_test.go
```

### Rationale 
same as `jws`. 