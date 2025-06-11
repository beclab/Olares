package jwt_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"olares-cli/pkg/web5/dids/didkey"
	"olares-cli/pkg/web5/jws"
	"olares-cli/pkg/web5/jwt"

	"github.com/alecthomas/assert/v2"
)

func TestClaims_MarshalJSON(t *testing.T) {
	claims := jwt.Claims{
		Issuer: "issuer",
		Misc:   map[string]interface{}{"foo": "bar"},
	}

	b, err := json.Marshal(&claims)
	assert.NoError(t, err)

	obj := make(map[string]interface{})
	err = json.Unmarshal(b, &obj)
	assert.NoError(t, err)

	assert.Equal(t, "issuer", obj["iss"])
	assert.False(t, obj["foo"] == nil)
}

func TestClaims_UnmarshalJSON(t *testing.T) {
	claims := jwt.Claims{
		Issuer: "issuer",
		Misc:   map[string]interface{}{"foo": "bar"},
	}

	b, err := json.Marshal(&claims)
	assert.NoError(t, err)

	claimsAgane := jwt.Claims{}
	err = json.Unmarshal(b, &claimsAgane)
	assert.NoError(t, err)

	assert.Equal(t, claims.Issuer, claimsAgane.Issuer)
	assert.False(t, claimsAgane.Misc["foo"] == nil)
	assert.Equal(t, claimsAgane.Misc["foo"], claims.Misc["foo"])
}

func TestSign(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	claims := jwt.Claims{
		Issuer: did.ID,
		Misc:   map[string]interface{}{"c_nonce": "abcd123"},
	}

	jwt, err := jwt.Sign(claims, did)
	assert.NoError(t, err)

	assert.False(t, jwt == "", "expected jwt to not be empty")
}

func TestSign_IssuerOverridden(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	claims := jwt.Claims{
		Issuer: "something-not-equal-to-did.URI", // this will be overridden by the call to jwt.Sign()
		Misc:   map[string]interface{}{"c_nonce": "abcd123"},
	}

	signed, err := jwt.Sign(claims, did)
	assert.NoError(t, err)

	decoded, err := jwt.Decode(signed)
	assert.NoError(t, err)

	assert.Equal(t, did.URI, decoded.Claims.Issuer)
}

func TestVerify(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	claims := jwt.Claims{
		Issuer: did.URI,
		Misc:   map[string]interface{}{"c_nonce": "abcd123"},
	}

	signedJWT, err := jwt.Sign(claims, did)
	assert.NoError(t, err)

	assert.False(t, signedJWT == "", "expected jwt to not be empty")

	decoded, err := jwt.Verify(signedJWT)
	assert.NoError(t, err)
	assert.NotEqual(t, decoded, jwt.Decoded{}, "expected decoded to not be empty")
}

func TestVerify_BadClaims(t *testing.T) {
	okHeader, err := jws.Header{ALG: "ES256K", KID: "did:web:abc#key-1"}.Encode()
	assert.NoError(t, err)

	input := fmt.Sprintf("%s.%s.%s", okHeader, "hehe", "hehe")

	decoded, err := jwt.Verify(input)
	assert.Error(t, err)
	assert.Equal(t, jwt.Decoded{}, decoded)
}

func Test_Decode_Empty(t *testing.T) {
	decoded, err := jwt.Decode("")
	assert.Error(t, err)
	assert.Equal(t, jwt.Decoded{}, decoded)
}

