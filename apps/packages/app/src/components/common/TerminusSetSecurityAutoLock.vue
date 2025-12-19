<template>
	<div>
		<TerminusCheckBox
			class="q-mt-md"
			:model-value="settings.autoLock"
			:label="t('autolock.title')"
			@update:modelValue="changeAutoLock(!settings.autoLock)"
		/>
		<div
			class="row items-center text-ink-1 text-body3 q-mt-md"
			v-if="settings.autoLock"
		>
			<span>10 {{ t('min') }}</span>
			<q-slider
				v-model="settings.lockTime"
				:min="10"
				:max="3 * 24 * 60"
				:step="5"
				label
				:label-value="formatMinutesTime(settings.lockTime)"
				color="yellow"
				style="flex: 1; width: auto; margin-left: 5px"
				class="q-mx-sm"
				label-text-color="color-title"
				@change="changeAutoLockDelay"
			/>
			<span>3 {{ t('time.days') }}</span>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { formatMinutesTime } from 'src/utils/utils';
import TerminusCheckBox from 'src/components/common/TerminusCheckBox.vue';
import { useAutoLockSettings } from 'src/composables/mobile/useAutoLockSettings';

const { t } = useI18n();

const { settings, changeAutoLock, changeAutoLockDelay } = useAutoLockSettings();
</script>

<style lang="scss" scoped></style>
