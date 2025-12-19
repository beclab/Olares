<template>
	<div class="rule-checker-root row justify-start items-center">
		<!-- <q-img class="rule-checker-root__img" :src="imgRef" /> -->
		<div
			class="rule-checker-root__img"
			:class="
				!checkPass
					? 'rule-checker-root__img_normal'
					: 'rule-checker-root__img_selected'
			"
		></div>
		<div class="rule-checker-root__label login-sub-title text-ink-2">
			{{ label }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { onMounted, watch, ref } from 'vue';
// import { getRequireImage } from '../../utils/imageUtils';

const props = defineProps({
	label: {
		type: String,
		required: true
	},
	value: {
		type: String,
		required: true,
		default: ''
	},
	pattern: {
		type: String,
		required: true,
		default: ''
	},
	result: {
		type: Boolean,
		required: true,
		default: false
	}
});

const emit = defineEmits('update:result');

let checkRule: RegExp;
// const imgRef = ref(getRequireImage('login/check_no_pass.svg'));
const checkPass = ref(false);

onMounted(() => {
	if (props.pattern) {
		checkRule = new RegExp(props.pattern);
	}
});

watch(
	() => props.value,
	(value) => {
		if (value) {
			const result = checkRule.test(value);
			checkPass.value = result;
			if (result) {
				// imgRef.value = getRequireImage('login/check_pass.svg');
			} else {
				// imgRef.value = getRequireImage('login/check_no_pass.svg');
			}
			emit('update:result', result);
		} else {
			checkPass.value = false;
			// imgRef.value = getRequireImage('login/check_no_pass.svg');
			emit('update:result', false);
		}
	}
);
</script>

<style scoped lang="scss">
.rule-checker-root {
	margin-top: 12px;
	height: 20px;
	width: 100%;

	&__img {
		width: 16px;
		height: 16px;
		gap: 0px;
		border-radius: 4px;
	}
	&__img_normal {
		border: 1px solid $radio-stroke;
		background: transparent;
	}
	&__img_selected {
		border: 1px solid transparent;
		background-image: url('../../assets/login/check_pass.svg');
		background-repeat: no-repeat;
		background-size: 100%;
	}

	&__label {
		margin-left: 8px;
		margin-top: 0;
	}
}
</style>
