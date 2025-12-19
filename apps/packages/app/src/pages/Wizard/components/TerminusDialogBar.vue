<template>
	<q-bar class="bar q-mb-md">
		<q-icon v-if="icon" :name="icon" size="18px" />
		<div class="title col" :class="titAlign ? titAlign : 'text-left'">
			{{ label }}
		</div>
		<q-space />
		<q-btn dense flat icon="close" @click="onCancel">
			<q-tooltip>{{ t('close') }}</q-tooltip>
		</q-btn>
	</q-bar>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

defineProps({
	label: {
		type: String,
		default: '',
		required: false
	},
	icon: {
		type: String,
		default: '',
		required: false
	},
	titAlign: {
		type: String,
		default: 'left',
		required: false
	}
});

const { t } = useI18n();
const emit = defineEmits(['close']);

const onCancel = () => {
	emit('close');
};
</script>

<style scoped lang="scss">
.mobile-title {
	color: #322f2d;
	font-size: 16px;
	text-align: center;
	font-weight: 500;
	margin: 20px 0;
}
.bar {
	background-color: #f8f8f8;
	.title {
		height: 18px;
		line-height: 18px;
		color: #322f2d;
		font-size: 14px;
	}
}
</style>
