<template>
	<base-entry-view
		:name="
			entry?.title
				? entry.title
				: entry?.url
				? entry.url
				: `[${t('base.no_summary')}]`
		"
		:desc="entry?.summary"
		:status="entry?.status"
		:image-url="entry?.image_url"
		:file-type="entry?.file_type"
		:show-read-status="showReadStatus"
		:selected="selected"
		:clickable="true"
		:loss="fileLost"
		:skeleton="skeleton"
		:read-status="!entry?.unread"
		@on-hover="onHover"
		@on-item-click="onEntryClick"
		:time="getTime(entry)"
		:time-prefix="getTimePrefix()"
		:percentage="entry?.progress"
	>
		<template v-slot:bottom>
			<!--			downloadableFileTypes(entry?.file_type)-->
			<div class="layout-feed-other">
				<div v-if="transferItem" class="row justify-start items-center q-mr-sm">
					<div
						v-if="transferItem?.status === TransferStatus.Error"
						class="row text-orange-default"
					>
						<q-icon size="16px" name="sym_r_error" />

						<div class="q-ml-sm text-body3">
							{{ t('base.download_failed') }}

							<bt-tooltip
								v-if="transferItem.message"
								:label="transferItem.message"
								max-width="240px"
								align="start"
							/>
						</div>
					</div>

					<div
						v-else-if="transferItem?.status === TransferStatus.Canceled"
						class="row text-orange-default"
					>
						<q-icon size="16px" name="sym_r_block" />

						<div class="q-ml-sm text-body3">
							{{ t('download.cancelled') }}
						</div>
					</div>
					<div
						v-else-if="transferItem?.isPaused"
						class="row text-orange-default"
					>
						<q-icon size="16px" name="sym_r_download" />

						<div class="q-ml-sm text-body3">
							{{
								transferItem.progress && transferItem.size
									? t('download.paused') +
									  ':' +
									  Math.round(transferItem.progress * 100) +
									  '%'
									: t('download.paused') +
									  ':' +
									  getValueByUnit(
											transferItem.bytes,
											getSuitableUnit(transferItem.bytes, 'memory')
									  ) +
									  ' ' +
									  getSuitableUnit(transferItem.bytes, 'memory')
							}}
						</div>
					</div>

					<div
						v-else-if="
							transferItem?.status === TransferStatus.Prepare ||
							transferItem?.status === TransferStatus.Pending
						"
						class="row"
					>
						<bt-loading :loading="true" size="16px" />

						<div class="q-ml-sm text-body3 text-orange-default">
							{{ t('pending') }}
						</div>
					</div>

					<div
						v-else-if="
							transferItem?.status === TransferStatus.Canceling ||
							transferItem?.status === TransferStatus.Checking ||
							transferItem?.status === TransferStatus.Resuming ||
							transferItem?.status === TransferStatus.Removing ||
							transferItem?.status === TransferStatus.Running
						"
						class="row"
					>
						<bt-loading :loading="true" size="16px" />

						<div class="q-ml-sm text-body3 text-orange-default">
							{{
								transferItem.progress && transferItem.size
									? t('base.downloading') +
									  ':' +
									  Math.round(transferItem.progress * 100) +
									  '%'
									: t('base.downloading') +
									  ':' +
									  getValueByUnit(
											transferItem.bytes,
											getSuitableUnit(transferItem.bytes, 'memory')
									  ) +
									  ' ' +
									  getSuitableUnit(transferItem.bytes, 'memory')
							}}
						</div>
					</div>
				</div>

				<feed-icon
					v-if="feedRef && feedTitle"
					:feed="feedRef"
					size="16px"
					class="q-mr-sm"
				/>

				<q-icon
					v-else-if="entry?.author"
					name="sym_r_account_circle"
					class="text-ink-3 q-mr-sm"
					size="16px"
				/>

				<q-icon
					v-else-if="entry?.local_file_path"
					name="sym_r_folder"
					class="text-ink-3 q-mr-sm"
					size="16px"
				/>

				<div
					v-if="feedRef && feedTitle"
					class="entry-feed-title text-ink-3 text-body3 q-mr-sm"
				>
					{{ feedTitle ? feedTitle : '' }}
				</div>
				<div
					v-if="feedRef && feedTitle && entry?.author"
					class="auth-linker bg-ink-3 text-body3 q-mr-sm"
				/>
				<div
					v-if="entry?.author"
					class="text-ink-3 entry-feed-author text-body3 q-mr-sm"
				>
					{{ entry?.author }}
				</div>

				<div
					v-if="entry?.author && entry?.local_file_path"
					class="auth-linker bg-ink-3 text-body3 q-mr-sm"
				/>
				<div
					v-if="entry?.local_file_path"
					class="text-ink-3 entry-feed-path text-body3 q-mr-xs"
					:class="
						entry?.local_file_path && entry?.title ? 'cursor-pointer' : ''
					"
					@click.stop="openFile(DRIVER_FILE_PREFIX + entry?.local_file_path)"
				>
					{{ getLocalPath }}
					<q-tooltip>{{ getLocalPath }}</q-tooltip>
				</div>

				<q-icon
					v-if="
						downloadedFileTypes &&
						configStore.menuChoice.type !== MenuType.Trend
					"
					class="q-mr-xs"
					name="sym_r_sell"
					size="16px"
					color="ink-2"
					@click.stop
				>
					<bt-tooltip :label="t('add_tag')" />
					<tag-edit-popup :entry="entry" />
				</q-icon>

				<div
					v-if="entryLabels.length > 0"
					class="ellipsis text-orange-default"
					style="flex: 2"
				>
					<create-view
						v-for="item in entryLabels"
						:key="item.id"
						:name="item.name"
						:selected="selected"
						class="q-mr-xs"
					/>
				</div>
			</div>
		</template>
		<template v-slot:float>
			<q-btn
				class="btn-size-sm btn-no-text btn-no-border btn-circle-border"
				color="ink-2"
				outline
				no-caps
				@click.stop
				icon="sym_r_inbox"
				:loading="inboxLoading"
				:disabled="isInbox || !downloadedFileTypes"
				@click="setReadLater(false)"
			>
				<bt-tooltip :label="t('main.inbox')" />
				<template v-slot:loading>
					<bt-loading :loading="inboxLoading" />
				</template>
			</q-btn>

			<q-btn
				class="btn-size-sm btn-no-text btn-no-border btn-circle-border"
				color="ink-2"
				outline
				no-caps
				@click.stop
				icon="sym_r_schedule"
				:loading="readLaterLoading"
				:disabled="isReadLater || !downloadedFileTypes"
				@click="setReadLater(true)"
			>
				<bt-tooltip :label="t('main.read_later')" />
				<template v-slot:loading>
					<bt-loading :loading="readLaterLoading" />
				</template>
			</q-btn>

			<q-btn
				class="btn-size-sm btn-no-text btn-no-border btn-circle-border"
				color="ink-2"
				outline
				no-caps
				@click.stop
				:loading="isRemoveLoading"
				@click="removeEntry"
				icon="sym_r_do_not_disturb_on"
			>
				<bt-tooltip
					:label="
						configStore.menuChoice.type === MenuType.Trend
							? t('main.not_interested')
							: t('base.remove')
					"
				/>
				<template v-slot:loading>
					<bt-loading :loading="isRemoveLoading" />
				</template>
			</q-btn>
		</template>
	</base-entry-view>
