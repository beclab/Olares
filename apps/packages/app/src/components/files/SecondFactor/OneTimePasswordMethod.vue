<template>
	<div class="security column items-center justify-center">
		<div class="column items-center justify-center">
			<div class="iconContainer q-mb-md">
				<q-circular-progress
					:value="progressPeriod"
					size="100px"
					:thickness="1"
					color="grey-8"
					class="q-ma-md"
					style="margin: 10px"
				/>
			</div>

			<div class="text-ink-1 text-h6">{{ t('security_verification') }}</div>
			<div class="text-ink-1 text-h7 q-mb-md">
				{{
					t('to_secure_your_account_please_complete_the_following_verification')
				}}
			</div>

			<OtpInput
				ref="otpInputRef"
				input-classes="otp-input"
				separator=""
				:num-inputs="6"
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
import { ref, onMounted, defineExpose } from 'vue';
import { useI18n } from 'vue-i18n';
import OtpInput from 'vue3-otp-input';

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

const emits = defineEmits(['handleOnChange']);

const otpInputRef = ref();
const currentPeriod = ref(0);
const progressPeriod = ref(0);
const currentSecond = ref(0);

const { t } = useI18n();

const handleOnChange = (value: any) => {
	emits('handleOnChange', value);
};

const handleOnComplete = () => {
	// emits('handleOnComplete', value);
};

const clearInput = () => {
	otpInputRef.value.clearInput();
};

defineExpose({
	clearInput,
	currentPeriod
});

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
</script>
<style lang="scss">
.security {
	padding: 20px 0;
	border-radius: 14px;
	.information-icon {
		border-radius: 60px;
		margin: 20px 0;
	}
}

.iconContainer {
	width: 100px;
	height: 100px;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 50px;
	border: 10px solid $white;
	box-sizing: border-box;
}

.otp-input {
	width: 52px;
	height: 64px;
	margin: 0 10px;
	font-size: 20px;
	text-align: center;
	border: none;
	border-radius: 18px;
	background: rgba(255, 255, 255, 0.3);
	box-shadow: 0px 2px 12px 0px rgba(0, 0, 0, 0.2),
		inset 0px 0px 6px 0px rgba(255, 255, 255, 0.7);
	opacity: 1;
	color: $ink-1;
	&:focus-visible {
		border: none;
		padding: 0 !important;
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
input::placeholder {
	font-size: 25px;
	text-align: center;
	font-weight: 600;
}
</style>
