<template>
	<page-container :model-value="false" :title-height="56">
		<template v-slot:title>
			<title-bar :show="true" title="" @onReturn="Router.back()" />
		</template>
		<template v-slot:page>
			<adaptive-layout>
				<template v-slot:pc>
					<div class="topic-detail-container">
						<div class="row justify-between topic-detail-content">
							<q-img
								v-if="topicRef?.detailimg && !$q.screen.xs && !$q.screen.sm"
								class="topic-detail-img"
								ratio="3/4"
								:src="topicRef?.detailimg"
								v-show="!$q.screen.xs && !$q.screen.sm"
							>
								<template v-slot:loading>
									<q-skeleton class="topic-detail-img" style="height: 100%" />
								</template>

								<div
									v-show="topicRef?.apps?.length === 1"
									class="topic-image-app"
								>
									<recommend-app-card
										class="topic-detail-apps"
										:app-name="topicRef?.apps[0]"
										layout="columns"
										:source-id="settingStore.marketSourceId"
										:is-last-line="true"
									/>
								</div>
							</q-img>

							<bt-scroll-area class="topic-detail-rich">
								<div v-html="topicRef?.richtext || ''" />

								<div
									class="topic-detail-apps q-mt-xl"
									v-if="topicRef?.apps?.length > 1"
								>
									<recommend-app-card
										v-for="(item, index) in topicRef?.apps"
										:key="item"
										:app-name="item"
										layout="columns"
										:source-id="settingStore.marketSourceId"
										:is-last-line="index === topicRef?.apps?.length - 1"
									/>
								</div>

								<div
									class="topic-detail-single-app q-mt-xl"
									v-if="topicRef?.apps?.length === 1"
								>
									<topic-app-card
										:app-name="topicRef?.apps[0]"
										layout="columns"
										:source-id="settingStore.marketSourceId"
									/>
								</div>

								<div style="height: 32px; width: 100%" />
							</bt-scroll-area>
						</div>
					</div>
				</template>
				<template v-slot:mobile>
					<div class="topic-detail-container-mobile">
						<bt-scroll-area class="topic-detail-rich">
							<q-img
								v-if="topicRef?.mobileDetailImg || topicRef?.detailimg"
								class="topic-detail-img"
								ratio="3/4"
								:src="topicRef?.mobileDetailImg || topicRef?.detailimg"
							>
								<template v-slot:loading>
									<q-skeleton class="topic-detail-img" style="height: 100%" />
								</template>
							</q-img>

							<div
								v-html="topicRef?.mobileRichtext || topicRef?.richtext || ''"
							/>

							<div
								class="topic-detail-apps q-mt-xl"
								v-if="topicRef?.apps?.length > 1"
							>
								<recommend-app-card
									v-for="(item, index) in topicRef?.apps"
									:key="item"
									:app-name="item"
									layout="columns"
									:source-id="settingStore.marketSourceId"
									:is-last-line="index === topicRef?.apps?.length - 1"
								/>
							</div>

							<div
								class="topic-detail-single-app q-mt-xl"
								v-if="topicRef?.apps?.length === 1"
							>
								<topic-app-card
									:app-name="topicRef?.apps[0]"
									layout="columns"
									:source-id="settingStore.marketSourceId"
								/>
							</div>

							<div style="height: 32px; width: 100%" />
						</bt-scroll-area>
					</div>
				</template>
			</adaptive-layout>
		</template>
	</page-container>
</template>

<script setup lang="ts">
import RecommendAppCard from '../../../components/appcard/RecommendAppCard.vue';
import AdaptiveLayout from '../../../components/settings/AdaptiveLayout.vue';
import TopicAppCard from '../../../components/appcard/TopicAppCard.vue';
import PageContainer from '../../../components/base/PageContainer.vue';
import TitleBar from '../../../components/base/TitleBar.vue';
import { useSettingStore } from '../../../stores/market/setting';
import { useCenterStore } from '../../../stores/market/center';
import SimpleWaiter from '../../../utils/simpleWaiter';
import { useRoute, useRouter } from 'vue-router';
import { onActivated } from 'vue-demi';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';
import {
	CONTENT_TYPE,
	getI18nValue,
	TopicInfo
} from '../../../constant/constants';

