<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef" @hide="onCancel">
		<q-card class="card-continer" flat>
			<terminus-dialog-bar
				:label="title"
				icon=""
				titAlign="text-left"
				@close="onCancel"
			/>

			<div
				class="dialog-desc"
				:style="{ textAlign: isMobile ? 'center' : 'left' }"
			>
				<div class="text-body2 text-ink-2 q-my-sm">
					{{ label }}
				</div>
				<q-input
					class="password-input"
					v-model="password"
					:type="isPwd ? 'password' : 'text'"
					borderless
					dense
					input-style="height: 40px;lineHeight: 40px; text-indent: 10px;"
				>
					<template v-slot:append>
						<q-icon
							size="16px"
							:name="isPwd ? 'visibility_off' : 'visibility'"
							class="cursor-pointer q-mr-sm"
							@click="isPwd = !isPwd"
						/>
					</template>
				</q-input>
			</div>

			<terminus-dialog-footer
				:okText="t('confirm')"
				:cancelText="t('cancel')"
				showCancel
				:loading="loading"
				@close="onCancel"
				@submit="onOKClick"
			/>
		</q-card>
	</q-dialog>

	<!-- <q-dialog ref="root" @hide="onDialogHide">
		<q-card class="q-dialog-plugin row root">
			<div class="text-color-title text-subtitle1 title">{{ title }}</div>
			<div class="text-color-title text-body1 content" v-if="message">
				{{ message }}
			</div>
			<div class="password">
				<q-input v-model="password" :label="label" borderless type="password" />
			</div>

			<div class="row iterm-center justify-between button">
				<q-btn class="bg-blue confirm" :label="confirmTxt" @click="onOKClick" />
				<q-btn
					class="bg-grey-11 text-ink-2 cancel"
					:label="cancelTxt"
					@click="onCancelClick"
				/>
			</div>
		</q-card>
	</q-dialog> -->
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { app } from '../../../globals';
import { useUserStore } from '../../../stores/user';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import { i18n } from '../../../boot/i18n';
import { useDialogPluginComponent, useQuasar } from 'quasar';

import TerminusDialogBar from '../../../components/common/TerminusDialogBar.vue';
import TerminusDialogFooter from '../../../components/common/TerminusDialogFooter.vue';

defineProps({
	confirmTxt: {
		type: String,
		default: i18n.global.t('ok'),
		required: false
	},
	cancelTxt: {
		type: String,
		default: i18n.global.t('cancel'),
		required: false
	}
});

const $q = useQuasar();
const show = ref(true);
const userStore = useUserStore();
const { t } = useI18n();

const password = ref('');
const newPassword = ref('');
const oldPassword = ref('');
const passwordStatus = ref('pending');
const loading = ref(false);
const isPwd = ref(true);

const title = ref(t('change_master_password'));
const message = ref(t('please_enter_your_current_password'));
const label = ref(t('enter_current_password'));

const { dialogRef, onDialogCancel } = useDialogPluginComponent();
const isMobile = ref(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile);

const onOKClick = async () => {
	if (!password.value) {
		return false;
	}
	switch (passwordStatus.value) {
		case 'pending':
			try {
				if (!userStore.users || userStore.users.locked) {
					notifyFailed(t('please_unlock_first'));
					return;
				}
				await userStore.users.unlock(password.value).then(() => {
					passwordStatus.value = 'newPassword';
					message.value = t('now_choose_a_new_master_password');
					label.value = t('enter_New_password');
					oldPassword.value = password.value;
					password.value = '';
				});
			} catch (error) {
				notifyFailed(t('wrong_password_please_try_again'));
			}
			break;

		case 'newPassword':
			message.value = t('please_confirm_your_new_password');
			label.value = t('repert_new_password');
			newPassword.value = password.value;
			password.value = '';
			passwordStatus.value = 'repertPassword';
			break;

		case 'repertPassword':
			if (newPassword.value === password.value) {
				try {
					loading.value = true;
					const resetPasswordStatus = await userStore.updateUserPassword(
						oldPassword.value,
						newPassword.value
					);
					if (resetPasswordStatus.status) {
						app.lock();
					} else {
						notifyFailed(resetPasswordStatus.message);
					}
					loading.value = false;
				} catch (error) {
					loading.value = false;
					if (error.message) {
						notifyFailed(error.message);
					}
				}
			} else {
				notifyFailed(t('wrong_password_please_try_again'));
			}
	}
};

const onCancel = () => {
	onDialogCancel();
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-continer {
		width: 400px;
		border-radius: 12px;

		.dialog-desc {
			padding-left: 20px;
			padding-right: 20px;

			.password-input {
				font-size: map-get($map: $body2, $key: size);
				border: 1px solid $input-stroke;
				border-radius: 8px;
				overflow: hidden;
			}
		}
	}
}
</style>
