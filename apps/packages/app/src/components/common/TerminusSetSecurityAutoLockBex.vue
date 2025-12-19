<template>
	<div v-if="userStore.passwordReseted">
		<div class="text-h5 text-ink-1">{{ $t('Auto Lock') }}</div>
		<TerminusCheckBox
			class="q-mt-md"
			:model-value="settings.autoLock"
			:label="t('autolock.title')"
			@update:modelValue="changeAutoLock(!settings.autoLock)"
		/>
		<div v-if="settings.autoLock">
			<div class="text-body3 text-ink-3 q-mt-lg">{{ $t('Lock After') }}</div>
			<bt-select
				class="q-mt-xs"
				v-model="lockTimeSelect"
				:options="lockTimeOptions"
				:border="true"
				option-value="value"
				option-label="label"
				@update:modelValue="changeAutoLockDelay"
			/>
			<div
				class="row items-center text-ink-1 text-body3 q-mt-lg"
				v-if="lockTimeSelect < 0"
			>
				<q-slider
					v-model="settings.lockTime"
					:min="10"
					:max="3 * 24 * 60"
					:step="5"
					size="2px"
					label
					:label-value="formatMinutesTime(settings.lockTime)"
					color="yellow-default"
					style="flex: 1; width: auto"
					class="q-mx-sm"
					label-text-color="color-title"
					@change="changeAutoLockDelay"
				/>
				<span>3 {{ t('time.days') }}</span>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { formatMinutesTime } from 'src/utils/utils';
import TerminusCheckBox from 'src/components/common/TerminusCheckBox.vue';
import BtSelect from 'src/components/base/BtSelect.vue';
import {
	lockTimeOptions,
	useAutoLockSettings
} from 'src/composables/mobile/useAutoLockSettings';
import { useUserStore } from 'src/stores/user';

const userStore = useUserStore();

const { t } = useI18n();

const { settings, changeAutoLock, changeAutoLockDelay, lockTimeSelect } =
	useAutoLockSettings();
</script>

<style lang="scss" scoped></style>
