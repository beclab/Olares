<template>
	<bt-scroll-area class="right-drawer-scroll">
		<div class="row justify-between items-center" style="width: 100%">
			<tab-item
				:transparent="true"
				style="margin-left: 20px"
				:title="t('base.info')"
				:index="1"
				:cur-index="1"
			/>
			<q-btn
				class="q-mr-sm btn-size-sm btn-no-text btn-no-border"
				icon="sym_r_right_panel_close"
				color="ink-2"
				outline
				no-caps
				@click="configStore.setRightDrawerOpen(false)"
			>
				<bt-tooltip :label="t('base.hide_panel')" />
			</q-btn>
		</div>
		<div class="column drawer-entry-layout" v-if="readerStore.hoverEntry">
			<div class="drawer-entry-title text-h6 text-ink-1">
				{{
					readerStore.hoverEntry?.title
						? readerStore.hoverEntry?.title
						: decodeURIComponent(readerStore.hoverEntry?.url)
				}}
			</div>
			<a
				:href="getUrl(readerStore.hoverEntry)"
				target="_blank"
				class="drawer-entry-domain text-body3 text-ink-3"
				>{{ getDomain(readerStore.hoverEntry) }}
				<q-icon
					class="cursor-pointer"
					size="16px"
					color="text-ink-3"
					name="sym_r_open_in_new"
				/>
			</a>
			<div
				v-if="readerStore.readingFeed"
				class="drawer-feed-layout q-mt-xl column justify-start"
			>
				<div class="drawer-feed-info row justify-start items-center">
					<q-img
						class="drawer-feed-icon"
						:src="getFeedIcon(readerStore.readingFeed)"
					/>
					<div class="drawer-feed-content column justify-between items-start">
						<div class="drawer-feed-title text-subtitle2 text-ink-2">
							{{ readerStore.readingFeed?.title }}
						</div>
						<div class="drawer-feed-url text-body3 text-ink-3">
							{{ readerStore.readingFeed?.feed_url }}
						</div>
					</div>
				</div>
				<feed-subscribe-btn />
			</div>

			<div
				v-if="
					entryLabels.length > 0 &&
					configStore.menuChoice.type !== MenuType.Trend
				"
				class="text-body3 drawer-title q-mt-xl"
			>
				{{ t('base.tags') }}
			</div>

			<div
				v-if="configStore.menuChoice.type !== MenuType.Trend"
				class="drawer-tag-info row"
			>
				<template v-for="item in entryLabels" :key="item.id">
					<create-view
						:border="true"
						class="drawer-tag-view"
						:name="item.name"
					/>
				</template>
			</div>

			<note-editor
				v-if="configStore.menuChoice.type !== MenuType.Trend"
				class="q-mt-xl"
			/>

			<div class="drawer-title text-body3 q-mt-xl">
				{{ t('base.metadata') }}
			</div>

			<div class="row drawer-metadata-info">
				<div class="text-body2 drawer-metadata-type">
					{{ t('base.author') }}
				</div>
				<div class="text-body2 drawer-metadata-value">
					{{
						readerStore.hoverEntry?.author
							? readerStore.hoverEntry?.author
							: t('base.unknown')
					}}
				</div>
			</div>

			<div class="drawer-metadata-info row">
				<div class="text-body2 drawer-metadata-type">
					{{ t('base.published') }}
				</div>
				<div class="text-body2 drawer-metadata-value">
					{{ formattedDate(readerStore.hoverEntry?.published_at) }}
				</div>
			</div>

			<div class="drawer-metadata-info row">
				<div class="text-body2 drawer-metadata-type">
					{{ t('base.domain') }}
				</div>
				<div class="text-body2 drawer-metadata-value">
					{{ getDomain(readerStore.hoverEntry) }}
				</div>
			</div>

			<div class="drawer-metadata-info row">
				<div class="text-body2 drawer-metadata-type">
					{{ t('base.created') }}
				</div>
				<div class="text-body2 drawer-metadata-value">
					{{ formattedUtc(readerStore.hoverEntry.createdAt) }}
				</div>
			</div>

			<div class="drawer-metadata-info row">
				<div class="text-body2 drawer-metadata-type">
					{{ t('main.read_later') }}
				</div>
				<div class="text-body2 drawer-metadata-value">
					{{ readerStore.readLater ? 'True' : 'False' }}
				</div>
			</div>

			<div class="drawer-metadata-info row">
				<div class="text-body2 drawer-metadata-type">
					{{ t('base.saved') }}
				</div>
				<div class="text-body2 drawer-metadata-value">
					{{ readerStore.inbox ? 'True' : 'False' }}
				</div>
			</div>
		</div>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import { computed, watch } from 'vue';
