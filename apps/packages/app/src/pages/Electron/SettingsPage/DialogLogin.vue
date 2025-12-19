<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('check_password')"
		:ok="t('unlock.title')"
		:cancel="t('cancel')"
		:size="isWeb ? 'medium' : 'small'"
		:platform="isWeb ? 'web' : 'mobile'"
		@onSubmit="onDialogOK"
	>
		<terminus-edit
			v-model="password"
			:label="t('password')"
			:show-password-img="true"
			class="terminus-unlock-box__edit"
			@keyup.enter="login"
		/>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useUserStore } from '../../../stores/user';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import TerminusEdit from '../../../components/common/TerminusEdit.vue';

const password = ref();
const userStore = useUserStore();
const { t } = useI18n();

const isWeb = ref(
	process.env.APPLICATION == 'VAULT' || process.env.PLATFORM == 'DESKTOP'
);

const login = () => {
	if (!password.value) {
		notifyFailed(t('password_empty'));
		return;
	}
	loginByPassword(password.value);
};

const loginByPassword = async (password: string) => {
	if (password === userStore.password) {
		onDialogOK();
	} else {
		notifyFailed(t('password_error'));
	}
};

const CustomRef = ref();

const onDialogOK = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-continer {
		width: 400px;
		border-radius: 8px;

		.card-content {
			padding: 0 20px;
			.input {
				border-radius: 5px;
				border: 1px solid $input-stroke;
				background-color: transparent;
				&:focus {
					border: 1px solid $yellow-disabled;
				}
			}
		}
	}
}
</style>
