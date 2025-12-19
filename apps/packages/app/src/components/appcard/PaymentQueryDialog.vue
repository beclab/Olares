<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="'Payment Result Detection Failed'"
		:ok="false"
		:cancel="false"
	>
		<div class="text-ink-2 text-body2">
			{{
				t(
					'Failed to detect the payment result. Please select an operation based on your actual payment status.'
				)
			}}
		</div>

		<div class="row justify-end items-center q-mt-md">
			<q-btn
				class="bg-blue-default text-white text-subtitle1 text-capitalize"
				:label="t('Unpaid, Pay Again')"
				@click="onOK('unpaid')"
			/>
			<q-btn
				class="bg-blue-default text-white text-subtitle1 q-ml-lg text-capitalize"
				:label="t('Paid, Refresh Result')"
				@click="onOK('paid')"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

const { t } = useI18n();
const customRef = ref();

const onOK = (status: string) => {
	customRef.value.onDialogOK({
		status
	});
};
</script>

<style scoped lang="scss">
::v-deep(.dialog-content) {
	margin: 20px 0 0 !important;
}
</style>
