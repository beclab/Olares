<template>
	<div class="welcome-page column justify-start items-center">
		<div class="welcome-carousel-parent coloum justify-start items-start">
			<q-carousel
				class="welcome-carousel"
				animated
				v-model="slide"
				swipeable
				transition-prev="slide-right"
				transition-next="slide-left"
			>
				<q-carousel-slide
					:name="1"
					img-src="../../../assets/login/termipass_mobile_slide_img.svg"
				/>
				<q-carousel-slide
					:name="2"
					img-src="../../../assets/login/termipass_mobile_slide_img.svg"
				/>
				<q-carousel-slide
					:name="3"
					img-src="../../../assets/login/termipass_mobile_slide_img.svg"
				/>
				<q-carousel-slide
					:name="4"
					img-src="../../../assets/login/termipass_mobile_slide_img.svg"
				/>
			</q-carousel>
			<div class="welcome-data text-h4">{{ carouselData }}</div>
			<div class="welcome-navigation row justify-start items-center">
				<template v-for="index in 4" :key="index">
					<div
						:class="index === slide ? 'navigation-active' : 'navigation-normal'"
						@click="changeActive(index)"
					/>
				</template>
			</div>

			<div class="welcome-control row justify-between items-center">
				<span @click="onSkip">{{ t('skip') }}</span>
				<div
					class="welcome-next text-grey-10 column justify-center items-center"
				>
					<q-icon size="20px" name="sym_r_arrow_forward_ios" @click="onNext" />
				</div>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { useRouter } from 'vue-router';
import { computed, ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { saveDefaultPassword } from '../../../utils/UnlockBusiness';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import TerminusTipDialog from '../../../components/dialog/TerminusTipDialog.vue';
import { useQuasar } from 'quasar';
import { displayAppServices, appServices } from '../../../utils/platform';
import AndroidPlugins from 'src/platform/interface/capacitor/android/androidPlugins';
import { getNativeAppPlatform } from 'src/application/platform';

const slide = ref(1);
const { t } = useI18n();
const router = useRouter();
const $q = useQuasar();

const carouselData = computed(() => {
	switch (slide.value) {
		case 1:
			return t('System state at a glance');
		case 2:
			return t('Secure, encrypted connection');
		case 3:
			return t('Password management with end-to-end encryption');
		case 4:
			return t('Portable, unified file system for your data');
		default:
			return '';
	}
});

const changeActive = (index: number) => {
	slide.value = index;
};

const onNext = () => {
	if (slide.value < 4) {
		slide.value++;
	} else {
		onSkip();
	}
};

const onSkip = async () => {
	await saveDefaultPassword({
		async onSuccess() {
			if ($q.platform.is.nativeMobile && $q.platform.is.android) {
				AndroidPlugins.AndroidUniversal.setAcceptedPrivacyTerms({
					accepted: true
				});
			}
			const jumpToRegisterDid = () => {
				if (process.env.IS_BEX) {
					router.replace({ path: '/import_mnemonic' });
					return;
				}
				router.replace({
					name: 'setupSuccess'
				});
			};
			jumpToRegisterDid();
		},
		onFailure(message: string) {
			notifyFailed(message);
		}
	});
};

onMounted(() => {
	if ($q.platform.is.nativeMobile && displayAppServices()) {
		$q.dialog({
			component: TerminusTipDialog,
			componentProps: {
				title: t('Terms of Use and Privacy Policy'),
				position: t('Agree'),
				message: t(
					'Thank you for using LarePass. Before getting started, please take a moment to read our {ServiceAgreement} and {PrivacyPolicy}, which outline our services and how we handle your personal information. By clicking “Agree,” you accept these terms.',
					{
						ServiceAgreement: `<a class='text-blue-4' href='${
							appServices().serviceAgreement
						}' target='_blank'>《${t('Service Agreement')}》</a>`,
						PrivacyPolicy: `<a class='text-blue-4' href='${
							appServices().privacyPolicy
						}' target='_blank'>《${t('Privacy Policy')}》</a>`
					}
				)
			}
		}).onCancel(() => {
			getNativeAppPlatform().finished();
		});
	}
});
</script>

<style lang="scss" scoped>
.welcome-page {
	width: 100%;
	height: 100%;

	.welcome-carousel-parent {
		width: 100%;
		height: 100%;
		position: relative;

		.welcome-carousel {
			width: 100%;
			height: calc(100% - 222px);
		}

		.welcome-data {
			color: $ink-1;
			position: absolute;
			bottom: 239px;
			width: calc(100% - 64px);
			left: 32px;
			right: 32px;
		}

		.welcome-navigation {
			width: calc(100% - 32px);
			margin-left: 32px;
			margin-top: 15px;
			height: 4px;

			.navigation-base {
				height: 4px;
				border-radius: 20px;
				margin-right: 8px;
			}

			.navigation-active {
				@extend .navigation-base;
				width: 20px;
				background: $yellow;
			}

			.navigation-normal {
				@extend .navigation-base;
				width: 4px;
				background: $grey-5;
			}
		}

		.welcome-control {
			width: calc(100% - 64px);
			margin-left: 32px;
			margin-top: 67px;

			.welcome-next {
				border-radius: 12px;
				width: 56px;
				height: 56px;
				background-color: $yellow;
			}
		}
	}
}
</style>
