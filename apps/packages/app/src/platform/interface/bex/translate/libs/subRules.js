import { GLOBAL_KEY } from '../config';
import {
	getSyncWithDefault,
	updateSync,
	setSubRules,
	getSubRules
} from './storage';
import { apiFetch } from '../apis';
import { checkRules } from './rules';
import { isAllchar } from './utils';
import { kissLog } from './log';
import {
	RULES_DATA,
	RULES_DATA_ON,
	RULES_DATA_OFF
} from '../config/rules-data';

/**
 * Check if the URL is a local rule
 * @param {*} url
 * @returns
 */
const isLocalRule = (url) => {
	return url && url.startsWith('local:');
};

/**
 * Get local rules data by URL identifier
 * @param {*} url
 * @returns
 */
const getLocalRulesData = (url) => {
	const ruleMap = {
		'local:olares-rules': RULES_DATA,
		'local:olares-rules-on': RULES_DATA_ON,
		'local:olares-rules-off': RULES_DATA_OFF
	};
	return ruleMap[url] || null;
};

/**
 * Update sync data cache timestamp
 * @param {*} url
 */
const updateSyncDataCache = async (url) => {
	const { dataCaches = {} } = await getSyncWithDefault();
	dataCaches[url] = Date.now();
	await updateSync({ dataCaches });
};

/**
 * Sync subscription rules
 * @param {*} url
 * @returns
 */
export const syncSubRules = async (url) => {
	// If it's a local rule, get data from local source
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

	// Remote rules, fetch from network
	const res = await apiFetch(url);
	const rules = checkRules(res).filter(
		({ pattern }) => !isAllchar(pattern, GLOBAL_KEY)
	);
	if (rules.length > 0) {
		await setSubRules(url, rules);
	}
	return rules;
};

/**
 * Sync all subscription rules
 * @param {*} url
 * @returns
 */
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

/**
 * Sync all subscription rules based on time
 * @param {*} url
 * @returns
 */
export const trySyncAllSubRules = async ({ subrulesList }) => {
	try {
		const { subRulesSyncAt } = await getSyncWithDefault();
		const now = Date.now();
		const interval = 24 * 60 * 60 * 1000; // One day interval
		if (now - subRulesSyncAt > interval) {
			// Sync subscription rules
			await syncAllSubRules(subrulesList);
			await updateSync({ subRulesSyncAt: now });
		}
	} catch (err) {
		kissLog(err, 'try sync all subrules');
	}
};

/**
 * Load subscription rules from cache or remote
 * @param {*} url
 * @returns
 */
export const loadOrFetchSubRules = async (url) => {
	let rules = await getSubRules(url);
	if (!rules || rules.length === 0) {
		rules = await syncSubRules(url);
		await updateSyncDataCache(url);
	}
	return rules || [];
};
