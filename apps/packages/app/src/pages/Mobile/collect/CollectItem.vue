<template>
	<div class="collect-item" :class="{ outline: outline }">
		<div class="q-pa-md" :class="[contentClass]">
			<div class="row items-center q-gutter-x-md no-wrap overflow-hidden">
				<div>
					<slot name="image" />
				</div>

				<div class="overflow-hidden">
					<div class="text-subtitle2 text-ink-1 ellipsis">{{ item.title }}</div>
					<!-- <div class="text-overline text-negative q-mt-xs">
						{{ $t('download.recommend_cookie_to_download') }}
					</div> -->
					<div class="text-body3 text-ink-3 q-mt-xs ellipsis">
						{{ item.url.replace(/.*?\/rss/, 'rss') }}
					</div>
					<q-item-label lines="1" class="text-overline text-ink-2">{{
						item.detail
					}}</q-item-label>
				</div>
			</div>
			<div
				class="row no-wrap items-center flex-gap-xs q-mt-md text-body3"
				v-if="link"
			>
				<q-icon name="sym_r_link" size="16px" />
				<div style="flex: 1" class="ellipsis">
					<a class="link-wrapper" :href="item.url">{{ item.url }}</a>
				</div>
			</div>
			<div side class="q-mt-md row justify-center">
				<slot name="side" />
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { PropType } from 'vue';
import { COOKIE_LEVEL } from '../../../utils/rss-types';
import { BaseCollectInfo } from './utils';
import { DownloadFileRecord } from 'src/platform/interface/bex/rss/utils';

defineProps({
	item: {
		type: Object as PropType<BaseCollectInfo>,
		required: true
	},
	outline: {
		type: Boolean
	},
	link: {
		type: Boolean
	},
	contentClass: {
		type: String
	}
});

const cookieRecommend = (item: DownloadFileRecord) =>
	item.file.cookie_require === COOKIE_LEVEL.RECOMMEND &&
	!item.file.cookie_exist;
console.log(cookieRecommend);
</script>

<style scoped lang="scss">
.collect-item {
	width: 100%;
	padding: 0;
	border-radius: 12px;
	.link-wrapper {
		&:link {
			color: $ink-2;
		}
	}
	&.outline {
		border: 1px solid $separator-2;
	}
}
</style>
