<template>
	<page-title-component :show-back="true" v-if="deviceStore.isMobile" />
	<bt-scroll-area
		class="scroll-area-height"
		:class="deviceStore.isMobile ? 'scroll-area-height-mobile' : ''"
	>
		<div class="q-mt-lg q-mb-md column items-center justify-center">
			<setting-avatar :size="72" />
			<div class="text-h4 text-ink-1 q-mt-md">
				{{ meStore.user.name }}
			</div>
			<div class="text-ink-3 text-body1 q-mt-xs">
				{{ '@' + terminusStore.olaresId.split('@')[1] }}
			</div>
		</div>
		<BtList>
			<bt-form-item
				v-if="terminusDStore.olaresInfo"
				:title="t('My hardware')"
				:description="t('View and manage your Olares hardware')"
				:chevron-right="true"
				@click="gotoPage('/hardware')"
			/>
			<bt-form-item
				:title="t('olares_space')"
				:description="t('Check your subscribed plan and usage on Olares Space')"
				:chevron-right="true"
				@click="gotoPage('/olares_space')"
			/>
			<bt-form-item
				@click="updatePassword"
				:title="t('change_password')"
				:description="t('Change your password for Olares.')"
				:chevron-right="true"
			/>
			<bt-form-item
				@click="gotoPage('/authority')"
				:title="t('Set network access policy')"
				:description="
					t('Manage who can connect to your Olares services and how.')
				"
				:chevron-right="true"
			/>
			<!--			<bt-form-item-->
			<!--				v-if="userInfo?.wizard_complete"-->
			<!--				@click="goLoginHistory"-->
			<!--				:title="t('view_login_history')"-->
			<!--				:chevron-right="true"-->
			<!--			/>-->
			<bt-form-item :width-separator="false">
				<template v-slot:title>
					<div class="q-my-lg">
						<div class="row items-center">
							{{ t('olares_os_version') }}
						</div>
					</div>
				</template>

				<div class="row items-center justify-end">
					<div>
						{{ terminusStore.terminusInfo.osVersion }}
					</div>
				</div>
			</bt-form-item>
		</BtList>

		<BtList>
			<bt-form-item
				:title="t('Communication and feedback')"
				:chevron-right="true"
				@click="gotoForum"
			/>

			<bt-form-item
				:title="t('Acknowledgements')"
				:chevron-right="true"
				@click="gotoAcknowledgment"
				:width-separator="false"
			/>
		</BtList>

		<BtList
			class="q-mb-lg"
			v-if="meStore.isAdmin || !isDemo"
			:label="t('devices')"
		>
			<device-item
				v-for="(device, index) in devicesStore.devices"
				:key="'device' + index"
				:device="device"
				:is-latest="index + 1 == devicesStore.devices.length"
				:is-first="index == 0"
			/>
		</BtList>
		<ListBottomFuncBtn
			@funcClick="signOut"
			:justify-end="false"
			:title="t('Sign out')"
			style="margin-bottom: 20px"
		/>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import UpdateUserPassworDialog from '../Account/dialog/UpdateUserPassworDialog.vue';
import SettingAvatar from 'src/components/settings/base/SettingAvatar.vue';
import DeviceItem from 'src/components/settings/person/DeviceItem.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useMeStore } from 'src/stores/settings/me';
import { useUsersStore } from 'src/stores/settings/users';
import { useTerminusStore } from 'src/stores/terminus';
import { useDevicesStore } from 'src/stores/settings/devices';
import { AccountInfo } from 'src/constant/global';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { checkDomainSuffix } from 'src/constant';
import { useTerminusDStore } from 'src/stores/settings/terminusd';
import ListBottomFuncBtn from 'src/components/settings/ListBottomFuncBtn.vue';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import { notifyFailed } from 'src/utils/settings/btNotify';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { useDeviceStore } from 'src/stores/settings/device';

const meStore = useMeStore();
const usersStore = useUsersStore();
const terminusStore = useTerminusStore();
const devicesStore = useDevicesStore();
const router = useRouter();
const deviceStore = useDeviceStore();
const quasar = useQuasar();
const { t } = useI18n();
const terminusDStore = useTerminusDStore();
const userInfo = ref<AccountInfo | undefined>();
const $q = useQuasar();
const isDemo = computed(() => {
	return !!process.env.DEMO;
});

onMounted(async () => {
	await usersStore.get_accounts();
	terminusDStore.system_status();
	userInfo.value = await usersStore.get_account_info(meStore.user.name);
});

function gotoPage(path: string) {
	router.push({ path });
}

const updatePassword = async () => {
	quasar.dialog({
		component: UpdateUserPassworDialog,
		componentProps: {
			name: meStore.user.name
		}
	});
};

const goLoginHistory = () => {
	router.push({
		name: 'loginHistory',
		params: {
			name: userInfo.value?.name
		}
	});
};

const gotoForum = () => {
	// checkDomainSuffix('olares.cn').then((isOlaresCN: boolean) => {
	// 	if (isOlaresCN) {
	// 		window.open('https://forum.olares.cn/');
	// 	} else {
	// 		window.open('https://forum.olares.com/');
	// 	}
	// });
	router.push('/feedback');
};

const gotoAcknowledgment = () => {
	// checkDomainSuffix('olares.cn').then((isOlaresCN: boolean) => {
	// 	if (isOlaresCN) {
	// 		window.open('https://forum.olares.cn/');
	// 	} else {
	// 		window.open('https://forum.olares.com/');
	// 	}
	// });
	router.push('/acknowledgment');
};

const signOut = () => {
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('Sign out'),
			message: t('Your changes may not be saved.'),
			useCancel: true,
			confirmText: t('confirm'),
			cancelText: t('cancel')
		}
	}).onOk(async () => {
		try {
			$q.loading.show();
			await meStore.logout();
			const auth_url = terminusStore.getAuthURL() + '?logout=1';
			if (window.top == window) {
				window.location.replace(auth_url);
			} else {
				window.parent.postMessage(
					{
						message: 'sign_out'
					},
					'*'
				);
			}
		} catch (err) {
			notifyFailed((err as Error).message);
		} finally {
			$q.loading.hide();
		}
	});
};
</script>

<style scoped lang="scss">
.scroll-area-height {
	width: 100%;
	height: 100%;
	max-height: 100%;
	padding-left: 20px;
	padding-right: 20px;
}

.scroll-area-height-mobile {
	height: calc(100% - 56px);
}
</style>
