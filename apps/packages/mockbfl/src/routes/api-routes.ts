// routes/api-routes.ts
import express, { Router } from 'express';
import { ApiControllers } from '../controllers/api-controllers';

const apiControllers = new ApiControllers();

export class ApiRouter {
	router: Router;
	constructor() {
		this.router = express.Router();
		this.initializeRoutes();
	}

	initializeRoutes() {
		this.router.post('/bfl/iam/v1alpha1/login', apiControllers.postLogin);
		this.router.post(
			'/bfl/settings/v1alpha1/binding-zone',
			apiControllers.postSetZone
		);

		this.router.post(
			'/bfl/settings/v1alpha1/activate',
			apiControllers.postSetSystem
		);

		// this.router.post(
		// 	'/bfl/settings/v1alpha1/launcher-acc-policy',
		// 	apiControllers.postNetwork
		// );

		this.router.get(
			'/bfl/settings/v1alpha1/ssl/task-state',
			apiControllers.sslTaskState
		);

		this.router.put(
			'/bfl/iam/v1alpha1/users/:user/password',
			apiControllers.resetPassword
		);
		this.router.get('/bfl/monitor/v1alpha1/cluster', apiControllers.getMoniter);

		this.router.get('/bfl/backend/v1/ip', apiControllers.getIP);

		this.router.get('/bfl/iam/v1alpha1/roles', apiControllers.getRoles);
		//this.router.get('/bfl/iam/v1alpha1/set_did', apiControllers.getSetDID);
		// this.router.get(
		// 	'/bfl/backend/v1/user-info',
		// 	apiControllers.getUserInfo
		// );
		this.router.get('/bfl/info/v1/olares-info', apiControllers.getUserInfo);
		this.router.get(
			'/bfl/settings/v1alpha1/default-reverse-proxy',
			apiControllers.defaultReverseProxy
		);
		this.router.post(
			'/api/secondfactor/totp/identity/bind',
			apiControllers.getMFACode
		);
		this.router.post('/api/firstfactor', apiControllers.getFirstFactor);
		this.router.post('/api/secondfactor/totp', apiControllers.getSecondFactor);
	}
}
