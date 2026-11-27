<template>
	<install-configuration
		v-if="isSharedApp"
		:name="t('base.scope')"
		src="sym_r_lan"
		:unit="t('base.cluster_app')"
	/>
	<install-configuration
		:name="t('detail.chart_version')"
		:data="chartVersionLabel"
		:unit="appVersionRelativeTime"
	/>
	<install-configuration
		:name="t('base.developer')"
		src="sym_r_group"
		:unit="appEntry?.developer || '-'"
	/>
	<install-configuration
		:name="t('base.submitter')"
		src="sym_r_verified_user"
		:unit="appEntry?.submitter || '-'"
	/>
	<install-configuration
		:name="t('base.language')"
		:data="language.toUpperCase()"
		:unit="languageUnit"
		:last="!showAccelerator"
	/>
	<install-configuration
		v-if="showAccelerator"
		:name="t('detail.accelerator')"
		:data="acceleratorPrimaryLabel"
		:unit="acceleratorExtraLabel"
		:last="true"
	/>
</template>

<script setup lang="ts">
import InstallConfiguration from '../../../../components/appintro/InstallConfiguration.vue';
import { convertLanguageCodeToName } from '../../../../utils/utils';
import { getPastTime } from '../../../../utils/rss-utils';
import { computed, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { AppEntry } from '../../../../constant/constants';
import { computeModeTitle } from '../../../../constant/compute';

const props = defineProps({
	appEntry: {
		type: Object as PropType<AppEntry>,
		require: true
	},
	appName: {
		type: String,
		require: true
	},
	sourceId: {
		type: String,
		require: true
	},
	isSharedApp: {
		type: Boolean,
		default: false
	}
});

const { t } = useI18n();

const primaryAcceleratorMode = computed(() => props.appEntry?.accelerator?.[0]);

const showAccelerator = computed(
	() => !!props.appEntry?.accelerator && props.appEntry.accelerator.length > 0
);

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

const languageUnit = computed(() => {
	if (props.appEntry?.locale && props.appEntry.locale.length > 1) {
		return t('detail.more_languages', {
			count: props.appEntry.locale.length - 1
		});
	}
	return convertLanguageCodeToName(language.value);
});

const chartVersionLabel = computed(() => props.appEntry?.version || '-');

const appVersionRelativeTime = computed(() => {
	const raw = props.appEntry?.updateTime;
	if (!raw) return '';

	const timestamp = raw < 1e12 ? raw * 1000 : raw;
	const updatedAt = new Date(timestamp);
	if (Number.isNaN(updatedAt.getTime())) return '';

	return getPastTime(new Date(), updatedAt);
});

const modeLabel = (mode: string) => computeModeTitle(mode);

const acceleratorPrimaryLabel = computed(() => {
	const first = primaryAcceleratorMode.value;
	return first ? modeLabel(first.mode) : '-';
});

const acceleratorExtraLabel = computed(() => {
	const count = props.appEntry?.accelerator?.length ?? 0;
	if (count <= 1) return '';
	return t('detail.more_modes', { count: count - 1 });
});
</script>

<style scoped lang="scss"></style>