import { date } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useConfigStore } from '../../stores/rss-config';
import BtTooltip from '../../components/base/BtTooltip.vue';
import FeedSubscribeBtn from '../../components/rss/FeedSubscribeBtn.vue';
import NoteEditor from '../../components/rss/NoteEditor.vue';
import TabItem from '../../components/rss/TabItem.vue';
import CreateView from '../../components/rss/CreateView.vue';
import { getFeedIcon } from '../../utils/rss-utils';
import { useRssStore } from '../../stores/rss';
import { Entry } from '../../utils/rss-types';
import { MenuType } from '../../utils/rss-menu';
import { useReaderStore } from '../../stores/rss-reader';
import { useTerminusStore } from 'src/stores/terminus';

const rssStore = useRssStore();
const readerStore = useReaderStore();
const configStore = useConfigStore();
const terminusStore = useTerminusStore();
const { t } = useI18n();

const entryLabels = computed(() => {
	if (readerStore.readingEntry) {
		return rssStore.getEntryLabels(readerStore.readingEntry);
	}
	return [];
});

const getDomain = (entry: Entry) => {
	try {
		if (entry.local_file_path && entry.url === entry.local_file_path) {
			return terminusStore.olares_device_id;
		} else {
			const url = new URL(entry.url);
			return url.hostname;
		}
	} catch (e) {
		// console.log(e);
		return decodeURIComponent(entry.url);
	}
};

const getUrl = (entry: Entry) => {
	try {
		if (entry.local_file_path && entry.url === entry.local_file_path) {
			return configStore.getModuleSever('files');
		} else {
			return decodeURIComponent(entry.url);
		}
	} catch (e) {
		// console.log(e);
		return decodeURIComponent(entry.url);
	}
};

const formattedDate = (datetime: number) => {
	if (!datetime) {
		return t('base.unknown');
	}
	const originalDate = new Date(datetime * 1000);
	return date.formatDate(originalDate, 'MMM Do YYYY');
};

const formattedUtc = (utc: string) => {
	if (!utc) {
		return t('base.unknown');
	}
	const originalDate = new Date(utc);
	return date.formatDate(originalDate, 'MMM Do YYYY');
};

watch(
	() => readerStore.hoverEntry,
	() => {
		if (!readerStore.hoverEntry) {
			configStore.setRightDrawerOpen(false);
		}
	},
	{
		immediate: true
	}
);
</script>

<style lang="scss">
.right-drawer-scroll {
	background: $background-6;
	height: 100vh;
	width: 320px;
	max-width: 320px;

	.right-drawer-close {
		margin-right: 12px;
	}

	.drawer-entry-layout {
		padding: 20px 32px;
		width: 320px;
		overflow: hidden;
		max-width: 320px;

		.drawer-entry-title {
			width: 100%;
			color: var(--Grey-10, #1f1814);
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-line-clamp: 3;
			-webkit-box-orient: vertical;
		}

		.drawer-tag-info {
			margin-top: 8px;

			.drawer-tag-view {
				margin-top: 4px;
				margin-right: 12px;
			}
		}

		.drawer-entry-domain {
			color: var(--Grey-05, #adadad);
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-line-clamp: 1;
			-webkit-box-orient: vertical;
		}

		.drawer-feed-layout {
			width: 100%;

			.drawer-feed-info {
				height: 44px;
				width: 100%;

				.drawer-feed-icon {
					width: 32px;
					height: 32px;
					border-radius: 8px;
				}

				.drawer-feed-content {
					width: calc(100% - 40px);
					margin-left: 8px;

					.drawer-feed-title {
						max-width: 200px;
						white-space: nowrap;
						overflow: hidden;
						text-overflow: ellipsis;
					}

					.drawer-feed-url {
						max-width: 200px;
						white-space: nowrap;
						overflow: hidden;
						text-overflow: ellipsis;
					}
				}
			}
		}

		.drawer-title {
			color: $ink-3;
		}

		.drawer-metadata-info {
			margin-top: 12px;
			width: 100%;

			.drawer-metadata-type {
				color: $ink-2;
				max-width: 50%;
				width: 50%;
			}

			.drawer-metadata-value {
				color: $ink-1;
				max-width: 50%;
				width: 50%;
				word-wrap: break-word;
			}
		}
	}
}
</style>
