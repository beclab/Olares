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
		:data="formatMemory(appEntry?.requiredMemory).value"
		:unit="formatMemory(appEntry?.requiredMemory).unit"
	/>
	<install-configuration
		:name="t('detail.require_disk')"
		:data="formatMemory(appEntry?.requiredDisk).value"
		:unit="formatMemory(appEntry?.requiredDisk).unit"
	/>
	<install-configuration
		:name="t('detail.require_cpu')"
		:data="formatCPU(appEntry?.requiredCPU).value"
		:unit="formatCPU(appEntry?.requiredCPU).unit"
	/>
	<install-configuration
		:name="t('detail.require_gpu')"
		:data="formatMemory(appEntry?.requiredGPU).value"
		:unit="formatMemory(appEntry?.requiredGPU).unit"
		:last="true"
	/>
</template>

<script setup lang="ts">
import InstallConfiguration from '../../../../components/appintro/InstallConfiguration.vue';
import { CPUResource, MemoryResource } from '@icebergtsn/k8s-resources';
import { convertLanguageCodeToName } from '../../../../utils/utils';
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

function splitK8sResource(str: string): { value: string; unit: string } {
	const trimmed = str.trim();
	if (!trimmed) {
		return { value: '-', unit: '-' };
	}

	const match = trimmed.match(/^(\d+(?:\.\d+)?)([a-zA-Z]+)?$/);
	if (!match) {
		return { value: trimmed, unit: '' };
	}

	let value = match[1];
	const unit = match[2] || '';

	const num = Number(value);
	if (Number.isFinite(num)) {
		const fixed = Number(num.toFixed(2));
		value = fixed.toString();
	}

	return { value, unit };
}

function formatCPU(raw: string | number | undefined | null) {
	if (raw === undefined || raw === null || raw === '') {
		return { value: '-', unit: '-' };
	}

	try {
		let res: CPUResource | null = null;
		if (typeof raw === 'number') {
			res = CPUResource.fromCores(raw);
		} else {
			const s = String(raw).trim();
			if (!s) return { value: '-', unit: '-' };
			if (/^\d+(\.\d+)?$/.test(s)) {
				res = CPUResource.fromCores(parseFloat(s));
			} else {
				res = new CPUResource(s);
			}
		}

		const str = res.toString();
		const { value, unit } = splitK8sResource(str);
		return {
			value,
			unit: unit || 'core'
		};
	} catch (e) {
		console.error('formatCPU error', e);
		return { value: '-', unit: '-' };
	}
}

function formatMemory(raw: string | number | undefined | null) {
	if (raw === undefined || raw === null || raw === '') {
		return { value: '-', unit: '-' };
	}

	try {
		let res: MemoryResource | null = null;
		if (typeof raw === 'number') {
			res = MemoryResource.fromBytes(raw);
		} else {
			const s = String(raw).trim();
			if (!s) return { value: '-', unit: '-' };
			if (/^\d+(\.\d+)?$/.test(s)) {
				res = MemoryResource.fromBytes(Number(s));
			} else {
				res = new MemoryResource(s);
			}
		}

		const str = res.toString();
		const { value, unit } = splitK8sResource(str);
		return {
			value,
			unit: unit || 'B'
		};
	} catch (e) {
		console.error('formatMemory error', e);
		return { value: '-', unit: '-' };
	}
}

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
