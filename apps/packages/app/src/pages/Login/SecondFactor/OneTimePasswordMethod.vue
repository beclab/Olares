<template>
	<div class="security column items-center justify-center">
		<div class="column items-center justify-center">
			<div
				class="iconContainer"
				:style="{ width: circularSize + 'px', height: circularSize + 'px' }"
			>
				<q-circular-progress
					:value="progress"
					:size="`${circularSize}px`"
					:thickness="1"
					color="white"
					class="q-ma-md"
					style="margin: 10px"
				/>
			</div>

			<div class="second-title q-mb-sm">
				{{ t('otp_title') }}
			</div>
			<div class="second-desc">
				{{ t('otp_message') }}
			</div>

			<OtpInput
				:class="passwordErr ? 'shake' : ''"
				ref="otpInputRef"
				:inputClasses="[
					'otp-input',
					tokenStore.deviceInfo.device === DeviceType.MOBILE
						? 'mobile-size'
						: 'pc-size'
				]"
				separator=""
				:num-inputs="digits"
				:should-auto-focus="true"
				:is-input-num="true"
				:conditionalClass="['one', 'two', 'three', 'four', 'five', 'six']"
				@on-change="handleOnChange"
				@on-complete="handleOnComplete"
			/>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { ref, onMounted, onUnmounted, defineExpose } from 'vue';
import OtpInput from 'vue3-otp-input';
import { useI18n } from 'vue-i18n';
import { useTokenStore } from '../../../stores/token';
import { DeviceType } from '@bytetrade/core';

const props = defineProps({
	digits: {
		type: Number,
		required: false,
		default: 6
	},
	period: {
		type: Number,
		required: false,
		default: 30
	},
	passwordErr: {
		type: Boolean,
		required: false,
		default: false
	}
});

const emits = defineEmits(['handleOnComplete']);
const tokenStore = useTokenStore();

const { t } = useI18n();
const otpInputRef = ref();
const circularSize = ref(
	tokenStore.deviceInfo.device === DeviceType.MOBILE ? 80 : 100
);

const handleOnChange = () => {
	console.log('handleOnChange');
};

const handleOnComplete = (value: any) => {
	emits('handleOnComplete', value);
};

const clearInput = () => {
	otpInputRef.value.clearInput();
};

const timer = ref(0);
const progress = ref(0);

const updateTimer = async (updInt = 2000) => {
	window.clearTimeout(timer.value);
	progress.value = (((Date.now() / 1000) % props.period) / props.period) * 100;

	if (updInt) {
		timer.value = window.setTimeout(() => updateTimer(updInt), updInt);
	}
};

onMounted(() => {
	updateTimer();
});

onUnmounted(() => {
	if (timer.value) {
		clearInterval(timer.value);
	}
});

defineExpose({
	clearInput
});
</script>
<style lang="scss">
.security {
	border-radius: 14px;
	.information-icon {
		border-radius: 60px;
		margin: 20px 0;
	}
}

.iconContainer {
	display: flex;
	align-items: center;
	justify-content: flex-start;
	border-radius: 50px;
	border: 5px solid #ffffff;
	box-sizing: border-box;
	opacity: 0.5;
	margin-bottom: 30px;
}

.second-title {
	font-weight: 600;
	font-size: 24px;
	line-height: 32px;
	color: #ffffff;
}
.second-desc {
	font-size: 14px;
	font-weight: 400;
	color: #ffffff;
	margin-bottom: 20px;
}

.mobile-size {
	width: 44px;
	height: 54px;
	font-size: 23px;
}

.pc-size {
	width: 32px;
	height: 40px;
	font-size: 20px;
}

.otp-input {
	line-height: 40px;
	margin: 0 4px;
	font-weight: 600;
	text-align: center;
	border: none;
	border-radius: 8px;
	background: rgba(255, 255, 255, 0.4);
	opacity: 1;
	color: #ffffff;

	&:focus-visible {
		border: none !important;
		padding: 0 !important;
		outline-style: none;
		outline-color: rgba(255, 255, 255, 0.4);
	}
}

.otp-input.is-complete {
	background: rgba(255, 255, 255, 0.3);
}
.otp-input::-webkit-inner-spin-button,
.otp-input::-webkit-outer-spin-button {
	-webkit-appearance: none;
	margin: 0;
}
</style>
