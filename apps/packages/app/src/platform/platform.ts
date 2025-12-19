import { Platform, StubPlatform, DeviceInfo } from '@didvault/sdk/src/core';
import { bytesToBase64 } from '@didvault/sdk/src/core';

import { LocalStorage } from '@didvault/sdk/src/storage';
import {
	AuthPurpose,
	AuthRequestStatus,
	AuthType
} from '@didvault/sdk/src/core';

import {
	StartRegisterAuthenticatorResponse,
	CompleteRegisterMFAuthenticatorParams,
	StartAuthRequestParams,
	CompleteAuthRequestParams,
	StartRegisterAuthenticatorParams
} from '@didvault/sdk/src/core/api';
import { StartAuthRequestResponse } from '@didvault/sdk/src/core';
import { Err, ErrorCode } from '@didvault/sdk/src/core';
import UAParser from 'ua-parser-js';
import { SSIAuthClient } from '../auth/ssi';
import { RSAPublicKey } from '@didvault/sdk/src/core';
import { app } from '../globals';

import { WebCryptoProvider } from '@didvault/sdk/src/crypto';
import WebCryptoProviderLocal from '@didvault/sdk/src/localcrypto';
// import { i18n } from 'src/boot/i18n';
import { translate as $l } from '@didvault/sdk/src/util';

let publicKey: RSAPublicKey;

const browserInfo = (async () => {
	return new UAParser(navigator.userAgent).getResult();
})();

export class WebPlatform extends StubPlatform implements Platform {
	private _clipboardTextArea: HTMLTextAreaElement;
	private _qrVideo: HTMLVideoElement;
	private _qrCanvas: HTMLCanvasElement;

	constructor() {
		super();

		const base_url = window.location.origin;
		if (base_url.startsWith('http://')) {
			if (base_url.startsWith('http://localhost')) {
				this.crypto = new WebCryptoProvider();
			} else {
				this.crypto = new WebCryptoProviderLocal();
			}
		} else {
			this.crypto = new WebCryptoProvider();
		}
	}

	storage = new LocalStorage();

	get supportedAuthTypes() {
		return [AuthType.SSI];
	}

	// Set clipboard text using `document.execCommand("cut")`.
	// NOTE: This only works in certain environments like Google Chrome apps with the appropriate permissions set
	async setClipboard(text: string): Promise<void> {
		this._clipboardTextArea =
			this._clipboardTextArea || document.createElement('textarea');
		this._clipboardTextArea.contentEditable = 'true';
		this._clipboardTextArea.readOnly = false;
		this._clipboardTextArea.value = text;
		document.body.appendChild(this._clipboardTextArea);
		const range = document.createRange();
		range.selectNodeContents(this._clipboardTextArea);

		const s = window.getSelection();
		s!.removeAllRanges();
		s!.addRange(range);
		this._clipboardTextArea.select();

		this._clipboardTextArea.setSelectionRange(
			0,
			this._clipboardTextArea.value.length
		);

		document.execCommand('cut');
		document.body.removeChild(this._clipboardTextArea);
	}

	// Get clipboard text using `document.execCommand("paste")`
	// NOTE: This only works in certain environments like Google Chrome apps with the appropriate permissions set
	async getClipboard(): Promise<string> {
		return new Promise<string>((resolve, reject) => {
			if (navigator.clipboard && window.isSecureContext) {
				navigator.clipboard
					.readText()
					.then((v) => {
						resolve(v);
					})
					.catch((v) => {
						reject(v);
					});
			} else {
				this._clipboardTextArea =
					this._clipboardTextArea || document.createElement('textarea');
				document.body.appendChild(this._clipboardTextArea);
				this._clipboardTextArea.value = '';
				this._clipboardTextArea.select();
				document.execCommand('paste');
				document.body.removeChild(this._clipboardTextArea);
				resolve(this._clipboardTextArea.value);
			}
		});
	}

