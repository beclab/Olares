<template>
	<div
		v-if="(!pdf && readerStore.topicArray.length > 0) || pdf"
		class="topic-root row justify-start"
	>
		<transition name="slide">
			<q-btn
				no-caps
				outline
				color="ink-2"
				icon="sym_r_sort"
				v-if="!configStore.articleTopicOpen"
				style="height: 32px; width: 32px"
				class="q-mr-sm btn-size-sm btn-no-text btn-no-border q-ma-md"
				@click="configStore.setArticleTopicOpen(true)"
			>
				<bt-tooltip :label="t('pdf.show_outline')" />
			</q-btn>
		</transition>

		<transition name="slide">
			<div class="topic-panel" v-show="configStore.articleTopicOpen">
				<q-btn
					outline
					no-caps
					color="ink-2"
					icon="sym_r_menu_open"
					style="height: 32px; width: 32px"
					class="q-mr-sm btn-size-sm btn-no-text btn-no-border q-ma-md"
					@click="configStore.setArticleTopicOpen(false)"
				>
					<bt-tooltip :label="t('pdf.hide_outline')" />
				</q-btn>

				<div
					v-if="readerStore.topicArray.length > 0 && !pdf"
					class="cursor-pointer single-line text-h6 level-0"
					:class="[
						readerStore.readingTopic &&
						readerStore.topicArray[0].id === readerStore.readingTopic.id
							? 'text-orange'
							: 'text-ink-1'
					]"
					@click="toggleSelection(readerStore.topicArray[0])"
				>
					{{ readerStore.topicArray[0].text }}
				</div>

				<bt-scroll-area
					class="full-width"
					v-if="!pdf"
					:style="{ height: articleScrollHeight + 'px' }"
				>
					<div class="article-topic-scroll">
						<template
							:key="topic.id"
							v-for="(topic, index) in readerStore.topicArray"
						>
							<div
								v-if="index > 0"
								class="cursor-pointer single-line text-body1"
								:class="[
									readerStore.readingTopic &&
									topic.id === readerStore.readingTopic.id
										? 'text-orange'
										: 'text-ink-2',
									` level-${topic.level}`
								]"
								@click="toggleSelection(topic)"
							>
								{{ topic.text }}
							</div>
						</template>

						<div style="width: 100%; height: 20px" />
					</div>
				</bt-scroll-area>

				<bt-scroll-area
					class="full-width"
					style="height: calc(100vh - 100px)"
					v-if="pdfConfigStore.topicLoad && pdf"
				>
					<div class="pdf-scroll">
						<pdfvuer
							:src="pdfConfigStore.source"
							v-for="i in pdfConfigStore.numPages"
							:key="i"
							:id="i"
							:page="i"
							class="cursor-pointer"
							:class="
								i === pdfConfigStore.pageNum ? 'selected-children' : 'children'
							"
							@click="pdfConfigStore.handleItemClick(i)"
						>
							<template v-slot:loading>
								<div class="loading column justify-center items-center">
									<bt-loading :loading="true" size="30px" />
								</div>
							</template>
						</pdfvuer>
					</div>
				</bt-scroll-area>
			</div>
		</transition>
	</div>
</template>

<script setup lang="ts">
import BtTooltip from '../../../../components/base/BtTooltip.vue';
import BtLoading from '../../../../components/base/BtLoading.vue';
import { useConfigStore } from '../../../../stores/rss-config';
import { useReaderStore } from '../../../../stores/rss-reader';
import { ArticleTopic } from '../../../../utils/rss-types';
import { usePDfStore } from '../../../../stores/pdf';
import { useI18n } from 'vue-i18n';
import pdfvuer from 'pdfvuer';
import { computed } from 'vue';
import { useQuasar } from 'quasar';

const pdfConfigStore = usePDfStore();
const configStore = useConfigStore();
const readerStore = useReaderStore();
const { t } = useI18n();
const $q = useQuasar();

defineProps({
	pdf: {
		type: Boolean,
		default: false
	}
});

function toggleSelection(topic: ArticleTopic) {
	readerStore.readingTopic = {
		...topic,
		jump: true
	};
}

const articleScrollHeight = computed(() => {
	if (readerStore.topicArray.length > 1) {
		const height =
			(readerStore.topicArray.length - 1) * 24 +
			(readerStore.topicArray.length - 2) * 12 +
			32;
		const maxScreenHeight = $q.screen.height - 100;

		return maxScreenHeight < height ? maxScreenHeight : height;
	}

	return 0;
});

console.log('topic ', readerStore.topicArray);
</script>

<style scoped lang="scss">
.topic-root {
	width: 240px;
	height: auto;
	position: absolute;
	z-index: 99;
	background: transparent;
	left: 0;
	top: 4px;

	.topic-panel {
		width: 100%;
		z-index: 100;
		position: absolute;
		border-radius: 12px;
		display: flex;
		flex-direction: column;
		max-height: calc(100vh - 60px);
		background: rgba(255, 255, 255, 0.8);
		backdrop-filter: blur(15px);

		.level-0 {
			max-width: 216px;
			margin-left: 12px;
			padding: 0 8px;
			margin-bottom: 20px;
			-webkit-line-clamp: 2 !important;
		}

		.article-topic-scroll {
			margin-left: 12px;
			height: 100%;
			width: calc(100% - 12px);
			display: flex;
			flex-direction: column;
			gap: 12px;

			.level-1 {
				max-width: 200px;
				margin-left: 8px;
			}

			.level-2 {
				max-width: 180px;
				margin-left: 28px;
			}

			.level-3 {
				max-width: 160px;
				margin-left: 48px;
			}

			.level-4 {
				max-width: 140px;
				margin-left: 68px;
			}

			.level-5 {
				max-width: 120px;
				margin-left: 88px;
			}

			.level-6 {
				max-width: 100px;
				margin-left: 108px;
			}
		}
	}

	.pdf-scroll {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-items: center;

		.children {
			width: 118px;
			background: $background-3;
			border-radius: 8px;
			border: 2px solid $separator;
			padding: 10px 20px;
			margin-bottom: 58px;
		}

		.selected-children {
			@extend .children;
			background: $separator;
			border: 2px solid $orange-default;
		}

		.loading {
			width: 78px;
			height: 100px;
		}
	}

	.slide-enter-active,
	.slide-leave-active {
		transition: transform 0.25s ease;
	}

	.slide-enter-from,
	.slide-leave-to {
		transform: translateX(-100%);
	}
}
</style>
