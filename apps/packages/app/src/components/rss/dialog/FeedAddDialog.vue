<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('dialog.add_rss_feed')"
		@onSubmit="onOK"
		:okLoading="isAddLoading ? t('loading') : ''"
		:cancel="t('base.cancel')"
		:okDisabled="!selectedFeed"
		:ok="t('dialog.add')"
	>
		<div class="column">
			<div class="prompt-name q-mb-xs text-body3">
				{{ t('main.rss_feeds') }}
			</div>
			<q-input
				class="prompt-input text-body3"
				v-model="inputText"
				borderless
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				:placeholder="t('dialog.enter_the_name_or_url_of_your_feed')"
			/>
		</div>
		<div class="prompt-name text-body3 q-mb-xs q-mt-lg">
			{{ t('main.feeds') }}
		</div>
		<div
			v-if="feedType === FEED_TYPE.NONE"
			class="column text-body3 text-ink-2"
		>
			{{ t('dialog.search_feed_none') }}
		</div>
		<div
			v-if="feedType === FEED_TYPE.LOADING"
			class="loading-item row justify-start items-center"
		>
			<bt-loading :loading="true" size="24px" />
			<div
				class="text-subtitle3 text-ink-2 q-pl-sm"
				style="width: calc(100% - 32px)"
			>
				{{ t('dialog.search_feed_loading') }}
			</div>
		</div>
		<div
			v-if="feedType === FEED_TYPE.EMPTY"
			class="row justify-start items-center"
		>
			<q-img
				class="feed-icon"
				:src="getRequireImage('feed_default_icon.svg')"
			/>
			<div
				class="text-subtitle3 text-ink-2 q-pl-sm"
				style="width: calc(100% - 32px)"
			>
				{{ t('dialog.search_feed_empty') }}
			</div>
		</div>
		<div
			v-if="feedType === FEED_TYPE.DEFAULT"
			class="row justify-start items-center"
		>
			<feed-subscribe-item
				:model-value="
					selectedFeed ? selectedFeed.feed_url === defaultFeed?.feed_url : false
				"
				@update:model-value="
						(value: any, _: Event) => onSelected(defaultFeed, value)
					"
				:feed="defaultFeed"
				:subscribed="defaultFeed?.is_subscribed"
			/>
		</div>
		<div
			v-if="feedType === FEED_TYPE.FEED"
			class="row justify-start items-center full-width"
		>
			<bt-scroll-area class="full-width" :style="`height: ${scrollBarHeight}`">
				<div
					v-for="feed in feedList"
					class="subscribe-item"
					:key="feed.feed_url"
				>
					<feed-subscribe-item
						:model-value="
							selectedFeed ? selectedFeed.feed_url === feed.feed_url : false
						"
						@update:model-value="
								(value: any, _: Event) => onSelected(feed, value)
							"
						:feed="feed"
						:subscribed="feed.is_subscribed"
					/>
				</div>
			</bt-scroll-area>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { getRequireImage } from '../../../utils/rss-utils';
import FeedSubscribeItem from '../FeedSubscribeItem.vue';
import { SearchFeed } from '../../../utils/rss-types';
import { searchFeed } from '../../../api/wise';
import { ref, watch, computed } from 'vue';
import { useRssStore } from '../../../stores/rss';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import BtLoading from '../../base/BtLoading.vue';
import debounce from 'lodash.debounce';
import { useI18n } from 'vue-i18n';

enum FEED_TYPE {
	NONE = 0,
	LOADING = 1,
	EMPTY = 2,
	DEFAULT = 3,
	FEED = 4
}

