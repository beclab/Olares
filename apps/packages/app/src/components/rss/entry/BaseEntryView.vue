<template>
	<div class="entry-root q-px-lg" v-if="skeleton">
		<div class="entry-hover row justify-start">
			<div
				class="entry-left-line"
				:style="{ '--lineBackgroundColor': 'transparent' }"
			/>

			<div class="entry-right-box">
				<div class="entry-inner-box row justify-start">
					<div class="entry-img-background row justify-center items-start">
						<q-skeleton class="entry-img" />
					</div>

					<div class="layout-right row justify-between items-start">
						<div class="layout-entry-info column justify-center">
							<q-skeleton width="40%" height="24px" />
							<q-skeleton
								v-if="desc"
								width="100%"
								height="20px"
								style="margin-top: 4px"
							/>
							<q-skeleton width="60%" height="16px" class="layout-feed-info" />
						</div>
						<q-skeleton class="entry-time" />
					</div>
				</div>
			</div>
		</div>
	</div>
	<div v-else class="entry-root q-px-lg">
		<div
			class="entry-hover row justify-start"
			:class="clickable ? 'cursor-pointer' : ''"
			@mouseenter="onHover(true)"
			@mouseleave="onHover(false)"
			:style="{
				'--backgroundColor': selected ? backgroundHover : 'transparent'
			}"
		>
			<div
				class="entry-left-line"
				:style="{
					'--lineBackgroundColor': selected ? orangeDefault : 'transparent'
				}"
			/>

			<div class="entry-right-box" @click="emit('onItemClick')">
				<div class="entry-inner-box row justify-start">
					<div class="entry-img-background row justify-center items-start">
						<q-img class="entry-img" fit="cover" :src="entryImage">
							<template v-slot:loading>
								<q-skeleton width="140px" height="88px" />
							</template>
							<template v-slot:error>
								<q-img
									class="entry-img"
									fit="cover"
									:src="getRequireImage('entry_default_img.svg')"
								/>
							</template>

							<div
								v-if="isFailed || isLoading"
								class="entry-grey bg-background-alpha"
								style="margin-top: 0"
							/>

							<div v-if="isLoading" class="entry-loading column justify-center">
								<bt-loading :loading="true" size="40px" />
							</div>
						</q-img>

						<div class="unread-circle" v-if="showReadStatus && !readStatus" />
					</div>

					<div class="layout-right row justify-between items-start">
						<div
							class="layout-entry-info column justify-center"
							:style="{
								paddingTop: downloadableFileTypes(fileType) ? '6px' : '0',
								width: isHover ? 'calc(100% - 180px)' : 'calc(100% - 120px)'
							}"
						>
							<div
								class="entry-title text-h5"
								:class="
									isLoading || isFailed
										? 'text-ink-3'
										: loss
										? 'text-line-through text-ink-3'
										: 'text-ink-1'
								"
							>
								{{ name }}
							</div>

							<div v-if="desc" class="entry-content text-body2 text-ink-2">
								{{ desc }}
							</div>

							<div class="layout-feed-info row">
								<slot name="bottom" />
							</div>

							<reading-progress-bar :percentage="percentage" />
						</div>

						<div
							class="entry-time text-body3 text-ink-3"
							:style="{ width: isHover ? '160px' : '100px' }"
						>
							{{
								timePrefix && isHover
									? timePrefix + ': ' + getTime()
									: getTime()
							}}
						</div>
					</div>

					<transition name="slide-fade">
						<div
							v-if="selected"
							class="entry-operate-layout row justify-start bg-background-1"
						>
							<slot name="float" />
						</div>
					</transition>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import ReadingProgressBar from '../ReadingProgressBar.vue';
import BtLoading from '../../base/BtLoading.vue';
import { ENTRY_STATUS, FILE_TYPE } from 'src/utils/rss-types';
import { useColor } from '@bytetrade/ui';
import { computed, ref } from 'vue';
import {
	downloadableFileTypes,
	getPastTime,
	getRequireImage
} from 'src/utils/rss-utils';

