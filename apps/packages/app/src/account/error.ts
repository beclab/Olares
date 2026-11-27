export enum ErrorCode {
	INVALID_ENCRYPTION_PARAMS = 'invalid_encryption_params',
	DECRYPTION_FAILED = 'decryption_failed',
	ENCRYPTION_FAILED = 'encryption_failed',
	ENCODING_ERROR = 'encoding_error',
	UNSUPPORTED_VERSION = 'unsupported_version'
}

export class Err extends Error {
	constructor(
		public code: ErrorCode,
		message?: string,
		{ error }: { error?: Error } = {}
	) {
		super(message || error?.message || '');
		this.originalError = error;
	}

	readonly originalError?: Error;
}
