<template>
	<div class="account-root">
		<div class="flur1"></div>
		<div class="flur2"></div>
		<terminus-title-bar
			right-icon="sym_r_more_horiz"
			@on-right-click="moreClick"
		/>
		<terminus-scroll-area class="account-scroll">
			<template v-slot:content>
				<q-list class="accountCentent">
					<div class="column items-center">
						<div class="accountCentent__avatar">
							<TerminusAvatar :info="userStore.terminusInfo()" :size="64" />
						</div>

						<div
							class="terminus-text-ellipsis accountCentent__name q-mt-xs text-h5"
						>
							{{ userStore.current_user?.local_name }}
						</div>
						<div
							class="terminus-text-ellipsis accountCentent__info text-body3 q-mt-xs"
						>
							@{{ userStore.current_user?.domain_name }}
						</div>
						<terminus-user-status class="q-mt-xs" />
					</div>

					<div class="row items-center justify-between q-mt-xl">
						<div
							class="accountCentent__action_base q-px-md q-pt-md q-pb-lg column justify-between"
							:class="
								!$q.dark.isActive
									? 'accountCentent__action_offline'
									: 'bg-background-1'
							"
						>
							<div class="row items-center justify-between">
								<div class="icon-bg row items-center justify-center">
									<q-icon name="sym_r_public_off" size="20px" />
								</div>
								<bt-switch
									size="sm"
									truthy-track-color="light-blue-default"
									v-model="offLineModeRef"
									@update:model-value="updateOffLineMode"
								/>
							</div>
							<div class="title text-subtitle2">
								{{
									offLineModeRef
										? t('user_current_status.offline_mode.enabled')
										: t('user_current_status.offline_mode.disabled')
								}}
							</div>
						</div>
						<div
							class="accountCentent__action_base q-px-md q-px-md q-pt-md q-pb-lg column justify-between"
							:class="
								!$q.dark.isActive
									? 'accountCentent__action_vpn'
									: 'bg-background-1'
							"
						>
							<div class="row items-center justify-between">
								<div class="icon-bg row items-center justify-center">
									<q-icon
										:name="TermiPassVpnStatusInfo[scaleStore.vpnStatus].icon"
										size="20px"
										:style="`color:${
											TermiPassVpnStatusInfo[scaleStore.vpnStatus].color
										};`"
									/>
								</div>
								<bt-switch
									size="sm"
									truthy-track-color="light-blue-default"
									v-model="vpnToggleStatus"
								/>
							</div>
							<div class="title text-subtitle2">
								{{
									t(TermiPassVpnStatusInfo[scaleStore.vpnStatus].description)
								}}
							</div>
						</div>
					</div>

					<terminus-item
						v-if="isBex || userStore.currentUserBackup"
						:show-board="true"
						img-bg-classes="bg-yellow"
						:item-height="88"
						icon-name="sym_r_fact_check"
						:whole-picture-size="48"
						class="q-mt-lg"
						@click="startBackUp"
					>
						<template v-slot:title> {{ t('mnemonic_phrase') }}</template>
						<template v-slot:detail>
							{{ t('view_your_mnemonic_phrase') }}
						</template>
						<template v-slot:side>
							<q-icon
								name="sym_r_keyboard_arrow_right"
								size="20px"
								color="ink-3"
							/>
						</template>
					</terminus-item>

					<div
						class="accountCentent__backup-mneminic q-mt-lg q-pl-lg q-py-lg row items-center justify-between"
						:class="
							!$q.dark.isActive ? 'backup-mnemonic-bg' : 'bg-background-1'
						"
						v-else
					>
						<div class="accountCentent__backup-mneminic__introduce">
							<div class="title text-h5">
								{{ t('please_backup_your_mnemonic_phrase') }}
							</div>
							<div class="detail text-body2">
								{{ t('back_up_your_mnemonic_phrase_immediately_to_safe') }}
							</div>
							<div
								class="backup text-subtitle3 q-mt-md row items-center justify-center"
								@click="startBackUp"
							>
								{{ t('start_backup') }}
							</div>
						</div>
						<div class="accountCentent__backup-mneminic__image">
							<img src="../../../assets/account/backup_mnemonic_image.svg" />
						</div>
					</div>

					<terminus-item
						v-if="!isBex"
						:show-board="true"
						img-bg-classes="bg-yellow"
						:item-height="88"
						icon-name="sym_r_fact_check"
						icon-color="grey-10"
						:whole-picture-size="48"
						class="q-mt-lg"
						@click="enterVCManagement"
					>
						<template v-slot:title> {{ t('vc_management') }}</template>
						<template v-slot:detail>
							{{ t('bind_your_social_account') }}
						</template>
						<template v-slot:side>
							<q-icon
								name="sym_r_keyboard_arrow_right"
								size="20px"
								color="ink-3"
							/>
						</template>
					</terminus-item>

					<!-- <terminus-item
						v-if="showCheckHistory"
						:show-board="true"
						img-bg-classes="bg-yellow"
						icon-color="grey-10"
						:item-height="88"
						icon-name="sym_r_fact_check"
						:whole-picture-size="48"
						class="q-mt-lg"
						@click="enterCheckHistory"
					>
						<template v-slot:title> CheckHistory </template>
						<template v-slot:side>
							<q-icon
								name="sym_r_keyboard_arrow_right"
								size="20px"
								color="ink-3"
							/>
						</template>
					</terminus-item> -->
				</q-list>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useUserStore } from '../../../stores/user';
import { useQuasar } from 'quasar';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import { useRouter } from 'vue-router';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import TerminusUserStatus from '../../../components/common/TerminusUserStatus.vue';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import { TermiPassVpnStatusInfo } from '../../../platform/terminusCommon/terminusCommonInterface';
import { useScaleStore } from '../../../stores/scale';
import { watch } from 'vue';
import AccoutMoreDialog from './AccoutMoreDialog.vue';
import { useI18n } from 'vue-i18n';
import { busEmit } from '../../../utils/bus';

const $router = useRouter();
const userStore = useUserStore();
const $q = useQuasar();
const scaleStore = useScaleStore();

const isBex = ref(process.env.IS_BEX);

const { t } = useI18n();

const startBackUp = async () => {
	if (!(await userStore.unlockFirst(undefined, { hide: true }))) {
		return;
	}
	if (!userStore.passwordReseted) {
		busEmit('configPassword');
		return;
	}
	$router.push({
		path: '/backup_mnemonics'
	});
};

const enterVCManagement = () => {
	// $router.push({ path: '/vc_manage' });
	$router.push({
		path: '/vc_manage'
	});
};

const offLineModeRef = ref(userStore.current_user?.offline_mode || false);

const updateOffLineMode = async () => {
	userStore.updateOfflineMode(offLineModeRef.value);
};

const moreClick = () => {
	$q.dialog({
		component: AccoutMoreDialog,
		componentProps: {}
	});
};

const vpnToggleStatus = computed({
	get: () => scaleStore.isOn || scaleStore.isConnecting,
	set: async (value) => {
		if (scaleStore.isDisconnecting) {
			return;
		}
		if (value) {
			await scaleStore.start();
		} else {
			await scaleStore.stop();
		}
	}
});

// const enterCheckHistory = () => {
// 	$router.push({
// 		path: isBex.value ? `${route.meta.root}/checkHistory` : '/checkHistory'
// 	});
// };

// const onLoginCloud = () => {
// 	$router.push({ path: '/space_management' });
// };
</script>

<style scoped lang="scss">
.account-root {
	width: 100%;
	height: 100%;
	position: relative;
	z-index: 0;
}

.flur1 {
	width: 135px;
	height: 135px;
	background: rgba(133, 211, 255, 0.3);
	filter: blur(70px);
	position: absolute;
	right: 20vw;
	top: 10vh;
	z-index: -1;
}

.flur2 {
	width: 110px;
	height: 110px;
	background: rgba(217, 255, 109, 0.2);
	filter: blur(70px);
	position: absolute;
	left: 20vw;
	top: 20vh;
	z-index: -1;
}

.account-scroll {
	width: 100%;
	height: calc(100% - 56px);
}

.accountCentent {
	width: 100%;
	padding-left: 20px;
	padding-right: 20px;

	&__avatar {
		width: 64px;
		height: 64px;
		border-radius: 32px;
		overflow: hidden;
	}

	&__name {
		color: $ink-1;
		max-width: calc(100% - 40px);
	}

	&__info {
		text-align: center;
		color: $ink-2;
	}

	&__action_base {
		border: 1px solid $separator;
		border-radius: 16px;
		width: calc(50% - 10px);
		height: 128px;

		.icon-bg {
			width: 32px;
			height: 32px;
			overflow: hidden;
			border-radius: 16px;
			background-color: $background-3;
			color: $ink-2;
		}

		.title {
			color: $ink-1;
		}
	}

	&__action_offline {
		background: linear-gradient(
			127.05deg,
			#fffef7 4.41%,
			rgba(249, 254, 199, 0.5) 49.41%,
			rgba(243, 254, 194, 0.5) 84.8%
		);
	}

	&__action_offline_dark {
		background: linear-gradient(127.05deg, #4f462f 4.41%, #2e311f 84.8%);
	}

	&__action_vpn {
		background: linear-gradient(
			125.16deg,
			#fcfff7 4.57%,
			rgba(216, 255, 222, 0.5) 51.18%,
			rgba(226, 255, 219, 0.5) 87.85%
		);
	}

	&__action_vpn_dark {
		background: linear-gradient(125.16deg, #243f2d 4.57%, #172814 87.85%);
	}

	&__backup-mneminic {
		border: 1px solid $separator;
		width: 100%;
		border-radius: 16px;
		position: relative;
		min-height: 190px;

		&__introduce {
			width: calc(100% - 135px);
			height: 100%;

			.title {
				color: $ink-1;
			}

			.detail {
				color: $ink-2;
			}

			.backup {
				background: $yellow;
				border-radius: 8px;
				height: 32px;
				text-align: center;
				color: $grey-10;
				width: 93px;
			}
		}

		&__image {
			width: 127px;
			position: absolute;
			right: 15px;
			bottom: 20px;
		}
	}

	.backup-mnemonic-bg {
		background: linear-gradient(
			127.05deg,
			#fffef7 4.41%,
			rgba(249, 254, 199, 0.5) 49.41%,
			rgba(243, 254, 194, 0.5) 84.8%
		);
	}
}
</style>
