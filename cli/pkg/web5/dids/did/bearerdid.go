package did

import (
	"fmt"

	"olares-cli/pkg/web5/crypto"
	"olares-cli/pkg/web5/dids/didcore"
	"olares-cli/pkg/web5/jwk"
)

// BearerDID is a composite type that combines a DID with a KeyManager containing keys
// associated to the DID. Together, these two components form a BearerDID that can be used to
// sign and verify data.
type BearerDID struct {
	DID
	crypto.KeyManager
	Document didcore.Document
}

// DIDSigner is a function returned by GetSigner that can be used to sign a payload with a key
// associated to a BearerDID.
type DIDSigner func(payload []byte) ([]byte, error)

// ToPortableDID exports a BearerDID to a portable format
func (d *BearerDID) ToPortableDID() (PortableDID, error) {
	portableDID := PortableDID{
		URI:      d.URI,
		Document: d.Document,
	}

	exporter, ok := d.KeyManager.(crypto.KeyExporter)
	if ok {
		privateKeys := make([]jwk.JWK, 0)

		for _, vm := range d.Document.VerificationMethod {
			keyAlias, err := vm.PublicKeyJwk.ComputeThumbprint()
			if err != nil {
				continue
			}

			key, err := exporter.ExportKey(keyAlias)
			if err != nil {
				// TODO: decide if we want to blow up or continue
				continue
			}

			privateKeys = append(privateKeys, key)
		}

		portableDID.PrivateKeys = privateKeys
	}

	return portableDID, nil
}

// GetSigner returns a sign method that can be used to sign a payload using a key associated to the DID.
// This function also returns the verification method needed to verify the signature.
//
// Providing the verification method allows the caller to provide the signature's recipient
// with a reference to the verification method needed to verify the payload. This is often done
// by including the verification method id either alongside the signature or as part of the header
// in the case of JSON Web Signatures.
//
// The verifier can dereference the verification method id to obtain the public key needed to verify the signature.
//
// This function takes a Verification Method selector that can be used to select a specific verification method
// from the DID Document if desired. If no selector is provided, the payload will be signed with the key associated
// to the first verification method in the DID Document.
//
// The selector can either be a Verification Method ID or a Purpose. If a Purpose is provided, the first verification
// method in the DID Document that has the provided purpose will be used to sign the payload.
//
// The returned signer is a function that takes a byte payload and returns a byte signature.
func (d *BearerDID) GetSigner(selector didcore.VMSelector) (DIDSigner, didcore.VerificationMethod, error) {
	vm, err := d.Document.SelectVerificationMethod(selector)
	if err != nil {
		return nil, didcore.VerificationMethod{}, err
	}

	keyAlias, err := vm.PublicKeyJwk.ComputeThumbprint()
	if err != nil {
		return nil, didcore.VerificationMethod{}, fmt.Errorf("failed to compute key alias: %s", err.Error())
	}

	signer := func(payload []byte) ([]byte, error) {
		return d.Sign(keyAlias, payload)
	}

	return signer, vm, nil
}

// FromPortableDID inflates a BearerDID from a portable format.
func FromPortableDID(portableDID PortableDID) (BearerDID, error) {
	did, err := Parse(portableDID.URI)
	if err != nil {
		return BearerDID{}, err
	}

	keyManager := crypto.NewLocalKeyManager()
	for _, key := range portableDID.PrivateKeys {
		_, err := keyManager.ImportKey(key)
		if err != nil {
			// todo what should we do here?
			return BearerDID{}, err
		}
	}

	return BearerDID{
		DID:        did,
		KeyManager: keyManager,
		Document:   portableDID.Document,
	}, nil
}
