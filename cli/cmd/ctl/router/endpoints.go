package router

// Every route this tree reaches, named once.
//
// Three surfaces answer on two hosts, and telling them apart matters more than
// it looks. Router's console plane and data plane share the entrance this tree
// discovers, but not the credential: the console plane runs on the profile's
// session, and /v1 rejects a session by design (see dataplane.go). The Model
// Console is a third surface on a *different* entrance — the model
// application's own — reached over the same session.
//
// Collecting them here is not tidiness. A path spelled inline is a path
// nothing can check: the backend moves a route, and the verb that used it
// fails at the one moment somebody needed it, with a 404 that reads like the
// application is missing rather than like this file is out of date. One
// declaration per route means a rename is a compile-time edit in a single
// place, and `endpoints_test.go` can hold the rest of the package to it.
//
// Paths carry no query string. A caller that needs one passes url.Values to
// withQuery, so the escaping rules live here too rather than being
// reimplemented next to each request.

import (
	"net/url"
	"strconv"
	"strings"
)

// Prefixes on the entrance host. The console plane is the management surface
// every configuration verb uses; the data plane is OpenAI-shaped and takes an
// `sk-*` key rather than the session this tree carries.
const (
	consoleAPI   = "/console/api"
	dataPlaneAPI = "/v1"
)

// isDataPlanePath reports whether a path this package already built addresses
// the data plane. It exists so that code deciding something about a request in
// hand — which credential paid for it, what a retry would cost — does not have
// to name the prefix itself, which is how a second spelling of it starts.
func isDataPlanePath(path string) bool {
	return strings.HasPrefix(path, dataPlaneAPI+"/")
}

// withQuery appends a query string when there is one. Called with empty values
// it returns the path untouched, so a caller assembling optional filters does
// not have to decide whether it ended up with any.
func withQuery(path string, q url.Values) string {
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}

// Identity. Admin-only, and read by no verb of its own: it is how `--user` and
// `--for-user` turn a name into the id the other routes take.
const epUsers = consoleAPI + "/users"

// API keys.
const epAPIKeys = consoleAPI + "/api-keys"

func epAPIKey(id string) string { return epAPIKeys + "/" + url.PathEscape(id) }

// Providers, and the per-provider verbs.
const epProviders = consoleAPI + "/providers"

func epProvider(id string) string { return epProviders + "/" + url.PathEscape(id) }

func epProviderCredentialsForm(id string) string { return epProvider(id) + "/credentials-form" }

func epProviderSyncModels(id string) string { return epProvider(id) + "/sync-models" }

func epProviderValidate(id string) string { return epProvider(id) + "/validate" }

// The same probe for a provider that does not exist yet. A hyphen rather than
// /providers/validate because a static segment cannot sit beside the
// /providers/:id wildcard — the same reason the two catalog routes are spelled
// that way.
const epProviderValidateDraft = consoleAPI + "/provider-validate"

// The app directory's cache. It answers whether an application is on this
// Olares at all, which no other route does: the model-app list says what may be
// installed, the provider list says what Router can call, and the caller_app
// dimension of the spend summary only knows an application that has already
// billed something. Readable by any authenticated console user, unlike almost
// everything else under /console/api.
const epInstalledApps = consoleAPI + "/installed-apps"

func epProviderCredentialHistory(id string) string { return epProvider(id) + "/credential-history" }

func epProviderRollback(id string, version int) string {
	return epProvider(id) + "/rollback/" + strconv.Itoa(version)
}

func epProviderPredefinedModels(id string) string { return epProvider(id) + "/predefined-models" }

func epProviderCustomizableModels(id string) string {
	return epProvider(id) + "/customizable-models"
}

func epProviderModel(providerID, modelID string) string {
	return epProvider(providerID) + "/models/" + url.PathEscape(modelID)
}

// The model card of a model application, reached through Router rather than at
// the application itself.
//
// Both routes address the app by the `model` string a caller would send to
// /v1 — it carries a slash, which is why it rides a query parameter — and both
// are admin-only. The Model Console's own equivalents are the epLocal* routes
// below; the difference between the two paths is not the host but the
// semantics, and `spec.go` is where that is written down.
const (
	epModelSpec     = consoleAPI + "/model-spec"
	epEngineRestart = consoleAPI + "/engine/restart"
)

// epForModel is the query every model-addressed console route takes.
func epForModel(path, model string) string {
	q := url.Values{}
	q.Set("model", model)
	return withQuery(path, q)
}

