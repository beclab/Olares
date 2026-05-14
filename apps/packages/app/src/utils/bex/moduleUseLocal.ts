const MODULE_NAMES_USE_REMOTE = ['wise'] as const;

export function useLocalForModule(moduleName: string): boolean {
	return !MODULE_NAMES_USE_REMOTE.includes(
		moduleName as (typeof MODULE_NAMES_USE_REMOTE)[number]
	);
}
