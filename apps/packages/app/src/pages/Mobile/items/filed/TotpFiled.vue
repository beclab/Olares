<template>
	<div class="totp-card row items-center justify-between">
		<div class="row items-start justify-start">
			<div class="text-h5 totp-card__code" v-if="_error">
				{{ _error }}
			</div>
			<div class="text-h5 totp-card__code text-cyan-4" v-else>
				{{ tokenRef.substring(0, 3) + ' ' + tokenRef.substring(3, 6) }}
			</div>
		</div>
		<q-circular-progress
			:value="_age"
			size="24px"
			:thickness="1"
			color="title"
			track-color="cyan-4"
		/>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { hotp } from '@didvault/sdk/src/core';
import { base32ToBytes } from '@didvault/sdk/src/core';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const props = defineProps({
	secret: {
		type: String,
		required: true
	}
});

const interval = ref(30);
const tokenRef = ref('');
const _error = ref('');
const _age = ref(0);
const _counter = ref(0);
const _updateTimeout = ref(-1);

const _update = async (updInt = 2000) => {
	window.clearTimeout(_updateTimeout.value);

	if (!props.secret) {
		tokenRef.value = '';
		_age.value = 0;
		return;
	}

	const time = Date.now();

	const counter = Math.floor(time / 1000 / interval.value);
	if (counter !== _counter.value) {
		try {
			tokenRef.value = await hotp(base32ToBytes(props.secret), counter);
			_error.value = '';
		} catch (e) {
			tokenRef.value = '';
			_error.value = t('errors.invalid_code');
			_age.value = 0;
			return;
		}
		_counter.value = counter;
	}

	_age.value = (((Date.now() / 1000) % interval.value) / interval.value) * 100;

	if (updInt) {
		_updateTimeout.value = window.setTimeout(() => _update(updInt), updInt);
	}
};

onMounted(() => {
	_update();
});

onUnmounted(() => {
	window.clearTimeout(_updateTimeout.value);
});
</script>

<style lang="scss" scoped>
.totp-card {
	height: 35px;
	width: 100%;
}
</style>
