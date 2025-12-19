<template>
	<bt-scroll-area class="scroll-area-height">
		<div class="q-mt-lg q-mb-md column items-center justify-center">
			<setting-avatar :size="72" />
			<div class="text-h4 text-ink-1 q-mt-md">
				{{ adminStore.user.name }}
			</div>
			<div class="text-ink-3 text-body1 q-mt-xs">
				{{ '@' + adminStore.olaresId.split('@')[1] }}
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
							{{ t('current_version') }}
						</div>
					</div>
				</template>

				<div class="row items-center justify-end">
					<div>
						{{ adminStore.terminus.osVersion }}
					</div>
				</div>
			</bt-form-item>
		</BtList>

		<BtList>
			<bt-form-item
				:title="t('Communication and Feedback')"
				:chevron-right="true"
				@click="gotoForum"
				:width-separator="false"
			/>
		</BtList>

		<BtList
			class="q-mb-lg"
			v-if="adminStore.isAdmin || !isDemo"
			:label="t('device')"
		>
			<device-item
				v-for="(device, index) in adminStore.devices"
				:key="'device' + index"
				:device="device"
				:is-latest="index + 1 == adminStore.devices.length"
				:is-first="index == 0"
			/>
		</BtList>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import UpdateUserPassworDialog from '../Account/dialog/UpdateUserPassworDialog.vue';
import SettingAvatar from 'src/components/settings/base/SettingAvatar.vue';
import DeviceItem from 'src/components/settings/person/DeviceItem.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useAdminStore } from 'src/stores/settings/admin';
import { useUserStore } from 'src/stores/settings/user';
import { AccountInfo } from 'src/constant/global';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { checkDomainSuffix } from 'src/constant';
import { useTerminusDStore } from 'src/stores/settings/terminusd';

const adminStore = useAdminStore();
const userStore = useUserStore();
const router = useRouter();
const quasar = useQuasar();
const { t } = useI18n();
const terminusDStore = useTerminusDStore();
const userInfo = ref<AccountInfo | undefined>();
const isDemo = computed(() => {
	return !!process.env.DEMO;
});

onMounted(async () => {
	await userStore.get_accounts();
	terminusDStore.system_status();
	userInfo.value = await userStore.get_account_info(adminStore.user.name);
});

function gotoPage(path: string) {
	router.push({ path });
}

const updatePassword = async () => {
	quasar.dialog({
		component: UpdateUserPassworDialog,
		componentProps: {
			name: adminStore.user.name
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
	checkDomainSuffix('olares.cn').then((isOlaresCN: boolean) => {
		if (isOlaresCN) {
			window.open('https://forum.olares.cn/');
		} else {
			window.open('https://forum.olares.com/');
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
</style>
