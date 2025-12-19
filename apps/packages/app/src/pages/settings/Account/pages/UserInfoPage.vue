<template>
	<page-title-component :show-back="true" :title="t('account_info')" />
	<bt-scroll-area class="nav-height-scroll-area-conf" v-if="userInfo">
		<div
			v-show="usage"
			class="text-ink-1 resource-title"
			:class="deviceStore.isMobile ? 'text-subtitle1-m' : 'text-subtitle1'"
		>
			{{ t('resource_usage') }}
		</div>
		<adaptive-layout v-show="usage">
			<template v-slot:pc>
				<div
					class="row justify-between"
					style="
						grid-column-gap: 20px;
						display: grid;
						grid-template-columns: repeat(2, minmax(0, 1fr));
					"
				>
					<resource-limit
						:total="usage?.user_cpu_total"
						:usage="usage?.user_cpu_usage"
						:label="t('cpu_core')"
						unit-key="cpu"
					/>
					<resource-limit
						:total="usage?.user_memory_total"
						:usage="usage?.user_memory_usage"
						:label="t('memory_gi')"
						unit-key="memory"
					/>
				</div>
			</template>
			<template v-slot:mobile>
				<div>
					<resource-limit
						:total="usage?.user_cpu_total"
						:usage="usage?.user_cpu_usage"
						:label="t('cpu_core')"
						unit-key="cpu"
					/>
					<resource-limit
						class="q-mt-lg"
						:total="usage?.user_memory_total"
						:usage="usage?.user_memory_usage"
						:label="t('memory_gi')"
						unit-key="memory"
					/>
				</div>
			</template>
		</adaptive-layout>

		<div
			class="text-ink-1"
			:class="
				deviceStore.isMobile
					? 'text-subtitle1-m q-mt-lg'
					: 'text-subtitle1 q-mt-md'
			"
		>
			{{ t('info') }}
		</div>

		<bt-list>
			<bt-form-item :title="t('profile_avatar')" :margin-top="false">
				<q-avatar :size="`40px`">
					<TerminusAvatar
						:info="{
							terminusName: userInfo?.terminusName,
							avatar: userInfo?.avatar
						}"
						:size="40"
					/>
				</q-avatar>
			</bt-form-item>
			<bt-form-item
				:title="t('olares_ID')"
				:data="userInfo ? userInfo.terminusName : ''"
			/>
			<!--			<bt-form-item-->
			<!--				:title="t('email')"-->
			<!--				:data="userInfo ? userInfo.email : ''"-->
			<!--			/>-->
			<bt-form-item
				:title="t('state')"
				:data="
					userInfo?.wizard_complete ? userInfo.state : t('waiting_onBoard')
				"
			/>
			<!-- <bt-form-item
				:title="t('last_login_time')"
				:data="
					userInfo.last_login_time
						? getLocalTime(userInfo.last_login_time * 1000).format(
								'YYYY-MM-DD HH:mm'
						  )
						: '--'
				"
			/> -->
			<bt-form-item
				:title="t('create_time')"
				:data="
					getLocalTime(userInfo.creation_timestamp * 1000).format(
						'YYYY-MM-DD HH:mm'
					)
				"
			/>
			<bt-form-item
				:title="t('roles')"
				:width-separator="false"
				:data="
					userInfo
						? userInfo.roles.length
							? getRoleName(userInfo.roles.join('/'))
							: ''
						: ''
				"
			/>
		</bt-list>

		<bt-list v-if="!userInfo?.wizard_complete && managerPermission">
			<bt-form-item
				:title="t('reset_password')"
				:chevronRight="true"
				:margin-top="false"
				:width-separator="false"
				@click="resetPassword"
			/>
		</bt-list>

		<bt-list v-if="!userInfo?.wizard_complete">
			<AdaptiveLayout>
				<template v-slot:pc>
					<bt-form-item
						:title="t('wizard_url')"
						:width-separator="false"
						:data="`https://wizard-${userInfo?.name}.${url_domain}`"
					/>
				</template>
				<template v-slot:mobile>
					<bt-form-item :title="t('wizard_url')" :width-separator="false">
						<div
							class="row items-center justify-end"
							@click="
								setCopyInfo(`https://wizard-${userInfo?.name}.${url_domain}`)
							"
						>
							<div
								style="
									max-width: 130px;
									overflow: hidden;
									white-space: nowrap;
									text-overflow: ellipsis;
								"
								class="text-body-3 text-ink-3"
							>
								{{ `https://wizard-${userInfo?.name}.${url_domain}` }}
							</div>
							<q-icon name="sym_r_content_copy" size="24px" />
						</div>
					</bt-form-item>
				</template>
			</AdaptiveLayout>
		</bt-list>

		<bt-list
			v-if="
				userInfo?.wizard_complete &&
				!isDemo &&
				(managerPermission || isCurrentUser)
			"
		>
			<bt-form-item
				v-if="managerPermission"
				:title="t('reset_password')"
				:margin-top="false"
				:width-separator="false"
				:chevron-right="true"
				@click="resetPassword"
			/>
			<bt-form-item
				v-if="isCurrentUser"
				:title="t('change_password')"
				:margin-top="false"
				:width-separator="false"
				:chevron-right="true"
				@click="updatePassword"
			/>
		</bt-list>

		<div class="row justify-end q-mb-lg" style="margin-top: 24px">
			<ListBottomFuncBtn
				v-if="managerPermission"
				@funcClick="deleteUser"
				:title="t('delete_user')"
				style="margin-right: 20px"
			/>

			<ListBottomFuncBtn
				v-if="managerPermission"
				@funcClick="changeQuota"
				:title="t('modify_limits')"
				style="margin-right: 20px"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ListBottomFuncBtn from 'src/components/settings/ListBottomFuncBtn.vue';
