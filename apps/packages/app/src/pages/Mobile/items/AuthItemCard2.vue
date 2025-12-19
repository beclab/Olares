<template>
	<div
		class="authenticator-card row items-center justify-between"
		:class="'bg-background-1 text-ink-1'"
		@click="onCardClick"
	>
		<q-circular-progress
			:value="ageRef"
			size="16px"
			:thickness="1"
			color="background-1"
			trackColor="ink-1"
		/>
		<div class="row items-center auth-code q-pr-sm">
			<span class="authenticator-card__code text-h5" v-if="errorMessageRef">{{
				errorMessageRef
			}}</span>
			<span class="code-common">
				<template v-if="isHideRef && !errorMessageRef">
					<i v-for="index in 6" :key="index" class="bg-ink-1"></i>
				</template>
				<template v-else>
					<span
						v-for="index in tokenRef.length"
						:key="index"
						class="text-subtitle2 text-ink-1"
					>
						{{ tokenRef[index - 1] }}
					</span>
				</template>
			</span>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue';
import { base32ToBytes, Field, hotp, VaultType } from '@didvault/sdk/src/core';
import { getVaultsByType } from '../../../utils/terminusBindUtils';
import { useI18n } from 'vue-i18n';
import { useUserStore } from '../../../stores/user';

const { t } = useI18n();

const interval = 30;
let counter = 0;

const tokenRef = ref('');
const errorMessageRef = ref('');
const ageRef = ref(0);
const updateTimeoutRef = ref(-1);

const totpFiledRef = ref<Field | undefined>();
const isHideRef = ref(true);

const userStore = useUserStore();

const onCardClick = async () => {
	await userStore.unlockFirst(
		async () => {
			if (!totpFiledRef.value) {
				const itemList = getVaultsByType(VaultType.TerminusTotp);
				if (itemList.length > 0) {
					totpFiledRef.value = itemList[0].fields.find((value) => {
						return value.type === 'totp';
					});
					await refreshTokenRef();
				}
			}
			isHideRef.value = !isHideRef.value;
		},
		{
			info: t('unlock.one_time_unlock_introduce')
		}
	);
};

const _update = async (updInt = 2000) => {
	window.clearTimeout(updateTimeoutRef.value);
	await refreshTokenRef();
	ageRef.value = (((Date.now() / 1000) % interval) / interval) * 100;

	if (updInt) {
		updateTimeoutRef.value = window.setTimeout(() => _update(updInt), updInt);
	}
};

const refreshTokenRef = async () => {
	const time = Date.now();
	const tempCounter = Math.floor(time / 1000 / interval);
	if (totpFiledRef.value && tempCounter !== counter) {
		try {
			tokenRef.value = await hotp(
				base32ToBytes(totpFiledRef.value!.value),
				tempCounter
			);
			errorMessageRef.value = '';
		} catch (e) {
			tokenRef.value = '';
			errorMessageRef.value = t('errors.invalid_code');
			ageRef.value = 0;
			return;
		}
		counter = tempCounter;
	}
};

onMounted(() => {
	_update();
});

onUnmounted(() => {
	window.clearTimeout(updateTimeoutRef.value);
});
</script>

<style lang="scss" scoped>
.authenticator-card {
	width: 100%;
	height: 20px;
	border-radius: 20px;
	// padding: 18px 20px;

	.auth-code {
		height: 100%;
	}

	&__code {
		display: flex;
		align-items: center;
	}

	.code-common {
		font-size: 40px;
		display: flex;
		align-items: center;
		justify-content: center;

		i {
			width: 8px;
			height: 8px;
			border-radius: 4px;
			display: inline-block;
			margin-left: 8px;
		}
		span {
			margin-left: 4px;
		}
	}
}
</style>
