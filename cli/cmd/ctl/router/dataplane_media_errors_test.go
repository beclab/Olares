package router

import (
	"strings"
	"testing"
)

// Every refusal the creative surface can answer with has a sentence of its own.
// The four field codes are the reason the canonical contract exists: before it,
// all four were one outcome — the field was forwarded to a provider that
// ignored it and the call was billed — so answering them with one message
// would put the caller back where they started.
func TestEveryCreativeRefusalNamesSomethingToChange(t *testing.T) {
	codes := []string{
		"media_field_unknown", "media_field_not_allowed", "media_field_conflict",
		"media_field_invalid", "media_input_unsupported", "media_input_required",
		"media_input_invalid", "media_input_too_large", "media_payload_too_large",
		"media_prompt_required", "media_operation_invalid", "media_operation_not_declared",
		"media_generation_unsupported_for_provider", "media_mode_unsupported",
		"media_model_required", "media_output_not_found", "media_generation_not_found",
		"media_content_unavailable", "image_generation_async_multiple_unsupported",
		"image_generation_async_required",
	}
	seen := make(map[string]string, len(codes))
	for _, code := range codes {
		advice := mediaAdvice(&RouterError{Status: 400, Code: code})
		if advice == "" {
			t.Errorf("%s has no advice", code)
			continue
		}
		if other, repeated := seen[advice]; repeated {
			t.Errorf("%s and %s are answered with the same sentence", code, other)
		}
		seen[advice] = code
	}
	if mediaAdvice(&RouterError{Code: "model_not_found"}) != "" {
		t.Error("a non-creative code was claimed by the creative table")
	}
}

// The four field codes reach the caller through callErr, and two of them end in
// a suffix callErr already matched for audio. A media family answered with
// audio's advice is pointed at the wrong list of engine images.
func TestCreativeAdviceWinsOverTheAudioSuffixMatch(t *testing.T) {
	err := callErr(&RouterError{
		Status: 422, Type: "invalid_request_error",
		Code:    "media_generation_unsupported_for_provider",
		Message: "provider does not expose music generation",
	})
	msg := err.Error()
	if !strings.Contains(msg, "does not serve this kind of generation") {
		t.Errorf("got %q", msg)
	}
	if strings.Contains(msg, "--mode audio") {
		t.Error("a music refusal was explained as an audio engine mismatch")
	}
}

// Router refuses most of these before it has spoken to a provider, and the
// status branches below the table would say the model's endpoint never
// answered. A code this build has no sentence for still has to read as Router's
// own refusal.
func TestAnUnknownCreativeCodeIsStillReportedAsRoutersOwnRefusal(t *testing.T) {
	msg := callErr(&RouterError{Status: 422, Code: "media_something_new"}).Error()
	if !strings.Contains(msg, "before it reached the provider") {
		t.Errorf("got %q", msg)
	}
}

// Router words a refusal in the keys the caller actually sent — a released
// route's caller is told `reference_images`, not `inputs.images`. Advice that
// restated the field in canonical names would undo that on the way out.
func TestAdviceDoesNotRespellTheFieldRouterAlreadyNamed(t *testing.T) {
	for _, code := range []string{
		"media_field_unknown", "media_field_not_allowed", "media_field_invalid",
		"media_input_unsupported", "media_input_required",
	} {
		advice := mediaAdvice(&RouterError{Code: code})
		for _, canonical := range []string{"inputs.", "output.", "music."} {
			if strings.Contains(advice, canonical) {
				t.Errorf("%s: advice spells %q, which the caller may not have written", code, canonical)
			}
		}
	}
}
