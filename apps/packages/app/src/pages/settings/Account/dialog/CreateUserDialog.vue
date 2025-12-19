<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('create_account')"
		:skip="false"
		:ok="t('save')"
		:cancel="t('cancel')"
		size="medium"
		:platform="deviceStore.platform"
		:ok-disabled="!enableCreate"
		@onSubmit="createUserName"
	>
		<div>
			<terminus-edit
				v-model="userName"
				:label="t('olares_ID')"
				:show-password-img="false"
				style="width: 100%"
				class=""
				:is-error="userName.length > 0 && usernameRule(userName).length > 0"
				:error-message="usernameRule(userName)"
			/>
			<div class="text-body3 text-ink-3 q-mt-md">
				{{ t('roles') }}
			</div>
			<bt-select
				v-model="newRole"
				:options="roleOptions"
				:border="true"
				color="text-blue-default"
			/>

			<terminus-edit
				v-model="cpuLimit"
				label="CPU"
				:show-password-img="false"
				class="q-mt-md"
				:is-error="cpuLimit.length > 0 && cpuLimitRule(cpuLimit).length > 0"
				:error-message="cpuLimitRule(cpuLimit)"
			>
				<template v-slot:right>
					<edit-number-right-slot v-model="cpuLimit" label="core" />
				</template>
			</terminus-edit>

			<terminus-edit
				v-model="memoryLimit"
				label="Memory"
				:show-password-img="false"
				class="q-mt-md"
				:is-error="
					memoryLimit.length > 0 && memoryLimitRule(memoryLimit).length > 0
				"
				:error-message="memoryLimitRule(memoryLimit)"
			>
				<template v-slot:right>
					<edit-number-right-slot v-model="memoryLimit" label="GB" />
				</template>
			</terminus-edit>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useQuasar, Loading } from 'quasar';
import { ref, onUnmounted, computed } from 'vue';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import { useUserStore } from 'src/stores/settings/user';
import { useDIDStore } from 'src/stores/settings/did';
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import EditNumberRightSlot from 'src/components/settings/EditNumberRightSlot.vue';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import { useDeviceStore } from 'src/stores/settings/device';
import { AccountModifyStatus, AccountStatus } from 'src/constant/global';
import { generatePasword } from '../utils';
import { useI18n } from 'vue-i18n';
import { passwordAddSort } from 'src/utils/account';
import { useAdminStore } from 'src/stores/settings/admin';
import { OLARES_ROLE, SelectorProps } from 'src/constant';
import BtSelect from 'src/components/base/BtSelect.vue';
import { getApplication } from 'src/application/base';

const { t } = useI18n();

const PASSWORD_RULE = {
	LENGTH_RULE: '^.{8,32}$',
	LOWERCASE_RULE: '^(?=.*[a-z])',
	UPPERCASE_RULE: '^(?=.*[A-Z])',
	DIGIT_RULE: '^(?=.*[0-9])',
	SYMBOL_RULE: '^(?=.*[@$!%*?&_.])',
	ALL_RULE:
		'^(.*[a-z].*[A-Z].*[0-9].*)$|^(.*[a-z].*[0-9].*[A-Z].*)$|^(.*[A-Z].*[a-z].*[0-9].*)$|^(.*[A-Z].*[0-9].*[a-z].*)$|^(.*[0-9].*[a-z].*[A-Z].*)$|^(.*[0-9].*[A-Z].*[a-z].*)$|^(\$2[ayb]\$.{56})$'
};
const allRule = new RegExp(PASSWORD_RULE.ALL_RULE);

const userName = ref('');
// const email = ref('');

const cpuLimit = ref('1');
const memoryLimit = ref('4');

const accountStore = useUserStore();
const didStore = useDIDStore();
const adminStore = useAdminStore();

const quasar = useQuasar();
const deviceStore = useDeviceStore();
const newRole = ref();
const roleOptions = computed(() => {
	if (adminStore.user.owner_role === OLARES_ROLE.OWNER) {
		return [
			{
				label: t('admin'),
				value: OLARES_ROLE.ADMIN
			},
			{
				label: t('members'),
				value: OLARES_ROLE.NORMAL
			}
		];
	} else {
		return [
			{
				label: t('members'),
				value: OLARES_ROLE.NORMAL
			}
		] as SelectorProps[];
	}
});

let password = '';

const CustomRef = ref();

let checkAccountCreateProgress: any = null;

