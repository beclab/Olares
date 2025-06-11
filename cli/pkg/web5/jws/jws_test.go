package jws_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"olares-cli/pkg/web5/dids/didkey"

	"olares-cli/pkg/web5/jws"

	"github.com/alecthomas/assert/v2"
)

func TestDecode(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	payload := []byte("hi")

	compactJWS, err := jws.Sign(payload, did)
	assert.NoError(t, err)

	decoded, err := jws.Decode(compactJWS)
	assert.NoError(t, err)

	assert.Equal(t, payload, decoded.Payload)
}

func TestDecode_SuccessWithTestJwtWithPayload(t *testing.T) {
	jwsString := "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaU" +
		"xDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbWRsWjI5YWNuWTVjemxuVWtwT1praFBlVGt5Tm" +
		"1oa1drNTBVMWxZWjJoaFlsOVJSbWhGTlRNM1lrMGlmUSMwIiwidHlwIjoiSldUIn0" +
		".eyJpc3MiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2ll" +
		"Q0k2SW1kbFoyOWFjblk1Y3psblVrcE9aa2hQZVRreU5taGtXazUwVTFsWVoyaGhZbDlSUm1oRk5UTT" +
		"NZazBpZlEiLCJqdGkiOiJ1cm46dmM6dXVpZDpjNWMzZGExMi02ODhmLTQxZDYtOTQzMC1lYzViNDAy" +
		"NTFmMzMiLCJuYmYiOjE3MTE2NTA4MjcsInN1YiI6IjEyMyIsInZjIjp7IkBjb250ZXh0IjpbImh0dHB" +
		"zOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOlsiVmVyaWZpYWJsZUNyZW" +
		"RlbnRpYWwiXSwiaXNzdWVyIjoiZGlkOmp3azpleUpyZEhraU9pSlBTMUFpTENKamNuWWlPaUpGWkRJMU" +
		"5URTVJaXdpZUNJNkltZGxaMjlhY25ZNWN6bG5Va3BPWmtoUGVUa3lObWhrV2s1MFUxbFlaMmhoWWw5UlJ" +
		"taEZOVE0zWWswaWZRIiwiY3JlZGVudGlhbFN1YmplY3QiOnsiaWQiOiIxMjMifSwiaWQiOiJ1cm46dmM" +
		"6dXVpZDpjNWMzZGExMi02ODhmLTQxZDYtOTQzMC1lYzViNDAyNTFmMzMiLCJpc3N1YW5jZURhdGUiOiIy" +
		"MDI0LTAzLTI4VDE4OjMzOjQ3WiJ9fQ" +
		".ydUiwf33dDCdk4RyPfoTdgbK3yTUpLCDpPBIECbn-rCGn_W3q5QxzAt43ClOIWibpOXHs-9T86UDBFPyd79vAQ"

	decoded, err := jws.Decode(jwsString)
	assert.NoError(t, err)

	assert.Equal(t, "EdDSA", decoded.Header.ALG)
	assert.Equal(t, "JWT", decoded.Header.TYP)
	assert.Equal(t, "did:jwk:eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6Imdl"+
		"Z29acnY5czlnUkpOZkhPeTkyNmhkWk50U1lYZ2hhYl9RRmhFNTM3Yk0ifQ", decoded.SignerDID.URI)
	var payloadMap map[string]interface{}

	err = json.Unmarshal(decoded.Payload, &payloadMap)
	assert.NoError(t, err)

	if iss, ok := payloadMap["iss"].(string); ok {
		assert.Equal(t, "did:jwk:eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6Imdl"+
			"Z29acnY5czlnUkpOZkhPeTkyNmhkWk50U1lYZ2hhYl9RRmhFNTM3Yk0ifQ", iss)
	} else {
		t.Fail()
	}
	if subject, ok := payloadMap["sub"].(string); ok {
		assert.Equal(t, "123", subject)
	} else {
		t.Fail()
	}
	if notBefore, ok := payloadMap["nbf"].(float64); ok {
		assert.Equal(t, 1711650827, notBefore)
	} else {
		t.Fail()
	}

}

