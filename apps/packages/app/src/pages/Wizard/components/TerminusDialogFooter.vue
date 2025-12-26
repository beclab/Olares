<template>
	<q-card-actions class="row justify-end items-center q-mt-md q-mb-sm">
		<q-item
			v-if="showCancel"
			clickable
			dense
			class="but-cancel row justify-center items-center q-px-md q-mr-md"
			@click="onCancel"
		>
			{{ cancelText }}
		</q-item>
		<q-item
			clickable
			dense
			class="but-create row justify-center items-center q-px-md q-mr-sm"
			@click="onOK"
			v-if="!loading"
		>
			{{ okText }}
		</q-item>
		<q-item
			v-else
			dense
			class="but-create row justify-center items-center q-px-md q-mr-sm"
		>
			{{ t('loading') }}
		</q-item>
	</q-card-actions>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

defineProps({
	okText: {
		type: String,
		default: 'Sumbit',
		required: false
	},
	cancelText: {
		type: String,
		default: 'Cancel',
		required: false
	},
	showCancel: {
		type: Boolean,
		default: true,
		required: false
	},
	loading: {
		type: Boolean,
		default: false,
		required: false
	}
});

const { t } = useI18n();

const emit = defineEmits(['close', 'submit']);

const onCancel = () => {
	emit('close');
};

const onOK = () => {
	emit('submit');
};
</script>

<style scoped lang="scss">
.but-create {
	border-radius: 8px;
	font-weight: 500;
	font-size: 12px;
	background: linear-gradient(90deg, #ffa337 0%, #ef438b 100%);
	color: #ffffff;
}
.but-cancel {
	border-radius: 8px;
	font-weight: 500;
	font-size: 12px;
	border: 1px solid #e0e0e0;
	color: #4d4742;
}

.card-action {
	margin: 20px;
	.cancel-button {
		width: 48%;
	}
}
</style>
