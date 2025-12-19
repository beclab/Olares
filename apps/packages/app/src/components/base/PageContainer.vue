<template>
	<div
		class="page-container-root column items-center"
		:class="
			isAppDetails
				? deviceStore.isMobile
					? 'app-detail-mobile'
					: $q.dark.isActive
					? 'app-detail-dark'
					: 'app-detail-light'
				: ''
		"
		:style="{
			'--titleHeight': `${titleHeight}px`,
			'--maxWidth': `${deviceStore.isMobile ? '100dvw' : '1920px - 208px'}`
		}"
	>
		<div class="page-container-title row justify-start">
			<slot name="title" />
		</div>
		<q-scroll-area
			class="page-container-scroll"
			@scroll="onScroll"
			:visible="false"
			:thumb-style="thumbStyle"
			:bar-style="barStyle"
			:style="{
				'--titleHeight': `${titleHeight}px`
			}"
		>
			<div
				class="page-container-scroll-area column justify-center items-center"
			>
				<div
					v-if="
						!isAppDetails &&
						globalConfig.isOfficial &&
						!deviceStore.isMobile &&
						!hideGradient
					"
					class="fill1"
				/>
				<div
					v-if="
						!isAppDetails &&
						globalConfig.isOfficial &&
						!deviceStore.isMobile &&
						!hideGradient
					"
					class="fill2"
				/>
				<div class="page-container-content">
					<slot name="page" />
				</div>
			</div>
		</q-scroll-area>
	</div>
</template>

<script setup lang="ts">
import globalConfig from '../../api/market/config';
import { useQuasar } from 'quasar';
import { useDeviceStore } from '../../stores/settings/device';

const prop = defineProps({
	modelValue: {
		type: Boolean,
		require: false
	},
	verticalPosition: {
		type: Number,
		default: 0
	},
	titleHeight: {
		type: Number,
		default: 0
	},
	isAppDetails: {
		type: Boolean,
		require: false
	},
	hideGradient: {
		type: Boolean,
		require: false
	}
});

const $q = useQuasar();

const thumbStyle = {
	right: '4px',
	borderRadius: '5px',
	backgroundColor: '#EBEBEB',
	width: '5px',
	opacity: 0.75
};

const barStyle = {
	right: '2px',
	borderRadius: '9px',
	backgroundColor: '#EBEBEB',
	width: '9px',
	opacity: 0.2
};

const emit = defineEmits(['update:modelValue', 'onLoadMore']);
// const slotTitle = useSlots().title

const deviceStore = useDeviceStore();

const onScroll = async (info: any) => {
	emit('update:modelValue', info.verticalPosition > prop.verticalPosition);
	if (info.verticalPercentage === 1) {
		emit('onLoadMore');
	}
};
</script>

<style scoped lang="scss">
.page-container-root {
	height: 100%;
	width: 100%;
	background: $background-1;

	.page-container-title {
		height: var(--titleHeight);
		width: 100%;
		max-width: var(--maxWidth);
	}

	.page-container-scroll {
		width: 100%;
		height: calc(100% - var(--titleHeight));

		.page-container-scroll-area {
			width: 100%;
			height: 100%;
			position: relative;

			.fill1 {
				position: absolute;
				top: -60px;
				right: 14.8%;
				width: 280px;
				height: 280px;
				border-radius: 280px;
				background: $blue-6;
				opacity: 10%;
				filter: blur(80px);
			}

			.fill2 {
				position: absolute;
				top: -43px;
				right: 2%;
				width: 230px;
				height: 230px;
				border-radius: 100px;
				background: $teal-6;
				opacity: 10%;
				filter: blur(80px);
			}
		}

		.page-container-content {
			height: 100%;
			width: 100%;
			max-width: var(--maxWidth);
		}
	}
}

.app-detail-dark {
	background: rgba(18, 25, 29, 1);
}

.app-detail-light {
	background: linear-gradient(175.85deg, #ffffff 3.38%, #f1f9ff 96.62%);
}

.app-detail-mobile {
	background: $background-1;
}
</style>
