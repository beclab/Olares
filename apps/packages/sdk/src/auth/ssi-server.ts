import {
	AuthClient,
	AuthServer,
	Authenticator,
	AuthRequest
} from '../core/auth';
import { AuthType } from '../core';
import {
	Serializable,
	AsBytes,
	AsSerializable,
	Exclude,
	bytesToBase64,
	base64ToBytes
} from '../core';
import { Err, ErrorCode } from '../core';
import { getCryptoProvider } from '../core';
import { checkJWS } from '../core/jws';

export class SSIAuthChallenge extends Serializable {
	@AsBytes()
	value!: Uint8Array;

	// @AsSerializable(RSASigningParams)
	// signingParams = new RSASigningParams();

	async init() {
		this.value = await getCryptoProvider().randomBytes(16);
	}
}

export class SSIAuthServer implements AuthServer {
	supportsType(type: AuthType) {
		return type === AuthType.SSI;
	}

	async initAuthenticator(authenticator: Authenticator) {
		const challenge = new SSIAuthChallenge();
		await challenge.init();
		authenticator.state = {
			activationChallenge: challenge.toRaw()
		};
		authenticator.description =
			authenticator.device?.description ||
			'Unknown Device Platform Authenticator';
		return {
			challenge: challenge.toRaw()
		};
	}

	async activateAuthenticator(
		authenticator: Authenticator<any>,
		{ jws }: { jws: string }
	): Promise<any> {
		try {
			const { terminus_name, body } = await checkJWS(jws);
			console.log('name ' + terminus_name);
			const challenge = new SSIAuthChallenge().fromRaw(
				authenticator.state.activationChallenge
			);

			console.log('challenge ' + bytesToBase64(challenge.value));
			console.log('body.challenge ' + body.challenge);

			if (
				!challenge ||
				!body.challenge ||
				body.challenge != bytesToBase64(challenge.value)
			) {
				throw new Err(ErrorCode.AUTHENTICATION_FAILED, 'Invalid signature.');
			}
		} catch (error) {
			console.log(error);
			if (error instanceof Err) {
				error.message = `[activateAuthenticator]:` + error.message;
			}
			throw error;
		}
	}

	async initAuthRequest(
		_authenticator: Authenticator<any>,
		request: AuthRequest<any>
	): Promise<any> {
		const challenge = new SSIAuthChallenge();
		await challenge.init();
		request.state = { challenge: challenge.toRaw() };
		return {
			challenge: challenge.toRaw()
		};
	}

	async verifyAuthRequest(
		_authenticator: Authenticator<any>,
		request: AuthRequest<any>,
		{ jws }: { jws: string }
	): Promise<void> {
		try {
			const { terminus_name, body } = await checkJWS(jws);
			console.log('name ' + terminus_name);
			const challenge = new SSIAuthChallenge().fromRaw(request.state.challenge);
			console.log('challenge ' + bytesToBase64(challenge.value));
			console.log('body.challenge ' + body.challenge);

			if (
				!challenge ||
				!body.challenge ||
				body.challenge != bytesToBase64(challenge.value)
			) {
				throw new Err(ErrorCode.AUTHENTICATION_FAILED, 'Invalid signature.');
			}
		} catch (error) {
			if (error instanceof Err) {
				error.message = `[verifyAuthRequest]:` + error.message;
			}
			throw error;
		}
	}
}