</template>

<script setup lang="ts">
import {
	computed,
	onBeforeUnmount,
	onMounted,
	PropType,
	ref,
	watch
} from 'vue';
import {
	downloadableFileTypes,
	downloadedFileTypes
} from 'src/utils/rss-utils';
import {
	DRIVER_FILE_PREFIX,
	SOURCE_TYPE,
	SORT_TYPE,
	Entry
} from 'src/utils/rss-types';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import FeedIcon from '../FeedIcon.vue';
import CreateView from '../CreateView.vue';
import TagEditPopup from '../TagEditPopup.vue';
import BtTooltip from '../../base/BtTooltip.vue';
import BaseEntryView from './BaseEntryView.vue';
import { busOff, busOn } from 'src/utils/bus';
import { utcToStamp } from 'src/utils/rss-utils';
import BtLoading from '../../base/BtLoading.vue';
import { MenuType } from 'src/utils/rss-menu';
import { useRssStore } from 'src/stores/rss';
import { useConfigStore } from 'src/stores/rss-config';
import { useReaderStore } from 'src/stores/rss-reader';
import BaseCheckBoxDialog from '../../base/BaseCheckBoxDialog.vue';
import { TransferStatus } from 'src/utils/interface/transfer';
import { getSuitableUnit, getValueByUnit } from 'src/utils/monitoring';
import TransferClient from 'src/services/transfer';
import { useTransferStore } from 'src/stores/rss-transfer';
import { useTransfer2Store } from 'src/stores/transfer2';

