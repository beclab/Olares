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
)

// Prefixes on the entrance host. The console plane is the management surface
// every configuration verb uses; the data plane is OpenAI-shaped and takes an
// `sk-*` key rather than the session this tree carries.
const (
	consoleAPI   = "/console/api"
	dataPlaneAPI = "/v1"
)

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

// Market: the model applications this Olares publishes, each row carrying what
// Router knows about the copy installed here.
//
// Router mounts lifecycle routes beside this one, and the CLI does not use them:
// installing, cloning, upgrading and removing an application are `olares-cli
// market`. This is read for one thing only — the provider id of an application
// that is still installing, which no other list names yet.
const epMarketCatalog = consoleAPI + "/market/catalog"

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
	epEmbeddings      = dataPlaneAPI + "/embeddings"
	epRerank          = dataPlaneAPI + "/rerank"
	epDataPlaneModels = dataPlaneAPI + "/models"
	epSearch          = dataPlaneAPI + "/search"
	epScrape          = dataPlaneAPI + "/scrape"
)

// Audio is one catch-all upstream, so every suffix here reaches the sibling
// audio engine unchanged. Which suffixes exist depends on the engine behind the
// model: recognition, synthesis, voice activity, diarization, enhancement and
// alignment are separate engine images, and a model that does one answers 404
// or refuses the mode for the others.
const (
	epAudioTranscriptions = dataPlaneAPI + "/audio/transcriptions"
	epAudioTranslations   = dataPlaneAPI + "/audio/translations"
	epAudioSpeech         = dataPlaneAPI + "/audio/speech"
	epAudioVoices         = dataPlaneAPI + "/audio/voices"
	epAudioVAD            = dataPlaneAPI + "/audio/vad"
	// Spelled in full. The engine serves `diarization` over HTTP and reserves
	// `diarize/stream` for the WebSocket, so the short form is a 404.
	epAudioDiarization = dataPlaneAPI + "/audio/diarization"
	epAudioEnhance     = dataPlaneAPI + "/audio/enhance"
	epAudioAlign       = dataPlaneAPI + "/audio/align"
)

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

// Translate mirrors the upstream's own service-root names under /v1. These four
// carry no model field: each resolves the translate default per call, so there
// is nothing for a caller to name and nothing to get wrong.
const (
	epTranslate      = dataPlaneAPI + "/translate"
	epTranslateBatch = epTranslate + "/batch"
	epLanguages      = dataPlaneAPI + "/languages"
	epDetect         = dataPlaneAPI + "/detect"
)

// OCR. The prefix is /v1/ocr rather than the upstream's bare /v1 because the
// engine's list-models route is /v1/models, which already names the catalogue.
//
// Router also mounts /v1/ocr/models, and it is deliberately absent here: it
// answers with the one model the engine behind the request was deployed with,
// which `router list --mode ocr` says for every OCR model at once.
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
	epLocalDiagPerf      = "/api/diag/perf"
	epLocalDiagPerfLast  = epLocalDiagPerf + "/last"
	epLocalRetry         = "/api/retry"
	epLocalEngineRestart = "/api/engine/restart"
)
