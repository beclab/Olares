<template>
	<q-card-actions class="row justify-end items-center q-mt-md q-mb-sm">
		<q-item
			v-if="showMore"
			clickable
			dense
			class="but-cancel row justify-center items-center q-px-md q-mr-md"
			@click="onMore"
		>
			{{ moreText }}
		</q-item>
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
			:disable="okDisable"
			clickable
			dense
			class="but-creat row justify-center items-center q-px-md q-mr-sm"
			@click="onOK"
			v-if="!loading"
		>
			{{ okText }}
		</q-item>
		<q-item
			v-else
			dense
			class="but-creat row justify-center items-center q-px-md q-mr-sm"
		>
			{{ t('loading') }}
		</q-item>
	</q-card-actions>
</template>

<script lang="ts" setup>
import { i18n } from '../../boot/i18n';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

defineProps({
	okText: {
		type: String,
		default: i18n.global.t('submit'),
		required: false
	},
	okDisable: {
		type: Boolean,
		default: false
	},
	cancelText: {
		type: String,
		default: i18n.global.t('cancel'),
		required: false
	},
	moreText: {
		type: String,
		default: i18n.global.t('base.more'),
		required: false
	},
	showCancel: {
		type: Boolean,
		default: true,
		required: false
	},
	showMore: {
		type: Boolean,
		default: false,
		required: false
	},
	loading: {
		type: Boolean,
		default: false,
		required: false
	}
});

const emit = defineEmits(['close', 'submit', 'more']);

const onCancel = () => {
	emit('close');
};

const onMore = () => {
	emit('more');
};

const onOK = (e: any) => {
	emit('submit', e);
};
</script>

<style scoped lang="scss">
.but-creat {
	border-radius: 8px;
	font-weight: 500;
	font-size: 12px;
	background: $orange-default;
	color: $ink-on-brand;
}

.but-cancel {
	border-radius: 8px;
	font-weight: 500;
	font-size: 12px;
	border: 1px solid $btn-stroke;
	color: $ink-2;
}

.card-action {
	margin: 20px;

	.cancel-button {
		width: 48%;
	}
}
</style>
