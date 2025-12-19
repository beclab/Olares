<template>
	<q-bar class="bg-background-3" :style="{ height: height }">
		<q-icon v-if="icon" :name="icon" size="18px" />
		<div class="col left text-ink-1" :class="titAlign">
			{{ label }}
		</div>
		<q-space />
		<q-btn dense flat icon="close" color="text-ink-2" @click="onCancel">
			<q-tooltip>{{ t('base.close') }}</q-tooltip>
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
		required: false
	},
	height: {
		type: String,
		default: '32px'
	}
});

const { t } = useI18n();
const emit = defineEmits(['close']);

const onCancel = () => {
	emit('close');
};
</script>

<style scoped lang="scss">
.q-bar--standard > div {
	font-size: small;
}
.q-bar > .q-icon:first-child,
.q-bar > .q-btn:first-child,
.q-bar > div:first-child {
	margin-left: 0 !important;
}
</style>