// Catalogs. epProviderModels is the aggregate model list across every
// provider; the other two describe what *could* be configured rather than what
// is.
const (
	epProviderModels    = consoleAPI + "/provider-models"
	epProviderCatalog   = consoleAPI + "/provider-catalog"
	epPredefinedCatalog = consoleAPI + "/predefined-catalog"
)

// The model applications the Market publishes, each row carrying what Router
// knows about the copy installed here.
//
// Installing, cloning, upgrading and removing an application are `olares-cli
// market`, and Router has no route for any of them. This is read for one thing
// only — the provider id of an application that is still installing, which no
// other list names yet.
const epModelApps = consoleAPI + "/model-apps"

// Named routes: every name a caller may send in `model` that is not a
// qualified `<provider>/<model>` reference. Aliases, groups and the system's
// default categories are rows of one table, so they are one route here too —
// there is no separate default-models surface, and asking for one gets a 404
// that reads like the deployment is broken.
const epModelRoutes = consoleAPI + "/model-routes"

func epModelRoute(id string) string { return epModelRoutes + "/" + url.PathEscape(id) }

func epModelRouteMember(routeID, modelID string) string {
	return epModelRoute(routeID) + "/members/" + url.PathEscape(modelID)
}

// What a default category answers with, when an administrator rather than
// reconciliation decides it.
//
// A subresource rather than a field on the route itself, and the same on this
// side as on Router's: pinning a category stops reconciliation maintaining it,
// which is not something a rename-and-enable patch has anything to say about.
// The candidates read is the other half — the categories select on capability,
// so which models one would accept is a question only Router can answer.
func epModelRouteTarget(routeID string) string { return epModelRoute(routeID) + "/target" }

func epModelRouteCandidates(routeID string) string { return epModelRoute(routeID) + "/candidates" }

// Spend: what was called, what it cost, the same rows as a download, and how
// long the per-call rows are kept.
const (
	epSpendLogs      = consoleAPI + "/spend-logs"
	epSpendSummary   = epSpendLogs + "/summary"
	epSpendExportCSV = epSpendLogs + "/export.csv"
	epSpendSettings  = consoleAPI + "/spend-settings"
)

// Who changed Router, and to what.
const epAuditLogs = consoleAPI + "/audit-logs"

func epAuditLog(id string) string { return epAuditLogs + "/" + url.PathEscape(id) }

// Ceilings on a key, a person, or a model.
const epQuotas = consoleAPI + "/quotas"

func epQuota(id int64) string { return epQuotas + "/" + strconv.FormatInt(id, 10) }

// There is no trace surface. Router accepted OTLP spans and served them back
// per person for a while; the tables were dropped and the routes withdrawn,
// because a spend row already carries the model, tokens, cost, latency, status
// and failure reason, and keeping request bodies to add to that bought
// compliance exposure rather than insight.

// Data plane. These take an `sk-*` bearer or a platform-injected caller
// identity, never the console session.
//
// Three shapes live here and behave differently enough to be worth naming. Most
// routes answer with the result. The two media routes can answer with a receipt
// instead, and the thing generated is then fetched from a `/content` route
// afterwards. OCR only ever answers with a receipt.
const (
	epChatCompletions = dataPlaneAPI + "/chat/completions"
	// One path, two protocols: POST is the whole answer, and a GET carrying an
	// Upgrade is the same call with the answer arriving as it is written. A
	// GET without the upgrade is refused with 426 rather than answered.
	epResponses       = dataPlaneAPI + "/responses"
	epEmbeddings      = dataPlaneAPI + "/embeddings"
	epRerank          = dataPlaneAPI + "/rerank"
	epDataPlaneModels = dataPlaneAPI + "/models"
	epSearch          = dataPlaneAPI + "/search"
	epScrape          = dataPlaneAPI + "/scrape"
)

// The Anthropic-shaped ingress. Router mounts it beside the OpenAI one so a
// client built for Claude reaches the same models over the same key, and both
// shapes share one dispatch and one spend path.
//
// Counting is the half worth having a verb for, and it is mounted apart from
// everything else: it sits above the quota line and records no spend, because
// it asks how large a turn would be rather than sending one.
const epMessagesCountTokens = dataPlaneAPI + "/messages/count_tokens"

