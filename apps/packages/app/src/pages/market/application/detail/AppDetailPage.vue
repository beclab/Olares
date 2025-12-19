<template>
	<div class="app-detail-root">
		<app-title-bar
			:app-title="appTitle"
			:app-version="appVersion"
			:app-icon="appIcon"
			:show-icon="showIcon(cfgType)"
			:show-header-bar="showHeaderBar"
			:absolute="true"
		>
			<install-button
				:item="appStatus"
				:app-name="appName"
				:source-id="sourceId"
				:version="appVersion"
				:larger="true"
				:image-size="false"
				@on-error="onHandleErrorGroup"
			/>
		</app-title-bar>
		<page-container
			:vertical-position="100"
			v-model="showHeaderBar"
			:is-app-details="true"
		>
			<template v-slot:page>
				<div class="app-detail-page column">
					<div
						class="full-width column"
						:class="
							deviceStore.isMobile
								? 'app-detail-mobile'
								: $q.dark.isActive
								? 'app-detail-top-dark'
								: 'app-detail-top-light'
						"
					>
						<title-bar
							:show="true"
							:show-back="true"
							@on-return="clickReturn"
						/>
						<div
							v-if="!appEntry"
							:style="{
								marginBottom: showIcon(cfgType) ? '0' : '20px',
								paddingLeft: paddingX + 'px',
								paddingRight: paddingX + 'px'
							}"
						>
							<large-app-card
								:source-id="sourceId"
								:app-name="appName"
								:type="cfgType"
								:skeleton="true"
								:is-last-line="true"
							/>
						</div>
						<div
							v-else
							:style="{
								marginBottom: showIcon(cfgType) ? '0' : '20px',
								paddingLeft: paddingX + 'px',
								paddingRight: paddingX + 'px'
							}"
						>
							<large-app-card
								:source-id="sourceId"
								:app-name="appName"
								:type="cfgType"
								:item="appStatus"
								:is-last-line="true"
								@on-error="onHandleErrorGroup"
							/>
						</div>
						<transition
							enter-active-class="animated slideInDown"
							leave-active-class="animated slideOutUp"
						>
							<div
								v-if="
									appStatus && errorGroup.length > 0 && !globalConfig.isOfficial
								"
								class="app-detail-warning bg-red-soft row justify-start"
								:style="{
									paddingLeft: paddingX + 'px',
									paddingRight: paddingX + 'px'
								}"
							>
								<q-icon
									style="margin-top: 2px"
									color="negative"
									size="16px"
									name="sym_r_error"
								/>
								<div
									class="column justify-start"
									style="max-width: calc(100% - 16px)"
								>
									<div class="text-body-3 text-negative q-ml-sm">
										{{ t('my.unable_to_install_app') }}
									</div>
									<template v-for="error in errorGroup" :key="error.code">
										<div class="row justify-start items-start">
											<div class="text-circle bg-negative" />
											<div
												class="text-negative text-body-3"
												style="max-width: calc(100% - 16px)"
											>
												{{ t(error.title, error.variables || {}) }}
											</div>
										</div>
										<template
											v-for="(sub, index) in error.subGroups"
											:key="index"
										>
											<div class="q-ml-lg row justify-start items-start">
												<div class="text-circle bg-negative" />
												<div
													class="text-negative text-body-3"
													style="max-width: calc(100% - 16px)"
												>
													{{ t(sub.title, sub.variables || {}) }}
												</div>
											</div>
										</template>
									</template>
								</div>
							</div>
						</transition>

						<q-separator
							color="separator"
							:style="{
								marginLeft: paddingX + 'px',
								width: `calc(100% - ${paddingX * 2}px)`
							}"
						/>
						<div
							class="app-detail-install-configuration row justify-start items-center"
							:style="{
								'--showSize':
									appEntry &&
									appEntry.options &&
									appEntry.options.appScope &&
									appEntry.options.appScope.clusterScoped
										? '7'
										: '6'
							}"
						>
							<app-install-config
								:app-entry="appEntry"
								:app-name="appName"
								:source-id="sourceId"
							/>
						</div>
					</div>

					<div
						class="full-width column"
						:class="
							deviceStore.isMobile
								? 'app-detail-mobile'
								: $q.dark.isActive
								? 'app-detail-bottom-dark'
								: 'app-detail-bottom-light'
						"
						:style="{ paddingBottom: deviceStore.isMobile ? '20px' : '56px' }"
					>
						<app-store-swiper
							v-if="appEntry && appEntry.promoteImage"
							:data-array="
								appEntry && appEntry.promoteImage ? appEntry.promoteImage : []
							"
							:padding-x="paddingX"
							show-size="3,3,2"
							style="margin-bottom: 20px"
						>
							<template v-slot:swiper="{ item, index }">
								<q-img
									@click="goImagePreview(index)"
									class="promote-img"
									ratio="1.6"
									:src="item"
								>
									<template v-slot:loading>
										<q-skeleton class="promote-img" style="height: 100%" />
									</template>
								</q-img>
							</template>
						</app-store-swiper>

						<adaptive-layout>
							<template v-slot:pc>
								<div
									v-if="appEntry"
									class="bottom-layout row justify-between"
									:style="{
										paddingLeft: paddingX + 'px',
										paddingRight: paddingX + 'px'
									}"
								>
									<div class="column" style="width: calc(70% - 10px)">
										<app-intro-card
											v-if="appEntry.fullDescription"
											:title="
												t('detail.about_this_type', {
													type: cfgType?.toUpperCase()
												})
											"
											:description="
												getI18nValue(appEntry.fullDescription, locale)
											"
										/>

										<app-intro-card
											v-if="appEntry.upgradeDescription"
											class="q-mt-lg"
											:title="t('detail.whats_new')"
											:description="
												getI18nValue(appEntry.upgradeDescription, locale)
											"
										/>

										<app-permission
											:app-entry="appEntry"
											:app-name="appName"
											:source-id="sourceId"
										/>

										<!--								<app-intro-card-->
										<!--									v-if="readMeHtml && appEntry"-->
										<!--									class="q-mt-lg"-->
										<!--									style="width: 100%"-->
										<!--									:title="t('detail.readme')"-->
										<!--								>-->
										<!--									<template v-slot:content>-->
										<!--										<div-->
										<!--											style="width: 100%; max-width: 1086px"-->
										<!--											id="readMe"-->
										<!--											v-html="readMeHtml"-->
										<!--										/>-->
										<!--									</template>-->
										<!--								</app-intro-card>-->
									</div>

									<div
										class="column justify-start items-start"
										style="width: calc(30% - 10px)"
									>
										<app-information
											:app-entry="appEntry"
											:app-name="appName"
											:source-id="sourceId"
										/>

										<app-client
											:app-entry="appEntry"
											:app-name="appName"
											:source-id="sourceId"
										/>

										<app-intro-card
											v-if="dependencies.length > 0"
											class="q-mt-lg"
											:title="t('detail.dependency')"
										>
											<template v-slot:content>
												<recommend-app-card
													:key="app"
													:source-id="sourceId"
													v-for="app in dependencies"
													:app-name="app"
													:is-last-line="true"
													:image-size="false"
												/>
											</template>
										</app-intro-card>

										<app-intro-card
											v-if="cloneApps.length > 0"
											class="q-mt-lg"
											:title="t('Clone')"
										>
											<template v-slot:content>
												<recommend-app-card
													:key="app"
													v-for="app in cloneApps"
													:app-name="app"
													:source-id="sourceId"
													:is-last-line="true"
													:image-size="false"
												/>
											</template>
										</app-intro-card>
									</div>
								</div>
							</template>

							<template v-slot:mobile>
								<div
									v-if="appEntry"
									class="bottom-layout column"
									:style="{
										paddingLeft: paddingX + 'px',
										paddingRight: paddingX + 'px'
									}"
								>
									<app-intro-card
										v-if="appEntry.fullDescription"
										:title="
											t('detail.about_this_type', {
												type: cfgType?.toUpperCase()
											})
										"
										:description="
											getI18nValue(appEntry.fullDescription, locale)
										"
									/>

									<app-intro-card
										v-if="appEntry.upgradeDescription"
										class="q-mt-lg"
										:title="t('detail.whats_new')"
										:description="
											getI18nValue(appEntry.upgradeDescription, locale)
										"
									/>

									<app-information
										class="q-mt-lg"
										:app-entry="appEntry"
										:app-name="appName"
										:source-id="sourceId"
									/>

									<app-permission
										class="q-mt-lg"
										:app-entry="appEntry"
										:app-name="appName"
										:source-id="sourceId"
									/>

									<!--								<app-intro-card-->
									<!--									v-if="readMeHtml && appEntry"-->
									<!--									class="q-mt-lg"-->
									<!--									style="width: 100%"-->
									<!--									:title="t('detail.readme')"-->
									<!--								>-->
									<!--									<template v-slot:content>-->
									<!--										<div-->
									<!--											style="width: 100%; max-width: 1086px"-->
									<!--											id="readMe"-->
									<!--											v-html="readMeHtml"-->
									<!--										/>-->
									<!--									</template>-->
									<!--								</app-intro-card>-->

									<app-client
										class="q-mt-lg"
										:app-entry="appEntry"
										:app-name="appName"
										:source-id="sourceId"
									/>

									<app-intro-card
										v-if="dependencies.length > 0"
										class="q-mt-lg"
										:title="t('detail.dependency')"
									>
										<template v-slot:content>
											<recommend-app-card
												:key="app"
												:source-id="sourceId"
												v-for="app in dependencies"
												:app-name="app"
												:is-last-line="true"
												:image-size="false"
											/>
										</template>
									</app-intro-card>

									<app-intro-card
										v-if="cloneApps.length > 0"
										class="q-mt-lg"
										:title="t('Clone')"
									>
										<template v-slot:content>
											<recommend-app-card
												:key="app"
												v-for="app in cloneApps"
												:app-name="app"
												:source-id="sourceId"
												:is-last-line="true"
												:image-size="false"
											/>
										</template>
									</app-intro-card>
								</div>
							</template>
						</adaptive-layout>
					</div>
				</div>
			</template>
		</page-container>
	</div>
