package didcore

import "context"

// MethodResolver is an interface that can be implemented for resolving specific DID methods.
// Each concrete implementation should adhere to the DID core specficiation defined here:
// https://www.w3.org/TR/did-core/#did-resolution
type MethodResolver interface {
	Resolve(uri string) (ResolutionResult, error)
	ResolveWithContext(ctx context.Context, uri string) (ResolutionResult, error)
}

// ResolutionResult represents the result of a DID (Decentralized Identifier)
// resolution.
//
// This class encapsulates the metadata and document information obtained as
// a result of resolving a DID. It includes the resolution metadata, the DID
// document (if available), and the document metadata.
//
// The `DidResolutionResult` can be initialized with specific metadata and
// document information, or it can be created with default values if no
// specific information is provided.
type ResolutionResult struct {
	// The metadata associated with the DID resolution process.
	//
	// This includes information about the resolution process itself, such as any errors
	// that occurred. If not provided in the constructor, it defaults to an empty object
	// as per the spec
	ResolutionMetadata ResolutionMetadata `json:"didResolutionMetadata,omitempty"`
	// The resolved DID document, if available.
	//
	// This is the document that represents the resolved state of the DID. It may be `null`
	// if the DID could not be resolved or if the document is not available.
	Document Document `json:"didDocument"`
	// The metadata associated with the DID document.
	//
	// This includes information about the document such as when it was created and
	// any other relevant metadata. If not provided in the constructor, it defaults to an
	// empty `DidDocumentMetadata`.
	DocumentMetadata DocumentMetadata `json:"didDocumentMetadata,omitempty"`
}

// ResolutionResultWithError creates a Resolution Result populated with all default values and the error code provided.
func ResolutionResultWithError(errorCode string) ResolutionResult {
	return ResolutionResult{
		ResolutionMetadata: ResolutionMetadata{
			Error: errorCode,
		},
		DocumentMetadata: DocumentMetadata{},
	}
}

// ResolutionResultWithDocument creates a Resolution Result populated with all default values and the document provided.
func ResolutionResultWithDocument(document Document) ResolutionResult {
	return ResolutionResult{
		ResolutionMetadata: ResolutionMetadata{},
		Document:           document,
		DocumentMetadata:   DocumentMetadata{},
	}
}

// GetError returns the error code associated with the resolution result. returns an empty string if no error code is present.
func (r *ResolutionResult) GetError() string {
	return r.ResolutionMetadata.Error
}

// ResolutionMetadata is a metadata structure consisting of values relating to the results of the
// DID resolution process which typically changes between invocations of the
// resolve and resolveRepresentation functions, as it represents data about
// the resolution process itself
//
// Spec: https://www.w3.org/TR/did-core/#dfn-didresolutionmetadata
type ResolutionMetadata struct {
	// The Media Type of the returned didDocumentStream. This property is
	// REQUIRED if resolution is successful and if the resolveRepresentation
	// function was called
	ContentType string `json:"contentType,omitempty"`

	// The error code from the resolution process. This property is REQUIRED
	// when there is an error in the resolution process. The value of this
	// property MUST be a single keyword ASCII string. The possible property
	// values of this field SHOULD be registered in the
	// [DID Specification Registries](https://www.w3.org/TR/did-spec-registries/#error)
	Error string `json:"error,omitempty"`
}

// ResolutionError represents the error field of a ResolutionMetadata object. This struct implements error and is used to
// surface the error code from the resolution process. it is returned as the error value from resolve as a means to
// support idiomatic go error handling while also remaining spec compliant. It's worth mentioning that the spec expects
// error to be returned within ResolutionMedata. Given this, the error code is also present on ResolutionMetadata whenever
// an error occurs
// well known code values can be found here: https://www.w3.org/TR/did-spec-registries/#error
type ResolutionError struct {
	Code string
}

func (e ResolutionError) Error() string {
	return e.Code
}