const $q = useQuasar();
const { t } = useI18n();
const router = useRouter();
const rssStore = useRssStore();
const configStore = useConfigStore();
const readerStore = useReaderStore();
const feedRef = ref(null);
const isInbox = ref(false);
const inboxLoading = ref(false);
const isRemoveLoading = ref(false);
const isReadLater = ref(false);
const readLaterLoading = ref(false);
const transfer2Store = useTransfer2Store();

const props = defineProps({
	entry: {
		type: Object as PropType<Entry>,
		require: false
	},
	skeleton: {
		type: Boolean,
		default: false
	},
	selected: {
		type: Boolean,
		default: false
	},
	showReadStatus: {
		type: Boolean,
		default: true,
		require: true
	},
	timeType: {
		type: String,
		required: true
	}
});

const feedTitle = ref();
const emit = defineEmits(['onSelectedChange', 'onEntryDelete']);
const entryLabels = ref(rssStore.getEntryLabels(props.entry));
const fileLost = computed(() => {
	return (
		downloadableFileTypes(props.entry?.file_type) &&
		props.entry?.crawler &&
		!props.entry.local_file_path
	);
});

const transferItem = computed(() => {
	if (props.entry?.task_ids && props.entry?.task_ids.length > 0) {
		const identify = TransferClient.client.clouder?.taskIdentify(
			String(props.entry?.task_ids[0])
		);
		console.log('entry identify', identify);
		let transferId = transfer2Store.filesCloudTransferMap[identify] || -1;
		if (transferId > 0) {
			const record = transfer2Store.transferMap[transferId];
			console.log(record);
			if (record) {
				return record;
			}
		}
	}

	return null;
});

const getLocalPath = computed(() => {
	try {
		return (
			DRIVER_FILE_PREFIX + decodeURIComponent(props.entry?.local_file_path)
		);
	} catch (e) {
		console.error(props.entry);
		console.error(e);
		return DRIVER_FILE_PREFIX + props.entry?.local_file_path;
	}
});

watch(
	() => rssStore.labels,
	() => {
		entryLabels.value = rssStore.getEntryLabels(props.entry);
	},
	{
		deep: true
	}
);

watch(
	() => props.entry,
	async (newValue) => {
		if (newValue) {
			updateEntry(newValue);
		}
	},
	{
		immediate: true
	}
);

onMounted(() => {
	feedUpdate();
	busOn('feedUpdate', feedUpdate);
});

onBeforeUnmount(() => {
	busOff('feedUpdate', feedUpdate);
});

const feedUpdate = async () => {
	if (props.entry && props.entry.feed_id) {
		const feed = await rssStore.getLocalFeed(props.entry.feed_id);
		if (feed) {
			feedRef.value = feed;
			feedTitle.value = feed.title;
		}
	}
};

const onHover = (hover: boolean) => {
	if (hover) {
		readerStore.entryUpdate(props.entry ? props.entry.id : '');
		emit('onSelectedChange', hover);
	}
};

const getTimePrefix = () => {
	switch (props.timeType) {
		case SORT_TYPE.PUBLISHED:
			return t('base.published_at');
		case SORT_TYPE.CREATED:
			return t('base.create_at');
		case SORT_TYPE.UPDATED:
			return t('base.last_updated');
	}
};

