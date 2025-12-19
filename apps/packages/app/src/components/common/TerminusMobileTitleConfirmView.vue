<template>
	<div class="title-confirm-root">
		<terminus-title-bar :title="title" />
		<q-scroll-area
			:class="
				btnHide ? 'scroll-content-close-hide-btn' : 'scroll-content-close'
			"
		>
			<slot name="content"></slot>
		</q-scroll-area>
		<div class="bottom-view-line" v-if="!btnHide && !bottomLineHide"></div>
		<div class="bottom-view row items-center q-pl-md q-pr-md" v-if="!btnHide">
			<template v-if="$slots.buttons">
				<slot name="buttons" />
			</template>
			<confirm-button
				v-else
				class="set-unlock-pwd-root-button"
				:btn-title="btnTitle"
				@onConfirm="onConfirm"
				@onError="onError"
				:btn-status="btnStatus"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import ConfirmButton from './ConfirmButton.vue';
import TerminusTitleBar from './TerminusTitleBar.vue';
import { i18n } from '../../boot/i18n';
import { ConfirmButtonStatus } from '../../utils/constants';

defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	marginLeft: {
		type: Number,
		required: false,
		default: 12
	},
	btnTitle: {
		type: String,
		required: false,
		default: i18n.global.t('confirm')
	},
	btnStatus: {
		type: Number,
		require: false,
		default: ConfirmButtonStatus.normal
	},
	btnHide: {
		type: Boolean,
		required: false,
		default: false
	},
	bottomLineHide: {
		type: Boolean,
		required: false,
		default: false
	}
});

const emit = defineEmits(['onConfirm', 'onError']);

const onConfirm = () => {
	emit('onConfirm');
};

const onError = () => {
	emit('onError');
};
</script>

<style scoped lang="scss">
.title-confirm-root {
	width: 100%;
	height: 100%;

	.scroll-content-close {
		width: 100%;
		height: calc(100% - 56px - 88px - 3px);
	}

	.scroll-content-close-hide-btn {
		width: 100%;
		height: calc(100% - 56px - 3px);
	}

	.bottom-view {
		height: 88px;
		width: 100%;
	}

	.bottom-view-line {
		background-color: $grey-2;
		height: 1px;
		width: 100%;
	}
}
</style>
