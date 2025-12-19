<template>
	<adaptive-layout>
		<template v-slot:pc>
			<div
				v-if="appAggregation"
				class="my-application-root column items-center"
				@click="goAppDetails"
				@mouseover="hoverRef = true"
				@mouseleave="hoverRef = false"
			>
				<app-featured-image :src="appFeaturedImage">
					<app-icon
						:src="appIcon"
						:skeleton="false"
						:size="108"
						:cs-app="clusterScopedApp"
					/>
				</app-featured-image>
				<div class="my-application-info column justify-start items-start">
					<div
						class="my-application-title-layout row justify-start items-center"
					>
						<div class="title-version-wrapper row justify-start items-center">
							<div class="my-application-title no-wrap text-h6 text-ink-1">
								{{ appTitle }}
							</div>

							<div class="q-ml-xs row justify-start items-center">
								<div
									v-if="
										versionDisplayMode === VERSION_DISPLAY_MODE.NONE &&
										!isUpdate
									"
								/>

								<template
									v-if="
										versionDisplayMode === VERSION_DISPLAY_MODE.NONE && isUpdate
									"
								>
									<div class="my-application-version text-caption text-ink-3">
										{{ myAppVersion }}
									</div>
									<div class="text-subtitle2 text-ink-3 q-ml-sm">→</div>
									<div
										class="my-application-version text-caption text-blue-default q-ml-sm"
									>
										{{ appVersion }}
									</div>
								</template>

								<div
									v-if="versionDisplayMode === VERSION_DISPLAY_MODE.ONLY_MY"
									class="my-application-version text-caption text-ink-3"
								>
									{{ myAppVersion }}
								</div>

								<div
									v-if="versionDisplayMode === VERSION_DISPLAY_MODE.ONLY_TARGET"
									class="my-application-version text-caption text-blue-default"
								>
									{{ appVersion }}
								</div>

								<div
									v-if="versionDisplayMode === VERSION_DISPLAY_MODE.PRIORITY_MY"
									class="my-application-version text-caption"
									:class="myAppVersion ? 'text-ink-3' : 'text-blue-default'"
								>
									{{ myAppVersion ? myAppVersion : appVersion }}
								</div>

								<div
									v-if="
										versionDisplayMode === VERSION_DISPLAY_MODE.PRIORITY_TARGET
									"
									class="my-application-version text-caption"
									:class="appVersion ? 'text-blue-default' : 'text-ink-3'"
								>
									{{ appVersion ? appVersion : myAppVersion }}
								</div>
							</div>
						</div>
					</div>

					<div class="my-application-content text-body3 text-ink-3">
						{{ appDesc }}
					</div>

					<div
						class="my-application-bottom-layout row justify-between items-center"
					>
						<install-button
							v-if="appAggregation"
							style="width: auto"
							:item="appAggregation.app_status_latest"
							:app-name="appName"
							:source-id="sourceId"
							:version="appVersion"
							:larger="true"
							:manager="manager"
						/>
						<app-tag
							v-if="isCloneApp(appAggregation.app_status_latest.status)"
							label="Clone"
							class="text-blue-default q-mt-xs"
						/>
						<app-tag v-else :label="sourceName" class="text-positive q-mt-xs" />
					</div>
				</div>
			</div>
			<div v-else class="my-application-root column items-center">
				<app-featured-image :skeleton="true" />
				<div class="my-application-info column justify-start items-start">
					<div
						class="my-application-title-layout row justify-start items-center"
					>
						<q-skeleton width="80px" height="24px" />
						<q-skeleton
							class="my-application-version"
							width="40px"
							height="20px"
						/>
					</div>

					<q-skeleton width="200px" height="32px" style="margin-top: 2px" />

					<div
						class="my-application-bottom-layout row justify-between items-center"
					>
						<q-skeleton width="88px" height="32px" />
					</div>
				</div>
			</div>
		</template>

		<template v-slot:mobile>
			<div
				v-if="appAggregation"
				class="column"
				:class="disabled ? '' : 'cursor-pointer'"
				@click="goAppDetails"
			>
				<div class="func-app-card-mobile row justify-between items-center">
					<app-icon :src="appIcon" :size="56" :cs-app="clusterScopedApp" />
					<div
						class="func-app-card-mobile-right row justify-between items-center"
					>
						<div
							class="func-app-card-mobile-text column justify-center"
							style="align-items: flex-start"
							:style="
								globalConfig.isOfficial
									? 'width: calc(100% - 22px);'
									: 'width: calc(100% - 108px - 22px);'
							"
						>
							<div
								class="func-app-card-mobile-title-layout text-subtitle2 text-ink-1 row justify-start items-center"
							>
								<div class="func-app-card-mobile-title">
									{{ appTitle }}
								</div>

								<div class="version-container items-center q-ml-xs">
									<div
										v-if="
											versionDisplayMode === VERSION_DISPLAY_MODE.NONE &&
											!isUpdate
										"
									/>

									<div
										v-if="versionDisplayMode === VERSION_DISPLAY_MODE.ONLY_MY"
										class="my-application-version text-caption text-ink-3"
									>
										{{ myAppVersion }}
									</div>

									<div
										v-if="
											versionDisplayMode === VERSION_DISPLAY_MODE.ONLY_TARGET
										"
										class="my-application-version text-caption text-blue-default"
									>
										{{ appVersion }}
									</div>

									<div
										v-if="
											versionDisplayMode === VERSION_DISPLAY_MODE.PRIORITY_MY
										"
										class="my-application-version text-caption"
										:class="myAppVersion ? 'text-ink-3' : 'text-blue-default'"
									>
										{{ myAppVersion ? myAppVersion : appVersion }}
									</div>

									<div
										v-if="
											versionDisplayMode ===
											VERSION_DISPLAY_MODE.PRIORITY_TARGET
										"
										class="my-application-version text-caption"
										:class="appVersion ? 'text-blue-default' : 'text-ink-3'"
									>
										{{ appVersion ? appVersion : myAppVersion }}
									</div>
								</div>
							</div>
							<div
								class="func-app-card-mobile-content text-overline text-ink-3"
							>
								{{ appDesc }}
							</div>

							<app-tag
								v-if="isCloneApp(appAggregation.app_status_latest.status)"
								label="Clone"
								class="text-blue-default q-mt-xs"
							/>
							<app-tag
								v-else
								:label="sourceName"
								class="text-positive q-mt-xs"
							/>
						</div>
						<install-button
							v-if="appAggregation"
							layout="column"
							:item="appAggregation.app_status_latest"
							:app-name="appName"
							:source-id="sourceId"
							:version="appVersion"
							:larger="true"
							:manager="manager"
						/>
					</div>
				</div>
				<template
					v-if="versionDisplayMode === VERSION_DISPLAY_MODE.NONE && isUpdate"
				>
					<div class="text-caption row items-center q-mt-sm">
						<span
							class="text-ink-3"
							v-html="
								t('upgrade_app_version', {
									currentVersion: `<span class='text-blue-default'>${myAppVersion}</span>`,
									targetVersion: `<span class='text-blue-default'>${appVersion}</span>`
								})
							"
						/>
					</div>
					<q-separator color="bg-separator q-mt-md" />
				</template>
			</div>
			<div v-else class="func-app-card-mobile row justify-between items-center">
				<app-icon :skeleton="true" :size="56" />
				<div
					class="func-app-card-mobile-right row justify-between items-center"
				>
					<div class="func-app-card-mobile-text column justify-center">
						<q-skeleton width="60px" height="18px" />
						<q-skeleton width="120px" height="12px" class="q-mt-xs" />
						<q-skeleton width="50px" height="18px" class="q-mt-xs" />
					</div>
					<q-skeleton width="108px" height="32px" />
				</div>
			</div>
		</template>
	</adaptive-layout>
