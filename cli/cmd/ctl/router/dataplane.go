package router

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/internal/keychain"
)

// Reaching the data plane.
//
// Router's /v1 accepts three identities and tries them in this order: an `sk-*`
// Bearer, the platform-injected `x-caller-appid` of a calling application, and
// the platform-injected `X-BFL-USER` of a person. The third arrived in Router
// v2.2.1 (ADR-49), and it is what makes this file short.
//
// Both injected headers are stamped by the Olares edge on the way in — the same
// edge, the same host and the same profile session that every management verb
// in this tree already travels on, since the console plane works precisely
// because that session becomes X-BFL-USER. So a /v1 request sent from here
// carrying no credential of its own is not anonymous: it arrives as whoever the
// platform says is calling. In a pod that is the application; from a laptop it
// is the person holding the profile.
//
// Hence two steps rather than five:
//
//  1. A key named explicitly — flag, then environment — wins. Someone naming a
//     credential is not asking to be second-guessed, and a key remains the way
//     to call with a model allowlist, a budget of its own, or from outside
//     Olares entirely.
//  2. Otherwise send nothing, and let the platform say who this is.
//
// What stood here before was a keychain lookup, a probe gated on running inside
// a container, and — for every laptop, on the first call of any kind — minting
// a key through the management plane and saving it. That key had no expiry, no
// quota and no model restriction, and `router call models` was enough to create
// one. Nothing in this tree issues a credential as a side effect of using one
// any more, which is also why a key saved by an older build is now ignored
// rather than preferred: it still works in Router, and `key current` says so.
//
// A Router older than v2.2.1 refuses the keyless request with
// `missing_credentials`. That is a message to answer rather than a case to mint
// around, and callErr answers it by naming the upgrade and --api-key.

// dataPlaneKeyEnv names a key for one run without saving it anywhere.
const dataPlaneKeyEnv = "OLARES_ROUTER_API_KEY"

// keychainAccountSuffix separates a data-plane key saved by an older build from
// the profile's own access token, which lives under the bare Olares ID in the
// same keychain service. Nothing writes this entry now; `key current` reads it
// so a key left behind can still be found and dealt with.
const keychainAccountSuffix = "#router-api-key"

// authMode records which identity a call ended up presenting.
type authMode string

const (
	authPlatform authMode = "platform"
	authKey      authMode = "key"
)

type dataPlaneAuth struct {
	Mode authMode
	Key  string
}

// dataPlane returns a client for /v1. Choosing the credential reads a flag and
// an environment variable and speaks to nothing, so there is no context to pass
// and no failure to report: without a key the request carries no Authorization
// at all and the platform vouches for the caller.
func dataPlane(pc *preparedClient, explicitKey string) *routerClient {
	auth := resolveDataPlaneAuth(explicitKey)
	if auth.Mode == authPlatform {
		return pc.router
	}
	return pc.router.withHeader("Authorization", "Bearer "+auth.Key)
}

func resolveDataPlaneAuth(explicitKey string) *dataPlaneAuth {
	if k := strings.TrimSpace(explicitKey); k != "" {
		return &dataPlaneAuth{Mode: authKey, Key: k}
	}
	if k := strings.TrimSpace(os.Getenv(dataPlaneKeyEnv)); k != "" {
		return &dataPlaneAuth{Mode: authKey, Key: k}
	}
	return &dataPlaneAuth{Mode: authPlatform}
}

func keychainAccount(olaresID string) string {
	return olaresID + keychainAccountSuffix
}

func cachedDataPlaneKey(olaresID string) (string, error) {
	if strings.TrimSpace(olaresID) == "" {
		return "", nil
	}
	return keychain.Get(keychain.OlaresCliService, keychainAccount(olaresID))
}

func forgetDataPlaneKey(olaresID string) error {
	if strings.TrimSpace(olaresID) == "" {
		return nil
	}
	return keychain.Remove(keychain.OlaresCliService, keychainAccount(olaresID))
}

// dataPlaneKeyFlagUsage is shared by every calling verb so the sentence a user
// reads about credentials is the same one wherever they meet it.
const dataPlaneKeyFlagUsage = "call with this `sk-*` key instead of the one this machine keeps"

