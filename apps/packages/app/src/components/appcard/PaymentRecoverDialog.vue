<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('Purchase App')"
		:ok="false"
		:cancel="false"
	>
		<div class="text-ink-2 text-body2">
			{{
				t('Do you want to buy app for', {
					name: appName,
					price: price,
					unit: unit
				})
			}}
		</div>

		<div class="row justify-end items-center q-mt-lg">
			<q-btn
				class="btn-border btn-class text-ink-2 text-subtitle1 text-capitalize"
				:label="t('Cancel')"
				@click="onCancel"
			/>
			<q-btn
				class="btn-border btn-class text-ink-2 text-subtitle1 text-capitalize q-ml-lg"
				:label="t('Restore purchase')"
				@click="onRestore"
				:loading="isLoading"
			/>
			<q-btn
				class="btn-class bg-blue-default text-white text-subtitle1 q-ml-lg text-capitalize"
				:label="t('Buy')"
				@click="onOK"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { getI18nValue, SupportToken } from 'src/constant/constants';
import { useCenterStore } from 'src/stores/market/center';
import { usePaymentStore } from 'src/stores/market/payment';
import { useAppStore } from 'src/stores/market/appStore';
import { onMounted, PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import BigNumber from 'bignumber.js';

const props = defineProps({
	appId: {
		type: String,
		required: true
	},
	sourceId: {
		type: String,
		required: true
	},
	tokenInfo: {
		type: Object as PropType<SupportToken>,
		required: true
	}
});

const appName = ref();
const price = ref('0');
const unit = ref('');
const $q = useQuasar();
const customRef = ref();
const { t, locale } = useI18n();
const appStore = useAppStore();
const centerStore = useCenterStore();
const paymentStore = usePaymentStore();
const isLoading = ref(false);

onMounted(() => {
	console.log(props.tokenInfo);
	const appFullInfoLatest = appStore.getAppFullInfo(
		props.appId,
		props.sourceId
	);
	if (appFullInfoLatest) {
		appName.value = getI18nValue(
			appFullInfoLatest.app_simple_info.app_title,
			locale
		);
	}
	unit.value = props.tokenInfo.token_symbol;
	try {
		price.value = new BigNumber(props.tokenInfo.token_amount)
			.dividedBy(
				new BigNumber(10).exponentiatedBy(props.tokenInfo.token_decimals)
			)
			.toString();
	} catch (e) {
		console.log(e);
		price.value = '0';
	}
});

const onCancel = () => {
	customRef.value.onDialogOK({
		status: 'cancel'
	});
};

const onRestore = async () => {
	isLoading.value = true;
	await paymentStore.recoverAppPurchase(props.appId, props.sourceId, t);
	isLoading.value = false;
	customRef.value.onDialogOK({
		status: 'restore'
	});
};

const onOK = () => {
	paymentStore.queryPaymentInfo(props.appId, props.sourceId, t, $q);
	customRef.value.onDialogOK({
		status: 'ok'
	});
};
</script>

<style scoped lang="scss">
::v-deep(.dialog-content) {
	margin: 20px 0 0 !important;
}
.btn-border {
	border: 1px solid $btn-stroke;
}
.btn-class {
	height: 40px;
}
</style>
