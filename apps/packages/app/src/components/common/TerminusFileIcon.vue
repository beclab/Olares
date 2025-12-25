<template>
	<div
		class="row items-center justify-center file-icon"
		:style="`height: ${iconSize}px`"
	>
		<img v-if="isDir || type === 'Folder'" :src="folderIcon(name)" />
		<template v-else-if="getFileIcon(name) === 'image' && isThumbsEnabled">
			<img
				style="border-radius: 4px"
				:src="thumbnailUrl"
				@error.once="
					(e) => {
						e.target.src = fileIcon(name);
					}
				"
			/>
		</template>
		<img
			v-else
			:src="fileIcon(name)"
			@error.once="
				(e) => {
					e.target.src = './img/file-blob.svg';
				}
			"
		/>
	</div>
</template>

<script lang="ts" setup>
import { getFileIcon } from '@bytetrade/core';
import { computed, PropType } from 'vue';
import { useRoute } from 'vue-router';
import { enableThumbs } from '../../utils/constants';
import { FileItem, useFilesStore } from '../../stores/files';
import { useOperateinStore } from './../../stores/operation';
import { DriveType } from '../../utils/interface/files';

const props = defineProps({
	name: {
		type: String,
		default: '',
		required: true
	},
	type: {
		type: String,
		default: '',
		required: true
	},
	modified: {
		type: Number,
		default: 0,
		required: false
	},
	path: {
		type: String,
		default: '',
		required: false
	},
	thumbnailLink: {
		type: String,
		default: '',
		required: false
	},
	isDir: {
		type: Boolean,
		default: false,
		required: false
	},
	iconSize: {
		type: Number,
		default: 32
	},
	driveType: {
		type: String as unknown as PropType<DriveType>,
		default: DriveType.Drive,
		required: false
	}
});

const filesStore = useFilesStore();
const route = useRoute();
const operateinStore = useOperateinStore();

const folderIcon = (name: any) => {
	let src = '/img/folder-';

	if (process.env.PLATFORM == 'DESKTOP') {
		src = './img/folder-';
	}

	let arr = ['Documents', 'Pictures', 'Movies', 'Downloads', 'Music'];
	if (arr.includes(name) && route.path === operateinStore.defaultPath) {
		src = src + name + '.svg';
	} else {
		src = src + 'default.svg';
	}
	return src;
};

const fileIcon = (name: any) => {
	let src = '/img/file-';
	let folderSrc = '/img/file-blob.svg';

	if (process.env.PLATFORM == 'DESKTOP') {
		src = './img/file-';
		folderSrc = './img/file-blob.svg';
	}

	if (name.split('.').length > 1) {
		src = src + getFileIcon(name) + '.svg';
	} else {
		src = folderSrc;
	}

	return src;
};

const isThumbsEnabled = computed(function () {
	return enableThumbs;
});

const thumbnailUrl = computed(function () {
	if (props.thumbnailLink) {
		return props.thumbnailLink;
	}
	const path = props.path.startsWith('/Files')
		? props.path.slice(6)
		: props.path;
	const file: FileItem = {
		extension: '',
		isDir: false,
		isSymlink: false,
		mode: 0,
		modified: props.modified,
		name: props.name,
		path: path,
		size: 0,
		type: 'image',
		driveType: props.driveType,
		param: '',
		index: 0,
		url: '',
		fileExtend: ''
	};

	return filesStore.getPreviewURL(file, 'thumb');
});
</script>

<style lang="scss" scoped>
.file-icon {
	overflow: hidden;
	img {
		height: 100%;
		object-fit: cover;
	}
}
</style>