func TestDecode_SuccessWithTestJwtWithDetachedPayload(t *testing.T) {
	jwsString := "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaU" +
		"xDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbWRsWjI5YWNuWTVjemxuVWtwT1praFBlVGt5Tm" +
		"1oa1drNTBVMWxZWjJoaFlsOVJSbWhGTlRNM1lrMGlmUSMwIiwidHlwIjoiSldUIn0" +
		"..ydUiwf33dDCdk4RyPfoTdgbK3yTUpLCDpPBIECbn-rCGn_W3q5QxzAt43ClOIWibpOXHs-9T86UDBFPyd79vAQ"

	payloadBase64Url := "eyJpc3MiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2ll" +
		"Q0k2SW1kbFoyOWFjblk1Y3psblVrcE9aa2hQZVRreU5taGtXazUwVTFsWVoyaGhZbDlSUm1oRk5UTT" +
		"NZazBpZlEiLCJqdGkiOiJ1cm46dmM6dXVpZDpjNWMzZGExMi02ODhmLTQxZDYtOTQzMC1lYzViNDAy" +
		"NTFmMzMiLCJuYmYiOjE3MTE2NTA4MjcsInN1YiI6IjEyMyIsInZjIjp7IkBjb250ZXh0IjpbImh0dHB" +
		"zOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOlsiVmVyaWZpYWJsZUNyZW" +
		"RlbnRpYWwiXSwiaXNzdWVyIjoiZGlkOmp3azpleUpyZEhraU9pSlBTMUFpTENKamNuWWlPaUpGWkRJMU" +
		"5URTVJaXdpZUNJNkltZGxaMjlhY25ZNWN6bG5Va3BPWmtoUGVUa3lObWhrV2s1MFUxbFlaMmhoWWw5UlJ" +
		"taEZOVE0zWWswaWZRIiwiY3JlZGVudGlhbFN1YmplY3QiOnsiaWQiOiIxMjMifSwiaWQiOiJ1cm46dmM" +
		"6dXVpZDpjNWMzZGExMi02ODhmLTQxZDYtOTQzMC1lYzViNDAyNTFmMzMiLCJpc3N1YW5jZURhdGUiOiIy" +
		"MDI0LTAzLTI4VDE4OjMzOjQ3WiJ9fQ"

	payloadByteArray, err := base64.StdEncoding.DecodeString(payloadBase64Url)
	if err != nil {
		fmt.Println("Error decoding base64 string:", err)
		return
	}

	decoded, err := jws.Decode(jwsString, jws.Payload(payloadByteArray))
	assert.NoError(t, err)

	assert.Equal(t, "EdDSA", decoded.Header.ALG)
	assert.Equal(t, "JWT", decoded.Header.TYP)
	assert.Equal(t, "did:jwk:eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6Imdl"+
		"Z29acnY5czlnUkpOZkhPeTkyNmhkWk50U1lYZ2hhYl9RRmhFNTM3Yk0ifQ", decoded.SignerDID.URI)
	var payloadMap map[string]interface{}

	err = json.Unmarshal(decoded.Payload, &payloadMap)
	assert.NoError(t, err)

	if iss, ok := payloadMap["iss"].(string); ok {
		assert.Equal(t, "did:jwk:eyJrdHkiOiJPS1AiLCJjcnYiOiJFZDI1NTE5IiwieCI6Imdl"+
			"Z29acnY5czlnUkpOZkhPeTkyNmhkWk50U1lYZ2hhYl9RRmhFNTM3Yk0ifQ", iss)
	} else {
		t.Fail()
	}
	if subject, ok := payloadMap["sub"].(string); ok {
		assert.Equal(t, "123", subject)
	} else {
		t.Fail()
	}
	if notBefore, ok := payloadMap["nbf"].(float64); ok {
		assert.Equal(t, 1711650827, notBefore)
	} else {
		t.Fail()
	}

}

func TestDecode_HeaderIsNotBase64Url(t *testing.T) {

	compactJWS := "lol." +
		"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ." +
		"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	decoded, err := jws.Decode(compactJWS)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to decode header")
	assert.Equal(t, jws.Decoded{}, decoded)
}

