import * as semver from 'semver';
export interface TerminusInfo {
	terminusName: string;
	wizardStatus: string;
	selfhosted: boolean;
	tailScaleEnable: boolean;
	osVersion: string;
	avatar: string;
	loginBackground: string;
	terminusId: string;
	did: string;
	terminusd: string;
}

export interface OlaresInfo extends TerminusInfo {
	olaresId: string;
	id: string;
	olaresd: string;
	style: string;
	// selfhosted
	enableReverseProxy: boolean;
}

export const DefaultTerminusInfo: TerminusInfo = {
	terminusName: '',
	wizardStatus: '',
	selfhosted: false,
	tailScaleEnable: false,
	osVersion: '',
	avatar: '',
	loginBackground: '',
	terminusId: '',
	did: '',
	terminusd: '0'
};

export const DefaultOlaresInfo: OlaresInfo = {
	olaresId: '',
	wizardStatus: '',
	selfhosted: false,
	tailScaleEnable: false,
	osVersion: '',
	avatar: '',
	loginBackground: '',
	style: '',
	id: '',
	did: '',
	olaresd: '0',
	terminusd: '0',
	terminusName: '',
	terminusId: '',
	enableReverseProxy: false
};

export const compareOlaresVersion = (version0: string, version1: string) => {
	if (!semver.valid(version0) || !semver.valid(version1)) {
		return {
			compare: 0
		};
	}

	const result = {
		compare: 0
	};
	result.compare = semver.compare(version0, version1);
	return result;
};

export const autoFuncWithRetry = async <T>(
	autoFunc: () => T,
	maxRetries = -1,
	retryDelay = 5000,
	delayFunc?: (attempt: number) => number,
	logger?: {
		log: (message: any, ...optionalParams: [...any, string?]) => void;
		error: (message: any, ...optionalParams: [...any, string?]) => void;
	}
): Promise<T> => {
	let attempt = 0;
	while (maxRetries === -1 || attempt < maxRetries) {
		try {
			attempt++;
			if (logger) {
				logger.log(`Attempting to autoFunc (attempt ${attempt})...`);
			}

			const t = await autoFunc();
			if (logger) {
				logger.log('Successfully autoFunc');
			}
			return t;
		} catch (error) {
			if (logger) {
				logger.error(`Connection attempt ${attempt} failed:`, error);
			}

			if (maxRetries !== -1 && attempt >= maxRetries) {
				if (logger) {
					logger.error(`Exceeded max retries (${maxRetries}), giving up`);
				}
				throw error;
			}
			let retryDelayNum = retryDelay;
			if (delayFunc) {
				retryDelayNum = delayFunc(attempt);
			}
			if (logger) {
				logger.log(`Waiting ${retryDelayNum}ms before next attempt...`);
			}

			await new Promise((resolve) => setTimeout(resolve, retryDelayNum));
		}
	}
	throw new Error('Unexpected exit from connection loop');
};

export const TerminusDefaultDomain = 'olares.com';
export const OlaresDefaultDomain = 'olares.com';
