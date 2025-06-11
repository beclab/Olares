package vc_test

import (
	"encoding/json"
	"fmt"
	"time"

	"olares-cli/pkg/web5/dids/didkey"
	"olares-cli/pkg/web5/vc"
)

// Demonstrates how to create, sign, and verify a Verifiable Credential using the vc package.
func Example() {
	// create sample issuer and subject DIDs
	issuer, err := didkey.Create()
	if err != nil {
		panic(err)
	}

	subject, err := didkey.Create()
	if err != nil {
		panic(err)
	}

	// creation
	claims := vc.Claims{"id": subject.URI, "name": "Randy McRando"}
	cred := vc.Create(claims)

	// signing
	vcJWT, err := cred.Sign(issuer)
	if err != nil {
		panic(err)
	}

	// verification
	decoded, err := vc.Verify[vc.Claims](vcJWT)
	if err != nil {
		panic(err)
	}

	fmt.Println(decoded.VC.CredentialSubject["name"])
	// Output: Randy McRando
}

type KnownCustomerClaims struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c KnownCustomerClaims) GetID() string {
	return c.ID
}

func (c *KnownCustomerClaims) SetID(id string) {
	c.ID = id
}

// Demonstrates how to use a strongly typed credential subject
func Example_stronglyTyped() {
	issuer, err := didkey.Create()
	if err != nil {
		panic(err)
	}

	subject, err := didkey.Create()
	if err != nil {
		panic(err)
	}

	claims := KnownCustomerClaims{ID: subject.URI, Name: "Randy McRando"}
	cred := vc.Create(&claims)

	vcJWT, err := cred.Sign(issuer)
	if err != nil {
		panic(err)
	}

	decoded, err := vc.Verify[*KnownCustomerClaims](vcJWT)
	if err != nil {
		panic(err)
	}

	fmt.Println(decoded.VC.CredentialSubject.Name)
	// Output: Randy McRando
}

// Demonstrates how to use a mix of strongly typed and untyped credential subjects with the vc package.
func Example_mixed() {
	issuer, err := didkey.Create()
	if err != nil {
		panic(err)
	}

	subject, err := didkey.Create()
	if err != nil {
		panic(err)
	}

	claims := KnownCustomerClaims{ID: subject.URI, Name: "Randy McRando"}
	cred := vc.Create(&claims)

	vcJWT, err := cred.Sign(issuer)
	if err != nil {
		panic(err)
	}

	decoded, err := vc.Verify[vc.Claims](vcJWT)
	if err != nil {
		panic(err)
	}

	fmt.Println(decoded.VC.CredentialSubject["name"])
	// Output: Randy McRando
}

// Demonstrates how to create a Verifiable Credential
func ExampleCreate() {
	claims := vc.Claims{"name": "Randy McRando"}
	cred := vc.Create(claims)

	bytes, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}

// Demonstrates how to create a Verifiable Credential with options
func ExampleCreate_options() {
	claims := vc.Claims{"id": "1234"}
	issuanceDate := time.Now().UTC().Add(10 * time.Hour)
	expirationDate := issuanceDate.Add(30 * time.Hour)

	cred := vc.Create(
		claims,
		vc.ID("hehecustomid"),
		vc.Contexts("https://nocontextisbestcontext.gov"),
		vc.Types("StreetCredential"),
		vc.IssuanceDate(issuanceDate),
		vc.ExpirationDate(expirationDate),
	)

	bytes, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}
