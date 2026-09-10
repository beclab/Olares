package clierr

// The codes a caller may branch on.
//
// This is a vocabulary, not a registry: nothing here enumerates every
// code the CLI can emit, because each subtree classifies its own
// failures and Router's codes come off the wire. What it does is stop
// the same failure from being named two ways in two trees, which is the
// only thing that makes a `.error.code` branch unwritable.
//
// The names come from pkg/dashboard's ErrorKind, which arrived at this
// split first and has been in the JSON envelope of every dashboard verb
// since. `timeout`, `transport`, `decode`, `http_4xx` and `http_5xx`
// are spelled identically there on purpose.
//
// The auth family is finer-grained than dashboard's single `auth`
// because the recoveries genuinely differ, and a code whose recovery a
// caller still has to read the message to find is not doing its job.
// The shared prefix is what lets a caller that only wants "this is a
// credentials problem" match on it.
const (
	// CodeTimeout is a deadline the CLI itself imposed, not an upstream
	// one. Retrying is meaningful; retrying unchanged usually is not.
	CodeTimeout = "timeout"

	// CodeTransport is a failure to reach the far side at all.
	CodeTransport = "transport"

	// CodeDecode is a response that arrived and was not what its
	// content type promised.
	CodeDecode = "decode"

	// CodeHTTP4xx and CodeHTTP5xx are for upstream statuses a tree has
	// no finer name for.
	CodeHTTP4xx = "http_4xx"
	CodeHTTP5xx = "http_5xx"

	// CodeAuthNoProfile: no profile is configured at all. `profile
	// login` or `profile import` -- there is nothing to repair yet.
	CodeAuthNoProfile = "auth_no_profile"

	// CodeAuthNotLoggedIn: a profile exists with no usable token. For a
	// platform-issued credential the recovery is to repair the
	// application that requested it, not to log in.
	CodeAuthNotLoggedIn = "auth_not_logged_in"

	// CodeAuthTokenExpired: the access token aged out. Almost never
	// reaches a caller -- the refreshing transport rotates it first.
	CodeAuthTokenExpired = "auth_token_expired"

	// CodeAuthTokenInvalidated: the grant itself was rejected upstream.
	// Distinct from expiry because no amount of retrying or refreshing
	// recovers it.
	CodeAuthTokenInvalidated = "auth_token_invalidated"
)
