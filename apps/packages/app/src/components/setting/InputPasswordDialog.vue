<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:ok="unlockTitle"
		:cancel="t('cancel')"
		size="small"
		platform="mobile"
		@onSubmit="login"
	>
		<div class="card-content">
			<terminus-edit
				v-model="password"
				:label="passwordTitle"
				:show-password-img="true"
				class="terminus-unlock-box__edit q-mt-md"
				@keyup.enter="login"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import TerminusEdit from '../common/TerminusEdit.vue';
import { i18n } from './../../boot/i18n';
import { notifyFailed } from '../../utils/notifyRedefinedUtil';

defineProps({
	title: {
		type: String,
		default: i18n.global.t('check_password'),
		required: false
	},
	passwordTitle: {
		type: String,
		default: i18n.global.t('password'),
		required: false
	},
	unlockTitle: {
		type: String,
		default: i18n.global.t('confirm'),
		required: false
	}
});

const password = ref();
const { t } = useI18n();
const CustomRef = ref();

const login = () => {
	// if (!password.value) {
	// 	notifyFailed(t('password_empty'));
	// 	return;
	// }
	loginByPassword(password.value || '');
};

const loginByPassword = async (password: string) => {
	CustomRef.value.onDialogOK(password);
};
</script>

<style lang="scss" scoped>
.card-content {
	padding: 0 0px;
	.input {
		border-radius: 5px;
		border: 1px solid $input-stroke;
		background-color: transparent;
		&:focus {
			border: 1px solid $yellow-disabled;
		}
	}
}
</style>
