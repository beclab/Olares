<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('base.recommend_reason')"
		@onSubmit="onOK"
		:ok="t('base.close')"
	>
		<div class="column">
			<div
				v-if="extra.reason_type === REASON_TYPE.KEYWORD"
				class="text-ink-2 text-body3"
			>
				{{ t('dialog.seeing_this_article_because_keyword') }}

				<div v-for="entry in extra.reason_data" :key="entry.id" class="row">
					<div class="bg-orange-default recommend-circle" />
					<div class="text-orange-default text-body3 recommend-content">
						{{ entry.keyword }}
					</div>
				</div>
			</div>
			<div
				v-if="extra.reason_type === REASON_TYPE.ARTICLE"
				class="text-ink-2 text-body3"
			>
				{{ t('dialog.seeing_this_article_because_entry') }}

				<div v-for="entry in extra.reason_data" :key="entry.id" class="row">
					<div class="bg-orange-default recommend-circle" />
					<div
						class="cursor-pointer text-orange-default text-body3 recommend-content"
						@click="goTrendEntry(entry.id)"
					>
						{{ entry.title }}
					</div>
				</div>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { PropType, ref } from 'vue';
import { EntryExtra, REASON_TYPE } from 'src/utils/rss-types';
import { MenuType } from 'src/utils/rss-menu';
import { useRouter } from 'vue-router';
import { getEntryById } from 'src/api/wise';
import { useRssStore } from '../../../stores/rss';

const { t } = useI18n();
const rssStore = useRssStore();
const router = useRouter();
const customRef = ref();

defineProps({
	extra: {
		type: Object as PropType<EntryExtra>
	}
});

const goTrendEntry = async (id: string) => {
	let entry = rssStore.getLocalRecommend(id);
	const pushToTrend = (id) => {
		router.push({
			name: MenuType.Entry,
			params: {
				path: MenuType.Trend,
				id
			}
		});
	};
	if (entry) {
		console.log('get local recommend : ' + entry.id);
		pushToTrend(entry.id);
	} else {
		entry = await getEntryById(id as string);
		if (entry) {
			console.log('get server recommend : ' + entry.id);
			rssStore.setLocalRecommendEntry(entry.sources[0], [entry]);
			pushToTrend(entry.id);
		} else {
			console.log('no entry get');
		}
	}
};

const onOK = () => {
	customRef.value.onDialogOK();
};
</script>

<style scoped lang="scss">
.recommend-circle {
	width: 4px;
	height: 4px;
	border-radius: 50%;
	margin: 5px;
}
.recommend-content {
	width: calc(100% - 20px);
}
</style>
