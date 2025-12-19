<template>
	<BtTooltip2>
		<slot></slot>
		<template #tooltip v-if="missingAbility">
			<div class="text-body3 text-ink-tooltip">
				<div>
					{{ missingAbility.message }}
				</div>
				<div
					class="text-info decoration-line cursor-pointer"
					@click="handleOpenMarket(missingAbility.app.title)"
				>
					{{ $t('bex.install_from_market') }}
				</div>
			</div>
		</template>
		<template #tooltip v-else-if="tooltip">
			<div class="text-body3 text-ink-tooltip custom-tooltip-wrapper">
				{{ tooltip }}
			</div>
		</template>
	</BtTooltip2>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import BtTooltip2 from './base/BtTooltip2.vue';
import { useWiseAbility } from 'src/composables/common/useWiseAbility';
import { useI18n } from 'vue-i18n';

defineProps<{
	tooltip?: string;
}>();

const { t } = useI18n();
const { openWiseInMarket, missingAbility } = useWiseAbility();

const handleOpenMarket = (appTitle: string) => {
	openWiseInMarket(appTitle);
};
</script>

<style lang="scss" scoped>
.custom-tooltip-wrapper {
	width: 200px;
	white-space: normal;
}
</style>
