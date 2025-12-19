<template>
	<div class="title-bar-root row justify-between items-center">
		<div v-if="progress > 0" class="progress-bar">
			<reading-progress-bar height="2px" :percentage="progress" />
		</div>
		<slot name="before">
			<div
				v-if="useBack"
				class="row justify-center items-center cursor-pointer title-bar-background"
			>
				<q-btn
					class="btn-size-sm btn-no-text btn-no-border"
					color="ink-2"
					outline
					no-caps
					icon="sym_r_arrow_back_ios_new"
					@click="onBackClick"
				>
					<bt-tooltip :label="t('base.back')" />
				</q-btn>
			</div>
			<div v-else />
		</slot>
		<slot name="after" />
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import BtTooltip from '../base/BtTooltip.vue';
import ReadingProgressBar from './ReadingProgressBar.vue';

defineProps({
	useBack: {
		type: Boolean,
		default: false
	},
	progress: {
		type: Number,
		default: 0
	}
});

const { t } = useI18n();

const emit = defineEmits(['onBackClick']);

const onBackClick = () => {
	emit('onBackClick');
};
</script>

<style scoped lang="scss">
.title-bar-root {
	height: 56px;
	width: 100%;
	box-shadow: none;
	min-width: 700px;
	position: relative;

	.progress-bar {
		position: absolute;
		bottom: 0;
		left: 0;
		width: 100%;
		height: 2px;
		background: #f5f5f5;
		box-shadow: 0 2px 4px 0 #0000001a;
		z-index: 10;
		overflow: hidden;

		.progress-bar-fill {
			height: 100%;
			background: linear-gradient(to right, $orange-default, $orange-default);
			width: 0;
			transition: width 0.2s ease;
		}
	}
}

.title-bar-background {
	height: 32px;
	width: 32px;
	margin-left: 12px;
	margin-right: 12px;
}
</style>