// callErr adds what a refused call needs next. The data plane refuses for
// reasons that live in three different places — the credential, the catalogue,
// and the budget — and the message alone rarely says which.
//
// Two credentials can fail one call, and they read almost identically: ours to
// Router, and Router's to the provider. Both come back 401 `invalid_api_key`,
// because Router forwards what the upstream said. The `type` is the only thing
// that separates them — Router's own refusals are `authentication_error`, an
// upstream's carry whatever the upstream calls it — and pointing at the wrong
// key sends someone to re-issue a credential that was never the problem.
// routerErrorOf unwraps to Router's own envelope, or nil when the failure
// happened before anything answered.
func routerErrorOf(err error) *RouterError {
	var re *RouterError
	if err != nil && errors.As(err, &re) {
		return re
	}
	return nil
}

// retryAdvice turns the Retry-After header into a sentence, or into nothing
// when the server did not send one. "Try again later" without a number is
// advice the reader already had.
func retryAdvice(after time.Duration) string {
	if after <= 0 {
		return ""
	}
	return fmt.Sprintf(" Router asks for %s before trying again.", after.Round(time.Second))
}

func callErr(err error) error {
	if err == nil {
		return nil
	}
	re := routerErrorOf(err)
	if re == nil {
		return err
	}
	ours := re.Type == "authentication_error"
	switch {
	case ours && re.Code == "missing_credentials":
		// Router saw no Bearer, no app header and no X-BFL-USER. From here the
		// third one is always sent by the edge, so the Router that says this is
		// one that predates reading it.
		return fmt.Errorf("%w\nThis Router does not yet accept a call without a key: taking the caller's "+
			"identity from the platform arrived in Router v2.2.1, and `olares-cli market upgrade` on the "+
			"Router application is what closes the gap. Until then pass --api-key or set %s; "+
			"`olares-cli router key list` shows the keys that exist and `router key issue <name>` "+
			"creates one", err, dataPlaneKeyEnv)
	case ours && re.Code == "unknown_bfl_user":
		// The platform vouched for a person Router has no row for. Router
		// deliberately does not create one from the data plane, and neither
		// does this: the console plane is where a person comes into existence.
		return fmt.Errorf("%w\nThe platform knows you, Router does not have you yet. It records a person "+
			"the first time they use the console plane, so any management verb — `olares-cli router model "+
			"list` will do — creates the row, and the call then works", err)
	case ours && (re.Code == "invalid_api_key" || re.Code == "key_disabled" || re.Code == "key_expired"):
		return fmt.Errorf("%w\nThe key this call presented is no longer usable. Dropping it is enough to "+
			"keep working: without a key the call goes as you, vouched for by the platform. Unset %s "+
			"if that is where it came from, leave --api-key off, and `olares-cli router key current "+
			"--forget` discards a copy saved by an older build", err, dataPlaneKeyEnv)
	case ours && (re.Code == "owner_disabled" || re.Code == "app_archived" || re.Code == "app_suspended"):
		return fmt.Errorf("%w\nThis is about who or what the key belongs to rather than the key itself. "+
			"For a person, `olares-cli settings users get <name>` shows their account; for an "+
			"application, `olares-cli market list --mine` says whether it is still here", err)
	case re.Code == "no_default_model":
		return fmt.Errorf("%w\nNothing installed can serve that category. `olares-cli router route list --kind default` "+
			"says where each one stands, and it fills once a model of that kind exists — a category is "+
			"maintained against what is configured rather than pointed at by hand. --model names one "+
			"directly in the meantime", err)
	case re.Code == "model_route_disabled":
		return fmt.Errorf("%w\nThe name resolves, but the route serving it is switched off. "+
			"`olares-cli router route get <name>` shows it, and `route enable <name>` puts it back", err)
	case re.Code == "model_not_allowed":
		return fmt.Errorf("%w\nThe credential is restricted to a list this model is not on. "+
			"`olares-cli router call models` shows what it may call, and `router key update` changes the list", err)
	case re.Code == "model_not_found":
		return fmt.Errorf("%w\n`olares-cli router call models` shows the names this credential may send", err)
	case re.Code == "ambiguous_model":
		return fmt.Errorf("%w\nQualify it as <provider>/<model>; `olares-cli router call models` shows the "+
			"qualified names", err)
	case re.Code == "model_at_capacity":
		// Not a quota: nobody set this number and no admin can raise it. The
		// engine was launched to serve so many requests at once, and Router
		// now holds a caller in a soft queue while they are all busy — ten
		// seconds for an interactive call, a minute for a generation. So this
		// code no longer means "full", it means "still full after the wait",
		// which is why the answer is to try again rather than to change a
		// setting, and why it is a 503 with a Retry-After and not the 429 a
		// budget produces.
		return fmt.Errorf("%w\nRouter waited for a slot and the model was still serving every request "+
			"it was launched to handle, so this one was refused rather than held any longer.%s "+
			"`olares-cli router provider get <provider>` shows how wide the engine is and how deep "+
			"its queue was when Router last looked", err, retryAdvice(re.RetryAfter))
	case re.Code == "model_not_ready":
		return fmt.Errorf("%w\nThe model is still coming up.%s `olares-cli router model status <model>` "+
			"follows the phase it is in", err, retryAdvice(re.RetryAfter))
	case mediaAdvice(re) != "":
		// Before the two suffix matches below, which would otherwise answer a
		// media family's "unsupported for provider" with audio's advice.
		return fmt.Errorf("%w\n%s", err, mediaAdvice(re))
	case re.Type == "quota_exceeded_error":
		return fmt.Errorf("%w\n`olares-cli router quota list` shows the limits, and `router usage summary` "+
			"what has been spent against them", err)
	case re.Code == "stream_unsupported_on_fallback_provider":
		// Router only reaches its legacy single-upstream path when no provider
		// claims the model, and complains about streaming from there. The name
		// is what was actually wrong.
		return fmt.Errorf("%w\nNo provider serves this model, which is what put the request on Router's "+
			"fallback path. `olares-cli router call models` shows the names it does serve", err)
	case strings.HasSuffix(re.Code, "_mode_mismatch"):
		return fmt.Errorf("%w\nThe model is configured for a different kind of work than this verb asks "+
			"for. `olares-cli router model list` shows each model's mode, and a route can only serve the mode "+
			"it was created with", err)
	case strings.HasSuffix(re.Code, "_unsupported_for_provider") || re.Code == "audio_path_unsupported":
		return fmt.Errorf("%w\nThis model's provider does not serve that route at all. For a model "+
			"running on this Olares that usually means a different engine image does this job: "+
			"`olares-cli router model list --mode audio` and `router provider get <provider>` show which "+
			"capability each one declares", err)
	case re.Code == "capability_not_supported" || re.Code == "stream_unsupported_for_model":
		return fmt.Errorf("%w\n`olares-cli router provider get <provider>` lists what the model supports", err)
	case re.Status == 404 && re.Code == "" && re.Type == "":
		// A 404 with no envelope on a route Router does mount. Audio, OCR and
		// translate are forwarded to the engine unchanged, and an engine serves
		// the capability it was built for and nothing else, so this is usually
		// the wrong engine rather than a wrong URL.
		return fmt.Errorf("%w\nRouter forwarded this and the model's own endpoint has no such route. "+
			"For audio that means the wrong engine: recognition, synthesis, voice activity, "+
			"diarization, enhancement and alignment are separate images, and each answers 404 for "+
			"the others. Leaving --model off picks a model that declares the capability", err)
	case re.Status == 502 || re.Status == 503 || re.Status == 504:
		return fmt.Errorf("%w\nThe model's own endpoint did not answer. For a model running on this Olares "+
			"that usually means the application is not serving yet: `olares-cli router provider get <provider>` "+
			"shows its Olares status, and a degraded one is a matter for `olares-cli market` rather than "+
			"anything here", err)
	case strings.HasPrefix(re.Code, "media_") || strings.HasPrefix(re.Code, "image_generation_async_"):
		// A media code this build has no sentence for. Naming it as one is
		// better than the two status branches above claiming the provider
		// never answered, which for a refusal made before dispatch is untrue.
		return fmt.Errorf("%w\nRouter refused this before it reached the provider. `olares-cli router "+
			"model get <model>` shows what the model declares", err)
	case re.Status == 401 || re.Status == 403:
		// Last resort, and it has to stay last: Router refuses with these two
		// statuses as well, and every refusal of its own is named above. A
		// status matched ahead of a code sends someone to rotate an upstream
		// credential that was working — which is what this said for
		// model_not_allowed, an allowlist decision Router made by itself.
		return fmt.Errorf("%w\nThis came back from the provider, not from Router: the credential Router "+
			"holds for it was refused. `olares-cli router provider validate <provider>` checks it against "+
			"the upstream, and `router provider update` replaces it", err)
	}
	return err
}

