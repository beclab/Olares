package vc

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"olares-cli/pkg/web5/jwt"
)

// Verify decodes and verifies the vc-jwt. It checks for the presence of required fields and verifies the jwt.
// It returns the decoded vc-jwt and the verification result.
func Verify[T CredentialSubject](vcJWT string) (DecodedVCJWT[T], error) {
	decoded, err := Decode[T](vcJWT)
	if err != nil {
		return decoded, err
	}

	return decoded, decoded.Verify()
}

// Decode decodes a vc-jwt as per the [spec] and returns [DecodedVCJWT].
//
// # Note
//
// This function uses certain fields from the jwt claims to eagrly populate the vc model as described
// in the encoding section of the spec. The jwt fields will clobber any values that exist in the vc model.
// While the jwt claims should match the counterpart values in the vc model, it's possible that they don't
// but there would be no way to know if they don't match given that they're overwritten.
//
// [spec]: https://www.w3.org/TR/vc-data-model/#json-web-token
func Decode[T CredentialSubject](vcJWT string) (DecodedVCJWT[T], error) {
	decoded, err := jwt.Decode(vcJWT)
	if err != nil {
		return DecodedVCJWT[T]{}, fmt.Errorf("failed to decode vc-jwt: %w", err)
	}

	if decoded.Claims.Misc == nil {
		return DecodedVCJWT[T]{}, errors.New("vc-jwt missing vc claim")
	}

	if _, ok := decoded.Claims.Misc["vc"]; ok == false {
		return DecodedVCJWT[T]{}, errors.New("vc-jwt missing vc claim")
	}

	bytes, err := json.Marshal(decoded.Claims.Misc["vc"])
	if err != nil {
		return DecodedVCJWT[T]{}, fmt.Errorf("failed to decode vc claim: %w", err)
	}

	var vc DataModel[T]
	if err := json.Unmarshal(bytes, &vc); err != nil {
		return DecodedVCJWT[T]{}, fmt.Errorf("failed to decode vc claim: %w", err)
	}

	if vc.Type == nil {
		return DecodedVCJWT[T]{}, errors.New("vc-jwt missing vc type")
	}

	// the following conditionals are included to conform with the jwt decoding section
	// of the specification defined here: https://www.w3.org/TR/vc-data-model/#jwt-decoding
	if decoded.Claims.Issuer != "" {
		vc.Issuer = decoded.Claims.Issuer
	}

	if decoded.Claims.JTI != "" {
		vc.ID = decoded.Claims.JTI
	}

	if decoded.Claims.Subject != "" {
		vc.CredentialSubject.SetID(decoded.Claims.Subject)
	}

	if decoded.Claims.Expiration != 0 {
		vc.ExpirationDate = time.Unix(decoded.Claims.Expiration, 0).UTC().Format(time.RFC3339)
	}

	if decoded.Claims.NotBefore != 0 {
		vc.IssuanceDate = time.Unix(decoded.Claims.NotBefore, 0).UTC().Format(time.RFC3339)
	}

	return DecodedVCJWT[T]{
		JWT: decoded,
		VC:  vc,
	}, nil

}

// DecodedVCJWT represents a decoded vc-jwt. It contains the decoded jwt and decoded vc data model
type DecodedVCJWT[T CredentialSubject] struct {
	JWT jwt.Decoded
	VC  DataModel[T]
}

// Verify verifies the decoded vc-jwt. It checks for the presence of required fields and verifies the jwt.
func (vcjwt DecodedVCJWT[T]) Verify() error {
	if vcjwt.JWT.Header.TYP != "JWT" {
		return errors.New("invalid typ")
	}

	if vcjwt.VC.Issuer == "" {
		return errors.New("missing issuer")
	}

	if vcjwt.VC.ID == "" {
		return errors.New("missing id")
	}

	if vcjwt.VC.IssuanceDate == "" {
		return errors.New("missing issuance date")
	}

	issuanceDate, err := time.Parse(time.RFC3339, vcjwt.VC.IssuanceDate)
	if err != nil {
		return fmt.Errorf("failed to parse issuance date: %w", err)
	}

	if time.Now().UTC().Before(issuanceDate.UTC()) {
		return fmt.Errorf("vc cannot be used before %s", vcjwt.VC.IssuanceDate)
	}

	if vcjwt.VC.ExpirationDate != "" {
		exp, err := time.Parse(time.RFC3339, vcjwt.VC.ExpirationDate)
		if err != nil {
			return fmt.Errorf("failed to parse expiration date: %w", err)
		}

		if time.Now().UTC().After(exp.UTC()) {
			return fmt.Errorf("vc expired on %s", vcjwt.VC.ExpirationDate)
		}
	}

	if vcjwt.VC.Type == nil || len(vcjwt.VC.Type) == 0 {
		return errors.New("missing type")
	}

	if slices.Contains(vcjwt.VC.Type, BaseType) == false {
		return fmt.Errorf("missing base type: %s", BaseType)
	}

	if vcjwt.VC.Context == nil || len(vcjwt.VC.Context) == 0 {
		return errors.New("missing @context")
	}

	if slices.Contains(vcjwt.VC.Context, BaseContext) == false {
		return fmt.Errorf("missing base @context: %s", BaseContext)
	}

	err = vcjwt.JWT.Verify()
	if err != nil {
		return fmt.Errorf("integrity check mismatch: %w", err)
	}

	return nil
}
