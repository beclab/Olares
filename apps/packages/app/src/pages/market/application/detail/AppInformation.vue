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
				:title="t('detail.app_version')"
				:content="appEntry?.versionName"
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
					appEntry.locale ? convertLanguageCodesToNames(appEntry.locale) : ''
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
				:link-array="
					appEntry.legal && appEntry.legal.length > 0 && appEntry.legal[0]
						? [appEntry.legal[0]]
						: []
				"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('detail.license')"
				:link-array="
					appEntry.license && appEntry.license.length > 0 && appEntry.license[0]
						? [appEntry?.license[0]]
						: []
				"
			/>
			<app-intro-item
				v-if="appEntry?.sourceCode"
				class="q-mt-lg"
				:title="t('detail.source_code')"
				:link-array="[
					{
						text: t('detail.public'),
						url: appEntry?.sourceCode
					}
				]"
			/>
			<app-intro-item
				class="q-mt-lg"
				:title="t('detail.chart_version')"
				:content="appEntry?.version"
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
	DEPENDENCIES_TYPE,
	TRANSACTION_PAGE
} from '../../../../constant/constants';
import {
	capitalizeFirstLetter,
	convertLanguageCodesToNames
} from '../../../../utils/utils';
import { useRouter } from 'vue-router';
import { useMenuStore } from '../../../../stores/market/menu';

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
const router = useRouter();
const menuStore = useMenuStore();

const compatible = computed(() => {
	const appItem = props.appEntry?.options.dependencies.find((appInfo) => {
		return appInfo.type === DEPENDENCIES_TYPE.system;
	});
	if (appItem) {
		return `${capitalizeFirstLetter(appItem.name)} ${appItem.version}`;
	}
	return '';
});

const createLink = (field: string, text?: string) =>
	computed(() => {
		const value = props.appEntry?.[field];
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

const getDocuments = createLink('doc', t('base.documents'));
const getWebsite = createLink('website');

const goVersionHistory = () => {
	router.push({
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
