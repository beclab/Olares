<template>
	<div class="mobile-title" v-if="platform === 'mobile'">
		{{ title }}
	</div>

	<div v-else class="bar text-ink-1">
		<div class="head row items-center justify-between">
			<q-icon v-if="icon" class="icon q-mr-sm" :name="icon" />
			<div class="title col text-ink-1 text-left" v-html="title"></div>
		</div>

		<q-btn
			class="close-icon btn-size-md btn-no-text"
			dense
			flat
			icon="sym_r_close"
			color="ink-2"
			v-close-popup
			@click="close"
		>
			<q-tooltip>Close</q-tooltip>
		</q-btn>
	</div>
</template>

<script lang="ts" setup>
defineProps({
	platform: {
		type: String,
		default: 'web',
		required: false
	},
	title: {
		type: String,
		default: '',
		required: false
	},
	icon: {
		type: String,
		default: '',
		required: false
	}
});

const emits = defineEmits(['close']);

const close = () => {
	emits('close');
};
</script>

<style scoped lang="scss">
.mobile-title {
	color: $ink-1;
	font-size: 16px;
	text-align: center;
	font-weight: 500;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.bar {
	height: 24px;
	line-height: 24px;
	width: 100%;
	display: flex;
	align-items: center;
	justify-content: space-between;

	.head {
		height: 24px;
		line-height: 24px;
		max-width: calc(100% - 40px);

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
	}

	.close-icon {
		margin-top: -12px;
		margin-right: -12px;
	}
}
</style>