import UpdateUserPassworDialog from '../dialog/UpdateUserPassworDialog.vue';
import ResourceLimit from 'src/components/settings/user/ResourceLimit.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import UpdateUserQutoaDialog from './dialog/UpdateUserQutoaDialog.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useUserStore } from 'src/stores/settings/user';
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { copyToClipboard, Loading, useQuasar } from 'quasar';
import { notifySuccess } from 'src/utils/settings/btNotify';
import { useDeviceStore } from 'src/stores/settings/device';
import { useAdminStore } from 'src/stores/settings/admin';
import { useTokenStore } from 'src/stores/settings/token';
import { getLocalTime } from 'src/utils/settings/utils';
import { getRoleName, OLARES_ROLE } from 'src/constant';
import { passwordAddSort } from 'src/utils/account';
import { useRoute, useRouter } from 'vue-router';
import { generatePasword } from '../utils';
import { useI18n } from 'vue-i18n';
import {
	AccountInfo,
	AccountModifyStatus,
	AccountStatus
} from 'src/constant/global';
import {
	get_cluster_resource,
	get_user_resource,
	UserUsage
} from 'src/types/resource';

const userStore = useUserStore();
const tokenStore = useTokenStore();
const adminStore = useAdminStore();
const deviceStore = useDeviceStore();
const { t } = useI18n();

const url_domain = computed(() => {
	return tokenStore.url?.split('.').slice(1).join('.');
});

const quasar = useQuasar();
const Route = useRoute();
const router = useRouter();

const userInfo = ref<AccountInfo | undefined>(
	userStore.getUserByName(Route.params.name as string)
);

console.log(userInfo.value);
console.log(adminStore.user.name);
const usage = ref<UserUsage | undefined>(undefined);

async function changeQuota() {
	quasar
		.dialog({
			component: UpdateUserQutoaDialog,
			componentProps: {
				cpu: userInfo.value?.cpu_limit,
				memory: Number(
					userInfo.value?.memory_limit.slice(
						0,
						userInfo.value?.memory_limit.length - 1
					)
				)
			}
		})
		.onOk(async (data) => {
			if (!userInfo.value) {
				return;
			}
			try {
				await userStore.update_account_quoto(userInfo.value.name, {
					memory_limit: data.memoryLimit + 'G',
					cpu_limit: '' + data.cpuLimit
				});
				userInfo.value.cpu_limit = data.cpuLimit;
				userInfo.value.memory_limit = data.memoryLimit + 'G';
			} catch (e: any) {
				console.log(e);
			}
		});
}

const managerPermission = computed(() => {
	return adminStore.canManageAccount(userInfo.value);
});

const isCurrentUser = computed(() => {
	return adminStore.isCurrentAccount(userInfo.value);
});

