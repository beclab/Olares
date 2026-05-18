<template>
	<div
		class="container"
		:class="
			$q.dark.isActive ? 'theme-desktop-dark-bg' : 'theme-desktop-light-bg'
		"
	>
		<DesktopDefaultHeaderView
			class="headerBar"
			:height="30"
			v-if="
				($q.platform.is.win || $q.platform.is.linux) && $q.platform.is.electron
			"
		/>
		<div
			class="contain-content"
			:class="
				($q.platform.is.win || $q.platform.is.linux) && $q.platform.is.electron
					? 'contain-content-win'
					: $q.platform.is.ipad
					? 'contain-content-ipad'
					: $q.platform.is.android
					? 'contain-content-android-pad'
					: 'contain-content-common'
			"
		>
			<TerminusMenu @openSearch="changeSearchDialog(true)" />

			<div
				class="contain-body"
				:class="$q.platform.is.android ? 'contain-body-android-pad' : ''"
			>
				<FilesMainLayout
					v-if="menuStore.terminusActiveMenu === LayoutMenuIdetify.FILES"
				/>

				<VaultMainLayout
					v-if="
						menuStore.terminusActiveMenu === LayoutMenuIdetify.VAULT &&
						userStore.isUnlocked
					"
				/>
				<TermipassUnlockContent
					v-if="
						!$q.platform.is.mobile &&
						menuStore.terminusActiveMenu === LayoutMenuIdetify.VAULT &&
						!userStore.isUnlocked
					"
					:logo="
						$q.dark.isActive
							? 'login/vault_brand_web_dark.png'
							: 'login/vault_brand_web_light.png'
					"
					:detail-text="t('unlock.vault_unlock_introduce')"
					:cancel="false"
					:logo-width="144"
				/>
				<div
					class="row items-center justify-center bg-background-1"
					style="width: 100%; height: 100%"
					v-if="
						$q.platform.is.mobile &&
						menuStore.terminusActiveMenu === LayoutMenuIdetify.VAULT &&
						!userStore.isUnlocked
					"
				>
					<TermipassMobileUnlockContent
						v-if="
							$q.platform.is.mobile &&
							menuStore.terminusActiveMenu === LayoutMenuIdetify.VAULT &&
							!userStore.isUnlocked
						"
						:cancel="false"
						:detailText="t('unlock.vault_unlock_introduce')"
						logo="login/vault_unlock.svg"
						:biometry-auto-unlock="true"
						class="bg-background-1"
					/>
				</div>

				<TransferLayout
					v-if="menuStore.terminusActiveMenu === LayoutMenuIdetify.TRANSMISSION"
				/>

				<SettingsPage
					v-if="
						menuStore.terminusActiveMenu === LayoutMenuIdetify.SYSTEM_SETTINGS
					"
				/>

				<AccountCenter
					v-if="
						menuStore.terminusActiveMenu === LayoutMenuIdetify.ACCOUNT_CENTER
					"
				/>
			</div>

			<div
				class="search_mask"
				v-if="showSearchDialog"
				@click.self="changeSearchDialog(false)"
			></div>
			<SearchPage v-if="showSearchDialog" @hide="changeSearchDialog" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import { getAppPlatform } from '../application/platform';
import { useUserStore } from '../stores/user';
import { useMenuStore } from '../stores/menu';
import { useSearchStore } from '../stores/search';

import FilesMainLayout from './files/LayoutPc.vue';
import VaultMainLayout from './MainLayout.vue';
import TransferLayout from './TransferLayout.vue';

import SettingsPage from './../pages/Electron/SettingsPage/SettingsPage.vue';
import AccountCenter from './../pages/Electron/SettingsPage/Account.vue';

import DesktopDefaultHeaderView from '../components/DesktopDefaultHeaderView.vue';
import { useDataStore } from '../stores/data';
import { useLarepassWebsocketManagerStore } from '../stores/larepassWebsocketManager';

import TerminusMenu from './../components/TerminusMenu.vue';
import SearchPage from './../components/search/LarePass/IndexPage.vue';

import { watch } from 'vue';
import TermipassUnlockContent from '../components/unlock/desktop/TermipassUnlockContent.vue';
import TermipassMobileUnlockContent from '../components/unlock/mobile/TermipassUnlockContent.vue';

import { LayoutMenuIdetify } from '../utils/constants';
import { useI18n } from 'vue-i18n';

import HotkeyManager from 'src/directives/hotkeyManager';
import { FILES_HOTKEY } from 'src/api/files/hotKeys';

const route = useRoute();
const userStore = useUserStore();
const menuStore = useMenuStore();
const socketStore = useLarepassWebsocketManagerStore();
const searchStore = useSearchStore();
const { t } = useI18n();

const showSearchDialog = ref(false);

onMounted(async () => {
	if (process.env.PLATFORM === 'DESKTOP' || process.env.PLATFORM == 'MOBILE') {
		import('../css/larepass/layout-desktop.scss').then(() => {});
	}

	menuStore.pushTerminusMenuCache(LayoutMenuIdetify.FILES);

	getAppPlatform().homeMounted();

	HotkeyManager.registerHotkeys({
		[FILES_HOTKEY.DESKTOP_SEARCH.DISPLAY]: () => {
			showSearchDialog.value = !showSearchDialog.value;
		}
	});
});

watch(
	() => route.path,
	() => {
		if (process.env.PLATFORM == 'DESKTOP') {
			socketStore.restart();
		}
	},
	{
		immediate: true
	}
);

onUnmounted(() => {
	getAppPlatform().homeUnMounted();
	HotkeyManager.unregisterHotkeys({
		[FILES_HOTKEY.DESKTOP_SEARCH.DISPLAY]: () => {
			showSearchDialog.value = !showSearchDialog.value;
		}
	});
});

const changeSearchDialog = (value: boolean) => {
	showSearchDialog.value = value;
};
</script>

<style lang="scss" scoped>
.container {
	width: 100vw;
	height: 100vh;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: space-between;
	overflow: hidden;

	.contain-content {
		width: 100vw;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding-right: 8px;
		padding-bottom: 8px;

		.contain-body {
			width: 100%;
			height: 100%;
			border-radius: 12px;
			display: flex;
			align-items: self-start;
			justify-content: space-between;
			overflow: hidden;
			// background-color: red;
		}

		.contain-body-android-pad {
			padding-top: 20px;
			padding-bottom: 8px;
		}
	}

	.contain-content-win {
		height: calc(100vh - 30px);
	}

	.contain-content-ipad {
		height: 100vh;
		padding-top: env(safe-area-inset-top);
	}

	.contain-content-android-pad {
		height: 100vh;
		padding-bottom: 0px;
	}

	.contain-content-common {
		height: 100vh;
		padding-top: 8px;
	}

	.headerBar {
		width: 100%;
		height: 30px;
	}

	.search_mask {
		position: absolute;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		z-index: 1000;
		background-color: rgba($color: #000000, $alpha: 0.5);
	}
}
</style>
