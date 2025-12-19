import { API, PromiseWithProgress } from './api';
import { Sender, Request, Response, RequestProgress } from './transport';
import { DeviceInfo } from './platform';
import { Session } from './session';
import { Account } from './account';
import { Err, ErrorCode } from './error';
import { Serializable } from './encoding';

/**
 * Client state, keeping track of [[session]], [[account]] and [[device]] info
 */
export interface ClientState {
	/** Current session */
	session: Session | null;
	/** Currently logged in account */
	account: Account | null;
	/** info about current device */
	device: DeviceInfo;
}

/**
 * Client-side interface for Client-Server communication. Manages serialization,
 * authentication and some state like current session and account.
 */
export class Client extends API {
	online = true;

	constructor(
		/** Object for storing state */
		public state: ClientState,
		/** [[Sender]] implementation used for sending/receiving requests */
		private sender: Sender,
		private hook?: (
			req: Request,
			res: Response | null,
			err: Err | null
		) => void
	) {
		super();

		for (const { method, output } of this.handlerDefinitions) {
			this[method] = (input: Serializable | string | undefined) => {
				const progress = new RequestProgress();
				const promise = this.call(
					method,
					typeof input === 'undefined'
						? []
						: [
								input instanceof Serializable
									? input.toRaw()
									: input
						  ],
					progress
				).then((res) => {
					return output
						? Array.isArray(res.result)
							? res.result.map((each) =>
									new output().fromRaw(each)
							  )
							: new output().fromRaw(res.result)
						: res.result;
				}) as PromiseWithProgress<any>;
				promise.progress = progress;
				return promise;
			};
		}
	}

	/** The current session */
	get session() {
		return this.state.session;
	}

	/** Generic method for making an RPC call */
	async call(method: string, params?: any[], progress?: RequestProgress) {
		const { session } = this.state;

		const req = new Request();
		req.method = method;
		req.params = params;
		req.device = this.state.device;

		if (session) {
			await session.authenticate(req);
		}

		let res;

		try {
			res = await this.sender.send(req, progress);
		} catch (e) {
			if (progress) {
				progress.error = e;
			}
			this.hook && this.hook(req, null, e);
			throw e;
		}

		if (res.error) {
			const err = new Err(
				res.error.code as any as ErrorCode,
				res.error.message
			);
			if (progress) {
				progress.error = err;
			}
			this.hook && this.hook(req, res, err);
			throw err;
		}

		if (session && !(await session.verify(res))) {
			const err = new Err(ErrorCode.INVALID_RESPONSE);
			if (progress) {
				progress.error = err;
			}
			this.hook && this.hook(req, res, err);
			throw err;
		}

		this.hook && this.hook(req, res, null);
		return res;
	}
}
