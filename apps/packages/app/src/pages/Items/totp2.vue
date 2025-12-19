<template>
	<div class="row items-center justify-start totp-item">
		<q-circular-progress
			class="q-mr-sm"
			:value="_age"
			size="20px"
			:thickness="0.3"
			color="grey-9"
			track-color="grey-3"
		/>
		<span class="error" v-if="_error">{{ _error }}</span>
		<span v-else class="totp-text text-light-blue-default"
			>{{ token.substring(0, 3) }} {{ token.substring(3, 6) }}
		</span>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, onUnmounted } from 'vue';

import { hotp } from '@didvault/sdk/src/core';
import { base32ToBytes } from '@didvault/sdk/src/core';
import { useI18n } from 'vue-i18n';

export default defineComponent({
	name: 'ToTp',
	props: {
		secret: {
			type: String,
			required: true
		}
	},
	components: {},

	emits: [],
	setup(props: any) {
		const { t } = useI18n();

		const interval = ref(30);
		const token = ref('');
		const _error = ref('');
		const _age = ref(0);
		const _counter = ref(0);
		const _updateTimeout = ref(-1);

		const _update = async (updInt = 2000) => {
			window.clearTimeout(_updateTimeout.value);

			if (!props.secret) {
				token.value = '';
				_age.value = 0;
				return;
			}

			const time = Date.now();

			const counter = Math.floor(time / 1000 / interval.value);
			if (counter !== _counter.value) {
				try {
					token.value = await hotp(base32ToBytes(props.secret), counter);
					_error.value = '';
				} catch (e) {
					token.value = '';
					_error.value = t('errors.invalid_code');
					_age.value = 0;
					return;
				}
				_counter.value = counter;
			}

			_age.value =
				(((Date.now() / 1000) % interval.value) / interval.value) * 100;

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

		return {
			token,
			_error,
			_age
		};
	}
});
</script>

<style lang="scss" scoped>
.totp-item {
	width: 140px;
}
.totp-text {
	font-family: Roboto;
	font-size: 24px;
	line-height: 32px;
	font-weight: 700;
}
</style>
