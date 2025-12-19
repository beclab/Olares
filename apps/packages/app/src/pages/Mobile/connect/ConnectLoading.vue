<template>
	<div class="connect-loading-root">
		<div class="wizard-content">
			<div class="float-wrap" v-if="$q.platform.is.mobile">
				<div class="float" style="background: #fffdc1"></div>
				<div class="float" style="background: #fbffed"></div>
			</div>

			<div class="boot_justify">
				<animationPage
					:picture="waiting_waikuang_image"
					:certificate="waiting_image"
					:isAnimated="true"
				/>
			</div>
			<p class="wizard-content__detail ink-2" style="margin-top: 32px">
				{{ t('please_wait_a_monent_checking_the_status_of_the_olares') }}
			</p>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { onMounted, onBeforeUnmount } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useUserStore } from '../../../stores/user';
import { OlaresInfo, TerminusInfo } from '@bytetrade/core';
import { getTerminusInfo } from '../../../utils/BindTerminusBusiness';
import animationPage from './activate/animation.vue';
import './activate/wizard.scss';
import waiting_image from '../../../assets/wizard/waiting.png';
import waiting_waikuang_image from '../../../assets/wizard/waiting_waikuang.png';
import { useI18n } from 'vue-i18n';
import { StatusBar } from '@capacitor/status-bar';
import { busEmit } from '../../../utils/bus';
import { getAppPlatform } from 'src/application/platform';
import { UserItem } from '@didvault/sdk/src/core';
import TerminusDesktopTipDialog from '../../../components/dialog/TerminusDesktopTipDialog.vue';
import AndroidPlugins from 'src/platform/interface/capacitor/android/androidPlugins';

const { t } = useI18n();

const userStore = useUserStore();
const router = useRouter();
const $q = useQuasar();

onMounted(async () => {
	if ($q.platform.is.nativeMobile && $q.platform.is.android) {
		// StatusBar.setOverlaysWebView({ overlay: true });
		await AndroidPlugins.EdgeToEdge.disable();
	}

	const user = userStore.current_user;
	if (user) {
		if (user.setup_finished) {
			busEmit('account_update');
			if (process.env.PLATFORM == 'DESKTOP' || getAppPlatform().isPad) {
				router.replace('/Files/Home/');
			} else if (process.env.APPLICATION == 'VAULT') {
				router.replace('/items');
			} else {
				router.replace({ path: '/home' });
			}
		} else {
			if (!user.wizard || user.wizard == '') {
				router.replace({ path: '/Activate' });
			} else {
				connectUser(user);
			}
		}
	} else {
		router.replace({ path: '/home' });
	}
});

const connectUser = async (user: UserItem) => {
	const info: OlaresInfo | null = await getTerminusInfo(user); //terminus_name
	if (process.env.PLATFORM != 'MOBILE') {
		if (!info || !info.wizardStatus || info.wizardStatus !== 'completed') {
			$q.dialog({
				component: TerminusDesktopTipDialog,
				componentProps: {
					title: t('tips'),
					message: t('user_current_status.reactivation.message'),
					isReactive: false,
					btnTitle: t('ignore_it'),
					confirmBtnTitle: t('i_got_it')
				}
			});
		}
		router.replace({ path: '/ConnectTerminus' });
		return;
	}
	if (info && info.wizardStatus == 'completed') {
		router.replace({ path: '/ConnectTerminus' });
	} else {
		router.replace({ path: '/Activate' });
	}
};

onBeforeUnmount(async () => {
	if ($q.platform.is.android) {
		// StatusBar.setOverlaysWebView({ overlay: false });
		await AndroidPlugins.EdgeToEdge.enable();
	}
});
</script>

<style lang="scss" scoped>
.connect-loading-root {
	width: 100%;
	height: 100%;
	background: $background-2;
}
</style>