func Test_Decode_Works(t *testing.T) {
	vcJwt := `eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbkY0VjFGS2F6RTJSbWhCZWtOQlRsRktaR1F5UTFkRldrcE9lbXBSYjNGSmRYWk5SbUpVWjFKTVNFRWlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjE3MjQ1MzQwNTAsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbkY0VjFGS2F6RTJSbWhCZWtOQlRsRktaR1F5UTFkRldrcE9lbXBSYjNGSmRYWk5SbUpVWjFKTVNFRWlmUSIsImp0aSI6InVybjp2Yzp1dWlkOjlkMzdmMzY3LWE4ZDctNDY4Zi05NGYwLTk1NzAxNzBkNzZhNCIsIm5iZiI6MTcyMTk0MjA1MCwidmMiOnsiQGNvbnRleHQiOlsiaHR0cHM6Ly93d3cudzMub3JnLzIwMTgvY3JlZGVudGlhbHMvdjEiXSwidHlwZSI6WyJWZXJpZmlhYmxlQ3JlZGVudGlhbCJdLCJpc3N1ZXIiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SW5GNFYxRkthekUyUm1oQmVrTkJUbEZLWkdReVExZEZXa3BPZW1wUmIzRkpkWFpOUm1KVVoxSk1TRUVpZlEiLCJjcmVkZW50aWFsU3ViamVjdCI6eyJpc3N1ZXIiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SW5GNFYxRkthekUyUm1oQmVrTkJUbEZLWkdReVExZEZXa3BPZW1wUmIzRkpkWFpOUm1KVVoxSk1TRUVpZlEifSwiaWQiOiJ1cm46dmM6dXVpZDo5ZDM3ZjM2Ny1hOGQ3LTQ2OGYtOTRmMC05NTcwMTcwZDc2YTQiLCJpc3N1YW5jZURhdGUiOiIyMDI0LTA3LTI1VDIxOjE0OjEwWiIsImV4cGlyYXRpb25EYXRlIjoiMjAyNC0wOC0yNFQyMToxNDoxMFoiLCJjcmVkZW50aWFsU2NoZW1hIjpbeyJ0eXBlIjoiSnNvblNjaGVtYSIsImlkIjoiaHR0cHM6Ly92Yy5zY2hlbWFzLmhvc3Qva2JjLnNjaGVtYS5qc29uIn1dfX0.VwvrU5Lmv3rn9rzXB0OCxe-MtE5R0876pXsXNLRuQjoqSNB5tBv_12NqrobwA-LkMzFwzdQ5-LWJni6grGdXCQ`
	decoded, err := jwt.Decode(vcJwt)
	assert.NoError(t, err)
	assert.Equal(t, decoded.Header.ALG, "EdDSA")
	assert.Equal(t, decoded.Header.KID, "did:jwk:eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6InF4V1FKazE2RmhBekNBTlFKZGQyQ1dFWkpOempRb3FJdXZNRmJUZ1JMSEEifQ#0")
	assert.NotZero(t, decoded.SignerDID)
}

func Test_Decode_Bad_Header(t *testing.T) {
	vcJwt := `kakaHeader.eyJleHAiOjE3MjQ1MzQwNTAsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbkY0VjFGS2F6RTJSbWhCZWtOQlRsRktaR1F5UTFkRldrcE9lbXBSYjNGSmRYWk5SbUpVWjFKTVNFRWlmUSIsImp0aSI6InVybjp2Yzp1dWlkOjlkMzdmMzY3LWE4ZDctNDY4Zi05NGYwLTk1NzAxNzBkNzZhNCIsIm5iZiI6MTcyMTk0MjA1MCwidmMiOnsiQGNvbnRleHQiOlsiaHR0cHM6Ly93d3cudzMub3JnLzIwMTgvY3JlZGVudGlhbHMvdjEiXSwidHlwZSI6WyJWZXJpZmlhYmxlQ3JlZGVudGlhbCJdLCJpc3N1ZXIiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SW5GNFYxRkthekUyUm1oQmVrTkJUbEZLWkdReVExZEZXa3BPZW1wUmIzRkpkWFpOUm1KVVoxSk1TRUVpZlEiLCJjcmVkZW50aWFsU3ViamVjdCI6eyJpc3N1ZXIiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SW5GNFYxRkthekUyUm1oQmVrTkJUbEZLWkdReVExZEZXa3BPZW1wUmIzRkpkWFpOUm1KVVoxSk1TRUVpZlEifSwiaWQiOiJ1cm46dmM6dXVpZDo5ZDM3ZjM2Ny1hOGQ3LTQ2OGYtOTRmMC05NTcwMTcwZDc2YTQiLCJpc3N1YW5jZURhdGUiOiIyMDI0LTA3LTI1VDIxOjE0OjEwWiIsImV4cGlyYXRpb25EYXRlIjoiMjAyNC0wOC0yNFQyMToxNDoxMFoiLCJjcmVkZW50aWFsU2NoZW1hIjpbeyJ0eXBlIjoiSnNvblNjaGVtYSIsImlkIjoiaHR0cHM6Ly92Yy5zY2hlbWFzLmhvc3Qva2JjLnNjaGVtYS5qc29uIn1dfX0.VwvrU5Lmv3rn9rzXB0OCxe-MtE5R0876pXsXNLRuQjoqSNB5tBv_12NqrobwA-LkMzFwzdQ5-LWJni6grGdXCQ`
	_, err := jwt.Decode(vcJwt)
	assert.Error(t, err)
}

