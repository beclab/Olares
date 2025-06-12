package pexv2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/beclab/Olares/cli/pkg/web5/vc"

	"github.com/PaesslerAG/jsonpath"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// PresentationDefinition represents a DIF Presentation Definition defined [here].
// Presentation Definitions are objects that articulate what proofs a Verifier requires
//
// [here]: https://identity.foundation/presentation-exchange/#presentation-definition
type PresentationDefinition struct {
	ID               string            `json:"id"`
	Name             string            `json:"name,omitempty"`
	Purpose          string            `json:"purpose,omitempty"`
	InputDescriptors []InputDescriptor `json:"input_descriptors"`
}

// InputDescriptor represents a DIF Input Descriptor defined [here].
// Input Descriptors are used to describe the information a Verifier requires of a Holder.
//
// [here]: https://identity.foundation/presentation-exchange/#input-descriptor
type InputDescriptor struct {
	ID          string      `json:"id"`
	Name        string      `json:"name,omitempty"`
	Purpose     string      `json:"purpose,omitempty"`
	Constraints Constraints `json:"constraints"`
}

type tokenizedField struct {
	name string
	path string
}

// SelectCredentials selects vcJWTs based on the constraints defined in the input descriptor
func (ind InputDescriptor) SelectCredentials(vcJWTs []string) ([]string, error) {
	jsonSchema := JSONSchema{
		Schema:     "http://json-schema.org/draft-07/schema#",
		Type:       "object",
		Properties: make(map[string]Filter, len(ind.Constraints.Fields)),
		Required:   make([]string, 0, len(ind.Constraints.Fields)),
	}

	// Each Field can have multiple Paths. Add a 'tokenizedField' for each Path, and add the Filter to the JSON Schema
	tokenizedFields := make([]tokenizedField, 0, len(ind.Constraints.Fields))
	for _, field := range ind.Constraints.Fields {
		name := strconv.FormatInt(rand.Int63(), 10) //nolint:gosec
		for _, path := range field.Path {
			tf := tokenizedField{name: name, path: path}
			tokenizedFields = append(tokenizedFields, tf)
		}

		if field.Filter != nil {
			jsonSchema.AddProperty(name, *field.Filter, true)
		}
	}

	sch, err := json.Marshal(jsonSchema)
	if err != nil {
		return nil, fmt.Errorf("error marshalling schema: %w", err)
	}

	schema, err := jsonschema.CompileString(ind.ID, string(sch))
	if err != nil {
		return nil, fmt.Errorf("error compiling schema: %w", err)
	}

	matched := make([]string, 0, len(vcJWTs))

	for _, vcJWT := range vcJWTs {
		tokensFound := make(map[string]bool, len(tokenizedFields))

		decoded, err := vc.Decode[vc.Claims](vcJWT)
		if err != nil {
			continue
		}

		var jwtPayload map[string]any
		payload, err := base64.RawURLEncoding.DecodeString(decoded.JWT.Parts[1])
		if err != nil {
			continue
		}

		if err := json.Unmarshal(payload, &jwtPayload); err != nil {
			continue
		}

		selectionCandidate := make(map[string]any)
		for _, tf := range tokenizedFields {
			if ok := tokensFound[tf.name]; ok {
				continue
			}

			value, err := jsonpath.Get(tf.path, jwtPayload)
			if err != nil {
				continue
			}

			if value != nil {
				selectionCandidate[tf.name] = value
				tokensFound[tf.name] = true
			}
		}

		if len(selectionCandidate) != len(ind.Constraints.Fields) {
			continue
		}

		if err := schema.Validate(selectionCandidate); err != nil {
			continue
		}

		matched = append(matched, vcJWT)
	}

	return matched, nil
}

// Constraints contains the requirements for a given Input Descriptor.
type Constraints struct {
	Fields []Field `json:"fields,omitempty"`
}

// Field contains the requirements for a given field within a proof
type Field struct {
	ID        string       `json:"id,omitempty"`
	Name      string       `json:"name,omitempty"`
	Path      []string     `json:"path,omitempty"`
	Purpose   string       `json:"purpose,omitempty"`
	Filter    *Filter      `json:"filter,omitempty"`
	Optional  bool         `json:"optional,omitempty"`
	Predicate *Optionality `json:"predicate,omitempty"`
}

// Optionality is a type alias for the possible values of the predicate field
type Optionality string

// Constants for Optionality values
const (
	Required  Optionality = "required"
	Preferred Optionality = "preferred"
)

// Filter is a JSON Schema that is applied against the value of a field.
type Filter struct {
	Type     string  `json:"type,omitempty"`
	Pattern  string  `json:"pattern,omitempty"`
	Const    string  `json:"const,omitempty"`
	Contains *Filter `json:"contains,omitempty"`
}

// JSONSchema represents a minimal JSON Schema
type JSONSchema struct {
	Schema     string            `json:"$schema"`
	Type       string            `json:"type"`
	Properties map[string]Filter `json:"properties"`
	Required   []string          `json:"required"`
}

// AddProperty adds the provided Filter with the provided name to the JsonSchema
func (j *JSONSchema) AddProperty(name string, value Filter, required bool) {
	j.Properties[name] = value

	if required {
		j.Required = append(j.Required, name)
	}
}