</template>

<script lang="ts" setup>
import RecommendAppCard from '../../../../components/appcard/RecommendAppCard.vue';
import AdaptiveLayout from '../../../../components/settings/AdaptiveLayout.vue';
import InstallButton from '../../../../components/appcard/InstallButton.vue';
import AppIntroCard from '../../../../components/appintro/AppIntroCard.vue';
import AppStoreSwiper from '../../../../components/base/AppStoreSwiper.vue';
import LargeAppCard from '../../../../components/appcard/LargeAppCard.vue';
import AppTitleBar from '../../../../components/appintro/AppTitleBar.vue';
import PageContainer from '../../../../components/base/PageContainer.vue';
import TitleBar from '../../../../components/base/TitleBar.vue';
import AppInstallConfig from './AppInstallConfig.vue';
import AppInformation from './AppInformation.vue';
import AppPermission from './AppPermission.vue';
import AppClient from './AppClient.vue';
import { useDeviceStore } from '../../../../stores/settings/device';
import { useCenterStore } from '../../../../stores/market/center';
import { CFG_TYPE, isCloneApp, showIcon } from '../../../../constant/config';
import { OnUpdateUITask, TaskHandler } from '../TaskHandler';
import globalConfig from '../../../../api/market/config';
import { onActivated, onDeactivated } from 'vue-demi';
import { useRoute, useRouter } from 'vue-router';
import { computed, ref, watch } from 'vue';
import ins_plugin from 'markdown-it-mark';
import markdownit from 'markdown-it';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import hljs from 'highlight.js';
import {
	AppEntry,
	DEPENDENCIES_TYPE,
	getI18nValue,
	TRANSACTION_PAGE
} from '../../../../constant/constants';

