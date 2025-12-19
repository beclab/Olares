<template>
	<div class="row items-center">
		<q-avatar :icon="SETTING_MENU.display.icon" @click="toggleDrawer" />
		<q-toolbar-title>{{ t(SETTING_MENU.display.name) }}</q-toolbar-title>
	</div>

	<q-list class="q-ma-md sItem" separator v-if="!$q.platform.is.mobile">
		<q-item class="titleBg">
			<q-item-section>Theme</q-item-section>
		</q-item>
		<q-item>
			<q-select
				class="full-width"
				borderless
				dense
				v-model="themeModel"
				:options="themeOption"
			/>
		</q-item>
	</q-list>

	<q-list class="q-ma-md sItem" separator>
		<q-item class="titleBg">
			<q-item-section>Favicons</q-item-section>
		</q-item>
		<q-item class="row items-center justify-between">
			<q-item-label>Enable Favicons</q-item-label>
			<q-item-label>
				<bt-switch
					size="sm"
					truthy-track-color="blue-default"
					v-model="faviconsStatus"
					@update:model-value="changeFavicons"
				/>
			</q-item-label>
		</q-item>
	</q-list>

	<q-list class="q-ma-md sItem" separator>
		<q-item class="titleBg row items-center justify-between">
			<q-item-section>Masked Fields</q-item-section>
			<BtIcon class="q-mr-md cursor-pointer" src="info">
				<q-tooltip
					anchor="bottom middle"
					self="top middle"
					style="width: 240px"
				>
					If this option is enabled, masked fields such as passwords or credit
					card numbers will be unmasked when you move your mouse over them.
					Disable this option if you would rather use an explicit button.
				</q-tooltip>
			</BtIcon>
		</q-item>
		<q-item class="row items-center justify-between">
			<q-item-label>Reveal On Hover</q-item-label>
			<q-item-label>
				<bt-switch
					size="sm"
					truthy-track-color="blue-default"
					v-model="unmaskStatus"
					@update:model-value="changeUnmaskStatus"
				/>
			</q-item-label>
		</q-item>
	</q-list>
</template>

<script lang="ts">
import { defineComponent, ref, watch } from 'vue';
import { useQuasar } from 'quasar';
import { SETTING_MENU } from '../../../utils/constants';
import { useMenuStore } from '../../../stores/menu';
import { app } from '../../../globals';

import { useI18n } from 'vue-i18n';

export default defineComponent({
	name: 'DisplayPage',
	components: {},
	setup() {
		const themeModel = ref(app.settings.theme);
		const themeOption = ref(['auto', 'light', 'dark']);
		const faviconsStatus = ref(app.settings.favicons);
		const unmaskStatus = ref(app.settings.unmaskFieldsOnHover);
		const meunStore = useMenuStore();
		const $q = useQuasar();

		const toggleDrawer = () => {
			meunStore.rightDrawerOpen = false;
		};

		const changeUnmaskStatus = (val) => {
			app.setSettings({ unmaskFieldsOnHover: val });
		};

		watch(
			() => app.settings.theme,
			(newVal) => {
				themeModel.value = newVal;
			}
		);

		watch(
			() => themeModel.value,
			async (newVal) => {
				if (newVal) {
					await app.setSettings({ theme: newVal });
					if (newVal === 'auto') {
						if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
							$q.dark.set(true);
						} else {
							$q.dark.set(false);
						}
					}
					if (newVal === 'dark') {
						$q.dark.set(true);
					}
					if (newVal === 'light') {
						$q.dark.set(false);
					}
				}
			}
		);

		const changeFavicons = (value) => {
			app.setSettings({ favicons: value });
		};

		const { t } = useI18n();

		return {
			SETTING_MENU,
			themeModel,
			themeOption,
			faviconsStatus,
			toggleDrawer,
			unmaskStatus,
			changeUnmaskStatus,
			changeFavicons,
			t
		};
	}
});
</script>

<style scoped lang="scss">
.sItem {
	border: 1px solid $separator;
	border-radius: 10px;
}
</style>
