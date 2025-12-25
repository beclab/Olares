<template>
	<div class="security column justify-center items-center">
		<div
			style="height: 48px; width: 48px; border-radius: 24px"
			class="bg-white row items-center justify-center"
		>
			<div
				style="height: 40px; width: 40px; border-radius: 20px"
				class="row items-center justify-center"
			>
				<q-circular-progress
					:value="progressPeriod"
					size="32px"
					:thickness="1"
					color="white"
				/>
			</div>
		</div>

		<div class="text-color-title text-subtitle2 one-time-security-verification">
			{{ t('security_verification') }}
		</div>

		<otp-input
			ref="otpInputRef"
			input-classes="otp-input text-color-title text-h6"
			separator=""
			:num-inputs="digits"
			:should-auto-focus="true"
			:is-input-num="true"
			:conditionalClass="['one', 'two', 'three', 'four', 'five', 'six']"
			@on-change="handleOnChange"
			@on-complete="handleOnComplete"
		/>
		<div class="text-color-title text-body3 one-time-reminder">
			{{ t('enter_one_time_password') }}
		</div>
	</div>
</template>
<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import OtpInput from 'vue3-otp-input';
import { useI18n } from 'vue-i18n';

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
	}
});

const { t } = useI18n();
const emits = defineEmits(['handleOnComplete', 'handleOnChange']);

const otpInputRef = ref();
const currentPeriod = ref(0);
const progressPeriod = ref(0);
const currentSecond = ref(0);

const handleOnChange = (value: any) => {
	emits('handleOnChange', value);
};

const handleOnComplete = (value: any) => {
	emits('handleOnComplete', value);
};

const clearInput = () => {
	otpInputRef.value.clearInput();
};

const initSecond = () => {
	currentSecond.value = new Date().getSeconds();
	currentPeriod.value =
		currentSecond.value > 30
			? Math.round(currentSecond.value / 2)
			: currentSecond.value;
};

onMounted(() => {
	initSecond();

	window.setInterval(() => {
		currentPeriod.value = currentPeriod.value + 1;
		if (currentPeriod.value >= props.period) {
			currentPeriod.value = 0;
		}
		progressPeriod.value = (currentPeriod.value * 100) / props.period;
	}, 1000);
});

defineExpose({
	clearInput,
	currentPeriod
});
</script>
<style lang="scss" scoped>
.security {
	width: 100%;

	.one-time-reminder {
		margin-top: 8px;
	}

	.one-time-security-verification {
		margin-bottom: 8px;
	}
}
</style>
