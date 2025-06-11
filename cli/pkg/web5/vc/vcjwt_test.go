package vc_test

import (
	"fmt"
	"testing"
	"time"

	"olares-cli/pkg/web5/dids/didkey"
	"olares-cli/pkg/web5/jwt"
	"olares-cli/pkg/web5/vc"

	"github.com/alecthomas/assert/v2"
	"github.com/decentralized-identity/web5-go"
)

type vector struct {
	description string
	input       string
	errors      bool
}

func TestDecode(t *testing.T) {
	// TODO: move these to web5-spec repo test-vectors (Moe - 2024-02-24)
	vectors := []vector{
		{
			description: "fail to decode jwt",
			input:       "doodoo",
			errors:      true,
		},
		{
			description: "no claims",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJa3RPTVZGRU5ETkhZVkpJTm1OeWJpMTFTWFk0ZDBoT1pqZHlSMlkyVUY5RFMxZzJkbmsyTUdjMWQyc2lmUSMwIn0.e30.1iq9_pDtMlzL22h6xVY77nRNfXnR3oFU2kNYDAM52dPAs0l8zLL6AJ18B8rz9HziYzRo4Zo_jyYhq4nlHE3lBw",
			errors:      true,
		}, {
			description: "no vc claim",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJa0YyZFMxVVNsTlRWakZKT0dob1NrUmFlWGw1YUhnMmVHSlZjamRQT0RWMWRFMWpaa3RLVDJOblFWVWlmUSMwIn0.eyJoZWhlIjoiaGkifQ.QQ5aottVrsHRisxx7vRzin9CnyOcxeScxLOIy5qI30pV2FkXXBe3BdyujLS7i7M0CHW0eS9XhaVKe76504RZCQ",
			errors:      true,
		}, {
			description: "vc claim wrong type",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbEZvUkhsdFprdzFaQzB0WjFGQ1prOVNOMlZFYjBrelprTjJOVUV0Y3pBMGFYZHlZMGRsTkVWd1lsa2lmUSMwIn0.eyJ2YyI6ImhpIn0.O_-xPUAZhi9W3OD1pJn4wN5Q9nZKYXcmtJPhWuk6WxlOXMca2jNXyjYpEKCJ1vFWZ4OHfSifErPvClLsH8-MCQ",
			errors:      true,
		}, {
			description: "legit",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbVJwVGkxRlEydGhaRTVEVlVKZlUxRkhTVFJtVUdOZlluVmZObmt3VWpKRFdEUllkMjlTUzBjNFVEZ2lmUSMwIn0.eyJpc3MiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SW1ScFRpMUZRMnRoWkU1RFZVSmZVMUZIU1RSbVVHTmZZblZmTm5rd1VqSkRXRFJZZDI5U1MwYzRVRGdpZlEiLCJqdGkiOiJ1cm46dmM6dXVpZDoxOGQ5OTZjZi03N2YwLTRkYjgtOGQ5MS0zNGI1ZDY1NzcwNmUiLCJuYmYiOjE3MDg3NTY3ODUsInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbVJwVGkxRlEydGhaRTVEVlVKZlUxRkhTVFJtVUdOZlluVmZObmt3VWpKRFdEUllkMjlTUzBjNFVEZ2lmUSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOlsiVmVyaWZpYWJsZUNyZWRlbnRpYWwiXSwiaXNzdWVyIjoiZGlkOmp3azpleUpyZEhraU9pSlBTMUFpTENKamNuWWlPaUpGWkRJMU5URTVJaXdpZUNJNkltUnBUaTFGUTJ0aFpFNURWVUpmVTFGSFNUUm1VR05mWW5WZk5ua3dVakpEV0RSWWQyOVNTMGM0VURnaWZRIiwiY3JlZGVudGlhbFN1YmplY3QiOnsiaWQiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SW1ScFRpMUZRMnRoWkU1RFZVSmZVMUZIU1RSbVVHTmZZblZmTm5rd1VqSkRXRFJZZDI5U1MwYzRVRGdpZlEifSwiaWQiOiJ1cm46dmM6dXVpZDoxOGQ5OTZjZi03N2YwLTRkYjgtOGQ5MS0zNGI1ZDY1NzcwNmUiLCJpc3N1YW5jZURhdGUiOiIyMDI0LTAyLTI0VDA2OjM5OjQ1WiJ9fQ.Y1-9dFop7bg_0jvgZMLyE3CPjnSXH9SGTHeA_jn5HosYbhST8y_pK7LcDeCYLSgDfiIOeVsvJFqOr3XT2J2cDA",
			errors:      false,
		},
	}

	for _, tt := range vectors {
		t.Run(tt.description, func(t *testing.T) {
			decoded, err := vc.Decode[vc.Claims](tt.input)

			if tt.errors == true {
				assert.Error(t, err)
				assert.Equal(t, vc.DecodedVCJWT[vc.Claims]{}, decoded)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, vc.DecodedVCJWT[vc.Claims]{}, decoded)
			}
		})
	}
}

