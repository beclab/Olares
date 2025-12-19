<template>
	<div class="active-terminus-root">
		<terminus-wizard-view
			:btn-title="t('scan_the_qr_code')"
			:bottom-more-height="0"
			:enable-btn="false"
			btn-icon="sym_r_qr_code_scanner"
		>
			<template v-slot:content>
				<div class="active-terminus-root__content column items-center">
					<div
						class="active-terminus-root__content__user q-mt-md column items-center justify-center"
					>
						<div class="active-terminus-root__content__user__image">
							<TerminusAvatar :info="userStore.terminusInfo()" :size="64" />
						</div>

						<div
							class="text-h6 text-ink-1 terminus-text-ellipsis q-mt-xs active-terminus-root__content__user__hide"
						>
							{{ userStore.current_user?.local_name }}
						</div>
						<div
							class="text-body2 text-ink-2 q-mt-xs terminus-text-ellipsis active-terminus-root__content__user__hide"
						>
							{{ userStore.current_user?.name }}
						</div>
						<div
							v-if="isCreate"
							class="active-terminus-root__content__user__success row items-center justify-center q-mt-sm"
						>
							<q-icon name="sym_r_check_circle" size="16px" color="positive" />
							<div class="text-overline text-ink-1 q-ml-xs">
								{{ t('Olares ID successfully created') }}
							</div>
						</div>
					</div>
					<div
						class="active-terminus-root__content__part_bg row items-center justify-between q-pa-md q-mt-md"
						:class="
							!$q.dark.isActive ? 'network-scan-local-bg' : 'bg-background-1'
						"
					>
						<div class="active-terminus-root__content__part_bg__introduce">
							<div class="title text-body2 text-ink-2">
								{{
									t(
										'Have an Olares in your LAN and ready to activate it as an admin?'
									)
								}}
							</div>
							<div
								class="text-subtitle3 q-mt-md row items-center justify-center detail"
								@click="discoverTerminus"
							>
								{{ t('Discover nearby Olares') }}
							</div>
						</div>
						<div class="active-terminus-root__content__part_bg__scan_local">
							<div class="active-terminus-root__content__scan_local__image">
								<img src="../../../../assets/wizard/local_network_scan.svg" />
							</div>
						</div>
					</div>

					<div
						class="active-terminus-root__content__part_bg row items-center justify-between q-pa-md q-mt-lg"
						:class="!$q.dark.isActive ? 'scan-qrcode-bg' : 'bg-background-1'"
					>
						<div class="active-terminus-root__content__part_bg__introduce">
							<div class="title text-body2 text-ink-2">
								{{
									t(
										'Ready to activate Olares on the Wizard page, or log in to Olares Space?'
									)
								}}
							</div>
							<div
								class="text-subtitle3 q-mt-md row items-center justify-center scan-qrcode"
								@click="goToScanPage"
							>
								{{ t('scan_qr_code') }}
							</div>
						</div>
						<div class="active-terminus-root__content__part_bg__scan_local">
							<div class="active-terminus-root__content__scan_local__image">
								<img src="../../../../assets/wizard/scan_qrcode.svg" />
							</div>
						</div>
					</div>
				</div>
			</template>
			<template v-slot:bottom-bottom>
				<TerminusExportMnemonicRoot />
			</template>
		</terminus-wizard-view>
		<div
			class="active-terminus-root__img row items-center justify-center"
			:style="`top:${
				$q.platform.is.ios ? 'calc(env(safe-area-inset-top) + 20px);' : '57px'
			}`"
		>
			<TerminusChangeUserHeader :scan="false">
				<template v-slot:avatar>
					<q-icon name="account_circle" size="24px" color="grey-8" />
				</template>
			</TerminusChangeUserHeader>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useUserStore } from '../../../../stores/user';
import { useRouter, useRoute } from 'vue-router';
import { useQuasar } from 'quasar';
import { UserItem } from '@didvault/sdk/src/core';
import TerminusWizardView from '../../../../components/common/TerminusWizardView.vue';
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { getTerminusInfo } from '../../../../utils/BindTerminusBusiness';
import { OlaresInfo, TerminusInfo } from '@bytetrade/core';
import TerminusExportMnemonicRoot from '../../../../components/common/TerminusExportMnemonicRoot.vue';
import TerminusChangeUserHeader from '../../../../components/common/TerminusChangeUserHeader.vue';
import UserStatusCommonDialog from '../../../../components/userStatusDialog/UserStatusCommonDialog.vue';

import { useI18n } from 'vue-i18n';
import { busEmit } from '../../../../utils/bus';
const { t } = useI18n();

const userStore = useUserStore();
const router = useRouter();
const route = useRoute();
const $q = useQuasar();

const disableLeave = route.params.disableLeave;
const isCreate = ref(userStore.isNewCreateUser);
userStore.isNewCreateUser = false;

const user: UserItem = userStore.users!.items.get(userStore.current_id!)!;
if (!disableLeave) {
	if (user.terminus_activate_status == 'completed' || user.setup_finished) {
		router.push({ path: '/home' });
	} else if (user.terminus_activate_status == 'wait_reset_password') {
		router.push({ path: '/ResetPassword' });
	} else if (
		user.terminus_activate_status == 'wait_activate_vault' ||
		user.terminus_activate_status == 'vault_activating' ||
		user.terminus_activate_status == 'vault_activate_failed'
	) {
		// do nothing
	} else {
		// if (!) {
		// 	return;
		// }
		// if ()
		userStore.unlockFirst().then((status) => {
			if (!status) {
				return;
			}
			router.push({ path: 'activateWizard' });
		});
	}
}

async function goToScanPage() {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	if (userStore.current_user?.access_token) {
		userStore.current_user.access_token = '';
	}
	if (process.env.IS_PC_TEST) {
		router.push({ path: '/scan_local' });
	} else {
		if ($q.platform.is.nativeMobile) {
			router.push({ path: '/scan' });
		} else {
			router.push({ path: '/scan_local' });
		}
	}
}

let timer: any | undefined = undefined;
let curCount = 0;
onMounted(async () => {
	if (!userStore.current_user?.setup_finished && !isCreate.value) {
		//
		timer = setInterval(async () => {
			const info: OlaresInfo | null = await getTerminusInfo(user); //terminus_name
			curCount = curCount + 1;
			if (curCount > 5) {
				clearInterval(timer);
			}
			if (info && info.wizardStatus == 'completed') {
				clearInterval(timer);
				router.replace({ path: '/ConnectTerminus' });
			}
		}, 5000);
	}

	if (isCreate.value) {
		if (!userStore.passwordReseted) {
			busEmit('configPassword');
			return;
		}
		if (!userStore.currentUserBackup) {
			$q.dialog({
				component: UserStatusCommonDialog,
				componentProps: {
					title: t('Security Tips'),
					message: t(
						'Back up your mnemonic phase to ensure the security of your account and data.'
					),
					addSkip: true,
					btnTitle: t('start_backup'),
					skipTitle: t('Do it later')
				}
			}).onOk(() => {
				router.push({
					path: '/backup_mnemonics',
					query: {
						backup: 1
					}
				});
			});
		}
	}
});

onBeforeUnmount(() => {
	if (timer) {
		clearInterval(timer);
		timer = undefined;
	}
});

const discoverTerminus = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	router.push({ path: '/discover/machine' });
};
</script>

<style lang="scss" scoped>
.active-terminus-root {
	width: 100%;
	height: 100%;
	position: relative;

	&__img {
		width: 100%;
		height: 40px;
		position: absolute;
		right: 0px;
		border-radius: 16px;
		overflow: hidden;
	}

	&__content {
		width: 100%;
		height: 100%;

		&__image {
			width: 124px;
			height: 124px;
		}

		&__reminder {
			text-align: center;
			color: $ink-2;

			&__detail {
				color: $blue;
				text-decoration: underline;
				cursor: pointer;
			}
		}

		&__user {
			// border: 1px solid $separator;
			width: 100%;
			border-radius: 12px;
			padding-top: 20px;
			padding-bottom: 20px;
			margin-top: 68px;

			&__hide {
				max-width: calc(100% - 40px);
			}

			&__image {
				width: 64px;
				height: 64px;
				border-radius: 32px;
				overflow: hidden;
			}

			&__success {
				border: 1px solid $separator;
				height: 28px;
				border-radius: 14px;
				padding: 0px 8px;
			}
		}

		&__part_bg {
			border: 1px solid $separator;
			width: 100%;
			border-radius: 16px;
			position: relative;

			&__introduce {
				width: calc(100% - 115px);
				height: 100%;

				.detail {
					background: $yellow;
					border-radius: 8px;
					height: 32px;
					text-align: center;
					color: $grey-10;
					padding: 0 10px;
					bottom: 20px;
					line-height: 32px;
					display: inline-block;
				}
				.scan-qrcode {
					background: var(---background11, #252525);
					border-radius: 8px;
					height: 32px;
					text-align: center;
					color: #fff;
					padding: 0 10px;
					bottom: 20px;
					line-height: 32px;
					display: inline-block;
				}
			}
			&__scan_local {
				height: 92px;
				width: 92px;
				&__image {
					height: 100%;
				}
			}
		}

		.network-scan-local-bg {
			background: linear-gradient(
				127.05deg,
				#fffef7 4.41%,
				rgba(249, 254, 199, 0.5) 49.41%,
				rgba(243, 254, 194, 0.5) 84.8%
			);
		}
		.scan-qrcode-bg {
			background: linear-gradient(
				125.16deg,
				#fcfff7 4.57%,
				rgba(216, 255, 222, 0.5) 51.18%,
				rgba(226, 255, 219, 0.5) 87.85%
			);
		}
	}
}
</style>
