<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('dialog.edit_rss_Feed')"
		@onSubmit="onOK"
		:okLoading="isLoading ? 'loading' : false"
		:cancel="t('base.cancel')"
		:ok="t('base.confirm')"
	>
		<div class="column">
			<div class="prompt-name q-mb-xs text-body3">
				{{ t('dialog.feed_title') }}
			</div>
			<q-input
				class="prompt-input text-body3"
				v-model="feedTitle"
				borderless
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				no-error-icon
				placeholder=""
			/>

			<div class="prompt-name q-mb-xs text-body3 q-mt-lg">
				{{ t('dialog.feed_url') }}
			</div>
			<q-input
				class="prompt-input text-body3"
				v-model="feedUrl"
				borderless
				readonly
				no-error-icon
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				placeholder=""
			/>

			<div class="prompt-name q-mb-xs text-body3 q-mt-lg">
				{{ t('dialog.feed_id') }}
			</div>
			<q-input
				class="prompt-input text-body3"
				v-model="feedId"
				borderless
				readonly
				no-error-icon
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				placeholder=""
			/>

			<div class="prompt-name q-mb-xs text-body3 q-mt-lg">
				{{ t('dialog.site_url') }}
			</div>
			<q-input
				class="prompt-input text-body3"
				v-model="siteUrl"
				borderless
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				placeholder=""
			/>

			<div class="prompt-name q-mb-xs text-body3 q-mt-lg">
				{{ t('base.description') }}
			</div>
			<edit-view
				height="52px"
				class="prompt-style q-mt-xs"
				v-model="description"
			/>
			<bt-check-box
				style="padding: 0"
				class="q-mt-lg"
				:label="t('dialog.automatically_download_from_the_feed')"
				v-model="autoDownload"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { Feed } from '../../../utils/rss-types';
import BtCheckBox from '../../rss/BtCheckBox.vue';
import { useRssStore } from '../../../stores/rss';
import { PropType } from 'vue/dist/vue';
import EditView from '../EditView.vue';
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';
import { updateFeed } from '../../../api/wise';

const props = defineProps({
	feed: {
		type: Object as PropType<Feed>,
		require: false
	}
});

const { t } = useI18n();
const feedTitle = ref<string>('');
const feedUrl = ref<string>('');
const feedId = ref<string>('');
const siteUrl = ref<string>('');
const description = ref<string>('');
const autoDownload = ref<boolean>(true);
const rssStore = useRssStore();
const isLoading = ref(false);
const customRef = ref();
onMounted(() => {
	if (props.feed) {
		feedTitle.value = props.feed.title;
		feedUrl.value = props.feed.feed_url;
		feedId.value = props.feed.id;
		siteUrl.value = props.feed.site_url;
		description.value = props.feed.description;
		autoDownload.value = props.feed.auto_download;
	}
});

const onOK = async () => {
	if (
		props.feed &&
		(feedTitle.value !== props.feed.title ||
			feedUrl.value !== props.feed.feed_url ||
			siteUrl.value !== props.feed.site_url ||
			description.value !== props.feed.description ||
			autoDownload.value !== props.feed.auto_download)
	) {
		isLoading.value = true;

		if (autoDownload.value !== props.feed.auto_download) {
			await rssStore.addFeed(feedUrl.value, autoDownload.value);
		}

		await updateFeed(
			props.feed.feed_url,
			feedTitle.value,
			description.value,
			siteUrl.value,
			autoDownload.value
		);
		await rssStore.syncFeeds();
		isLoading.value = false;
	}
	customRef.value.onDialogOK();
};
</script>

<style scoped lang="scss">
.prompt-name {
	color: $ink-3;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}
.prompt-style {
	border: 1px solid $input-stroke;
	border-radius: 8px;
	color: $ink-3;
}

.prompt-input {
	padding-left: 7px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	color: $ink-3;
	height: 32px;
}
</style>