func TestDecode_SetClaims(t *testing.T) {
	issuer, err := didkey.Create()
	assert.NoError(t, err)

	subject, err := didkey.Create()
	assert.NoError(t, err)

	subjectClaims := vc.Claims{
		"firstName": "Randy",
		"lastName":  "McRando",
	}

	issuanceDate := time.Now().UTC()

	// missing issuer
	jwtClaims := jwt.Claims{
		JTI:        "abcd123",
		Issuer:     issuer.URI,
		Subject:    subject.URI,
		NotBefore:  issuanceDate.Unix(),
		Expiration: issuanceDate.Add(time.Hour).Unix(),
		Misc: map[string]any{
			"vc": vc.DataModel[vc.Claims]{
				CredentialSubject: subjectClaims,
				Type:              []string{"Something"},
			},
		},
	}

	vcJWT, err := jwt.Sign(jwtClaims, issuer)
	assert.NoError(t, err)

	decoded, err := vc.Decode[vc.Claims](vcJWT)
	assert.NoError(t, err)

	assert.Equal(t, jwtClaims.JTI, decoded.VC.ID)
	assert.Equal(t, jwtClaims.Issuer, decoded.VC.Issuer)
	assert.Equal(t, jwtClaims.Subject, decoded.VC.CredentialSubject.GetID())
	assert.Equal(t, issuanceDate.Format(time.RFC3339), decoded.VC.IssuanceDate)
	assert.NotZero(t, decoded.VC.ExpirationDate)
}