</template>

<script lang="ts" setup>
import AppFeaturedImage from '../../components/appcard/AppFeaturedImage.vue';
import InstallButton from '../../components/appcard/InstallButton.vue';
import AppIcon from '../../components/appcard/AppIcon.vue';
import AppTag from '../../components/appcard/AppTag.vue';
import { VERSION_DISPLAY_MODE } from '../../constant/constants';
import AdaptiveLayout from '../settings/AdaptiveLayout.vue';
import globalConfig from '../../api/market/config';
import useAppCard from './useAppCard';
import { PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { isCloneApp } from '../../constant/config';

const props = defineProps({
	appName: {
		type: String,
		required: false
	},
	sourceId: {
		type: String,
		required: true
	},
	disabled: {
		type: Boolean,
		default: false
	},
	versionDisplayMode: {
		type: Object as PropType<VERSION_DISPLAY_MODE>,
		default: VERSION_DISPLAY_MODE.NONE
	},
	isUpdate: {
		type: Boolean,
		required: false,
		default: false
	},
	manager: {
		type: Boolean,
		required: false,
		default: false
	}
});

const hoverRef = ref(false);
const { t } = useI18n();
const {
	appAggregation,
	clusterScopedApp,
	appIcon,
	appTitle,
	appDesc,
	appVersion,
	myAppVersion,
	appFeaturedImage,
	goAppDetails,
	sourceName
} = useAppCard(props);
</script>
<style lang="scss" scoped>
.my-application-root {
	width: 100%;
	height: auto;
	overflow: hidden;

	.my-application-info {
		width: 100%;
		height: 112px;

		.my-application-title-layout {
			width: 100%;
			display: flex;
			align-items: center;
			margin-top: 12px;

			.title-version-wrapper {
				flex: 1;
				display: flex;
				align-items: center;
				overflow: hidden;

				.my-application-title {
					max-width: calc(100% - 120px);
					white-space: nowrap;
					margin-right: 8px;
					overflow: hidden;
					text-overflow: ellipsis;
					white-space: nowrap;
					display: -webkit-box;
					-webkit-line-clamp: 1;
					-webkit-box-orient: vertical;
					word-break: break-all;
				}

				.version-container {
					display: inline-flex;
					align-items: center;
					flex-shrink: 0;
					min-width: 120px;

					.my-application-version {
						white-space: nowrap;
					}
				}
			}
		}

		.my-application-content {
			height: 32px;
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-line-clamp: 2;
			-webkit-box-orient: vertical;
		}

		.my-application-bottom-layout {
			width: 100%;
			height: 32px;
			margin-top: 10px;
		}
	}
}

.func-app-card-mobile {
	width: 100%;
	height: 80px;
	overflow: hidden;

	.func-app-card-mobile-right {
		width: calc(100% - 56px);
		height: 100%;
		position: relative;

		.func-app-card-mobile-text {
			height: 56px;
			padding-left: 12px;

			.func-app-card-mobile-title-layout {
				width: 100%;
				overflow: hidden;

				.func-app-card-mobile-title {
					max-width: calc(100% - 40px);
					text-overflow: ellipsis;
					display: -webkit-box;
					-webkit-line-clamp: 1;
					-webkit-box-orient: vertical;
					word-break: break-all;
				}

				.my-application-version {
					flex: 1;
					white-space: nowrap;
				}
			}

			.func-app-card-mobile-content {
				overflow: hidden;
				text-overflow: ellipsis;
				display: -webkit-box;
				-webkit-line-clamp: 1;
				-webkit-box-orient: vertical;
			}
		}
	}
}
</style>
