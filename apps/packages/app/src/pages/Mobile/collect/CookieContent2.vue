<template>
	<BtTooltip2 anchor="top middle" self="center middle">
		<template #tooltip v-if="cookieIcon.tooltip">
			<div style="max-width: 220px">{{ cookieIcon.tooltip }}</div>
		</template>
		<div>
			<CustomButton
				class="q-px-md"
				outline
				:disable="cookieList.length === 0"
				:loading="
					appAbilitiesStore.loading || collectStore.loading || pushLoading
				"
				@click="() => browserCookieStore.pushCookie()"
			>
				<template #label>
					<div class="row items-center flex-gap-xs no-wrap ellipsis">
						<img :src="cookieIcon.icon" style="height: 16px" />
						<span class="text-ink-1 text-subtitle3 ellipsis" style="flex: 1">
							{{ $t('bex.cookie') }}
						</span>
					</div>
				</template>
			</CustomButton>
		</div>
	</BtTooltip2>
</template>

<script setup lang="ts">
import cookieUploadIconDarkDark from 'src/assets/plugin/cookie-upload-dark.svg';
import cookieUploadedIconDark from 'src/assets/plugin/cookie-uploaded-dark.svg';
import cookieUploadedIconLight from 'src/assets/plugin/cookie-uploaded-white.svg';
import cookieExpiredfromIconLight from 'src/assets/plugin/cookie-expired-light.svg';
import cookieExpiredfromIconDark from 'src/assets/plugin/cookie-expired-dark.svg';
import cookieUploadIconLight from 'src/assets/plugin/cookie-upload.svg';
import { computed, onMounted, onUnmounted } from 'vue';
import { useQuasar } from 'quasar';
import { useBrowserCookieStore } from 'src/stores/settings/browserCookie';
import {
	createTabChangeListenerInCurrentWindow,
	getCurrentTabInfo
} from 'src/utils/bex/tabs';
import { browser } from 'src/platform/interface/bex/browser/target';
import { useI18n } from 'vue-i18n';
import BtTooltip2 from 'src/components/base/BtTooltip2.vue';
import CustomButton from 'src/pages/Plugin/components/CustomButton.vue';
import { useCookieStore } from 'src/stores/settings/cookie';
import { useUserStore } from 'src/stores/user';
import { useCollect } from 'src/composables/bex/useCollect';
import { COOKIE_LEVEL } from 'src/utils/rss-types';

const $q = useQuasar();
const { t } = useI18n();
const browserCookieStore = useBrowserCookieStore();
const cookieStore = useCookieStore();
const { cookieStatusCode, item, appAbilitiesStore, collectStore } =
	useCollect();

const cookieList = computed(() => browserCookieStore.cookieList);
const pushLoading = computed(() => browserCookieStore.pushLoading);
const getLoading = computed(() => cookieStore.loading);

const cookieUploadIcon = computed(() =>
	$q.dark.isActive ? cookieUploadIconDarkDark : cookieUploadIconLight
);

const cookieUploadedIcon = computed(() =>
	$q.dark.isActive ? cookieUploadedIconDark : cookieUploadedIconLight
);

const cookieExpiredfromIcon = computed(() =>
	$q.dark.isActive ? cookieExpiredfromIconDark : cookieExpiredfromIconLight
);

const cookieIcon = computed(() => {
	const icons = [
		cookieUploadIcon.value,
		cookieExpiredfromIcon.value,
		cookieUploadedIcon.value
	];

	const tooltip = [
		t('bex.cookie_upload_tooltip'),
		t('bex.cookie_expired_reupload')
	];
	return {
		icon: icons[cookieStatusCode.value],
		tooltip: tooltip[cookieStatusCode.value]
	};
});

let listener: any;

onMounted(async () => {
	const tab = await getCurrentTabInfo();
	const userStore = useUserStore();
	const url = userStore.getModuleSever('settings');
	if (userStore.current_user?.name) {
		browserCookieStore.init(
			tab,
			userStore.current_user?.name.split('@')[0],
			url
		);
	}
	listener = createTabChangeListenerInCurrentWindow(async (activeInfo) => {
		const tab2 = await browser.tabs.get(activeInfo.tabId);
		if (userStore.current_user?.name) {
			browserCookieStore.init(
				tab2,
				userStore.current_user?.name.split('@')[0],
				url
			);
		}
	});
});

onUnmounted(() => {
	listener && listener.remove();
});
</script>

<style lang="scss" scoped>
.message-container {
	border-radius: 12px;
}

.message-footer {
	text-decoration-line: underline;
	cursor: pointer;
}

::v-deep(.cookie-expansion-item .q-item) {
	padding-left: 4px;
	padding-right: 4px;
}
</style>
<style lang="scss" scoped>
.cookie-icon {
	width: 20px;
	height: 20px;
}

.submit-icon {
	position: absolute;
	bottom: 0;
	right: 0;
}
</style>