const feedType = ref<FEED_TYPE>(FEED_TYPE.NONE);
const feedList = ref<SearchFeed[]>([]);
const { t } = useI18n();
const inputText = ref<string>('');
const rssStore = useRssStore();
const defaultFeed = ref<SearchFeed | undefined>();
const selectedFeed = ref<SearchFeed | undefined>();
const isAddLoading = ref(false);
const customRef = ref();
const onOK = () => {
	if (!selectedFeed.value) {
		console.log(selectedFeed.value);
		return;
	}
	switch (feedType.value) {
		case FEED_TYPE.DEFAULT:
			isAddLoading.value = true;
			rssStore
				.addFeed(selectedFeed.value?.id)
				.then(() => {
					BtNotify.show({
						type: NotifyDefinedType.SUCCESS,
						message: t('dialog.add_rss_feed_success')
					});
					customRef.value.onDialogOK();
				})
				.catch((e) => {
					console.log(e.message);
					BtNotify.show({
						type: NotifyDefinedType.FAILED,
						message: t('dialog.add_rss_feed_failed')
					});
				})
				.finally(() => {
					isAddLoading.value = false;
				});
			break;
		case FEED_TYPE.FEED:
			isAddLoading.value = true;
			rssStore
				.addFeed(selectedFeed.value?.feed_url)
				.then(() => {
					BtNotify.show({
						type: NotifyDefinedType.SUCCESS,
						message: t('dialog.add_rss_feed_success')
					});
					customRef.value.onDialogOK();
				})
				.catch((e) => {
					console.log(e.message);
					BtNotify.show({
						type: NotifyDefinedType.FAILED,
						message: t('dialog.add_rss_feed_failed')
					});
				})
				.finally(() => {
					isAddLoading.value = false;
				});
			break;
		default:
			break;
	}
};
const scrollBarHeight = computed(() => {
	if (feedList.value.length <= 0) {
		return '0px';
	}
	if (feedList.value.length > 5) {
		return '232px';
	}
	return feedList.value.length * 48 - 8 + 'px';
});
const search = () => {
	searchFeed(inputText.value?.trim())
		.then((feeds: SearchFeed[]) => {
			feedList.value = feeds;
		})
		.catch((e) => {
			console.log(e.message);
		})
		.finally(async () => {
			if (inputText.value) {
				if (feedList.value.length === 0) {
					if (
						inputText.value.startsWith('http') ||
						inputText.value.startsWith('rsshub') ||
						inputText.value.startsWith('wechat')
					) {
						const find = rssStore.feeds.find(
							(item) => item.url === inputText.value
						);
						defaultFeed.value = {
							id: inputText.value,
							feed_url: t('dialog.still_add_it_and_manager'),
							site_url: '',
							title: t('dialog.no_feeds_were_detected'),
							description: '',
							icon_content: '',
							is_subscribed: !!find,
							icon_type: '',
							create_at: '0',
							updated_at: '0'
						};
						selectedFeed.value = defaultFeed.value;
						feedType.value = FEED_TYPE.DEFAULT;
					} else {
						feedType.value = FEED_TYPE.EMPTY;
					}
				} else {
					const find = feedList.value.find((item) => !item.is_subscribed);
					if (find) {
						selectedFeed.value = find;
					}
					feedType.value = FEED_TYPE.FEED;
				}
			} else {
				feedType.value = FEED_TYPE.NONE;
			}
		});
};
const debounceSearch = debounce(search, 500);
const onSelected = (feed: SearchFeed, value: boolean) => {
	if (value) {
		selectedFeed.value = feed;
	} else {
		selectedFeed.value = undefined;
	}
};
watch(
	() => inputText.value,
	(newValue) => {
		if (newValue) {
			defaultFeed.value = undefined;
			selectedFeed.value = undefined;
			feedType.value = FEED_TYPE.LOADING;
			debounceSearch();
		} else {
			feedType.value = FEED_TYPE.NONE;
		}
	}
);
</script>

<style scoped lang="scss">
.feed-icon {
	width: 32px;
	height: 32px;
}

.loading-item {
	height: 40px;
}

.subscribe-item {
	margin-bottom: 8px;
}

.subscribe-item:last-child {
	margin-bottom: 0;
}

.prompt-name {
	color: $ink-3;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.prompt-input {
	padding-left: 7px;
	padding-right: 7px;
	height: 32px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	color: $ink-3;
}
</style>
