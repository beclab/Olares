<template>
	<div class="password-validator-root column justify-start items-center">
		<terminus-edit
			v-model="passwordRef"
			:label="t('password')"
			:transaction="edtTransaction"
			:show-password-img="true"
			class="password-validator-root__edit"
			@update:model-value="setButtonStatus"
		/>

		<terminus-rule-checker
			:value="passwordRef"
			:pattern="PASSWORD_RULE.LENGTH_RULE"
			:label="t('8-32 characters')"
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
		/>

		<terminus-edit
			v-if="repeatEnable"
			v-model="rePwdRef"
			:transaction="edtTransaction"
			:label="t('repeat_password')"
			:show-password-img="true"
			:is-error="!samePasswordRef"
			:error-message="t('password_not_match')"
			class="password-validator-root__edit"
			@update:model-value="setButtonStatus"
		/>
	</div>
</template>

<script setup lang="ts">
import TerminusEdit from './TerminusEdit.vue';
import TerminusRuleChecker from './TerminusRuleChecker.vue';
import { ConfirmButtonStatus, PASSWORD_RULE } from '../../utils/constants';
import { PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	buttonStatus: {
		type: Object as PropType<ConfirmButtonStatus>,
		require: false
	},
	buttonText: {
		type: String,
		require: ''
	},
	repeatEnable: {
		type: Boolean,
		required: false,
		default: true
	},
	edtTransaction: {
		type: Boolean,
		default: false,
		require: false
	}
});

const emit = defineEmits(['update:buttonStatus', 'update:buttonText']);

const { t } = useI18n();
const passwordRef = ref();
const rePwdRef = ref();
const samePasswordRef = ref(true);

const symbolResult = ref(false);
const digitsResult = ref(false);
const uppercaseResult = ref(false);
const lowercaseResult = ref(false);
const lengthResult = ref(false);
const allRule = new RegExp(PASSWORD_RULE.ALL_RULE);

function setButtonStatus() {
	if (!passwordRef.value || (props.repeatEnable && !rePwdRef.value)) {
		emit('update:buttonStatus', ConfirmButtonStatus.disable);
		emit('update:buttonText', t('confirm'));
		return;
	}

	if (props.repeatEnable) {
		samePasswordRef.value = passwordRef.value === rePwdRef.value;
	}

	if (
		!digitsResult.value ||
		!symbolResult.value ||
		!lowercaseResult.value ||
		!uppercaseResult.value ||
		!lengthResult.value ||
		!samePasswordRef.value
	) {
		emit('update:buttonStatus', ConfirmButtonStatus.disable);
		emit('update:buttonText', t('reset'));
	} else {
		emit('update:buttonStatus', ConfirmButtonStatus.normal);
		emit('update:buttonText', t('confirm'));
	}
}

function clearPassword() {
	// passwordRef.value = '';
	// rePwdRef.value = '';
}

function getValidPassword(): string | null {
	if (!passwordRef.value || (props.repeatEnable && !rePwdRef.value)) {
		return null;
	}

	if (props.repeatEnable && !samePasswordRef.value) {
		return null;
	}

	if (!allRule.test(passwordRef.value)) {
		return null;
	}

	return passwordRef.value;
}

defineExpose({ clearPassword, getValidPassword });
</script>

<style scoped lang="scss">
.password-validator-root {
	width: 100%;
	height: auto;

	&__edit {
		margin-top: 10px;
		width: 100%;
	}
}
</style>
