package didcore_test

import (
	"testing"

	"olares-cli/pkg/web5/dids/didcore"

	"github.com/alecthomas/assert/v2"
)

func TestAddVerificationMethod(t *testing.T) {
	doc := didcore.Document{
		Context: []string{"https://www.w3.org/ns/did/v1"},
		ID:      "did:example:123456789abcdefghi",
	}

	vm := didcore.VerificationMethod{
		ID:         "did:example:123456789abcdefghi#keys-1",
		Type:       "Ed25519VerificationKey2018",
		Controller: "did:example:123456789abcdefghi",
	}

	doc.AddVerificationMethod(vm, didcore.Purposes("authentication"))

	assert.Equal(t, 1, len(doc.VerificationMethod))
	assert.Equal(t, 1, len(doc.Authentication))
	assert.Equal(t, vm.ID, doc.Authentication[0])
}

func TestWoo(t *testing.T) {
	doc := didcore.Document{
		ID: "did:example:123456789abcdefghi",
	}

	doc.AddVerificationMethod(didcore.VerificationMethod{
		ID:         "did:example:123456789abcdefghi#keys-1",
		Type:       "Ed25519VerificationKey2018",
		Controller: "did:example:123456789abcdefghi",
	}, didcore.Purposes("authentication"))

	vm, err := doc.SelectVerificationMethod(didcore.Purpose("authentication"))
	assert.NoError(t, err)
	assert.Equal(t, "did:example:123456789abcdefghi#keys-1", vm.ID)
}
