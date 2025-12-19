import {
	BiometricOptions,
	BiometryType
} from '@capgo/capacitor-native-biometric';
import { i18n } from '../boot/i18n';

export const BIOMETRIC_TYPE_OPTION_RECORD = (
	type: BiometryType
): BiometricOptions => {
	switch (type) {
		case BiometryType.NONE:
			return {
				reason: 'for_easy_login',
				title: i18n.global.t('login.title')
			};
		case BiometryType.FACE_ID:
			return {
				reason: i18n.global.t('face_id_verification')
			};
		case BiometryType.TOUCH_ID:
			return {
				reason: i18n.global.t('touch_id_verification')
			};
		case BiometryType.FINGERPRINT:
			return {
				title: i18n.global.t('fingerprint_verification'),
				subtitle: i18n.global.t(
					'scan_your_fingerprint_to_verify_your_identity'
				),
				maxAttempts: 2
			};
		case BiometryType.FACE_AUTHENTICATION:
			return {
				title: i18n.global.t('face_authentication'),
				subtitle: i18n.global.t('scan_your_face_to_verify_your_identity'),
				maxAttempts: 2
			};
		case BiometryType.IRIS_AUTHENTICATION:
			return {
				title: i18n.global.t('iris_authentication'),
				subtitle: i18n.global.t('scan_your_iris_to_verify_your_identity'),
				maxAttempts: 2
			};
		case BiometryType.MULTIPLE:
			return {
				title: i18n.global.t('authentication'),
				subtitle: i18n.global.t(
					'scan_your_authentication_to_verify_your_identity'
				),
				maxAttempts: 2
			};
	}
};
