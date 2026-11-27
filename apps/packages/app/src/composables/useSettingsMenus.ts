import { computed, ComputedRef } from 'vue';
import { useMeStore } from 'src/stores/settings/me';
import { MENU_TYPE, MenuItem, useMenuItem } from 'src/constant';

export function useSettingsMenus(): ComputedRef<(MenuItem | undefined)[][]> {
	const meStore = useMeStore();
	return computed(() => {
		if (meStore.isAdmin) {
			return [
				[
					useMenuItem(MENU_TYPE.Users),
					useMenuItem(MENU_TYPE.Appearance),
					useMenuItem(MENU_TYPE.Application),
					useMenuItem(MENU_TYPE.Integration)
				],
				[
					useMenuItem(MENU_TYPE.VPN),
					useMenuItem(MENU_TYPE.Network),
					useMenuItem(MENU_TYPE.GPU),
					useMenuItem(MENU_TYPE.Video),
					useMenuItem(MENU_TYPE.Search)
				],
				[
					useMenuItem(MENU_TYPE.Backup),
					useMenuItem(MENU_TYPE.Restore),
					useMenuItem(MENU_TYPE.Developer)
				]
			];
		}
		return [
			[
				useMenuItem(MENU_TYPE.Appearance),
				useMenuItem(MENU_TYPE.Application),
				useMenuItem(MENU_TYPE.Integration)
			],
			[useMenuItem(MENU_TYPE.Network)],
			[useMenuItem(MENU_TYPE.VPN)],
			[useMenuItem(MENU_TYPE.Video)],
			[useMenuItem(MENU_TYPE.Search)],
			[
				useMenuItem(MENU_TYPE.Backup),
				useMenuItem(MENU_TYPE.Restore),
				useMenuItem(MENU_TYPE.Developer)
			]
		];
	});
}