const props = defineProps({
	name: {
		type: String,
		require: true
	},
	desc: {
		type: String,
		default: ''
	},
	time: {
		type: Number,
		default: 0
	},
	clickable: {
		type: Boolean,
		default: false
	},
	imageUrl: {
		type: String,
		default: ''
	},
	fileType: {
		type: String,
		default: ''
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
		default: false
	},
	readStatus: {
		type: Boolean,
		default: false
	},
	loss: {
		type: Boolean,
		default: false
	},
	status: {
		type: String,
		default: ENTRY_STATUS.Waiting
	},
	timePrefix: {
		type: String,
		default: ''
	},
	percentage: {
		type: Number,
		default: 0
	}
});

const isHover = ref(false);
const { color: backgroundHover } = useColor('background-hover');
const { color: orangeDefault } = useColor('orange-default');

const isLoading = computed(() => {
	return (
		props.status === ENTRY_STATUS.Waiting ||
		props.status === ENTRY_STATUS.Empty ||
		props.status === ENTRY_STATUS.Crawling ||
		props.status === ENTRY_STATUS.Extracting
	);
});

const isFailed = computed(() => {
	return props.status === ENTRY_STATUS.Failed;
});

const onHover = (hover: boolean) => {
	if (isHover.value != hover) {
		isHover.value = hover;
		emit('onHover', hover);
	}
};

const entryImage = computed(() => {
	if (props.imageUrl) {
		return props.imageUrl;
	}

	switch (props.fileType) {
		case FILE_TYPE.VIDEO:
			return getRequireImage('filetype/video.svg');
		case FILE_TYPE.AUDIO:
			return getRequireImage('filetype/radio.svg');
		case FILE_TYPE.PDF:
			return getRequireImage('filetype/pdf.svg');
		case FILE_TYPE.EBOOK:
			return getRequireImage('filetype/ebook.svg');
		case FILE_TYPE.GENERAL:
			return getRequireImage('filetype/general.svg');
		default:
			return getRequireImage('entry_default_img.svg');
	}
});

const emit = defineEmits(['onHover', 'onItemClick']);

function getTime() {
	if (props.time !== 0) {
		return getPastTime(new Date(), new Date(props.time * 1000));
	}
	return '';
}
</script>

<style lang="scss" scoped>
.entry-root {
	height: 132px;
	width: 100%;
	background: var(--backgroundColor);

	.entry-hover {
		height: 100%;
		width: 100%;
		background: var(--backgroundColor);
		border-radius: 4px 12px 12px 4px;
		overflow: hidden;

		.entry-left-line {
			height: 100%;
			width: 4px;
			background: var(--lineBackgroundColor);
		}

		.entry-right-box {
			height: 100%;
			width: calc(100% - 4px);
			padding: 20px;

			.entry-inner-box {
				width: 100%;
				height: 100%;
				position: relative;

				.entry-operate-layout {
					position: absolute;
					right: 0;
					top: calc(50% - 20px);
					width: auto;
					height: 40px;
					border-radius: 20px;
					box-shadow: 0 2px 4px 0 #0000001a;
					padding: 4px;
				}

				.entry-img-background {
					width: 80px;
					height: 100%;

					.unread-circle {
						width: 12px;
						height: 12px;
						position: absolute;
						top: 0;
						margin-top: 2px;
						left: 0;
						background-color: $positive;
						border: 2px solid $background-1;
						border-radius: 50%;
						display: inline-block;
					}

					.entry-img {
						width: 72px;
						height: 72px;
						margin-top: 6px;
						border-radius: 8px;
						position: relative;

						.entry-grey {
							width: 72px;
							height: 72px;
						}

						.entry-loading {
							position: absolute;
							background: #ffffff99;
							width: 100%;
							height: 100%;
						}
					}
				}

				.layout-right {
					margin-left: 20px;
					height: 100%;
					width: calc(100% - 116px);

					.layout-entry-info {
						width: calc(100% - 120px);
						overflow: hidden;
						position: relative;

						.entry-title {
							overflow: hidden;
							display: -webkit-box;
							-webkit-line-clamp: 1;
							-webkit-box-orient: vertical;
							text-overflow: ellipsis;
						}

						.entry-content {
							overflow: hidden;
							text-overflow: ellipsis;
							display: -webkit-box;
							-webkit-line-clamp: 1;
							-webkit-box-orient: vertical;
						}

						.layout-feed-info {
							margin-top: 12px;
							margin-bottom: 12px;
							width: 100%;
							display: flex;
							align-items: center;
						}
					}

					.entry-time {
						margin-left: 20px;
						width: 100px;
						text-align: right;
					}
				}
			}
		}
	}
}
</style>