const openFile = (path: string) => {
	if (!path) {
		return;
	}
	const suffix = decodeURIComponent(path);
	const configStore = useConfigStore();
	let url = configStore.getModuleSever('files', 'https:', suffix);
	window.open(url);
};

const getTime = (entry: Entry) => {
	switch (props.timeType) {
		case SORT_TYPE.PUBLISHED:
			return entry.published_at;
		case SORT_TYPE.CREATED:
			return utcToStamp(entry.createdAt);
		case SORT_TYPE.UPDATED:
			return utcToStamp(entry.updatedAt);
	}
};

const setReadLater = async (readLater: boolean) => {
	if (props.entry) {
		if (readLater) {
			readLaterLoading.value = true;
		} else {
			inboxLoading.value = true;
		}
		try {
			await rssStore.markEntryReadLater([props.entry.id], readLater);
		} catch (e) {
			readLaterLoading.value = false;
			inboxLoading.value = false;
			console.log(e);
		}
	}
};

const removeEntry = async () => {
	if (props.entry) {
		if (configStore.menuChoice.type === MenuType.Trend) {
			if (configStore.trendRemoveNotNotify) {
				isRemoveLoading.value = true;
				emit('onEntryDelete', props.entry.url, true);
				isRemoveLoading.value = false;
			} else {
				$q.dialog({
					component: BaseCheckBoxDialog,
					componentProps: {
						label: t('dialog.block_article'),
						content: t('dialog.block_article_desc'),
						modelValue: configStore.trendRemoveNotNotify,
						showCheckbox: true,
						boxLabel: t('dialog.don_not_show_this_confirmation_again')
					}
				})
					.onOk(async (selected) => {
						console.log('remove task ok');
						if (props.entry) {
							isRemoveLoading.value = true;
							configStore.setTrendRemoveNotNotify(selected);
							console.log(props.entry);
							emit('onEntryDelete', props.entry.url, selected);
							isRemoveLoading.value = false;
						} else {
							console.log(props.entry);
							isRemoveLoading.value = false;
						}
					})
					.onCancel(() => {
						console.log('remove task cancel');
					});
			}
		} else {
			console.log(fileLost.value);
			$q.dialog({
				component: BaseCheckBoxDialog,
				componentProps: {
					label: t('dialog.remove_document'),
					content: t('dialog.remove_document_desc'),
					modelValue: configStore.entryRemoveWithFile,
					showCheckbox: props.entry.attachment,
					boxLabel: t('dialog.delete_the_files')
				}
			})
				.onOk(async (selected) => {
					console.log('remove task ok');
					if (props.entry) {
						isRemoveLoading.value = true;
						configStore.setEntryRemoveWithFile(selected);
						console.log(props.entry);
						emit('onEntryDelete', props.entry.url, selected);
						isRemoveLoading.value = false;
					} else {
						console.log(props.entry);
						isRemoveLoading.value = false;
					}
				})
				.onCancel(() => {
					console.log('remove task cancel');
				});
		}
	}
};

async function updateEntry(entry: Entry | null | undefined) {
	if (!entry) {
		return;
	}
	readLaterLoading.value = false;
	inboxLoading.value = false;
	isReadLater.value =
		entry.readlater && entry.sources.includes(SOURCE_TYPE.LIBRARY);
	isInbox.value =
		!entry.readlater && entry.sources.includes(SOURCE_TYPE.LIBRARY);
}

function onEntryClick() {
	if (props.entry) {
		console.log(props.entry);
		router.push({
			name: MenuType.Entry,
			params: {
				path: configStore.menuChoice.type,
				id: props.entry.id
			}
		});
	}
}
</script>

<style lang="scss" scoped>
.layout-feed-other {
	max-width: calc(100% - 26px);
	display: flex;
	align-items: center;
	overflow: hidden;

	.entry-feed-title {
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
		min-width: 20px;
	}

	.entry-feed-path {
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
		min-width: 50px;
		flex: 3;
	}

	.auth-linker {
		width: 4px;
		height: 4px;
		border-radius: 50%;
		min-width: 4px;
	}

	.entry-feed-author {
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
		min-width: 20px;
	}
}
</style>