func Test_Decode_Bad_Signature(t *testing.T) {
	vcJwt := `eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbkY0VjFGS2F6RTJSbWhCZWtOQlRsRktaR1F5UTFkRldrcE9lbXBSYjNGSmRYWk5SbUpVWjFKTVNFRWlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjE3MjQ1MzQwNTAsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbkY0VjFGS2F6RTJSbWhCZWtOQlRsRktaR1F5UTFkRldrcE9lbXBSYjNGSmRYWk5SbUpVWjFKTVNFRWlmUSIsImp0aSI6InVybjp2Yzp1dWlkOjlkMzdmMzY3LWE4ZDctNDY4Zi05NGYwLTk1NzAxNzBkNzZhNCIsIm5iZiI6MTcyMTk0MjA1MCwidmMiOnsiQGNvbnRleHQiOlsiaHR0cHM6Ly93d3cudzMub3JnLzIwMTgvY3JlZGVudGlhbHMvdjEiXSwidHlwZSI6WyJWZXJpZmlhYmxlQ3JlZGVudGlhbCJdLCJpc3N1ZXIiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SW5GNFYxRkthekUyUm1oQmVrTkJUbEZLWkdReVExZEZXa3BPZW1wUmIzRkpkWFpOUm1KVVoxSk1TRUVpZlEiLCJjcmVkZW50aWFsU3ViamVjdCI6eyJpc3N1ZXIiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SW5GNFYxRkthekUyUm1oQmVrTkJUbEZLWkdReVExZEZXa3BPZW1wUmIzRkpkWFpOUm1KVVoxSk1TRUVpZlEifSwiaWQiOiJ1cm46dmM6dXVpZDo5ZDM3ZjM2Ny1hOGQ3LTQ2OGYtOTRmMC05NTcwMTcwZDc2YTQiLCJpc3N1YW5jZURhdGUiOiIyMDI0LTA3LTI1VDIxOjE0OjEwWiIsImV4cGlyYXRpb25EYXRlIjoiMjAyNC0wOC0yNFQyMToxNDoxMFoiLCJjcmVkZW50aWFsU2NoZW1hIjpbeyJ0eXBlIjoiSnNvblNjaGVtYSIsImlkIjoiaHR0cHM6Ly92Yy5zY2hlbWFzLmhvc3Qva2JjLnNjaGVtYS5qc29uIn1dfX0.kakaSignature`
	_, err := jwt.Decode(vcJwt)
	assert.Error(t, err)
}

func Test_Decode_HeaderKID_InvalidDID(t *testing.T) {
	vcJwt := `eyJhbGciOiJFZERTQSIsImtpZCI6Imtha2EiLCJ0eXAiOiJKV1QifQ.eyJleHAiOjE3MjQ1MzQwNTAsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbkY0VjFGS2F6RTJSbWhCZWtOQlRsRktaR1F5UTFkRldrcE9lbXBSYjNGSmRYWk5SbUpVWjFKTVNFRWlmUSIsImp0aSI6InVybjp2Yzp1dWlkOjlkMzdmMzY3LWE4ZDctNDY4Zi05NGYwLTk1NzAxNzBkNzZhNCIsIm5iZiI6MTcyMTk0MjA1MCwidmMiOnsiQGNvbnRleHQiOlsiaHR0cHM6Ly93d3cudzMub3JnLzIwMTgvY3JlZGVudGlhbHMvdjEiXSwidHlwZSI6WyJWZXJpZmlhYmxlQ3JlZGVudGlhbCJdLCJpc3N1ZXIiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SW5GNFYxRkthekUyUm1oQmVrTkJUbEZLWkdReVExZEZXa3BPZW1wUmIzRkpkWFpOUm1KVVoxSk1TRUVpZlEiLCJjcmVkZW50aWFsU3ViamVjdCI6eyJpc3N1ZXIiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SW5GNFYxRkthekUyUm1oQmVrTkJUbEZLWkdReVExZEZXa3BPZW1wUmIzRkpkWFpOUm1KVVoxSk1TRUVpZlEifSwiaWQiOiJ1cm46dmM6dXVpZDo5ZDM3ZjM2Ny1hOGQ3LTQ2OGYtOTRmMC05NTcwMTcwZDc2YTQiLCJpc3N1YW5jZURhdGUiOiIyMDI0LTA3LTI1VDIxOjE0OjEwWiIsImV4cGlyYXRpb25EYXRlIjoiMjAyNC0wOC0yNFQyMToxNDoxMFoiLCJjcmVkZW50aWFsU2NoZW1hIjpbeyJ0eXBlIjoiSnNvblNjaGVtYSIsImlkIjoiaHR0cHM6Ly92Yy5zY2hlbWFzLmhvc3Qva2JjLnNjaGVtYS5qc29uIn1dfX0.VwvrU5Lmv3rn9rzXB0OCxe-MtE5R0876pXsXNLRuQjoqSNB5tBv_12NqrobwA-LkMzFwzdQ5-LWJni6grGdXCQ`
	_, err := jwt.Decode(vcJwt)
	assert.Error(t, err)
}
