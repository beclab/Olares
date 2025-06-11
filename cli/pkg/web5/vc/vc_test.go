package vc_test

import (
	"slices"
	"testing"
	"time"

	"olares-cli/pkg/web5/dids/didkey"
	"olares-cli/pkg/web5/vc"

	"github.com/alecthomas/assert/v2"
)

func TestCreate_Defaults(t *testing.T) {
	cred := vc.Create(vc.Claims{"id": "1234"})

	assert.Equal(t, 1, len(cred.Context))
	assert.Equal(t, vc.BaseContext, cred.Context[0])

	assert.Equal(t, 1, len(cred.Type))
	assert.Equal(t, vc.BaseType, cred.Type[0])

	assert.Contains(t, cred.ID, "urn:vc:uuid:")

	assert.NotZero(t, cred.IssuanceDate)

	_, err := time.Parse(time.RFC3339, cred.IssuanceDate)
	assert.NoError(t, err)

	assert.Equal(t, "1234", cred.CredentialSubject["id"])
}

func TestCreate_Options(t *testing.T) {
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
		vc.Schemas("https://example.org/examples/degree.json"),
		vc.Evidences(vc.Evidence{
			ID:   "evidenceID",
			Type: "Insufficient",
			AdditionalFields: map[string]interface{}{
				"kind":   "circumstantial",
				"checks": []string{"motive", "cell_tower_logs"},
			},
		}),
	)

	assert.Equal(t, 2, len(cred.Context))
	assert.True(t, slices.Contains(cred.Context, "https://nocontextisbestcontext.gov"))
	assert.True(t, slices.Contains(cred.Context, vc.BaseContext))

	assert.Equal(t, 2, len(cred.Type))
	assert.True(t, slices.Contains(cred.Type, "StreetCredential"))
	assert.True(t, slices.Contains(cred.Type, vc.BaseType))

	assert.Equal(t, "hehecustomid", cred.ID)

	assert.NotZero(t, cred.ExpirationDate)

	assert.Equal(t, 1, len(cred.CredentialSchema))
	assert.Equal(t, "https://example.org/examples/degree.json", cred.CredentialSchema[0].ID)
	assert.Equal(t, "JsonSchema", cred.CredentialSchema[0].Type)

	assert.Equal(t, 1, len(cred.Evidence))
	assert.Equal(t, "evidenceID", cred.Evidence[0].ID)
	assert.Equal(t, "Insufficient", cred.Evidence[0].Type)
	assert.Equal(t, "circumstantial", cred.Evidence[0].AdditionalFields["kind"])
	assert.Equal(t, []string{"motive", "cell_tower_logs"}, cred.Evidence[0].AdditionalFields["checks"].([]string)) // nolint:forcetypeassert
}

func TestSign(t *testing.T) {
	issuer, err := didkey.Create()
	assert.NoError(t, err)

	subject, err := didkey.Create()
	assert.NoError(t, err)

	claims := vc.Claims{"id": subject.URI, "name": "Randy McRando"}
	cred := vc.Create(claims)

	vcJWT, err := cred.Sign(issuer)
	assert.NoError(t, err)
	assert.NotZero(t, vcJWT)

	// TODO: make test more reliable by not depending on another function in this package (Moe - 2024-02-25)
	decoded, err := vc.Verify[vc.Claims](vcJWT)

	assert.NoError(t, err)
	assert.NotZero(t, decoded)
}
