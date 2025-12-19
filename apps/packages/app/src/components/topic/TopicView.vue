<template>
	<adaptive-layout>
		<template v-slot:pc>
			<div v-if="skeleton" class="column justify-start items-start">
				<div class="topic-skeleton-square">
					<q-skeleton class="topic-skeleton-item" />
				</div>
				<q-skeleton width="100px" height="12px" style="margin-top: 12px" />
				<q-skeleton width="200px" height="22px" style="margin-top: 2px" />
				<q-skeleton width="180px" height="18px" style="margin-top: 2px" />
			</div>
			<div v-else class="column justify-start items-start">
				<q-img
					class="topic-item-img"
					:src="item.iconimg ? item.iconimg : '../appIntro.svg'"
					:alt="item.iconimg"
					ratio="1.6"
				>
					<template v-slot:loading>
						<q-skeleton class="topic-item-img" />
					</template>
				</q-img>
				<span class="topic-item-overflow text-ink-3 text-caption q-mt-md">{{
					item.group
				}}</span>
				<span class="topic-item-overflow text-ink-1 text-h6">{{
					item.title
				}}</span>
				<span class="topic-item-overflow text-ink-3 text-body2">{{
					item.des
				}}</span>
			</div>
		</template>

		<template v-slot:mobile>
			<div
				class="column justify-start items-start relative-position"
				:style="{
					'--topicBackground': item.backgroundColor
						? item.backgroundColor
						: '#7856f1'
				}"
			>
				<q-img
					class="topic-item-img-mobile"
					:src="item.iconimg ? item.iconimg : '../appIntro.svg'"
					:alt="item.iconimg"
					ratio="1.6"
				>
					<template v-slot:loading>
						<q-skeleton class="topic-item-img-mobile" />
					</template>
				</q-img>
				<div class="topic-item-color-mobile">
					<div class="topic-item-overflow text-ink-on-brand text-subtitle2-m">
						{{ item.title }}
					</div>
					<div class="topic-item-overflow text-ink-on-brand text-overline-m">
						{{ item.des }}
					</div>
				</div>

				<div class="topic-item-top-mobile">
					<div class="topic-item-overflow text-ink-on-brand text-overline-m">
						{{ item.group }}
					</div>
				</div>
			</div>
		</template>
	</adaptive-layout>
</template>

<script setup lang="ts">
import { PropType } from 'vue';
import { TopicInfo } from '../../constant/constants';
import AdaptiveLayout from '../settings/AdaptiveLayout.vue';

defineProps({
	item: {
		type: Object as PropType<TopicInfo>,
		require: false
	},
	skeleton: {
		type: Boolean,
		default: false
	}
});
</script>

<style scoped lang="scss">
.topic-skeleton-square {
	width: 100%;
	padding-top: 62.5%;
	position: relative;

	.topic-skeleton-item {
		position: absolute;
		border-radius: 12px;
		top: 0;
		width: 100%;
		height: 100%;
	}
}

.topic-item-img {
	border-radius: 12px;
	width: 100%;
	height: 100%;
}

.topic-item-img-mobile {
	border-top-left-radius: 12px;
	border-top-right-radius: 12px;
	width: 100%;
	height: 100%;
}

.topic-item-top-mobile {
	position: absolute;
	border-top-left-radius: 12px;
	border-bottom-right-radius: 12px;
	background-color: var(--topicBackground);
	padding: 4px 8px;
}

.topic-item-color-mobile {
	border-bottom-left-radius: 12px;
	border-bottom-right-radius: 12px;
	background-color: var(--topicBackground);
	padding: 12px 20px;
	width: calc(100% - 40px);
	max-height: 64px;
	overflow: hidden;
}

.topic-item-overflow {
	text-align: left;
	font-style: normal;
	overflow: hidden;
	text-overflow: ellipsis;
	display: -webkit-box;
	-webkit-line-clamp: 1;
	-webkit-box-orient: vertical;
}
</style>