func TestVerify(t *testing.T) {
	vectors := []vector{
		{
			description: "no typ header",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJa1JXZDFwWmFWY3lTalZETkc1cmQyRkdRV0pKVUY5MlIzTnJTamhKT1VKRk5IcE9RVGgxUkdZMVZsVWlmUSMwIn0.eyJleHAiOjI2NTUyMTM3MDQsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJa1JXZDFwWmFWY3lTalZETkc1cmQyRkdRV0pKVUY5MlIzTnJTamhKT1VKRk5IcE9RVGgxUkdZMVZsVWlmUSIsImp0aSI6ImFiY2QxMjMiLCJuYmYiOjE3MDkxMzM3MDQsInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbWRpZVhwcGNuTmthekpOVW14WVV5MWtURkU0TkZWbVRrRlZjbUp5T0hZd2FESkViblZVTUdSV1kxRWlmUSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOlsiVmVyaWZpYWJsZUNyZWRlbnRpYWwiXSwiaXNzdWVyIjoiIiwiY3JlZGVudGlhbFN1YmplY3QiOnsiZmlyc3ROYW1lIjoiUmFuZHkiLCJsYXN0TmFtZSI6Ik1jUmFuZG8ifSwiaXNzdWFuY2VEYXRlIjoiIn19.7trsEIJxKlQhvCH3F-w4ZTessbGaCG6X_6di8sl3qTRdEk8QFyv7xvFSFXBcX4XC6i_DfWndlhj1cdEtL9B1CA",
			errors:      true,
		}, {
			description: "invalid typ header",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbmxXTFhvMGNqZGlNMmRtZEV0U016ZEpNR3N3VVhsSk0yZHBTa0Z3ZFhsU1FtMXpZVXBSWjI0eWVUUWlmUSMwIiwidHlwIjoiS2FrYW1pbWkifQ.eyJleHAiOjI2NTUyMTM3ODMsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbmxXTFhvMGNqZGlNMmRtZEV0U016ZEpNR3N3VVhsSk0yZHBTa0Z3ZFhsU1FtMXpZVXBSWjI0eWVUUWlmUSIsImp0aSI6ImFiY2QxMjMiLCJuYmYiOjE3MDkxMzM3ODMsInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbmRuTTFGUVJsSmplUzAxVXpaNlNqZEVWMmx4U0Vwd1RHTlJaRmhVVWsxWk1td3hhRTEyWm1reFVXTWlmUSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOlsiVmVyaWZpYWJsZUNyZWRlbnRpYWwiXSwiaXNzdWVyIjoiIiwiY3JlZGVudGlhbFN1YmplY3QiOnsiZmlyc3ROYW1lIjoiUmFuZHkiLCJsYXN0TmFtZSI6Ik1jUmFuZG8ifSwiaXNzdWFuY2VEYXRlIjoiIn19.fE58Vtqg5-oOQKvRCiJHCspZaqmGOtEIlUTf8TqWpviWGndpZWj1XofcUfcNFLWTHnk6H-2ku9FA7x_t4ymgAA",
			errors:      true,
		}, {
			description: "no id",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJakl5T1RoZldFSkhPRk41Y1drNFJuQnlWekF4U0RGblpHdzRlRXRZVERaUlNsaDJhR05mWW5sek1EZ2lmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjM2MDEyOTMyODIsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJakl5T1RoZldFSkhPRk41Y1drNFJuQnlWekF4U0RGblpHdzRlRXRZVERaUlNsaDJhR05mWW5sek1EZ2lmUSIsIm5iZiI6MTcwOTEzMzI4Miwic3ViIjoiZGlkOmp3azpleUpyZEhraU9pSlBTMUFpTENKamNuWWlPaUpGWkRJMU5URTVJaXdpZUNJNkluUXdhMmRYUldaSU5scG5TVGxGVW5wMFIxazBMV3RzU2paUFlVVTFTSGhrWlZOUFRUUmtPRVE0WW1NaWZRIiwidmMiOnsiQGNvbnRleHQiOlsiaHR0cHM6Ly93d3cudzMub3JnLzIwMTgvY3JlZGVudGlhbHMvdjEiXSwidHlwZSI6WyJWZXJpZmlhYmxlQ3JlZGVudGlhbCJdLCJpc3N1ZXIiOiIiLCJjcmVkZW50aWFsU3ViamVjdCI6eyJmaXJzdE5hbWUiOiJSYW5keSIsImxhc3ROYW1lIjoiTWNSYW5kbyJ9LCJpc3N1YW5jZURhdGUiOiIifX0.e7ITj0XXqXBqulsz5Bv0ACtBY9T_EW2jdJbM1Cv5C1Rg1uy_fVWN0asYOOgT75oU87W8dV8bKM_2YMns1JoOCA",
			errors:      true,
		}, {
			description: "no issuer",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJblYzUTJkV2IxVmplbTFpV1ZoWFVFbGFOMDgyTkdRMkxYTXhkamhPYkRacVoyUnpSMEpOV1d0cVV6QWlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjM2MDEyOTMyOTksImp0aSI6ImFiY2QxMjMiLCJuYmYiOjE3MDkxMzMyOTksInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJazVQUkRKQ01qSnVOaTFZTkdOTVNHc3hha2hEWDFaNFptVkNWVWd3UVdoYVJrcE1XRFJ1TTNwWFRWVWlmUSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOlsiVmVyaWZpYWJsZUNyZWRlbnRpYWwiXSwiaXNzdWVyIjoiIiwiY3JlZGVudGlhbFN1YmplY3QiOnsiZmlyc3ROYW1lIjoiUmFuZHkiLCJsYXN0TmFtZSI6Ik1jUmFuZG8ifSwiaXNzdWFuY2VEYXRlIjoiIn19.DZs9KgQLD8k0ouL1W8ENBzWXOKSJZ_plm7UOmA2VQtmyqUISnB_KxqncY-MIWzTXfPObQzMALY-ZVjYgqABRDA",
			errors:      true,
		}, {
			description: "no issuance date",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJa0pNYVhaT1Iza3hSV0Z5VDJGZlUzUkVPSEY1TjFaT1F6TnhhRW90TFZSdlJYTjVMVXBCYm5WdlJtOGlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjM2MDEyOTMzMTksImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJa0pNYVhaT1Iza3hSV0Z5VDJGZlUzUkVPSEY1TjFaT1F6TnhhRW90TFZSdlJYTjVMVXBCYm5WdlJtOGlmUSIsImp0aSI6ImFiY2QxMjMiLCJzdWIiOiJkaWQ6andrOmV5SnJkSGtpT2lKUFMxQWlMQ0pqY25ZaU9pSkZaREkxTlRFNUlpd2llQ0k2SWsxM1VtVndiWFo2T1hwaVVsSm5Za1l4UVc1cFluSnlhelZpVlV4NlUxWkNibEl0YWxGM2JYQlRVazBpZlEiLCJ2YyI6eyJAY29udGV4dCI6WyJodHRwczovL3d3dy53My5vcmcvMjAxOC9jcmVkZW50aWFscy92MSJdLCJ0eXBlIjpbIlZlcmlmaWFibGVDcmVkZW50aWFsIl0sImlzc3VlciI6IiIsImNyZWRlbnRpYWxTdWJqZWN0Ijp7ImZpcnN0TmFtZSI6IlJhbmR5IiwibGFzdE5hbWUiOiJNY1JhbmRvIn0sImlzc3VhbmNlRGF0ZSI6IiJ9fQ.1O8nUSZIUrbfqp0ulLabhyME7b7nQSx9lPwkkLvNmJGHaNCgF3EodDM88V8Zzke1meDWRM2tAUeZrTDCUWPGBQ",
			errors:      true,
		}, {
			description: "issuance date in future",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbFZUV0dSNlNVSkdNekZ0YWtkVWNrSTBPWGRNZWt4MGMyNWlibmxrTlVwbGVHOTBkR1prV0RkNVlqQWlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjQ1NDczNzMzNDQsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbFZUV0dSNlNVSkdNekZ0YWtkVWNrSTBPWGRNZWt4MGMyNWlibmxrTlVwbGVHOTBkR1prV0RkNVlqQWlmUSIsImp0aSI6ImFiY2QxMjMiLCJuYmYiOjI2NTUyMTMzNDQsInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbnBIYzJFMk5rTTVWRXh1TmxscE1FdE5RbnBST1ZOQ1gyMVBZa1Y0UTJrdGVHaEdaWGRzVDBGQkxUZ2lmUSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOlsiVmVyaWZpYWJsZUNyZWRlbnRpYWwiXSwiaXNzdWVyIjoiIiwiY3JlZGVudGlhbFN1YmplY3QiOnsiZmlyc3ROYW1lIjoiUmFuZHkiLCJsYXN0TmFtZSI6Ik1jUmFuZG8ifSwiaXNzdWFuY2VEYXRlIjoiIn19.0VxukH5AOhQ1WGzInDFXkOgr5gzvGCKLhFkbSsD76GcKqcDsuek1uQz-IoTOoZIA97d1yczFJCLMNMRV05ARCA",
			errors:      true,
		}, {
			description: "no context",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJalYyUnpGUlEyOHlTR3BNTjFCWlJraERNM2t3ZDE5VGQzTkZRV2xCYldGU1lXRkZWRmM1V2sxRk56QWlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjM2MDEyOTMzNzQsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJalYyUnpGUlEyOHlTR3BNTjFCWlJraERNM2t3ZDE5VGQzTkZRV2xCYldGU1lXRkZWRmM1V2sxRk56QWlmUSIsImp0aSI6ImFiY2QxMjMiLCJuYmYiOjE3MDkxMzMzNzQsInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbU5HVEhwSmMwOUVWME0wTVRKVGJuWnpSSFJuVkVSWGJuTlplSGgyT0dNMGFEZHVTRE5YZUZoVFprRWlmUSIsInZjIjp7IkBjb250ZXh0IjpudWxsLCJ0eXBlIjpbIlZlcmlmaWFibGVDcmVkZW50aWFsIl0sImlzc3VlciI6IiIsImNyZWRlbnRpYWxTdWJqZWN0Ijp7ImZpcnN0TmFtZSI6IlJhbmR5IiwibGFzdE5hbWUiOiJNY1JhbmRvIn0sImlzc3VhbmNlRGF0ZSI6IiJ9fQ.pixi8ODIn1TpUwTiQ_GJhP8vLgr8XNX1CBuXhtk9VnU4DJ137zHX30Z_BVgW7oFlrMcDpzq67A3PHaw6JHgdDA",
			errors:      true,
		}, {
			description: "missing base context",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbDlpY3pZMGN6SnBkMkp4ZFc5RGNrRmxSbkE0WHprMk1UaHdjblZ3WVdGMldtWmtjMDFpZW1oaFVuTWlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjM2MDEyOTM0MDAsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbDlpY3pZMGN6SnBkMkp4ZFc5RGNrRmxSbkE0WHprMk1UaHdjblZ3WVdGMldtWmtjMDFpZW1oaFVuTWlmUSIsImp0aSI6ImFiY2QxMjMiLCJuYmYiOjE3MDkxMzM0MDAsInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbWh2Tlc5SE1YazBjVFpQWVcxWVlWaEpUV3hVVlhScVYzRXhPRkF0TlcxRk5HNTRPVk5EVkVsMlZXOGlmUSIsInZjIjp7IkBjb250ZXh0IjpbIlN0cmVldENyZWRlbnRpYWwiXSwidHlwZSI6WyJWZXJpZmlhYmxlQ3JlZGVudGlhbCJdLCJpc3N1ZXIiOiIiLCJjcmVkZW50aWFsU3ViamVjdCI6eyJmaXJzdE5hbWUiOiJSYW5keSIsImxhc3ROYW1lIjoiTWNSYW5kbyJ9LCJpc3N1YW5jZURhdGUiOiIifX0.opzudBm-S5dyAWrKrNMF0XR0Ol_98OeH7Zvt-xaYlLQPRrGzKsFFTsr_crfQHQaX27WPTKacLBDc4C2ocB-1Aw",
			errors:      true,
		}, {
			description: "no type",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbTVUVHpOalMydFJkelpCWWs5dGEyazNZa1ZPYVZGeVRXc3lkVE0xT0hSSGQxcHFaRFpSTm5CeVUyOGlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjM2MDEyOTM0MzEsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbTVUVHpOalMydFJkelpCWWs5dGEyazNZa1ZPYVZGeVRXc3lkVE0xT0hSSGQxcHFaRFpSTm5CeVUyOGlmUSIsImp0aSI6ImFiY2QxMjMiLCJuYmYiOjE3MDkxMzM0MzEsInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJa1JIUm1KRWVVVXpZalJHYUZwNVpXNWtlV3h2TTBwbWRsUnVaMkZWV0Y5b1ltWm9lR2szV1RSNmJYY2lmUSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOm51bGwsImlzc3VlciI6IiIsImNyZWRlbnRpYWxTdWJqZWN0Ijp7ImZpcnN0TmFtZSI6IlJhbmR5IiwibGFzdE5hbWUiOiJNY1JhbmRvIn0sImlzc3VhbmNlRGF0ZSI6IiJ9fQ.yrvOZc58oFqEXpMs6rk4E0QDLv28gjjunNFSafx0yV6tmn0nYO2btJnawPusrTcHt0tTjxB5SMUEyo6m7kWsAw",
			errors:      true,
		}, {
			description: "missing base type",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbHBOYmt0aGRYZERkbWx3YUd4V1NFVXRjVlZ0T1VveVQyWk5RWE5JVkc5cFNtcEZVVkl4Y1RKMFZUUWlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjM2MDEyOTM0ODYsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbHBOYmt0aGRYZERkbWx3YUd4V1NFVXRjVlZ0T1VveVQyWk5RWE5JVkc5cFNtcEZVVkl4Y1RKMFZUUWlmUSIsImp0aSI6ImFiY2QxMjMiLCJuYmYiOjE3MDkxMzM0ODYsInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbGcwY0MxSk5qTnlYM1JZTVhsUVdXTXdjMFZZV1VwVWRuQTFOWFE0VVRkVU4yMWpaVmhhZVd4V1VtOGlmUSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOlsiU3RyZWV0Q3JlZGVudGlhbCJdLCJpc3N1ZXIiOiIiLCJjcmVkZW50aWFsU3ViamVjdCI6eyJmaXJzdE5hbWUiOiJSYW5keSIsImxhc3ROYW1lIjoiTWNSYW5kbyJ9LCJpc3N1YW5jZURhdGUiOiIifX0.VeXTzzoy8krwdg5-6zBqDueZ0l5RH3EwdhMO4n9BQ8ba8qe3Xx6nemnrLHpTYI7glMVg-ynTHugCgQXsspKsAg",
			errors:      true,
		}, {
			description: "expired",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbUUxVDA4NVl6TjVVVXd0UW5RNWIwMXpSM1JEYUdWck5tdE9OVXAyUkU4d2VGRXRlVkpuTjBOYWNXTWlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjE3MDkxMzM1OTcsImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbUUxVDA4NVl6TjVVVXd0UW5RNWIwMXpSM1JEYUdWck5tdE9OVXAyUkU4d2VGRXRlVkpuTjBOYWNXTWlmUSIsImp0aSI6ImFiY2QxMjMiLCJuYmYiOjE3MDkxMzM1OTYsInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJa04yZFhCMFpVaHBMWE5OUW1wb1N6Tk1VbXBrYW1OeFVVTkRjVWxzTTB4cGNIVmlNMUpYWVc5NmJsVWlmUSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOlsiVmVyaWZpYWJsZUNyZWRlbnRpYWwiXSwiaXNzdWVyIjoiIiwiY3JlZGVudGlhbFN1YmplY3QiOnsiZmlyc3ROYW1lIjoiUmFuZHkiLCJsYXN0TmFtZSI6Ik1jUmFuZG8ifSwiaXNzdWFuY2VEYXRlIjoiIn19.BhKQc5Q96ZAISp_qJ0qni0NfZhde3Z0A9hJET8Twhzu2XNA89OgHu8lKyp0M9Fj8WGFVZGOrpfnZtoA13mKMAQ",
			errors:      true,
		}, {
			description: "legit",
			input:       "eyJhbGciOiJFZERTQSIsImtpZCI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbGhCWmpkZmJWOXJZa2RsWVVGcVh6Wm9RMnBWVGtOaFRGUmlRek5DY1ZkaE5UWkVSSG93YlhwTVVYTWlmUSMwIiwidHlwIjoiSldUIn0.eyJleHAiOjI2NTUyMTM2MzksImlzcyI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbGhCWmpkZmJWOXJZa2RsWVVGcVh6Wm9RMnBWVGtOaFRGUmlRek5DY1ZkaE5UWkVSSG93YlhwTVVYTWlmUSIsImp0aSI6ImFiY2QxMjMiLCJuYmYiOjE3MDkxMzM2MzksInN1YiI6ImRpZDpqd2s6ZXlKcmRIa2lPaUpQUzFBaUxDSmpjbllpT2lKRlpESTFOVEU1SWl3aWVDSTZJbXROUmxWVVUwWlpia0pEUVZoRldVZEdXalozWjNSVFdWWTNlblF0Y0ZORlpVdHNOMU52YVROclgyTWlmUSIsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIl0sInR5cGUiOlsiVmVyaWZpYWJsZUNyZWRlbnRpYWwiXSwiaXNzdWVyIjoiIiwiY3JlZGVudGlhbFN1YmplY3QiOnsiZmlyc3ROYW1lIjoiUmFuZHkiLCJsYXN0TmFtZSI6Ik1jUmFuZG8ifSwiaXNzdWFuY2VEYXRlIjoiIn19.l3N5-G9sMInwzjQCFhyYNwjRMRd9ojdgsQEsH8S_reYShP_H0BwLbzHSXkIbeFMVTivUziMzOhq0pMJTr-0sAw",
			errors:      false,
		},
	}

	for _, tt := range vectors {
		t.Run(tt.description, func(t *testing.T) {
			_, err := vc.Verify[vc.Claims](tt.input)

			if tt.errors == true {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVector_Decode(t *testing.T) {
	testVectors, err :=
		web5.LoadTestVectors[string, any]("../web5-spec/test-vectors/vc_jwt/decode.json")
	assert.NoError(t, err)
	fmt.Println("Running test vectors: ", testVectors.Description)

	for _, vector := range testVectors.Vectors {
		t.Run(vector.Description, func(t *testing.T) {
			fmt.Println("Running test vector: ", vector.Description)

			_, err := vc.Decode[vc.Claims](vector.Input)

			if vector.Errors {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVector_Verify(t *testing.T) {
	testVectors, err :=
		web5.LoadTestVectors[string, any]("../web5-spec/test-vectors/vc_jwt/verify.json")
	assert.NoError(t, err)
	fmt.Println("Running test vectors: ", testVectors.Description)

	for _, vector := range testVectors.Vectors {
		t.Run(vector.Description, func(t *testing.T) {
			fmt.Println("Running test vector: ", vector.Description)

			_, err := vc.Verify[vc.Claims](vector.Input)

			if vector.Errors {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
