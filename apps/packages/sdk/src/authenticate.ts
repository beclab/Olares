import { AuthType, AccountStatus, AuthPurpose, Err } from './core';

import { completeAuthRequest, startAuthRequest } from './core';
import { AccountProvisioning } from './core';
import { StartAuthRequestResponse } from './core';

import { ErrorCode } from './core';

export async function _authenticate({
	did,
	type,
	purpose,
	pendingRequest: req,
	authenticatorIndex = 0,
	caller = ''
}: {
	did: string;
	type: AuthType;
	purpose: AuthPurpose;
	authenticatorIndex?: number;
	pendingRequest?: StartAuthRequestResponse;
	caller?: string;
}): Promise<{
	did: string;
	token: string;
	accountStatus: AccountStatus;
	provisioning: AccountProvisioning;
	deviceTrusted: boolean;

	//legacyData?: PBES2Container;
} | null> {
	let step = 0;
	step = 1;
	if (!req) {
		console.log(
			`[${caller}] Step ${step}: req is empty, starting auth request...`
		);
		try {
			req = await startAuthRequest({
				type: type,
				purpose: purpose,
				did: did,
				authenticatorIndex
			});
			console.log(
				`[${caller}] Step ${step}: Auth request started successfully. Request details: ${JSON.stringify(
					req
				)}`
			);
		} catch (authRequestError) {
			console.error(
				`[${caller}] Step ${step}: Error occurred while starting auth request:`,
				authRequestError
			);
			throw new Err(
				ErrorCode.AUTHENTICATION_FAILED,
				`[${caller}] Step ${step}: An error occurred: ${authRequestError.message}`,
				{ error: authRequestError }
			);
		}
	} else {
		console.log(
			`[${caller}] Step ${step}: req already exists. Skipping auth request.`
		);
	}

	step = 2;
	try {
		console.log(
			`[${caller}] Step ${step}: Completing auth request with req: ${JSON.stringify(
				req
			)}`
		);
		const res = await completeAuthRequest(req);
		console.log(
			`[${caller}] Step ${step}: Auth request completed successfully. Response details: ${JSON.stringify(
				res
			)}`
		);
		return res;
	} catch (completeAuthError) {
		console.error(
			`[${caller}] Step ${step}: Error occurred while completing auth request:`,
			completeAuthError
		);
		throw new Err(
			ErrorCode.AUTHENTICATION_FAILED,
			`[${caller}] Step ${step}: An error occurred: ${completeAuthError.message}`,
			{ error: completeAuthError }
		);
	}

	// try {
	// 	if (!req) {
	// 		console.log('req  1');
	// 		req = await startAuthRequest({
	// 			//type: AuthType.PublicKey,
	// 			type: type,
	// 			purpose: purpose,
	// 			did: did,
	//
	// 			authenticatorIndex
	// 		});
	// 		console.log('req  ' + JSON.stringify(req));
	// 		//await this.app.storage.save(req);
	// 		// this.router.setParams({ pendingAuth: req.id });
	// 	}
	//
	// 	try {
	// 		const res = await completeAuthRequest(req);
	// 		console.log('res  ' + JSON.stringify(res));
	// 		return res;
	// 	} finally {
	// 		// this.router.setParams({ pendingAuth: undefined });
	// 		// this.app.storage.delete(req);
	// 	}
	// } catch (e: any) {
	// 	console.log(e);
	//
	// 	if (e.code === ErrorCode.NOT_FOUND) {
	// 		// await alert(e.message, { title: $l("Authentication Failed"), options: [$l("Cancel")] });
	// 		return null;
	// 	}
	//
	// 	// const choice = await alert(e.message, {
	// 	//     title: $l("Authentication Failed"),
	// 	//     options: [$l("Try Again"), $l("Try Another Method"), $l("Cancel")],
	// 	// });
	// 	// switch (choice) {
	// 	//     case 0:
	// 	//         return this._authenticate({ email, authenticatorIndex });
	// 	//     case 1:
	// 	//         return this._authenticate({ email, authenticatorIndex: authenticatorIndex + 1 });
	// 	//     default:
	// 	//         return null;
	// 	// }
	// 	return null;
	// }
}