	async getDeviceInfo() {
		const { os, browser } = await browserInfo;
		const platform = (os.name && os.name.replace(' ', '')) || '';
		return new DeviceInfo({
			platform,
			osVersion: (os.version && os.version.replace(' ', '')) || '',
			id: '',
			// appVersion: process.env.PL_VERSION || "",
			// vendorVersion: process.env.PL_VENDOR_VERSION || "",
			manufacturer: '',
			model: '',
			browser: browser.name || '',
			browserVersion: browser.version,
			userAgent: navigator.userAgent,
			locale: navigator.language || 'en',
			description:
				browser.name && browser.name !== 'Electron'
					? $l('{browser} on {platform}', {
							browser: browser.name,
							platform: platform
					  })
					: $l('{platform} device', {
							platform
					  }),
			runtime: 'web'
		});
	}

	async scanQR() {
		return new Promise<string>((resolve, reject) => {
			const tick = async () => {
				if (this._qrVideo.readyState !== this._qrVideo.HAVE_ENOUGH_DATA) {
					requestAnimationFrame(() => tick());
					return;
				}

				const { default: jsQR } = await import(
					/* webpackChunkName: "jsqr" */ 'jsqr'
				);

				const canvas = this._qrCanvas.getContext('2d')!;
				this._qrCanvas.height = this._qrVideo.videoHeight;
				this._qrCanvas.width = this._qrVideo.videoWidth;
				canvas.drawImage(
					this._qrVideo,
					0,
					0,
					this._qrCanvas.width,
					this._qrCanvas.height
				);
				const imageData = canvas.getImageData(
					0,
					0,
					this._qrCanvas.width,
					this._qrCanvas.height
				);
				const code = jsQR(imageData.data, imageData.width, imageData.height, {
					inversionAttempts: 'dontInvert'
				});
				if (code) {
					resolve(code.data);
				}
				requestAnimationFrame(() => tick());
			};

			if (!this._qrVideo) {
				this._qrVideo = document.createElement('video');
				this._qrVideo.setAttribute('playsinline', '');
				this._qrVideo.setAttribute('muted', '');
				this._qrVideo.setAttribute('autoplay', '');
			}

			if (!this._qrCanvas) {
				this._qrCanvas = document.createElement('canvas');
				Object.assign(this._qrCanvas.style, {
					position: 'absolute',
					top: '0',
					left: '0',
					width: '100%',
					height: '100%',
					objectFit: 'cover',
					zIndex: '-1'
				});
				document.body.appendChild(this._qrCanvas);
			}

			this._qrCanvas.style.display = 'block';

			navigator.mediaDevices
				.getUserMedia({
					audio: false,
					video: { facingMode: 'environment' }
				})
				.then((stream) => {
					// Use facingMode: environment to attemt to get the front camera on phones
					this._qrVideo.srcObject = stream;
					this._qrVideo.play();
					requestAnimationFrame(() => tick());
				}, reject);
		});
	}

	async stopScanQR() {
		const stream: MediaStream | null =
			this._qrVideo && (this._qrVideo.srcObject as MediaStream);
		if (stream) {
			for (const track of stream.getTracks()) {
				track.stop();
			}
		}

		this._qrVideo && (this._qrVideo.srcObject = null);
		this._qrCanvas.style.display = 'none';
	}

	async composeEmail(addr: string, subj: string, msg: string) {
		window.open(
			`mailto:${addr}?subject=${encodeURIComponent(
				subj
			)}&body=${encodeURIComponent(msg)}`,
			'_'
		);
	}

	openExternalUrl(url: string) {
		window.open(url, '_blank');
	}

	async saveFile(name: string, type: string, contents: Uint8Array) {
		const a = document.createElement('a');
		a.href = `data:${type};base64,${bytesToBase64(contents, false)}`;
		a.download = name;
		a.rel = 'noopener';
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
	}

	private async _getAuthClient(type: AuthType) {
		switch (type) {
			// case AuthType.WebAuthnPlatform:
			// case AuthType.WebAuthnPortable:
			// 	return webAuthnClient;
			// // case AuthType.DID:
			// // 	return new EmailAuthClient();
			// case AuthType.Totp:
			// 	return new TotpAuthCLient();
			// case AuthType.OpenID:
			// 	return new OpenIDClient();
			// case AuthType.OsPassword:
			// 	return new OsPasswordAuthClient();
			// case AuthType.OsToken:
			// 	return new OsTokenAuthClient();
			// case AuthType.PublicKey:
			// 	const pdata = new PublicKeyAuthClientData();
			// 	await pdata.generateKeys();

			// 	if (!pdata.privateKey || !pdata.privateKey) {
			// 		return null;
			// 	}
			// 	publicKey = pdata.publicKey;
			// 	//let private_key = new RSAPrivateKey(pdata.privateKey);

			// 	return new PublicKeyAuthClient2(pdata.privateKey);
			case AuthType.SSI:
				return new SSIAuthClient();

			default:
				return null;
		}
	}