// mediaAdvice is what a refused creative request needs next, or "" for a code
// that is not one of theirs.
//
// The four field codes are the whole point of the canonical contract, and they
// are worth keeping apart. A field Router has no name for, a field that belongs
// to another family, two fields describing the same thing, and a value outside
// its domain are four different things to change; before the contract they were
// one thing — forwarded to a provider that ignored them, and billed.
//
// Nothing here restates which field was refused. Router words these in the
// caller's own spelling — a request on a released route is told
// `reference_images` rather than `inputs.images` — and repeating it in
// canonical names would take that back.
func mediaAdvice(re *RouterError) string {
	switch re.Code {
	case "media_field_unknown":
		return "Router parses a creative body strictly, so a field it has no name for is refused " +
			"rather than passed on and charged for. A vendor's own parameter belongs in " +
			"`--provider-option key=value`, which is forwarded untouched; anything else is a misspelling."
	case "media_field_not_allowed":
		return "The field is real and belongs to another family: a duration describes a video or a " +
			"track, a polygon budget a mesh. Each verb offers only its own family's fields, so this " +
			"usually means one arrived through --provider-option."
	case "media_field_conflict":
		return "Two fields describe the same thing and Router will not guess which was meant. " +
			"--size against --aspect-ratio or --resolution is the usual pair: ask in pixels or in a shape."
	case "media_field_invalid":
		return "The field is right and the value is outside what it accepts. --size is <width>x<height> " +
			"and --aspect-ratio is <w>:<h>; a provider's own vocabulary, \"2K\" or \"auto\", goes through " +
			"--provider-option."
	case "media_input_unsupported":
		return "The request is well formed and this model has no parameter for what it names. Router " +
			"refuses instead of dropping the field and billing for the rest, which is what used to " +
			"happen. `olares-cli router model get <model>` lists what it declares; without that field " +
			"the request runs as it stands."
	case "media_input_required":
		return "The operation works from something, and nothing was given. --image, --audio and " +
			"--source-generation are the three ways to name it, and which one applies follows from " +
			"the operation."
	case "media_prompt_required":
		return "This family is asked in words. Only a mesh can be asked for with a picture alone."
	case "media_input_invalid":
		return "The input reached Router and could not be read as an image or a recording. A file on " +
			"this machine is encoded before it is sent, so this is usually a link that answers with " +
			"something other than the media it names."
	case "media_input_too_large":
		return "One of the inputs is over the cap on a single reference. A file on this machine is " +
			"encoded into the request, which makes it about a third larger than on disk; a link the " +
			"provider can fetch is passed through as written and costs nothing here."
	case "media_payload_too_large":
		return "The request as a whole is over the cap, which is usually several encoded inputs rather " +
			"than one large one. Passing them as links instead leaves the body small."
	case "media_operation_invalid":
		return "That is not an operation this family has. Video is the only one with more than " +
			"generate, which is why --operation is video's alone."
	case "media_operation_not_declared", "video_operation_not_declared", "image_operation_not_declared":
		return "The operation exists and this model does not offer it. `olares-cli router model get " +
			"<model>` lists the ones it declares, and another model of the same mode may serve it."
	case "media_generation_unsupported_for_provider", "image_generation_unsupported_for_provider",
		"video_generation_unsupported_for_provider":
		return "This model's provider does not serve this kind of generation at all. `olares-cli router " +
			"model list --mode <mode>` shows which models do, and for a model running on this Olares " +
			"the mode follows from the application that was installed."
	case "media_mode_unsupported":
		return "The model resolves and is not a creative model: this verb creates a generation, and " +
			"only image, video, music and 3D models produce one. `olares-cli router model list` shows " +
			"each model's mode."
	case "media_model_required":
		return "This route resolves no default, so the model has to be named. `olares-cli router model " +
			"list --mode <mode>` shows the names this credential can send."
	case "media_output_not_found":
		return "--output-id names an output that is not this generation's. The ids are printed once a " +
			"generation has more than one, and leaving the flag off writes the first."
	case "media_generation_not_found":
		return "A generation belongs to the credential that created it and expires on its own, so an id " +
			"that used to work is either past its expiry or being collected with another key. " +
			"--no-wait prints the expiry when the work is submitted."
	case "media_content_unavailable":
		return "The generation exists and its bytes do not: it has not finished, or the provider " +
			"binding it was created against has moved. Collecting it again with --id reports which."
	case "image_generation_async_multiple_unsupported":
		return "A generation is one file behind one content route, so a persisted image is exactly one " +
			"output. Several pictures are several calls."
	case "image_generation_async_required":
		return "This provider serves image generation only as a generation to come back for, which is " +
			"what this verb asks for. Seeing it here means the request reached Router without that " +
			"preference — an older build of this CLI, or another client."
	}
	return ""
}