const Router = useRouter();
const route = useRoute();
const $q = useQuasar();
const settingStore = useSettingStore();
const topicRef = ref<TopicInfo>();
const centerStore = useCenterStore();
const topicId = route.params.topicId;
const category = route.params.category;
const { locale } = useI18n();
const waiter = new SimpleWaiter();

onActivated(async () => {
	waiter.waitForCondition(
		() => {
			return Array.from(centerStore.appSimpleInfoMap.keys()).length > 0;
		},
		() => {
			const pageData = centerStore.pagesMap.get(category);
			const topicData = pageData.find(
				(item) => item.type === CONTENT_TYPE.TOPIC
			);
			const findIndex = topicData.ids.findIndex((item) => item === topicId);
			if (
				findIndex > -1 &&
				topicData &&
				topicData.content &&
				topicData.content[findIndex]
			) {
				topicRef.value = getI18nValue<TopicInfo>(
					topicData.content[findIndex],
					locale.value
				);
			}
		}
	);
});
</script>

<style scoped lang="scss">
.topic-detail-container {
	width: 100%;
	height: calc(100dvh - 56px);
	overflow: hidden;

	.topic-detail-content {
		flex-wrap: nowrap;
		width: 100%;
		height: 100%;
		box-sizing: border-box;
		padding: 0;
		transition: opacity 0.2s ease;

		.topic-detail-img {
			height: calc(100% - 32px);
			border-radius: 12px;
			padding-bottom: 32px;
			flex-shrink: 0;
			margin: 0 32px 0;
			transition: all 0.2s ease;
			opacity: 1;

			::v-deep(.q-img__content > div) {
				pointer-events: all;
				position: absolute;
				width: 100%;
				bottom: 0;
				padding-left: 20px;
				padding-right: 20px;
				color: transparent;
				background: transparent;
			}
		}

		.topic-detail-rich {
			overflow: hidden;
			flex-grow: 1;
		}
	}

	@media (min-width: 1024px) and (max-width: 1600px) {
		.topic-detail-content {
			.topic-detail-img {
				min-width: 480px;
				max-width: 640px;
				width: 33%;
			}

			.topic-detail-rich {
				min-width: 540px;
				width: 67%;
				padding-left: 32px;
				padding-right: 80px;
			}
		}
	}

	@media (min-width: 1601px) {
		.topic-detail-content {
			.topic-detail-img {
				min-width: 640px;
				max-width: 810px;
				width: 33%;
			}

			.topic-detail-rich {
				min-width: 960px;
				width: 67%;
				padding-left: 32px;
				padding-right: 80px;
			}
		}
	}

	@media (max-width: 1023px) {
		.topic-detail-content {
			flex-direction: column;

			.topic-detail-rich {
				width: 100%;
				min-width: auto;
				padding: 0 80px;
			}
		}
	}
}

.topic-detail-container-mobile {
	width: 100%;
	height: calc(100dvh - 56px);
	overflow: hidden;

	.topic-detail-rich {
		width: 100%;
		height: 100%;
		padding: 0 20px;

		.topic-detail-img {
			width: 100%;
			min-height: 209px;
			border-radius: 12px;
			margin-bottom: 32px;
			height: auto;
		}
	}
}

.topic-detail-apps {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 0 16px;
	border-radius: 8px;
	border: 1px solid $separator;
	background: linear-gradient(90deg, #ebf1ff 0%, #ffffff 100%);
	backdrop-filter: blur(6.07811px);

	.dark-mode & {
		background: linear-gradient(90deg, #222637 0%, #1f1f1f 100%);
	}
}

.topic-detail-single-app {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 0 16px;
	border-radius: 8px;
	border: 1px solid $separator;
	background: linear-gradient(180deg, #f2f6ff 0%, white 100%);
	backdrop-filter: blur(6.07811px);

	.dark-mode & {
		background: linear-gradient(90deg, #222637 0%, #1f1f1f 100%);
	}
}
</style>
