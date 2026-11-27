<template>
	<app-intro-card v-if="appEntry" :title="t('detail.information')">
		<template v-slot:content>
			<app-intro-item
				:title="t('detail.get_support')"
				:link-array="getDocuments"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('detail.website')"
				:link-array="getWebsite"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('detail.chart_version')"
				:content="appEntry?.version"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('base.category')"
				separator=" · "
				:content-array="
					appEntry.categories.map((category) =>
						menuStore.getCategoryName(category)
					)
				"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('base.developer')"
				:content="appEntry?.developer"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('base.submitter')"
				:content="appEntry?.submitter"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('base.language')"
				:content-array="
					appEntry.locale ? convertLanguageCodesToNames(appEntry.locale) : []
				"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('detail.compatibility')"
				:content="compatible"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('detail.platforms')"
				:content-array="appEntry?.supportArch"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('detail.legal')"
				:link-array="legalLinks"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('detail.license')"
				:link-array="licenseLinks"
			/>
			<app-intro-item
				v-if="appEntry?.sourceCode"
				class="q-mt-lg"
				:title="t('detail.source_code')"
				:link-array="sourceCodeLinks"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('detail.app_version')"
				:content="appVersionLabel"
			/>
			<app-intro-item
				v-if="appEntry.versionHistory && appEntry.versionHistory.length > 0"
				class="q-mt-lg"
				:title="t('detail.version_history')"
				@on-link-click="goVersionHistory"
				:link="t('detail.see_all_version')"
			/>
		</template>
	</app-intro-card>
</template>

<script setup lang="ts">
import AppIntroItem from '../../../../components/appintro/AppIntroItem.vue';
import AppIntroCard from '../../../../components/appintro/AppIntroCard.vue';
import { computed, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import {
	AppEntry,
	DEPENDENCIES_TYPE,
	Dependency,
	License,
	TRANSACTION_PAGE
} from '../../../../constant/constants';
import {
	capitalizeFirstLetter,
	convertLanguageCodesToNames
} from '../../../../utils/utils';
import { useRouter } from 'vue-router';
import { useMenuStore } from '../../../../stores/market/menu';
import { pushWithParent } from '../../../../utils/marketBack';

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
	}
});

const { t } = useI18n();
const router = useRouter();
const menuStore = useMenuStore();

const compatible = computed(() => {
	const appItem = props.appEntry?.options.dependencies.find(
		(appInfo: Dependency) => {
			return appInfo.type === DEPENDENCIES_TYPE.system;
		}
	);
	if (appItem) {
		return `${capitalizeFirstLetter(appItem.name)} ${appItem.version}`;
	}
	return '';
});

const appVersionLabel = computed(
	() => props.appEntry?.versionName || props.appEntry?.version || ''
);

type IntroLink = {
	text: string;
	url: string;
};

const createTextLink = (getValue: () => string | undefined, text?: string) =>
	computed<IntroLink[]>(() => {
		const value = getValue();
		if (!value) return [];

		const displayText =
			text ||
			(() => {
				try {
					return new URL(value).hostname;
				} catch {
					return value;
				}
			})();

		return [{ text: displayText, url: value }];
	});

const getDocuments = createTextLink(
	() => props.appEntry?.doc,
	t('base.documents')
);
const getWebsite = createTextLink(() => props.appEntry?.website);
const legalLinks = computed<IntroLink[]>(() => {
	const legal = props.appEntry?.legal as unknown;
	if (!Array.isArray(legal) || legal.length === 0) return [];
	return legal[0] ? [legal[0] as IntroLink] : [];
});
const licenseLinks = computed<License[]>(() => {
	const license = props.appEntry?.license;
	return license && license.length > 0 && license[0] ? [license[0]] : [];
});
const sourceCodeLinks = computed<IntroLink[]>(() => {
	const sourceCode = props.appEntry?.sourceCode;
	if (!sourceCode) return [];
	return [
		{
			text: t('detail.public'),
			url: sourceCode
		}
	];
});

const goVersionHistory = () => {
	pushWithParent(router, {
		name: TRANSACTION_PAGE.Version,
		params: {
			appName: props.appName,
			sourceId: props.sourceId
		},
		query: {
			...router.currentRoute.value.query
		}
	});
};
</script>

<style scoped lang="scss"></style>
