# `jws` <!-- omit in toc -->


# Table of Contents <!-- omit in toc -->
- [Features](#features)
- [Usage](#usage)
  - [Signing:](#signing)
  - [Detached Content](#detached-content)
  - [Verifying](#verifying)
  - [Directory Structure](#directory-structure)
    - [Rationale](#rationale)


# Features
* Signing a JWS (JSON Web Signature) with a DID
* Verifying a JWS with a DID

# Usage

## Signing:

```go
package main

import (
    "fmt"
    "github.com/beclab/Olares/cli/pkg/web5/didjwk"
    "github.com/beclab/Olares/cli/pkg/web5/jws"
)

func main() {	
    did, err := didjwk.Create()
    if err != nil {
        fmt.Printf("failed to create did: %v", err)
        return
    }

    payload := map[string]interface{}{"hello": "world"}
    
    compactJWS, err := jws.Sign(payload, did)
    if err != nil {
        fmt.Printf("failed to sign: %v", err)
        return
    }

    fmt.Printf("compact JWS: %s", compactJWS)
}
```

## Detached Content

returning a JWS with detached content can be done like so:

```go
package main

import (
    "fmt"
    "github.com/beclab/Olares/cli/pkg/web5/didjwk"
    "github.com/beclab/Olares/cli/pkg/web5/jws"
)

func main() {	
    did, err := didjwk.Create()
    if err != nil {
        fmt.Printf("failed to create did: %v", err)
        return
    }

    payload := map[string]interface{}{"hello": "world"}
    
    compactJWS, err := jws.Sign(payload, did, Detached(true))
    if err != nil {
        fmt.Printf("failed to sign: %v", err)
        return
    }

    fmt.Printf("compact JWS: %s", compactJWS)
}
```

specifying a specific category of key associated with the provided did to sign with can be done like so:

```go
package main

import (
    "fmt"
    "github.com/beclab/Olares/cli/pkg/web5/didjwk"
    "github.com/beclab/Olares/cli/pkg/web5/jws"
)

func main() {	
    bearerDID, err := didjwk.Create()
    if err != nil {
        fmt.Printf("failed to create did: %v", err)
        return
    }

    payload := map[string]interface{}{"hello": "world"}
    
    compactJWS, err := jws.Sign(payload, did, Purpose("authentication"))
    if err != nil {
        fmt.Printf("failed to sign: %v", err)
    }

    fmt.Printf("compact JWS: %s", compactJWS)
}
```


## Verifying

```go
package main

import (
    "fmt"
    "github.com/beclab/Olares/cli/pkg/web5/didjwk"
    "github.com/beclab/Olares/cli/pkg/web5/jws"
)

func main() {	
    compactJWS := "SOME_JWS"
    ok, err := jws.Verify(compactJWS)
    if (err != nil) {
        fmt.Printf("failed to verify JWS: %v", err)
    }

    if (!ok) {
        fmt.Errorf("integrity check failed")
    }
}
```

> [!NOTE]
> an error is returned if something in the process of verification failed whereas `!ok` means the signature is actually shot


## Directory Structure

```
jws
├── jws.go
└── jws_test.go
```

### Rationale
bc i wanted `jws.Sign` and `jws.Verify` hipster vibes
