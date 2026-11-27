import { ref, Ref } from 'vue';
import { BaseEnv } from 'src/constant';
import { getSystemEnvList, getUserEnvList } from 'src/api/settings/env';

function effectiveEnvLiteral(e: BaseEnv): string {
	const v = e.value;
	if (v !== undefined && v !== null && String(v).trim() !== '') {
		return String(v);
	}
	return String(e.default ?? '');
}

/**
 * The env list endpoints are typed as `Promise<BaseEnv[]>`, but depending on
 * the active application channel the axios interceptor may return either the
 * unwrapped array (SETTINGS) or the raw response object carrying `.data`
 * (MARKET). Normalize both shapes into a `BaseEnv[]` without leaking `any`.
 */
function extractEnvArray(res: unknown): BaseEnv[] {
	if (Array.isArray(res)) {
		return res as BaseEnv[];
	}
	const data = (res as { data?: unknown } | null)?.data;
	return Array.isArray(data) ? (data as BaseEnv[]) : [];
}

export function useSystemEnvRefs(): {
	refNames: Ref<string[]>;
	refValueByName: Ref<Record<string, string>>;
	loadRefEnvLists: () => Promise<void>;
} {
	const refNames = ref<string[]>([]);
	const refValueByName = ref<Record<string, string>>({});

	async function loadRefEnvLists() {
		try {
			const [systemList, userList] = await Promise.all([
				getSystemEnvList(),
				getUserEnvList()
			]);
			const sys = extractEnvArray(systemList);
			const usr = extractEnvArray(userList);

			const valueMap: Record<string, string> = {};
			for (const e of sys) {
				if (e?.envName) {
					valueMap[e.envName] = effectiveEnvLiteral(e);
				}
			}
			for (const e of usr) {
				if (e?.envName) {
					valueMap[e.envName] = effectiveEnvLiteral(e);
				}
			}
			refValueByName.value = valueMap;

			const names = new Set<string>();
			sys.forEach((e) => e?.envName && names.add(e.envName));
			usr.forEach((e) => e?.envName && names.add(e.envName));
			refNames.value = Array.from(names).sort((a, b) =>
				a.localeCompare(b, undefined, { sensitivity: 'base' })
			);
		} catch (e) {
			console.error(e);
			refNames.value = [];
			refValueByName.value = {};
		}
	}

	return { refNames, refValueByName, loadRefEnvLists };
}
