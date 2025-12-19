<template>
	<div class="article-header column justify-start items-center">
		<div
			class="row justify-start items-center full-width q-py-md"
			v-if="readerStore.readingFeed"
		>
			<feed-icon :feed="readerStore.readingFeed" size="20px" />
			<span class="article-header__feed text-body3 text-ink-1 q-ml-sm">{{
				readerStore.readingFeed.title
			}}</span>
		</div>
		<div
			class="article-header__title text-h4 text-ink-1 q-py-md"
			id="wise-article-header-title"
		>
			{{ readerStore.readingEntry?.title }}
		</div>
		<q-separator class="article-header__line" />
		<div class="article-header__metadata text-body2 text-ink-3 q-py-md">
			<div
				class="article-header__metadata__info row justify-between items-center"
			>
				<span class="article-header__metadata__info__author">{{
					readerStore.readingEntry?.author
				}}</span>

				<span class="article-header__metadata__date">{{
					formattedDate(readerStore.readingEntry?.published_at)
				}}</span>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import FeedIcon from '../../../../components/rss/FeedIcon.vue';
import { useReaderStore } from '../../../../stores/rss-reader';
import { useI18n } from 'vue-i18n';
import { date } from 'quasar';

const { t } = useI18n();
const readerStore = useReaderStore();
const formattedDate = (datetime: number) => {
	if (!datetime) {
		return t('base.unknown');
	}
	const originalDate = new Date(datetime * 1000);
	return date.formatDate(originalDate, 'YYYY-MM-DD HH:mm:ss');
};
</script>

<style scoped lang="scss">
.article-header {
	width: 100%;

	&__feed {
		max-width: calc(100% - 22px);
		text-transform: uppercase;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	&__title {
		width: 100%;
		-webkit-hyphens: none;
		hyphens: none;
	}

	&__line {
		background: $separator;
		height: 1px;
		width: 100%;
	}

	&__metadata {
		word-break: break-word;
		overflow: hidden;
		text-align: right;
		width: 100%;

		&__info {
			white-space: nowrap;
			margin-right: 8px;

			&__author {
				text-overflow: ellipsis;
				white-space: nowrap;
				overflow: hidden;
			}

			&__separator {
				display: inline-flex;
				width: 4px;
				height: 4px;
				min-width: 4px;
				min-height: 4px;
				background: $separator;
				border-radius: 100%;
				margin: 0 6px 3px;
			}

			&__time {
				text-overflow: ellipsis;
				white-space: nowrap;
				overflow: hidden;
			}
		}

		&__date {
			float: right;
			white-space: nowrap;
		}
	}
}
</style>
