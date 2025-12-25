<template>
	<install-configuration
		v-if="
			appEntry &&
			appEntry.options &&
			appEntry.options.appScope &&
			appEntry.options.appScope.clusterScoped
		"
		:name="t('base.scope')"
		src="sym_r_lan"
		:unit="t('base.cluster_app')"
	/>
	<install-configuration
		:name="t('base.developer')"
		src="sym_r_group"
		:unit="appEntry && appEntry.developer"
	/>
	<install-configuration
		:name="t('base.language')"
		:data="language.toUpperCase()"
		:unit="
			appEntry && appEntry.locale && appEntry.locale.length > 1
				? `+ ${appEntry.locale.length - 1} more`
				: convertLanguageCodeToName(language)
		"
	/>
	<install-configuration
		:name="t('detail.require_memory')"
		:data="
			appEntry && appEntry.requiredMemory
				? getValueByUnit(
						appEntry.requiredMemory,
						getSuitableUnit(appEntry.requiredMemory, 'memory')
				  )
				: '-'
		"
		:unit="
			appEntry && appEntry.requiredMemory
				? getSuitableUnit(appEntry.requiredMemory, 'memory')
				: '-'
		"
	/>
	<install-configuration
		:name="t('detail.require_disk')"
		:data="
			appEntry && appEntry.requiredDisk
				? getValueByUnit(
						appEntry.requiredDisk,
						getSuitableUnit(appEntry.requiredDisk, 'memory')
				  )
				: '-'
		"
		:unit="
			appEntry && appEntry.requiredDisk
				? getSuitableUnit(appEntry.requiredDisk, 'memory')
				: '-'
		"
	/>
	<install-configuration
		:name="t('detail.require_cpu')"
		:data="
			appEntry && appEntry.requiredCPU
				? getValueByUnit(
						appEntry.requiredCPU,
						getSuitableUnit(appEntry.requiredCPU, 'cpu')
				  )
				: '-'
		"
		:unit="
			appEntry && appEntry.requiredCPU
				? getSuitableUnit(appEntry.requiredCPU, 'cpu')
				: '-'
		"
	/>
	<install-configuration
		:name="t('detail.require_gpu')"
		:data="
			appEntry && appEntry.requiredGPU
				? getValueByUnit(
						appEntry.requiredGPU,
						getSuitableUnit(appEntry.requiredGPU, 'memory')
				  )
				: '-'
		"
		:unit="
			appEntry && appEntry.requiredGPU
				? getSuitableUnit(appEntry.requiredGPU, 'memory')
				: '-'
		"
		:last="true"
	/>
</template>

<script setup lang="ts">
import InstallConfiguration from '../../../../components/appintro/InstallConfiguration.vue';
import { convertLanguageCodeToName } from '../../../../utils/utils';
import { getSuitableUnit, getValueByUnit } from '../../../../utils/monitoring';
import { computed, PropType } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	appEntry: {
		type: Object as PropType<any>,
		require: true
	},
	appName: {
		type: String,
		require: true
	},
	sourceId: {
		type: String,
		require: true
	}
});

const { t } = useI18n();
const language = computed(() => {
	if (
		props.appEntry &&
		props.appEntry.locale &&
		props.appEntry.locale.length > 0
	) {
		return props.appEntry.locale[0];
	}
	return 'en';
});
</script>

<style scoped lang="scss"></style>
