import { Serializable, AsSerializable, AsDate, AsBytes } from './encoding';
import { Server } from './server';
import { DeviceInfo } from './platform';
import { EventEmitter } from './event-target';
import { wait } from './util';

/** Authentication data */
export class RequestAuthentication extends Serializable {
	/** The id of the [[Session]] used for authentication */
	session: string = '';
	/** A time stamp of the request/response time */

	@AsDate()
	time!: Date;

	/** The signature used to verify the authentiation */
	@AsBytes()
	signature!: Uint8Array;

	constructor(vals: Partial<RequestAuthentication> = {}) {
		super();
		Object.assign(this, vals);
	}
}

/** RPC request object */
export class Request extends Serializable {
	/** Name of the method call */
	method: string = '';
	/** Arguments for the method */

	params?: any[] = undefined;

	/** Data used to authenticate the request */
	@AsSerializable(RequestAuthentication)
	auth?: RequestAuthentication;

	/** Info about the device the request is coming from */
	@AsSerializable(DeviceInfo)
	device?: DeviceInfo;

	ipAddress?: string;

	location?: {
		city?: string;
		country?: string;
	};
}

/** RPC response object */
export class Response extends Serializable {
	/** The result of the RPC call */
	result: any = null;

	/** Error info, if an error occurred while processing the request */
	error?: Error = undefined;

	/** Data used to authenticate the response */
	@AsSerializable(RequestAuthentication)
	auth?: RequestAuthentication;
}

/** Error info */
export interface Error {
	/** Error code */
	code: number | string;
	/** Error message */
	message: string;
}

/** Generic interface for sending [[Request]]s */
export interface Sender {
	send(req: Request, progress?: RequestProgress): Promise<Response>;
}

/** Generic interface for receiving [[Request]]s and processing them into a [[Response]] */
export interface Receiver {
	listen(handler: (req: Request) => Promise<Response>): void;
}

/**
 * Stub implementation of the [[Sender]] interface, passing requests directly
 * into a [[Server]] instance.  this is useful for testing, where client and
 * server instances are managed by the same process
 */
export class DirectSender implements Sender {
	constructor(private server: Server) {}

	async send(req: Request) {
		await wait(10);
		return this.server.handle(req);
	}
}

interface Progress {
	loaded: number;
	total: number;
}

export class RequestProgress extends EventEmitter {
	completed: Promise<void> = new Promise((resolve, reject) => {
		this._resolveCompleted = resolve;
		this._rejectCompleted = reject;
	});

	get uploadProgress() {
		return this._uploadProgress;
	}

	set uploadProgress(progress: Progress) {
		this._uploadProgress = progress;
		this.dispatch('upload-progress', progress);
		this.dispatch('progress', this.progress);
	}

	get downloadProgress() {
		return this._downloadProgress;
	}

	set downloadProgress(progress: Progress) {
		this._downloadProgress = progress;
		this.dispatch('download-progress', progress);
		this.dispatch('progress', this.progress);
	}

	get progress() {
		return {
			loaded: this.uploadProgress.loaded + this.downloadProgress.loaded,
			total: this.uploadProgress.total + this.downloadProgress.total
		};
	}

	set error(error: Error | undefined) {
		this._error = error;
		if (error) {
			this.dispatch('error', { error });
			this.uploadProgress = this.downloadProgress = {
				loaded: 0,
				total: 0
			};
			this._rejectCompleted(error);
		}
	}

	get error() {
		return this._error;
	}

	private _uploadProgress: Progress = { loaded: 0, total: 0 };
	private _downloadProgress: Progress = { loaded: 0, total: 0 };

	private _error?: Error;

	private _resolveCompleted!: () => void;
	private _rejectCompleted!: (err: Error) => void;

	complete() {
		this._resolveCompleted();
	}
}
