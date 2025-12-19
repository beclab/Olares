<template>
	<div
		class="fileItem q-pa-sm q-mb-xs"
		:class="activeItem === file.id ? 'active' : ''"
		v-for="(file, index) in fileData"
		:key="file.id"
		@click="clickItem(file, index)"
	>
		<div class="item-icon">
			<img
				v-if="item?.name === 'Wise' && file.meta?.image_url"
				class="icon"
				:src="
					file.meta?.image_url
						? file.meta?.image_url
						: '../../assets/desktop/search-wise-default.png'
				"
				alt="file"
			/>

			<img
				v-else-if="item?.name === 'Wise' && !file.meta?.image_url"
				class="icon"
				style="width: 48px; height: 48px"
				src="../../assets/desktop/search-wise-default.png"
				alt="file"
			/>

			<img
				v-else
				class="icon"
				:src="
					!file.isDir
						? `./files/file-${file.fileIcon}.svg`
						: `./files/folder-default.svg`
				"
				alt="file"
			/>
		</div>
		<div class="item-content q-mx-md">
			<div
				class="title"
				v-if="file.highlight_field.includes('title')"
				v-html="
					file.highlight
						? file.highlight[file.highlight_field.indexOf('title')]
						: ''
				"
			></div>
			<div class="title" v-else>
				{{ file.title }}
			</div>

			<div class="desc q-my-xs ellipsis" v-if="item?.name === 'Wise'">
				<span>{{ t('file_author') }}: {{ file.author || '-' }}</span>
				<span v-if="file.meta && file.meta.published_at">
					{{ t('file_published') }}:
					{{
						date.formatDate(
							file.meta?.published_at * 1000,
							'MMM Do YYYY, HH:mm:ss'
						)
					}}
				</span>
			</div>

			<div class="desc q-my-xs ellipsis" v-else>
				<span v-if="file.owner_userid"
					>{{ t('file_owner') }}: {{ file.owner_userid }}</span
				>
				<span v-if="file.meta && file.meta.updated">
					{{ t('file_modified') }}:
					{{
						date.formatDate(file.meta.updated * 1000, 'MMM Do YYYY, HH:mm:ss')
					}}
				</span>
				<span>{{ decodeUrl(file.path) }}</span>
			</div>

			<div
				class="context"
				v-if="file.highlight_field.includes('content')"
				v-html="
					file.highlight
						? file.highlight[file.highlight_field.indexOf('content')]
						: ''
				"
			></div>
		</div>
		<div class="item-search">
			<q-icon
				v-if="file?.name === 'Wise'"
				class="icon cursor-pointer"
				name="sym_r_share_windows"
				size="20px"
				color="ink-2"
				@click="open(file)"
			/>
			<q-icon
				v-else
				class="icon cursor-pointer"
				name="sym_r_search"
				size="20px"
				color="ink-2"
				@click="open(file)"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { PropType } from 'vue';
import { TextSearchItem } from '../../utils/interface/search';
import { useI18n } from 'vue-i18n';
import { date } from 'quasar';
import { decodeUrl } from 'src/utils/encode';

defineProps({
	activeItem: {
		type: [String, Number],
		default: '',
		required: false
	},
	fileData: {
		type: Object as PropType<TextSearchItem[]>,
		required: true
	},
	item: {
		type: Object,
		require: false
	}
});

const clickItem = (file: TextSearchItem, index: number) => {
	emits('clickItem', file, index);
};

const open = (file: TextSearchItem) => {
	emits('open', file);
};

const { t } = useI18n();

const emits = defineEmits(['clickItem', 'open']);
</script>

<style scoped lang="scss">
.fileItem {
	width: 100%;
	display: flex;
	border-radius: 8px;

	&.active {
		background-color: rgba(0, 0, 0, 0.1);
	}

	&:hover {
		background: $background-hover;
	}

	.item-icon {
		width: 40px;
		display: flex;
		align-items: flex-start;
		justify-content: center;
		img {
			width: 40px;
		}
	}
	.item-content {
		flex: 1;
		overflow: hidden;
		.title {
			width: 656px;
			color: $ink-1;
			font-size: 14px;
			font-style: normal;
			font-weight: 400;
			line-height: 20px;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
		.desc {
			color: $ink-3;
			font-size: 12px;
			font-style: normal;
			font-weight: 400;
			line-height: 16px;

			> span:not(:last-child) {
				padding-right: 6px;
				margin-right: 6px;
				border-right: 1px solid rgba(0, 0, 0, 0.1);
			}
		}
		.context {
			overflow: hidden;
			color: $ink-2;
			text-overflow: ellipsis;
			font-family: Roboto;
			font-size: 12px;
			font-style: normal;
			font-weight: 400;
			line-height: 16px;
			-webkit-box-orient: vertical;
			-webkit-line-clamp: 2;
			display: -webkit-box;
		}
	}
	.item-search {
		width: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
	}
}
</style>
