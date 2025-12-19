<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Reset SSH Password')"
		:ok="position ?? t('confirm')"
		:cancel="showCancel ? navigation ?? t('cancel') : false"
		size="small"
		platform="mobile"
		:okDisabled="false"
		@onSubmit="onSubmit"
	>
		<!-- <terminus-edit
			v-model="passwordRef"
			:label="passwordTitle"
			:show-password-img="true"
			class="terminus-unlock-box__edit q-mt-md"
			@keyup.enter="onSubmit"
		/>
		<terminus-rule-checker
			:value="passwordRef"
			:pattern="SSH_PASSWORD_RULE.LENGTH_RULE"
			:label="t('10-32 characters')"
			v-model:result="lengthResult"
			@update:result="setButtonStatus"
		/>
		<terminus-rule-checker
			:value="passwordRef"
			:pattern="PASSWORD_RULE.LOWERCASE_RULE"
			:label="t('Contains lowercase letters')"
			v-model:result="lowercaseResult"
			@update:result="setButtonStatus"
		/>
		<terminus-rule-checker
			:value="passwordRef"
			:pattern="PASSWORD_RULE.UPPERCASE_RULE"
			:label="t('Contains uppercase letters')"
			v-model:result="uppercaseResult"
			@update:result="setButtonStatus"
		/>
		<terminus-rule-checker
			:value="passwordRef"
			:pattern="PASSWORD_RULE.DIGIT_RULE"
			:label="t('Contains numbers')"
			v-model:result="digitsResult"
			@update:result="setButtonStatus"
		/>
		<terminus-rule-checker
			:value="passwordRef"
			:pattern="PASSWORD_RULE.SYMBOL_RULE"
			:label="t('Contains symbols')"
			v-model:result="symbolResult"
			@update:result="setButtonStatus"
		/> -->
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import TerminusEdit from '../common/TerminusEdit.vue';
import TerminusRuleChecker from '../common/TerminusRuleChecker.vue';
import {
	PASSWORD_RULE,
	SSH_PASSWORD_RULE,
	ConfirmButtonStatus
} from '../../utils/constants';

import { ref } from 'vue';

defineProps({
	message: String,
	navigation: String,
	position: String,
	showCancel: {
		default: true,
		type: Boolean
	},
	descAlign: {
		default: 'center',
		type: String as () => 'left' | 'center' | 'right'
	}
});

const { t } = useI18n();
const passwordRef = ref();
const CustomRef = ref();
const passwordTitle = t('password');

const symbolResult = ref(false);
const digitsResult = ref(false);
const uppercaseResult = ref(false);
const lowercaseResult = ref(false);
const lengthResult = ref(false);

const buttonStatus = ref(ConfirmButtonStatus.disable);

function setButtonStatus() {
	if (!passwordRef.value) {
		buttonStatus.value = ConfirmButtonStatus.disable;
		return;
	}
	if (
		!digitsResult.value ||
		!symbolResult.value ||
		!lowercaseResult.value ||
		!uppercaseResult.value ||
		!lengthResult.value
	) {
		buttonStatus.value = ConfirmButtonStatus.disable;
	} else {
		buttonStatus.value = ConfirmButtonStatus.normal;
	}
}

const onSubmit = () => {
	if (buttonStatus.value != ConfirmButtonStatus.normal) {
		return;
	}
	CustomRef.value.onDialogOK(passwordRef.value);
};
</script>

<style lang="scss" scoped>
.biometric-unlock-dialog {
	padding: 20px;

	&__title {
		color: $ink-1;
	}

	&__desc {
		width: 100%;
		margin-top: 12px;
	}

	&__group {
		margin-top: 40px;
		width: 100%;

		&__btn {
			width: calc(50% - 6px);
		}
		.cancel {
			border: 1px solid $separator;
		}
	}
}
</style>