const $q = useQuasar();
const route = useRoute();
const router = useRouter();
const deviceStore = useDeviceStore();

const paddingX = computed(() => {
	return deviceStore.isMobile ? 20 : 44;
});

const onHandleErrorGroup = (value) => {
	errorGroup.value = value;
};

const loadStyle = async () => {
	if ($q.dark.isActive) {
		await import('highlight.js/styles/github-dark.css');
	} else {
		await import('highlight.js/styles/github.css');
	}
};
loadStyle();
const md = markdownit({
	html: true,
	linkify: true,
	typographer: true,
	highlight: function (str, lang) {
		// console.log('str ===> ', str);
		// console.log('lang ===> ', lang);
		if (lang && hljs.getLanguage(lang)) {
			try {
				return (
					// '<div class="code-container">' +
					// '<button class="copy-button text-caption q-ml-lg text-grey-8" @click="copyCode(\'' +
					// str +
					// '\')">Copy</button>' +
					'<pre><code class="hljs">' +
					hljs.highlight(str, { language: lang, ignoreIllegals: true }).value +
					'</code></pre>'
					// '</div>'
				);
			} catch (e) {
				console.log('error ===> ', e);
			}
		}

		return (
			'<pre><code class="hljs">' + md.utils.escapeHtml(str) + '</code></pre>'
		);
	}
}).use(ins_plugin);

const showHeaderBar = ref(false);
const dependencies = ref<string[]>([]);
const cloneApps = ref<string[]>([]);

const cfgType = ref<string>(CFG_TYPE.APPLICATION);
// const readMeHtml = ref('');
const { t, locale } = useI18n();
const centerStore = useCenterStore();
const sourceId = route.params.sourceId as string;
const appName = route.params.appName as string;
const taskHandler = new TaskHandler();
const errorGroup = ref([]);

