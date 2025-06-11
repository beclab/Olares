package did_test

import (
	"testing"

	"olares-cli/pkg/web5/crypto/dsa"
	"olares-cli/pkg/web5/dids/did"
	"olares-cli/pkg/web5/dids/didcore"
	"olares-cli/pkg/web5/dids/didkey"
	"olares-cli/pkg/web5/jwk"
	"olares-cli/pkg/web5/jws"

	"github.com/alecthomas/assert/v2"
)

func TestToPortableDID(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	portableDID, err := did.ToPortableDID()
	assert.NoError(t, err)

	assert.Equal[string](t, did.URI, portableDID.URI)
	assert.True(t, len(portableDID.PrivateKeys) == 1, "expected 1 key")

	key := portableDID.PrivateKeys[0]

	assert.NotEqual(t, jwk.JWK{}, key, "expected key to not be empty")
}

func TestFromPortableDID(t *testing.T) {
	bearerDID, err := didkey.Create()
	assert.NoError(t, err)

	portableDID, err := bearerDID.ToPortableDID()
	assert.NoError(t, err)

	importedDID, err := did.FromPortableDID(portableDID)
	assert.NoError(t, err)

	payload := []byte("hi")

	compactJWS, err := jws.Sign(payload, bearerDID)
	assert.NoError(t, err)

	compactJWSAgane, err := jws.Sign(payload, importedDID)
	assert.NoError(t, err)

	assert.Equal[string](t, compactJWS, compactJWSAgane, "failed to produce same signature with imported did")
}

func TestGetSigner(t *testing.T) {
	bearerDID, err := didkey.Create()
	assert.NoError(t, err)

	sign, vm, err := bearerDID.GetSigner(nil)
	assert.NoError(t, err)

	assert.NotEqual(t, vm, didcore.VerificationMethod{}, "expected verification method to not be empty")

	payload := []byte("hi")
	signature, err := sign(payload)
	assert.NoError(t, err)

	legit, err := dsa.Verify(payload, signature, *vm.PublicKeyJwk)
	assert.NoError(t, err)

	assert.True(t, legit, "expected signature to be valid")
}
