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

// Ops. Unauthenticated, and served by Router itself rather than by the console
// plane, which is why `router status` can report a reachable-but-unauthorized
// deployment instead of one blanket failure.
const epHealth = "/healthz"

// Identity. epMe is the one console route any user may read; everything else on
// this plane is admin-only.
const (
	epMe    = consoleAPI + "/users/me"
	epUsers = consoleAPI + "/users"
)

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

// Catalogs. epProviderModels is the aggregate model list across every
// provider; the other two describe what *could* be configured rather than what
// is.
const (
	epProviderModels       = consoleAPI + "/provider-models"
	epProviderCatalog      = consoleAPI + "/provider-catalog"
	epPredefinedCatalog    = consoleAPI + "/predefined-catalog"
	epCapabilitiesSupports = consoleAPI + "/capabilities/supports"
)

// Market: the model applications Router can install on this machine, and the
// lifecycle verbs on one already installed.
//
// The two collection routes sit directly under /market and the two that address
// one row under /market/providers/:id, so a catalog can never occupy the same
// path position as an id. There is no route for progress: a lifecycle POST
// answers with the provider row to watch, and the app directory keeps that
// row's olares_status current.
const (
	epMarketCatalog = consoleAPI + "/market/catalog"
	epMarketInstall = consoleAPI + "/market/install"
)

func epMarketProviderAction(providerID, action string) string {
	return consoleAPI + "/market/providers/" + url.PathEscape(providerID) + "/" + action
}

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

// What is installed on this Olares, from the app directory's cache. Open to any
// console session, and read-only in both directions: the directory is the only
// writer, and installing is the Market's decision.
//
// There is no route for the applications that call Router, and none for
// archiving one. An appid is the platform's own identity for an app rather than
// something Router issues, so there is no row to create and nothing to revoke;
// what an app may spend is a quota on the appid, and what it has spent is the
// caller_app dimension of the spend summary.
const epInstalledApps = consoleAPI + "/installed-apps"

// Spend: what was called, what it cost, and the same rows as a download.
const (
	epSpendLogs      = consoleAPI + "/spend-logs"
	epSpendSummary   = epSpendLogs + "/summary"
	epSpendExportCSV = epSpendLogs + "/export.csv"
)

// Who changed Router, and to what.
const epAuditLogs = consoleAPI + "/audit-logs"

func epAuditLog(id string) string { return epAuditLogs + "/" + url.PathEscape(id) }

// Ceilings on a key, a person, or a model.
const epQuotas = consoleAPI + "/quotas"

func epQuota(id int64) string { return epQuotas + "/" + strconv.FormatInt(id, 10) }

// The spans an agent framework reported for a call.
const (
	epTraces      = consoleAPI + "/observability/traces"
	epCapturePref = consoleAPI + "/observability/capture-pref"
)

func epTrace(id string) string { return epTraces + "/" + url.PathEscape(id) }

// Data plane. These take an `sk-*` bearer or a platform-injected caller
// identity, never the console session.
const (
	epChatCompletions     = dataPlaneAPI + "/chat/completions"
	epEmbeddings          = dataPlaneAPI + "/embeddings"
	epDataPlaneModels     = dataPlaneAPI + "/models"
	epAudioTranscriptions = dataPlaneAPI + "/audio/transcriptions"
	epAudioTranslations   = dataPlaneAPI + "/audio/translations"
	epAudioSpeech         = dataPlaneAPI + "/audio/speech"
	epOCR                 = dataPlaneAPI + "/ocr"
	epOCRTasks            = epOCR + "/tasks"
)

func epOCRTask(id string) string { return epOCRTasks + "/" + url.PathEscape(id) }

// The Model Console inside a model application: a different host, addressed
// through the same session. Its /healthz is epHealth, which both surfaces
// serve on their own entrance.
const (
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