const appAggregation = computed(() => {
	return centerStore.getAppAggregationInfo(appName, sourceId);
});

onActivated(() => {
	taskHandler.addTask(dependenciesTask).addTask(appScopeTask, true);
	// .addTask(markdownTask);
});

onDeactivated(() => {
	errorGroup.value = [];
	taskHandler.reset();
});

watch(
	() => appAggregation.value,
	() => {
		if (appAggregation.value && appAggregation.value.app_full_info) {
			taskHandler.withApp(
				appAggregation.value.app_full_info.app_info.app_entry
			);
		}
	},
	{
		immediate: true,
		deep: true
	}
);

const simpleInfo = computed(
	() => appAggregation.value?.app_simple_latest?.app_simple_info
);
const appEntry = computed(
	() => appAggregation.value?.app_full_info?.app_info?.app_entry
);
const appStatus = computed(() => appAggregation.value?.app_status_latest);

const appIcon = computed(() => simpleInfo.value?.app_icon ?? '');
const appTitle = computed(() => {
	if (isCloneApp(appStatus.value?.status)) {
		return appStatus.value.status.title;
	} else {
		return getI18nValue(simpleInfo.value?.app_title, locale.value) ?? '';
	}
});
const appVersion = computed(() => simpleInfo.value?.app_version ?? '');

const dependenciesTask: OnUpdateUITask = {
	onInit() {
		dependencies.value = [];
	},
	onUpdate(app: AppEntry) {
		if (
			app?.options &&
			app?.options.dependencies &&
			app?.options.dependencies.length > 0
		) {
			app?.options.dependencies.forEach((appInfo) => {
				if (
					appInfo.type === DEPENDENCIES_TYPE.application ||
					appInfo.type === DEPENDENCIES_TYPE.middleware
				) {
					dependencies.value.push(appInfo.name);
				}
			});
		}
	}
};

const appScopeTask: OnUpdateUITask = {
	onInit() {
		cloneApps.value = [];
	},
	onUpdate(app: AppEntry) {
		cloneApps.value = centerStore.getCloneApp(appName, sourceId);
	}
};

// const markdownTask: OnUpdateUITask = {
// 	onInit() {
// 		readMeHtml.value = '';
// 	},
// 	onUpdate(app: AppEntry) {
// 		getMarkdown(app.name).then((result) => {
// 			if (result) {
// 				// console.log(result);
// 				readMeHtml.value = md.render(result);
// 				// console.log(readMeHtml.value);
// 			}
// 		});
// 	}
// };

const goImagePreview = (index: number) => {
	router.push({
		name: TRANSACTION_PAGE.Preview,
		params: {
			appName: appName,
			sourceId: sourceId,
			index: index
		},
		query: {
			...router.currentRoute.value.query
		}
	});
};

const clickReturn = () => {
	router.back();
};
</script>
<style lang="scss" scoped>
.app-detail-root {
	width: 100%;
	height: 100%;

	.app-detail-page {
		width: 100%;
		height: 100%;
		position: relative;

		.text-circle {
			margin: 6px;
			width: 4px;
			height: 4px;
			border-radius: 50%;
		}

		.app-detail-warning {
			margin-left: 44px;
			width: calc(100% - 88px);
			padding: 12px;
			margin-bottom: 20px;
			border-radius: 12px;
		}

		.promote-img {
			border-radius: 20px;
			width: 100%;
		}

		.app-detail-install-configuration {
			width: calc(100% - 40px);
			margin-left: 20px;
			max-width: 100%;
			overflow: scroll;
			height: 120px;
			display: grid;
			align-items: start;
			justify-items: start;
			justify-content: start;
			grid-template-columns: repeat(var(--showSize), minmax(100px, 1fr));
			padding-bottom: 20px;
			padding-top: 20px;

			scrollbar-width: none;
			-ms-overflow-style: none;
		}

		.app-detail-install-configuration::-webkit-scrollbar {
			display: none;
		}

		.bottom-layout {
			height: 100%;
			width: 100%;
		}
	}
}

.app-detail-top-dark {
	background: linear-gradient(175.85deg, #1f1f1f 3.38%, #12191d 96.62%);
}

.app-detail-bottom-dark {
	background: rgba(18, 25, 29, 1);
}

.app-detail-top-light {
	background: linear-gradient(175.85deg, #ffffff 3.38%, #f1f9ff 96.62%);
}

.app-detail-bottom-light {
	background: rgba(241, 249, 255, 1);
}
</style>