const isDemo = computed(() => {
	return !!process.env.DEMO;
});

let checkAccountDeleteProgress: any = null;

const deleteUser = () => {
	if (!userInfo.value) {
		return;
	}

	quasar
		.dialog({
			component: ReminderDialogComponent,
			componentProps: {
				title: t('delete_item', {
					item: userInfo.value.name
				}),
				message: t('delete_user_message', {
					account: userInfo.value.name
				}),
				useCancel: true
			}
		})
		.onOk(() => {
			deleteUserSureAction();
		});
};

const deleteUserSureAction = async () => {
	if (!userInfo.value) {
		return;
	}
	const userName = userInfo.value.name;
	Loading.show();
	try {
		await userStore.delete_account(userName);
		checkAccountDeleteProgress = setInterval(async () => {
			checkAccountDelete(userName);
		}, 4 * 1000);
	} catch (error: any) {
		console.log(error);
		Loading.hide();
	}
};

async function checkAccountDelete(username: string) {
	try {
		const data: AccountModifyStatus = await userStore.get_account_status(
			username
		);
		if (data.status == AccountStatus.Deleted) {
			if (checkAccountDeleteProgress) {
				clearInterval(checkAccountDeleteProgress);
			}
			userStore.removeLocalAccount(username);
			setTimeout(() => {
				Loading.hide();
				router.replace('/user');
			}, 4 * 1000);
		}
	} catch (e) {
		/* empty */
	}
}

const updatePassword = () => {
	quasar.dialog({
		component: UpdateUserPassworDialog,
		componentProps: Route.params
	});
};

const resetPassword = async () => {
	if (!userInfo.value) {
		return;
	}
	const password = generatePasword();
	try {
		await userStore.reset_account_password({
			password: passwordAddSort(password, adminStore.terminus.osVersion),
			current_password: '',
			username: userInfo.value?.name
		});
		quasar.dialog({
			component: ReminderDialogComponent,
			componentProps: {
				title: t('reset_password_successfully'),
				message: t('new_password_is', {
					password
				}),
				useCancel: false
			}
		});
	} catch (error: any) {
		if (error) {
			// quasar.notify(error.message);
		}
	}
};

let updateUserInfoInterval: number | null = null;

async function executeUpdateAccount(username: string) {
	try {
		await userStore.update_account_info(username);
		userInfo.value = userStore.getUserByName(Route.params.name as string);

		if (userInfo.value?.wizard_complete) {
			stopUserInfoPolling();
		}
	} catch (e) {
		console.log('Account info update failed:', e);
	}
}

function stopUserInfoPolling() {
	if (updateUserInfoInterval) {
		clearInterval(updateUserInfoInterval);
		updateUserInfoInterval = null;
	}
}

function startUserInfoPolling(username: string) {
	if (updateUserInfoInterval || userInfo.value?.wizard_complete) {
		return;
	}

	updateUserInfoInterval = setInterval(() => {
		executeUpdateAccount(username);
	}, 30000);
}

async function updateUserInfo(username: string) {
	try {
		userInfo.value = userStore.getUserByName(Route.params.name as string);
		executeUpdateAccount(username);

		if (userInfo.value?.roles.find((r) => r === OLARES_ROLE.OWNER)) {
			get_cluster_resource().then((res) => {
				usage.value = res;
			});
		} else {
			get_user_resource(username).then((res) => {
				usage.value = res;
			});
		}

		if (!userInfo.value?.wizard_complete) {
			startUserInfoPolling(username);
		}
	} catch (e) {
		console.log(e);
	}
}

watch(
	() => Route.params.name,
	async (val) => {
		const username = val as string;
		stopUserInfoPolling();
		await updateUserInfo(username);
	}
);

onMounted(() => {
	const username = Route.params.name as string;
	updateUserInfo(username);
});

onUnmounted(() => {
	if (checkAccountDeleteProgress) {
		clearInterval(checkAccountDeleteProgress);
		checkAccountDeleteProgress = null;
	}
	stopUserInfoPolling();
});

const setCopyInfo = (info: string) => {
	copyToClipboard(info).then(() => {
		notifySuccess(t('copy_successfully'));
	});
};
</script>

<style scoped lang="scss">
.resource-title {
	margin-top: 12px;
	margin-bottom: 8px;
}
</style>
