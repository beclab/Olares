/**
 * Error codes used within Padloc
 */
export enum ErrorCode {
	// Crypto Errors
	INVALID_ENCRYPTION_PARAMS = 'invalid_encryption_params',
	DECRYPTION_FAILED = 'decryption_failed',
	ENCRYPTION_FAILED = 'encryption_failed',
	NOT_SUPPORTED = 'not_supported',
	MISSING_ACCESS = 'missing_access',
	VERIFICATION_ERROR = 'verification_error',

	// Client Errors
	FAILED_CONNECTION = 'failed_connection',
	UNEXPECTED_REDIRECT = 'unexpected_redirect',

	// Server Errors
	BAD_REQUEST = 'bad_request',
	INVALID_SESSION = 'invalid_session',
	SESSION_EXPIRED = 'session_expired',
	INSUFFICIENT_PERMISSIONS = 'insufficient_permissions',
	INVALID_CREDENTIALS = 'invalid_credentials',
	ACCOUNT_EXISTS = 'account_exists',
	INVALID_RESPONSE = 'invalid_response',
	INVALID_REQUEST = 'invalid_request',
	OUTDATED_REVISION = 'merge_conflict',
	MAX_REQUEST_SIZE_EXCEEDED = 'max_request_size_exceeded',
	MAX_REQUEST_AGE_EXCEEDED = 'max_request_age_exceeded',
	JWS_INVALID = 'jws_invalid',

	// Quota errors
	// ORG_FROZEN = "org_frozen",
	// ORG_QUOTA_EXCEEDED = "org_quota_exceeded",
	// MEMBER_QUOTA_EXCEEDED = "member_quota_exceeded",
	// GROUP_QUOTA_EXCEEDED = "group_quota_exceeded",
	// VAULT_QUOTA_EXCEEDED = "vault_quota_exceeded",
	// STORAGE_QUOTA_EXCEEDED = "storage_quota_exceeded",

	// Provisioning Errors
	PROVISIONING_QUOTA_EXCEEDED = 'provisioning_quota_exceeded',
	PROVISIONING_NOT_ALLOWED = 'provisioning_not_allowed',

	// Generic Errors
	CLIENT_ERROR = 'client_error',
	SERVER_ERROR = 'server_error',
	UNKNOWN_ERROR = 'unknown_error',

	// Encoding errors
	ENCODING_ERROR = 'encoding_error',
	UNSUPPORTED_VERSION = 'unsupported_version',

	NOT_FOUND = 'not_found',
	INVALID_CSV = 'invalid_csv',
	INVALID_1PUX = 'invalid_1pux',
	INVALID_BITWARDEN = 'invalid_bitwarden',

	BILLING_ERROR = 'billing_error',

	// MFA Errors
	AUTHENTICATION_REQUIRED = 'email_verification_required',
	AUTHENTICATION_FAILED = 'email_verification_failed',
	JWS_AUTH_FAILED = 'jws_auth_failed',
	AUTHENTICATION_TRIES_EXCEEDED = 'email_verification_tries_exceeded',

	SERVER_NOT_EXIST = 'server_not_exist',
	SERVER_AUTH_FAILED = 'server_auth_failed',
	TOKE_INVILID = 'token_invalid',
    DOMAIN_RESOLUTION_ERROR = 'domain_resolution_error'
}

export interface ErrorOptions {
	report?: boolean;
	display?: boolean;
	status?: number;
	error?: Error;
}

/**
 * Custom error class augmenting the built-in `Error` with some additional properties
 */
export class Err extends Error {
	/** Error code used for more precise error segmentation */
	code: ErrorCode;
	/** Wether or not this error should be reported to an admin, if that option exists */
	report: boolean;
	/** Wether or not this error should be displayed to the user */
	display: boolean;
	/** The original error, if available */
	originalError?: Error;
	/** Time when the error was created */
	time = new Date();

	constructor(
		code: ErrorCode,
		message?: string,
		{ report = false, display = false, error }: ErrorOptions = {}
	) {
		super(message || (error && error.message) || '');
		this.code = code;
		this.report = report;
		this.display = display;
		this.originalError = error;
	}

	toRaw() {
		return {
			code: this.code,
			message: this.message,
			stack: this.originalError ? this.originalError.stack : this.stack
		};
	}

	toString() {
		return `Time: ${this.time.toISOString()}
Error Code: ${this.code}
Error Message: ${this.message}
Stack Trace:\n${this.originalError ? this.originalError.stack : this.stack}
`;
	}
}
