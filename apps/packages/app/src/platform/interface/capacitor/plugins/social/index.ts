import { registerPlugin } from '@capacitor/core';
import { LarePassSocialLoginPlugin } from './definitions';
const LarePassSocialLogin = registerPlugin<LarePassSocialLoginPlugin>(
	'LarePassSocialLoginPlugin'
);
export { LarePassSocialLogin };