	protected async _prepareRegisterAuthenticator({
		data,
		type
	}: StartRegisterAuthenticatorResponse): Promise<any> {
		const client = await this._getAuthClient(type);
		if (!client) {
			throw new Err(
				ErrorCode.AUTHENTICATION_FAILED,
				$l('Authentication type not supported!')
			);
		}
		return client.prepareRegistration(data);
	}

	async registerAuthenticator({
		purposes,
		type,
		data,
		device
	}: {
		purposes: AuthPurpose[];
		type: AuthType;
		data?: any;
		device?: DeviceInfo;
	}) {
		const res = await app.api.startRegisterAuthenticator(
			new StartRegisterAuthenticatorParams({
				purposes,
				type,
				data,
				device
			})
		);
		try {
			const prepData = await this._prepareRegisterAuthenticator(res);
			if (!prepData) {
				throw new Err(ErrorCode.AUTHENTICATION_FAILED, $l('Setup Canceled'));
			}
			await app.api.completeRegisterAuthenticator(
				new CompleteRegisterMFAuthenticatorParams({
					id: res.id,
					data: prepData
				})
			);
			return res.id;
		} catch (e) {
			await app.api.deleteAuthenticator(res.id);
			throw e;
		}
	}

	protected async _prepareCompleteAuthRequest({
		data,
		type
	}: StartAuthRequestResponse): Promise<any> {
		const client = await this._getAuthClient(type);
		if (!client) {
			throw new Err(
				ErrorCode.AUTHENTICATION_FAILED,
				$l('Authentication type not supported!')
			);
		}

		return client.prepareAuthentication(data);
	}

	async startAuthRequest({
		purpose,
		type,
		did = app.account?.did,
		authenticatorId,
		authenticatorIndex
	}: {
		purpose: AuthPurpose;
		type?: AuthType;
		did?: string;
		authenticatorId?: string;
		authenticatorIndex?: number;
	}) {
		return app.api.startAuthRequest(
			new StartAuthRequestParams({
				did,
				type,
				supportedTypes: this.supportedAuthTypes,
				purpose,
				authenticatorId,
				authenticatorIndex
			})
		);
	}

	async completeAuthRequest(req: StartAuthRequestResponse) {
		if (req.requestStatus === AuthRequestStatus.Verified) {
			return {
				did: req.did,
				token: req.token,
				deviceTrusted: req.deviceTrusted,
				accountStatus: req.accountStatus!,
				provisioning: req.provisioning!
			};
		}

		const data = await this._prepareCompleteAuthRequest(req);

		if (!data) {
			throw new Err(
				ErrorCode.AUTHENTICATION_FAILED,
				$l('The request was canceled.')
			);
		}

		if (req.type == AuthType.PublicKey) {
			data.publickey = bytesToBase64(publicKey);
		}

		const { accountStatus, deviceTrusted, provisioning /*, legacyData*/ } =
			await app.api.completeAuthRequest(
				new CompleteAuthRequestParams({
					id: req.id,
					data,
					did: req.did
				})
			);

		return {
			did: req.did,
			token: req.token,
			deviceTrusted,
			accountStatus,
			provisioning
		};
	}

	readonly platformAuthType: AuthType | null = AuthType.WebAuthnPlatform;

	async supportsPlatformAuthenticator() {
		return this.supportedAuthTypes.includes(AuthType.WebAuthnPlatform);
	}

	async registerPlatformAuthenticator(purposes: AuthPurpose[]) {
		if (!this.platformAuthType) {
			throw new Err(ErrorCode.NOT_SUPPORTED);
		}
		return this.registerAuthenticator({
			purposes,
			type: this.platformAuthType,
			device: app.state.device
		});
	}
}
