<template>
	<q-dialog
		v-model="newShowDialog"
		class="thumbnail-view"
		full-width
		full-height
		persistent
	>
		<div class="thumbnail-card">
			<div class="flex justify-between items-center thumbnail-header">
				<div></div>
				<div class="title">Add Thumbnail</div>
				<img class="close-icon" src="/svg/close.svg" @click="closeDialog()" />
			</div>
			<div class="flex justify-between items-center options">
				<div
					class="option-item flex items-center justify-center"
					@click="emits('setThumbnailUrl', '')"
				>
					<img class="item-icon" src="/svg/images.svg" />
					None
				</div>

				<BtUploader
					class="option-item"
					type="avator"
					:size="5"
					fileName="image"
					accept=".jpg, .jpeg, .png, .gif, image/*"
					action="/images/upload/v1"
					@ok="ok"
					@fail="fail"
				>
					<div class="option-item flex items-center justify-center">
						<img class="item-icon" src="/svg/up.svg" />
						Select Files to Upload
					</div>
				</BtUploader>
			</div>
			<div class="choice-images scroll">
				<div class="title">Choice from Tabler Icons</div>
				<div class="row images-view q-col-gutter-lg">
					<div
						class="col-2 image-item"
						v-for="(item, index) in thumbnails"
						:key="index"
					>
						<div class="item" @click="emits('setThumbnailUrl', item.url)">
							<q-avatar size="4.4rem">
								<img class="image" :src="item.url" />
								<div
									class="absolute-bottom-right stutus-option"
									:class="{ select: item.selected }"
								>
									<div class="outer-circle">
										<div class="inner-circle"></div>
									</div>
								</div>
							</q-avatar>
							<div class="image-name">{{ item.name }}</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</q-dialog>
</template>

<script lang="ts" setup>
import { defineProps, ref, watchEffect, computed, ComputedRef } from 'vue';
import { updateThumbnail, PathObject } from '../utils/helper';

const props = defineProps({
	thumbnailUrl: {
		type: String,
		default: ''
	},
	showDialog: {
		type: Boolean,
		required: true
	}
});

const emits = defineEmits(['setShowDialog', 'setThumbnailUrl']);
const newShowDialog = ref<boolean>(props.showDialog);
const newthumbnailUrl = ref<string>(props.thumbnailUrl);

const thumbnails: ComputedRef<PathObject[]> = computed(() => {
	let str = props.thumbnailUrl;
	return updateThumbnail(str, thumbnails.value);
});

watchEffect(() => {
	newShowDialog.value = props.showDialog;
	newthumbnailUrl.value = props.thumbnailUrl;
});

const ok = (response: { success: string; path: string }) => {
	emits('setThumbnailUrl', response.path);
};

const fail = (response: unknown) => {
	console.log('fail', response);
};

const closeDialog = () => {
	emits('setShowDialog', false);
};
</script>
<style lang="scss">
.thumbnail-view {
	width: 53.8rem !important;
	height: 47rem !important;
	.thumbnail-card {
		background-color: #ffffff;
		padding: 2.4rem 0;
	}
	.thumbnail-header {
		margin-right: 2rem;
		margin-left: 2rem;
		.title {
			font-family: 'Roboto';
			font-style: normal;
			font-weight: 500;
			font-size: 1.6rem;
			line-height: 2rem;
			color: #242424;
		}
		.close-icon {
			cursor: pointer;
			width: 1.42rem;
			height: 1.42rem;
		}
	}
	.options {
		margin-top: 3.2rem;
		margin-left: 2rem;
		margin-right: 2rem;
		.option-item {
			height: 4.4rem;
			width: 24.1rem;
			border: 0.1rem solid #e0e0e0;
			border-radius: 0.4rem;
			background: #ffffff;
			cursor: pointer;
			.item-icon {
				width: 1.2rem;
				height: 1.2rem;
				margin-right: 0.8rem;
			}
			.wrap input {
				width: 24.1rem;
				height: 4.4rem;
			}
		}
	}
	.choice-images {
		width: 49rem;
		height: 31rem;
		border: 0.1rem solid #e0e0e0;
		border-radius: 0.6rem;
		margin: auto;
		margin-top: 1.2rem;
		padding: 0 2rem;

		.q-col-gutter-y-lg,
		.q-col-gutter-lg {
			margin-top: 0;
			margin-left: -2rem !important;
		}
		.q-col-gutter-x-lg > *,
		.q-col-gutter-lg > * {
			padding-top: 0px;
			padding-left: 0rem !important;
		}
		.item {
			// background: #7E7E7E;
			height: 7.4rem;
			width: 7.4rem;
			margin-right: 0px;
			text-align: center;
			padding-top: 0.4rem;
		}
		.item:hover {
			background: #f5f5f5;
		}
		.title {
			font-family: 'Roboto';
			font-style: normal;
			font-weight: 500;
			font-size: 1.4rem;
			line-height: 1.6rem;
			color: #7e7e7e;
			margin-top: 2rem;
			margin-bottom: 2rem;
		}
		// 2.32rem 1.16

		.image-item {
			height: auto;
			width: 20% !important;
			margin-bottom: 1.6rem;
			display: flex;
			justify-content: flex-end;
			.image-name {
				margin-top: 0.8rem;
				text-align: center;
				font-size: 1.2rem;
				line-height: 1.4rem;
				font-family: 'Roboto';
				font-style: normal;
				font-weight: 400;
				color: #7e7e7e;
			}
			.image {
				border: 0.1rem solid #e0e0e0;
			}
			.stutus-option {
				padding: 0px;
			}
			.select {
				.outer-circle {
					background-color: #ffffff;
					border: 0.1rem solid #4666f6;
				}
				.inner-circle {
					position: absolute;
					width: 0.8rem;
					height: 0.8rem;
					background-color: #4666f6;
					border-radius: 50%;
				}
			}
			.outer-circle {
				position: relative;
				width: 1.6rem;
				height: 1.6rem;
				background-color: #ffffff;
				border-radius: 50%;
				display: flex;
				justify-content: center;
				align-items: center;
				border: 0.1rem solid #e0e0e0;
			}
		}
	}
	.q-dialog__inner--fullheight > div {
		height: 47rem !important;
	}
	.q-dialog__inner--fullwidth > div {
		width: 53.8rem !important;
	}
}
</style>
