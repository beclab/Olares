<template>
	<div class="bar text-ink-1">
		<div class="head row items-center justify-center">
			<q-icon v-if="icon" class="icon q-mr-sm" :name="icon" />
			<div class="title col text-ink-1 text-left">
				{{ label }}
			</div>
		</div>

		<q-btn
			class="close-icon btn-size-md btn-no-text"
			dense
			flat
			icon="sym_r_close"
			color="ink-2"
			v-close-popup
			@click="onCancel"
		>
			<q-tooltip>{{ t('close') }}</q-tooltip>
		</q-btn>
	</div>
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
.bar {
	height: 24px;
	line-height: 24px;
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-bottom: 20px;
	.head {
		height: 24px;
		line-height: 24px;
	}

	.icon {
		font-size: 24px;
	}

	.title {
		font-size: 16px;
		font-weight: 700;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.close-icon {
		margin-top: -12px;
		margin-right: -12px;
	}
}
</style>
