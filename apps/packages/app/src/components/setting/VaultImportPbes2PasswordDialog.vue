<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Enter password')"
		:ok="i18n.global.t('confirm')"
		:cancel="t('cancel')"
		size="medium"
		platform="web"
		@onSubmit="unlock"
	>
		<div class="card-content">
			<div class="text-body2 text-ink-2">
				{{ t('This file is protected by a password.') }}
			</div>
			<terminus-edit
				v-model="password"
				:label="t('password')"
				:show-password-img="true"
				class="terminus-unlock-box__edit q-mt-md"
				@keyup.enter="unlock"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import TerminusEdit from '../common/TerminusEdit.vue';
import { i18n } from '../../boot/i18n';
import { notifyFailed } from '../../utils/notifyRedefinedUtil';
import * as imp from '@didvault/sdk/src/import';

const props = defineProps({
	file: {
		type: File,
		required: true
	}
});

const password = ref();
const { t } = useI18n();
const CustomRef = ref();

const unlock = async () => {
	try {
		const items = await imp.asPBES2Container(props.file, password.value);
		CustomRef.value.onDialogOK(items);
	} catch (e) {
		notifyFailed(t('password_error'));
	}
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