func TestDecode_PayloadIsNotBase64Url(t *testing.T) {
	compactJWS := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"{woohoo}." +
		"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	decoded, err := jws.Decode(compactJWS)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to decode payload")
	assert.Equal(t, jws.Decoded{}, decoded)
}

func TestDecode_SignatureIsNotBase64Url(t *testing.T) {
	compactJWS := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ." +
		"{woot}"

	decoded, err := jws.Decode(compactJWS)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to decode signature")
	assert.Equal(t, jws.Decoded{}, decoded)
}

func TestDecode_MissingHeaderKid(t *testing.T) {
	compactJWS := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
		"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ." +
		"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	decoded, err := jws.Decode(compactJWS)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Expected header to contain kid")
	assert.Equal(t, jws.Decoded{}, decoded)
}

func TestDecode_Bad(t *testing.T) {
	badHeader := base64.RawURLEncoding.EncodeToString([]byte("hehe"))
	vectors := []string{
		"",
		"..",
		"a.b.c",
		fmt.Sprintf("%s.%s.%s", badHeader, badHeader, badHeader),
	}

	for _, vector := range vectors {
		decoded, err := jws.Decode(vector)

		assert.Error(t, err, "expected verification error. vector: %s", vector)
		assert.Equal(t, jws.Decoded{}, decoded, "expected empty DecodedJWS")
	}
}

func TestSign_Detached(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	payload := map[string]any{"hello": "world"}
	payloadBytes, err := json.Marshal(payload)
	assert.NoError(t, err)

	compactJWS, err := jws.Sign(payloadBytes, did, jws.DetachedPayload(true))
	assert.NoError(t, err)

	assert.True(t, compactJWS != "", "expected signature to be non-empty")

	parts := strings.Split(compactJWS, ".")
	assert.Equal(t, 3, len(parts), "expected 3 parts in compact JWS")
	assert.Equal(t, parts[1], "", "expected empty payload")
}

func TestSign_CustomType(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	payload := map[string]any{"hello": "world"}
	payloadBytes, err := json.Marshal(payload)
	assert.NoError(t, err)

	customType := "openid4vci-proof+jwt"

	compactJWS, err := jws.Sign(payloadBytes, did, jws.Type(customType))
	assert.NoError(t, err)

	parts := strings.Split(compactJWS, ".")
	encodedHeader := parts[0]
	header, err := jws.DecodeHeader(encodedHeader)
	assert.NoError(t, err)

	assert.Equal(t, customType, header.TYP)
}

func TestDecoded_Verify(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	payload := map[string]any{"hello": "world"}
	payloadBytes, err := json.Marshal(payload)
	assert.NoError(t, err)

	compactJWS, err := jws.Sign(payloadBytes, did)
	assert.NoError(t, err)

	decoded, err := jws.Decode(compactJWS)
	assert.NoError(t, err)
	assert.NotEqual(t, jws.Decoded{}, decoded, "expected decoded to not be empty")
}

func TestDecoded_Verify_Bad(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	header, err := jws.Header{
		ALG: "ES256K",
		KID: did.Document.VerificationMethod[0].ID,
	}.Encode()
	assert.NoError(t, err)

	payloadJSON := map[string]any{"hello": "world"}
	payloadBytes, _ := json.Marshal(payloadJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	compactJWS := fmt.Sprintf("%s.%s.%s", header, payload, payload)

	_, err = jws.Verify(compactJWS)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature")
}

func TestVerify(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	payload := map[string]any{"hello": "world"}
	payloadBytes, err := json.Marshal(payload)
	assert.NoError(t, err)

	compactJWS, err := jws.Sign(payloadBytes, did)
	assert.NoError(t, err)

	_, err = jws.Verify(compactJWS)
	assert.NoError(t, err)
}

func TestVerify_Detached(t *testing.T) {
	did, err := didkey.Create()
	assert.NoError(t, err)

	payload := []byte("hi")

	compactJWS, err := jws.Sign(payload, did, jws.DetachedPayload(true))
	assert.NoError(t, err)

	decoded, err := jws.Verify(compactJWS, jws.Payload(payload))
	assert.NoError(t, err)

	assert.Equal(t, payload, decoded.Payload)
}
