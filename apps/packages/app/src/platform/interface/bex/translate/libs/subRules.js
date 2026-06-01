import { GLOBAL_KEY } from '../config';
import {
	getSyncWithDefault,
	updateSync,
	setSubRules,
	getSubRules,
	getSubRulesVersion,
	setSubRulesVersion
} from './storage';
import { apiFetch } from '../apis';
import { checkRules } from './rules';
import { isAllchar } from './utils';
import { kissLog } from './log';
import {
	RULES_DATA,
	RULES_DATA_ON,
	RULES_DATA_OFF,
	RULES_DATA_VERSION,
	RULES_DATA_ON_VERSION,
	RULES_DATA_OFF_VERSION
} from '../config/rules-data';

const isLocalRule = (url) => {
	return url && url.startsWith('local:');
};

const getLocalRulesData = (url) => {
	const ruleMap = {
		'local:olares-rules': RULES_DATA,
		'local:olares-rules-on': RULES_DATA_ON,
		'local:olares-rules-off': RULES_DATA_OFF
	};
	return ruleMap[url] || null;
};

const getLocalRulesVersion = (url) => {
	const versionMap = {
		'local:olares-rules': RULES_DATA_VERSION,
		'local:olares-rules-on': RULES_DATA_ON_VERSION,
		'local:olares-rules-off': RULES_DATA_OFF_VERSION
	};
	return versionMap[url] || '1.0.0';
};

const updateSyncDataCache = async (url) => {
	const { dataCaches = {} } = await getSyncWithDefault();
	dataCaches[url] = Date.now();
	await updateSync({ dataCaches });
};

export const syncSubRules = async (url) => {
	if (isLocalRule(url)) {
		const localData = getLocalRulesData(url);
		if (localData) {
			const rules = checkRules(localData).filter(
				({ pattern }) => !isAllchar(pattern, GLOBAL_KEY)
			);
			if (rules.length > 0) {
				await setSubRules(url, rules);
			}
			return rules;
		}
		return [];
	}

	const res = await apiFetch(url);
	const rules = checkRules(res).filter(
		({ pattern }) => !isAllchar(pattern, GLOBAL_KEY)
	);
	if (rules.length > 0) {
		await setSubRules(url, rules);
	}
	return rules;
};

export const syncAllSubRules = async (subrulesList) => {
	for (const subrules of subrulesList) {
		try {
			await syncSubRules(subrules.url);
			await updateSyncDataCache(subrules.url);
		} catch (err) {
			kissLog(err, `sync subrule error: ${subrules.url}`);
		}
	}
};

export const trySyncAllSubRules = async ({ subrulesList }) => {
	try {
		const { subRulesSyncAt } = await getSyncWithDefault();
		const now = Date.now();
		const interval = 24 * 60 * 60 * 1000;
		if (now - subRulesSyncAt > interval) {
			await syncAllSubRules(subrulesList);
			await updateSync({ subRulesSyncAt: now });
		}
	} catch (err) {
		kissLog(err, 'try sync all subrules');
	}
};

export const loadOrFetchSubRules = async (url) => {
	let rules = await getSubRules(url);

	if (isLocalRule(url)) {
		const cachedVersion = await getSubRulesVersion(url);
		const currentVersion = getLocalRulesVersion(url);

		if (!rules || rules.length === 0 || cachedVersion !== currentVersion) {
			rules = await syncSubRules(url);
			await setSubRulesVersion(url, currentVersion);
			await updateSyncDataCache(url);
		}
	} else {
		if (!rules || rules.length === 0) {
			rules = await syncSubRules(url);
			await updateSyncDataCache(url);
		}
	}

	return rules || [];
};
