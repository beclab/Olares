<template>
	<div
		class="bottom-root"
		:style="{
			'--btn_height': '48px'
		}"
	>
		<q-btn
			:class="getCurrentStatusClass()"
			flat
			padding="0px"
			@click="confirmAction()"
			:disable="btnStatus === ConfirmButtonStatus.disable"
			no-caps
			class="row items-center justify-center"
		>
			<q-icon
				v-if="btnIcon.length"
				:name="btnIcon"
				size="20px"
				class="q-mr-sm"
				color="grey-10"
			/>
			<div>
				{{ btnTitle }}
			</div>
		</q-btn>
	</div>
</template>

<script setup lang="ts">
import { ConfirmButtonStatus } from 'src/utils/constants';
import { useDeviceStore } from 'src/stores/settings/device';
import { i18n } from 'src/boot/i18n';
const deviceStore = useDeviceStore();

const props = defineProps({
	btnTitle: {
		type: String,
		required: false,
		default: i18n.global.t('confirm')
	},
	btnStatus: {
		type: Number,
		required: false,
		default: ConfirmButtonStatus.normal
	},
	bgClasses: {
		type: String,
		required: false,
		default: 'bg-yellow'
	},
	bgDisabledClasses: {
		type: String,
		required: false,
		default: 'bg-yellow-2'
	},
	textClasses: {
		type: String,
		required: false,
		default: 'text-ink-1'
	},
	btnIcon: {
		type: String,
		required: false,
		default: ''
	}
});

const confirmAction = () => {
	if (props.btnStatus === ConfirmButtonStatus.error) {
		emit('onError');
	} else if (props.btnStatus === ConfirmButtonStatus.normal) {
		emit('onConfirm');
	}
};

const emit = defineEmits(['onConfirm', 'onError']);

const getCurrentStatusClass = () => {
	const classes = [props.textClasses];

	classes.push(
		deviceStore.isMobile
			? 'confirm-btn-mobile text-subtitle1'
			: 'confirm-btn-desktop text-subtitle2'
	);

	if (props.btnStatus === ConfirmButtonStatus.error) {
		classes.push('bg-color-error');
	}

	if (props.btnStatus === ConfirmButtonStatus.normal) {
		classes.push(props.bgClasses);
	}

	if (props.btnStatus === ConfirmButtonStatus.disable) {
		classes.push(props.bgDisabledClasses);
	}

	return classes.join(' ');
};
</script>

<style lang="scss" scoped>
.bottom-root {
	height: var(--btn_height);
	width: 100%;
	border-radius: 8px;

	.confirm-btn-mobile {
		width: 100%;
		height: 100%;
		text-align: center;
		border-radius: 8px;
	}

	.confirm-btn-desktop {
		width: 100%;
		height: 100%;
		border-radius: 8px;
	}

	.bg-color-error {
		background-color: $red;
	}

	.confirm-btn[disabled] {
		.q-btn {
			padding: 0;
		}
	}
}
</style>
