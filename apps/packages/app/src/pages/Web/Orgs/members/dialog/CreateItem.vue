<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('create')"
		:ok="t('create')"
		:cancel="t('cancel')"
		size="medium"
		@onSubmit="onOKClick"
	>
		<q-card-section class="q-pt-xs q-px-none">
			<div class="text-h5 q-my-lg">
				{{ t('invite_new_members') }}
			</div>
			<div class="text-subtitle1 text-grey text-center">
				{{ t('invite_new_members_message') }}
			</div>
		</q-card-section>

		<q-card-section class="row">
			<div class="email-wrap">
				<input
					type="text"
					placeholder="Enter DID"
					v-model="email"
					maxlength="50"
				/>
				<span>{{ email.length }}/50</span>
			</div>
		</q-card-section>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { notifyWarning } from '../../../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const email = ref('');

const CustomRef = ref();

async function onOKClick() {
	if (!email.value.length) {
		notifyWarning(t('please_enter_at_least_one_did'));
		return;
	}
	CustomRef.value.onDialogOK(email.value);
}
</script>

<style lang="scss" scoped>
.email-wrap {
	width: 90%;
	height: 40px;
	margin: 0 auto;
	border-radius: 6px;
	border: 1px solid $separator;
	display: flex;
	align-items: center;
	justify-content: space-between;
	overflow: hidden;

	&:hover {
		border: 1px solid $blue;
	}

	input {
		width: 84%;
		height: 40px;
		outline: none;
		border: none;
		text-indent: 10px;
	}
	span {
		display: inline-block;
		width: 16%;
		text-align: center;
	}
}

.but-creat-web {
	border-radius: 10px;
}
.but-cancel-web {
	border-radius: 10px;
}
</style>