const createUserName = async () => {
	if (userName.value.length === 0) {
		return;
	}

	if (!newRole.value) {
		return;
	}

	password = generatePasword();

	while (!allRule.test(password)) {
		password = generatePasword();
	}
	Loading.show();

	const currentUserName =
		userName.value + '@' + adminStore.olaresId.split('@')[1];

	if (adminStore.olaresd) {
		const data = await didStore.resolve_name(currentUserName);
		if (!data) {
			Loading.hide();
			return;
		}
	} else {
		const data = await didStore.resolve_name_by_did(currentUserName);
		if (!data) {
			Loading.hide();
			return;
		}
	}

	try {
		await accountStore.create_account({
			name: userName.value,
			owner_role: newRole.value,
			password: passwordAddSort(password, adminStore.terminus.osVersion),
			cpu_limit: '' + cpuLimit.value,
			memory_limit: '' + memoryLimit.value + 'G'
		});

		checkAccountCreateProgress = setInterval(async () => {
			checkAccountCreate(userName.value);
		}, 4 * 1000);
	} catch (error: any) {
		console.log(error);
		Loading.hide();
	}
};
/**
 * check wait creat account state
 * @param username
 */
async function checkAccountCreate(username: string) {
	try {
		const data: AccountModifyStatus = await accountStore.get_account_status(
			username
		);

		if (data.status == AccountStatus.Created) {
			Loading.hide();
			if (checkAccountCreateProgress) {
				clearInterval(checkAccountCreateProgress);
				checkAccountCreateProgress = null;
			}

			const message =
				t('olares_ID') +
				':' +
				userName.value +
				'<br>' +
				t('original_password') +
				':' +
				password +
				'<br>' +
				t('wizard_url') +
				':' +
				'https://' +
				data.address.wizard;

			quasar
				.dialog({
					component: ReminderDialogComponent,
					componentProps: {
						title: t('user_had_been_created', {
							username
						}),
						message,
						useCancel: false,
						confirmText: t('copy'),
						hasBorder: true
					}
				})
				.onOk(() => {
					getApplication()
						.copyToClipboard(message.replace(/<br>/g, '\r\n'))
						.then(() => {
							notifySuccess(t('copy_successfully'));
						});
					onDialogOK();
				});
		} else if (data.status === AccountStatus.Failed) {
			Loading.hide();
			notifyFailed(
				data.message || t('An error occurred while creating the user.')
			);
		}
	} catch (e: any) {
		Loading.hide();
		if (checkAccountCreateProgress) {
			clearInterval(checkAccountCreateProgress);
			checkAccountCreateProgress = null;
		}
	}
}

const usernameRule = (val: string) => {
	if (val.length === 0) {
		return t('errors.username_is_empty');
	}
	let u = accountStore.accounts.find((item) => item.name == val);
	if (u) {
		return t('errors.username_already_registered');
	} else {
		return '';
	}
};

// const emailRule = (val: string) => {
// 	if (val.length === 0) {
// 		return t('errors.email_is_empty');
// 	}
// 	let a = accountStore.accounts.find((item) => item.email == val);
// 	if (a) {
// 		return t('errors.email_already_registered');
// 	}
// 	return '';
// };

const cpuLimitRule = (val: string) => {
	if (val.length === 0) {
		return t('errors.cpu_limit_is_empty');
	}
	let rule = /^[+-]?(\d+\.?\d*|\.\d+)$/;
	if (!rule.test(val)) {
		return t('errors.only_valid_numbers_can_be_entered');
	}
	return '';
};

const memoryLimitRule = (val: string) => {
	if (val.length === 0) {
		return t('errors.memory_limit_is_empty');
	}
	let rule = /^[+-]?(\d+\.?\d*|\.\d+)$/;
	if (!rule.test(val)) {
		return t('errors.only_valid_numbers_can_be_entered');
	}
	return '';
};

const enableCreate = computed(() => {
	return (
		usernameRule(userName.value).length == 0 &&
		// emailRule(email.value).length == 0 &&
		cpuLimitRule(cpuLimit.value).length == 0 &&
		memoryLimitRule(memoryLimit.value).length == 0 &&
		!!newRole.value
	);
});

onUnmounted(() => {
	if (checkAccountCreateProgress) {
		clearInterval(checkAccountCreateProgress);
		checkAccountCreateProgress = null;
	}
});

const onDialogOK = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style scoped lang="scss">
.cpu-core {
	text-align: right;
}
</style>
