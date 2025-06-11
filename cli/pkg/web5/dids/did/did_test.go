package did_test

import (
	"testing"

	"olares-cli/pkg/web5/dids/did"

	"github.com/alecthomas/assert/v2"
)

type vector struct {
	input  string
	output map[string]interface{}
	error  bool
}

func TestParse(t *testing.T) {
	vectors := []vector{
		{input: "", error: true},
		{input: "did:", error: true},
		{input: "did:uport", error: true},
		{input: "did:uport:", error: true},
		{input: "did:uport:1234_12313***", error: true},
		{input: "2nQtiQG6Cgm1GYTBaaKAgr76uY7iSexUkqX", error: true},
		{input: "did:method:%12%1", error: true},
		{input: "did:method:%1233%Ay", error: true},
		{input: "did:CAP:id", error: true},
		{input: "did:method:id::anotherid%r9", error: true},
		{
			input: "did:example:123456789abcdefghi",
			output: map[string]interface{}{
				"method": "example",
				"id":     "123456789abcdefghi",
				"uri":    "did:example:123456789abcdefghi",
			},
		},
		{
			input: "did:example:123456789abcdefghi;foo=bar;baz=qux",
			output: map[string]interface{}{
				"alternate": "did:example:123456789abcdefghi;baz=qux;foo=bar",
				"method":    "example",
				"id":        "123456789abcdefghi",
				"uri":       "did:example:123456789abcdefghi",
				"params": map[string]string{
					"foo": "bar",
					"baz": "qux",
				},
			},
		},
		{
			input: "did:example:123456789abcdefghi?foo=bar&baz=qux",
			output: map[string]interface{}{
				"method": "example",
				"id":     "123456789abcdefghi",
				"uri":    "did:example:123456789abcdefghi",
				"query":  "foo=bar&baz=qux",
			},
		},
		{
			input: "did:example:123456789abcdefghi#keys-1",
			output: map[string]interface{}{
				"method":   "example",
				"id":       "123456789abcdefghi",
				"uri":      "did:example:123456789abcdefghi",
				"fragment": "keys-1",
			},
		},
		{
			input: "did:example:123456789abcdefghi?foo=bar&baz=qux#keys-1",
			output: map[string]interface{}{
				"method":   "example",
				"id":       "123456789abcdefghi",
				"uri":      "did:example:123456789abcdefghi",
				"query":    "foo=bar&baz=qux",
				"fragment": "keys-1",
			},
		},
		{
			input: "did:example:123456789abcdefghi;foo=bar;baz=qux?p1=v1&p2=v2#keys-1",
			output: map[string]interface{}{
				"alternate": "did:example:123456789abcdefghi;baz=quxfoo=bar;?p1=v1&p2=v2#keys-1",
				"method":    "example",
				"id":        "123456789abcdefghi",
				"uri":       "did:example:123456789abcdefghi",
				"params":    map[string]string{"foo": "bar", "baz": "qux"},
				"query":     "p1=v1&p2=v2",
				"fragment":  "keys-1",
			},
		},
	}

	for _, v := range vectors {
		t.Run(v.input, func(t *testing.T) {
			did, err := did.Parse(v.input)

			if v.error && err == nil {
				t.Errorf("expected error, got nil")
			}

			if err != nil {
				if !v.error {
					t.Errorf("failed to parse did: %s", err.Error())
				}
				return
			}

			// The Params map doesn't have a reliable order, so check both
			alt, ok := v.output["alternate"]
			if ok {
				firstOrder := v.input == did.URL()
				secondOrder := alt == did.URL()
				assert.True(t, firstOrder || secondOrder, "expected one of the orders to match")
			} else {
				assert.Equal[interface{}](t, v.input, did.URL())
			}
			assert.Equal[interface{}](t, v.output["method"], did.Method)
			assert.Equal[interface{}](t, v.output["id"], did.ID)
			assert.Equal[interface{}](t, v.output["uri"], did.URI)

			if v.output["params"] != nil {
				params, ok := v.output["params"].(map[string]string)
				assert.True(t, ok, "expected params to be map[string]string")

				for k, v := range params {
					assert.Equal[interface{}](t, v, did.Params[k])
				}
			}

			if v.output["query"] != nil {
				assert.Equal[interface{}](t, v.output["query"], did.Query)
			}

			if v.output["fragment"] != nil {
				assert.Equal[interface{}](t, v.output["fragment"], did.Fragment)
			}
		})
	}
}

func TestDID_ScanValueRoundtrip(t *testing.T) {
	tests := []struct {
		object  did.DID
		raw     string
		alt     string
		wantErr bool
	}{
		{
			raw:    "did:example:123456789abcdefghi",
			object: did.MustParse("did:example:123456789abcdefghi"),
		},
		{
			raw:    "did:example:123456789abcdefghi;foo=bar;baz=qux",
			alt:    "did:example:123456789abcdefghi;baz=qux;foo=bar",
			object: did.MustParse("did:example:123456789abcdefghi;foo=bar;baz=qux"),
		},
		{
			raw:    "did:example:123456789abcdefghi?foo=bar&baz=qux",
			object: did.MustParse("did:example:123456789abcdefghi?foo=bar&baz=qux"),
		},
		{
			raw:    "did:example:123456789abcdefghi#keys-1",
			object: did.MustParse("did:example:123456789abcdefghi#keys-1"),
		},
		{
			raw:    "did:example:123456789abcdefghi?foo=bar&baz=qux#keys-1",
			object: did.MustParse("did:example:123456789abcdefghi?foo=bar&baz=qux#keys-1"),
		},
		{
			raw:    "did:example:123456789abcdefghi;foo=bar;baz=qux?foo=bar&baz=qux#keys-1",
			alt:    "did:example:123456789abcdefghi;baz=qux;foo=bar?foo=bar&baz=qux#keys-1",
			object: did.MustParse("did:example:123456789abcdefghi;foo=bar;baz=qux?foo=bar&baz=qux#keys-1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			var d did.DID
			if err := d.Scan(tt.raw); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.object, d)

			value, err := d.Value()
			assert.NoError(t, err)
			actual, ok := value.(string)
			assert.True(t, ok)
			if tt.alt != "" {
				assert.True(t, actual == tt.raw || actual == tt.alt)
			} else {
				assert.Equal(t, tt.raw, actual)
			}
		})
	}
}