// Audio is one catch-all upstream, so every suffix here reaches the sibling
// audio engine unchanged. Which suffixes exist depends on the engine behind the
// model: recognition, streaming recognition, synthesis, voice cloning,
// dialogue, sound effects, voice activity, diarization, streaming diarization,
// speaker embedding, enhancement and alignment are separate engine images, and
// a model that does one answers 404 or refuses the mode for the others.
const (
	epAudioTranscriptions    = dataPlaneAPI + "/audio/transcriptions"
	epAudioTranslations      = dataPlaneAPI + "/audio/translations"
	epAudioSpeech            = dataPlaneAPI + "/audio/speech"
	epAudioSpeechClone       = epAudioSpeech + "/clone"
	epAudioVoices            = dataPlaneAPI + "/audio/voices"
	epAudioVAD               = dataPlaneAPI + "/audio/vad"
	epAudioSpeakerEmbeddings = dataPlaneAPI + "/audio/embeddings"
	// Spelled in full. The engine serves `diarization` over HTTP and reserves
	// `diarize/stream` for the WebSocket, so the short form is a 404.
	epAudioDiarization = dataPlaneAPI + "/audio/diarization"
	epAudioEnhance     = dataPlaneAPI + "/audio/enhance"
	epAudioAlign       = dataPlaneAPI + "/audio/align"
)

// The audio WebSocket routes wrapped by CLI commands. Router's full protocol
// surface has three: these two input streams plus /v1/audio/speech/stream for
// streaming TTS. `call speak` uses the HTTP synthesis dialects, so it does not
// need a third socket constant here. These are separate constants rather than
// suffixes on the HTTP routes because a typo would silently arrive as a POST.
const (
	epAudioStreamWS        = dataPlaneAPI + "/audio/stream"
	epAudioDiarizeStreamWS = dataPlaneAPI + "/audio/diarize/stream"
)

// Audio tasks. Batch HTTP operations whose operation catalogue declares async
// support can answer `--async` with a receipt instead of a result; WebSocket
// and HTTP chunked streams cannot. These commands read the receipt.
//
// /v1/tasks is the canonical prefix and /v1/audio/tasks is the alias Router
// retains for the clients that were written before it. This tree stayed on the
// alias while the canonical side was still settling; it no longer is, and the
// receipt Router writes now names the canonical one in its own `poll` and
// `result_url`. Following a receipt to a path it does not name is how a client
// ends up being the reason an alias cannot be retired.
//
// A receipt is not always in hand — an id can be pasted from a terminal a day
// later — so these build the same paths from an id alone.
const epTasks = dataPlaneAPI + "/tasks"

func epTask(id string) string { return epTasks + "/" + url.PathEscape(id) }

func epTaskResult(id string) string { return epTask(id) + "/result" }

// onDataPlane reports whether a path handed back in a response addresses the
// data plane. A receipt names its own follow-up routes and Router writes them
// relative, so anything else is not Router redirecting a client — it is a
// response steering one, and the id alone already reaches the task.
func onDataPlane(p string) bool { return strings.HasPrefix(p, dataPlaneAPI+"/") }

// The voice library and the log of what has been read out.
//
// These sit at the root of /v1 rather than under /audio because they are the
// ElevenLabs shape, which a synthesis engine serves alongside the OpenAI one.
// It is not merely a second spelling: these routes expose durable voices and,
// on ElevenLabs-shaped engines, synthesis history with retained audio. The
// OpenAI-shaped /v1/audio/speech contract does not itself promise history.
//
// Which default each reaches is Router's decision and it is not uniform:
// reading or editing the voice table is default-tts, creating a voice from a
// recording is default-tts-clone, and creating one from a description is
// default-tts-design. History reaches no default at all, so a history verb has
// to name a model.
const (
	epVoices            = dataPlaneAPI + "/voices"
	epVoicesAdd         = epVoices + "/add"
	epVoiceSettings     = epVoices + "/settings/default"
	epTextToVoice       = dataPlaneAPI + "/text-to-voice"
	epTextToVoiceDesign = epTextToVoice + "/design"
	epTextToSpeech      = dataPlaneAPI + "/text-to-speech"
	epHistory           = dataPlaneAPI + "/history"
)

func epVoice(id string) string { return epVoices + "/" + url.PathEscape(id) }

// epSpeakAs is /v1/audio/speech in the other dialect, where the voice is the
// address rather than a field. Router forwards both without translating
// between them for a locally installed engine, so which one answers is a
// property of the image the model application was built from.
func epSpeakAs(voice string) string { return epTextToSpeech + "/" + url.PathEscape(voice) }

func epVoiceOwnSettings(id string) string { return epVoice(id) + "/settings" }

func epHistoryItem(id string) string { return epHistory + "/" + url.PathEscape(id) }

func epHistoryAudio(id string) string { return epHistoryItem(id) + "/audio" }

// Images and video. A generation is a row Router keeps, so it can be asked
// about after the request that started it has gone, and the bytes come from the
// `/content` route rather than the record — a video is not something to carry
// through a JSON field.
const (
	epImageGenerations = dataPlaneAPI + "/images/generations"
	epVideos           = dataPlaneAPI + "/videos"
)

func epImageGeneration(id string) string {
	return epImageGenerations + "/" + url.PathEscape(id)
}

func epImageGenerationContent(id string) string {
	return epImageGeneration(id) + "/content"
}

func epVideo(id string) string { return epVideos + "/" + url.PathEscape(id) }

func epVideoContent(id string) string { return epVideo(id) + "/content" }

// The unified media route. It creates into the same table the two routes above
// do and answers with the same record, so the difference is the request rather
// than the resource: this one takes the canonical body as written, and it is the
// only route music and 3D have.
//
// It is always a receipt. There is no synchronous branch, which is why an image
// stays on the released route -- a provider that keeps no generations answers
// there with the picture itself, and here it cannot be served at all.
const epGenerations = dataPlaneAPI + "/generations"

func epGeneration(id string) string { return epGenerations + "/" + url.PathEscape(id) }

func epGenerationContent(id string) string { return epGeneration(id) + "/content" }

// Music. The same table and the same record as the unified route above, on a
// surface of its own, because a track is asked for in a shape the canonical
// body has no room for: a flat request with a title and words, and two things
// that happen before any audio is generated at all.
//
// A format pass runs a model's own language model over a caption and lyrics and
// hands back what it would actually sing, so the words can be confirmed before
// the minutes of compute; a draft writes both from one sentence. Neither is a
// generation to download, and both are asynchronous, which is why they are
// resources rather than fields.
//
// Cancel is here and nowhere else in this tree: a track is the one family whose
// upstream can be stopped mid-run and settled at what it used.
const (
	epMusicGenerations = dataPlaneAPI + "/music/generations"
	epMusicFormats     = dataPlaneAPI + "/music/formats"
	epMusicDrafts      = dataPlaneAPI + "/music/drafts"
)

func epMusicGeneration(id string) string { return epMusicGenerations + "/" + url.PathEscape(id) }

func epMusicGenerationContent(id string) string { return epMusicGeneration(id) + "/content" }

func epMusicLyricsAlignment(id string) string { return epMusicGeneration(id) + "/lyrics-alignment" }

func epMusicFormat(id string) string { return epMusicFormats + "/" + url.PathEscape(id) }

func epMusicDraft(id string) string { return epMusicDrafts + "/" + url.PathEscape(id) }

// Translate mirrors the upstream's own service-root names under /v1. These five
// carry no model field: each resolves the translate default per call, so there
// is nothing for a caller to name and nothing to get wrong.
//
// The transcript route is the one that is not MTran-compatible. The other four
// translate a text at a time; this one takes a stretch of dialogue, lets the
// model read the turns around each line, and answers one result per turn under
// the ids it was given.
const (
	epTranslate           = dataPlaneAPI + "/translate"
	epTranslateBatch      = epTranslate + "/batch"
	epTranslateTranscript = epTranslate + "/transcript"
	epLanguages           = dataPlaneAPI + "/languages"
	epDetect              = dataPlaneAPI + "/detect"
)

// OCR. The prefix is /v1/ocr rather than the upstream's bare /v1 because the
// engine's list-models route is /v1/models, which already names the catalogue.
//
// Router also mounts /v1/ocr/models, and it is deliberately absent here: it
// answers with the one model the engine behind the request was deployed with,
// which `router model list --mode ocr` says for every OCR model at once.
const (
	epOCR      = dataPlaneAPI + "/ocr"
	epOCRTasks = epOCR + "/tasks"
)

func epOCRTask(id string) string { return epOCRTasks + "/" + url.PathEscape(id) }

func epOCRTaskResult(id string) string { return epOCRTask(id) + "/result" }

// The Model Console inside a model application: a different host, addressed
// through the same session. Router serves a /healthz of its own on its own
// entrance, and no verb here reads it — an unreachable Router is reported by
// whichever verb was trying to reach it.
const (
	epHealth             = "/healthz"
	epLocalBuildInfo     = "/api/build-info"
	epLocalProgress      = "/api/progress"
	epLocalModelSpec     = "/api/model-spec"
	epLocalModelSpecFile = epLocalModelSpec + "/file"
	epLocalConfig        = "/api/config"
	epLocalEndpoints     = "/api/endpoints"
	epLocalDiagGPU       = "/api/diag/gpu"
	epLocalRetry         = "/api/retry"
	epLocalEngineRestart = "/api/engine/restart"
)
