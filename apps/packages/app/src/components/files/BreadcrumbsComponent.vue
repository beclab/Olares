<template>
	<div
		class="breadcrumbs q-ml-xs"
		ref="breadcrumbsRef"
		:style="{
			height: origin_id === FilesIdType.PAGEID ? '42px' : '36px'
		}"
	>
		<span class="breadcrumbs-item" v-for="(link, index) in items" :key="index">
			<q-icon
				v-if="
					index > 0 &&
					(origin_id === FilesIdType.PAGEID || origin_id === FilesIdType.SHARE)
				"
				class="q-mx-sm text-ink-3"
				name="sym_r_keyboard_arrow_right"
				style="font-size: 20px"
			/>
			<span
				v-else-if="index > 0 && origin_id !== FilesIdType.PAGEID"
				class="q-mx-xs text-ink-2"
			>
				/
			</span>
			<span
				class="cursor-pointer link-text"
				:class="
					index === items.length - 1
						? origin_id === FilesIdType.PAGEID
							? 'text-h6 text-ink-1'
							: 'text-body2 text-ink-1'
						: origin_id === FilesIdType.PAGEID
						? 'text-body1 text-ink-2'
						: 'text-body2 text-ink-2'
				"
				:style="{
					maxWidth: breadcrumbsRef ? `${breadcrumbsRef.clientWidth - 120}px` : 0
				}"
				@click="go(link.url, link.query)"
				>{{ link.name }}</span
			>
			<template v-if="link.children">
				<q-menu
					v-model="showFoldingMenu"
					style="
						box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2) !important;
						border-radius: 8px;
					"
				>
					<q-list class="menu-list">
						<q-item
							class="row items-center justify-start menu-item q-px-sm"
							clickable
							dense
							v-close-popup
							v-for="(foldLink, i) in link.children"
							:key="i"
							@click="go(foldLink.url, foldLink.query)"
						>
							<q-img
								class="q-mr-sm"
								src="./../../assets/images/folder.svg"
								style="width: 18px; height: 13px"
							/>
							<div
								class="text-ink-2 foldTxt"
								:class="[
									origin_id === FilesIdType.PAGEID
										? 'text-body-1'
										: 'text-body-2'
								]"
							>
								{{ foldLink.name }}
							</div>
						</q-item>
					</q-list>
				</q-menu>
			</template>
		</span>
	</div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue';
import { useFilesStore, FilesIdType } from '../../stores/files';
import { dataAPIs, common } from './../../api';
import { MenuItem } from './../../utils/contact';
import { useI18n } from 'vue-i18n';
import { DriveType } from '../../utils/interface/files';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	}
});

const { t } = useI18n();
const filesStore = useFilesStore();
const breadcrumbsRef = ref();
const showFoldingMenu = ref(false);
const breadcrumbsWidth = ref(0);

const items = computed(function () {
	const currentPath = filesStore.currentPath[props.origin_id];

	if (!currentPath) return false;

	const driveType = common().formatUrltoDriveType(currentPath.path);

	const dataAPI = dataAPIs(driveType, props.origin_id);

	let parts = currentPath.path.split('/');

	if (parts[0] === '') {
		parts.shift();
	}

	const shiftFirstParts = ['files', 'seahub', 'drive'];

	const shiftDriveParts = ['google', 'dropbox', 'awss3', 'tencent'];

	if (shiftFirstParts.includes(parts[0].toLowerCase())) {
		parts.shift();
	}

	if (shiftDriveParts.includes(parts[0].toLowerCase())) {
		parts.shift();
	}

	if (parts[parts.length - 1] === '') {
		parts.pop();
	}

	let breadcrumbs: any[] = [];

	for (let i = 0; i < parts.length; i++) {
		if (i === 0) {
			breadcrumbs.push({
				name: translateFolderName(decodeURIComponent(parts[i])),
				url: dataAPI.breadcrumbsBase + '/' + parts[i] + '/',
				query: currentPath.param
			});
		} else {
			let name = decodeURIComponent(parts[i]);
			if (currentPath.path.startsWith('/Files/Home')) {
				if (i === 1) {
					name = translateFolderName(name);
				}
			}

			breadcrumbs.push({
				name,
				url: breadcrumbs[i - 1].url + parts[i] + '/',
				query: currentPath.param
			});
		}
	}
	if (process.env.APPLICATION == 'SHARE') {
		breadcrumbs.shift();
		// breadcrumbs.shift();

		if (breadcrumbs.length > 0) {
			breadcrumbs[0].name = 'all';
		}
	}

	if (breadcrumbs.length >= 3) {
		return formatBreadcrumbs(breadcrumbs);
	} else {
		return breadcrumbs;
	}
});

const translateFolderName = (name: string) => {
	const homeValueExists = Object.values(MenuItem).includes(name);
	if (homeValueExists) {
		return t(`files_menu.${name}`);
	}
	return name;
};

const go = async (url: string, query: any) => {
	if (!url) {
		return false;
	}
	const driveType =
		(await common().formatUrltoDriveType(url)) || DriveType.Drive;
	// filesStore.setBrowserUrl(url + query, driveType, props.origin_id);
	filesStore.setFilePath(
		{
			path: url,
			isDir: true,
			driveType,
			param: query
		},
		false,
		true,
		props.origin_id
	);
};

const formatBreadcrumbs = (parts) => {
	let newParts: any[] = [];
	const lastPart = JSON.parse(JSON.stringify(parts))[parts.length - 1];
	const firstPart = JSON.parse(JSON.stringify(parts))[0];

	if (!breadcrumbsRef.value) return false;

	if (breadcrumbsWidth.value < getTextWidth(lastPart.name) + 20) {
		parts.pop();
		newParts = [{ name: '···', url: '', query: '', children: parts }, lastPart];
	} else if (
		breadcrumbsWidth.value <
		getTextWidth(lastPart) + getTextWidth(firstPart.name) + 20 * 2
	) {
		parts.pop();
		parts.shift();
		newParts = [
			firstPart,
			{ name: '···', url: '', query: '', children: parts },
			lastPart
		];
	} else {
		const midParts = parts.slice(1, -1).reverse();
		const midTemParts: any[] = [];

		for (let i = 0; i < midParts.length; i++) {
			const part = midParts[i];
			const midWidth = midTemParts.reduce((accumulator, currentValue: any) => {
				return accumulator + getTextWidth(currentValue.name) + 20;
			}, 0);

			const hasAbbrIndex = midTemParts.findIndex((item) => item.children);
			if (hasAbbrIndex > -1) {
				midTemParts[hasAbbrIndex].children.unshift(part);
			} else {
				// console.log('breadcrumbsWidth', breadcrumbsWidth.value);
				// console.log('part', getTextWidth(part.name));
				// console.log('firstPart', getTextWidth(firstPart.name));
				// console.log('lastPart', getTextWidth(lastPart.name));
				// console.log('midWidth', midWidth);
				if (
					breadcrumbsWidth.value <
					getTextWidth(firstPart.name) +
						getTextWidth(part.name) +
						midWidth +
						getTextWidth(lastPart.name) +
						30 * 4
				) {
					const foldParts = {
						name: '···',
						url: '',
						query: '',
						children: []
					};
					foldParts.children.push(part);
					midTemParts.unshift(foldParts);
				} else {
					midTemParts.unshift(part);
				}
			}
		}

		newParts = [firstPart, ...midTemParts, lastPart];
	}

	return newParts;
};

const getTextWidth = (text) => {
	const canvas = document.createElement('canvas');
	const context: any = canvas.getContext('2d');
	context.font = '16px Roboto';
	return context.measureText(text).width;
};

const updateDivWidth = () => {
	if (breadcrumbsRef.value) {
		breadcrumbsWidth.value = breadcrumbsRef.value.clientWidth;
	}
};

onMounted(() => {
	updateDivWidth();
	window.addEventListener('resize', updateDivWidth);
});

onUnmounted(() => {
	window.removeEventListener('resize', updateDivWidth);
});
</script>

<style lang="scss" scoped>
.breadcrumbs {
	width: calc(100% - 80px);
}

.foldTxt {
	max-width: 200px;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.breadcrumbs {
	display: flex;
	align-items: center;
	justify-content: flex-start;
	color: #6f6f6f;

	.breadcrumbs-item {
		display: inline-block;
		white-space: nowrap;
		display: flex;
		align-items: center;
		justify-content: flex-start;
	}

	.link-text {
		cursor: pointer;
		display: inline-block;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		&:hover {
			color: $info !important;
		}
	}
}

.menu-list {
	min-width: 100px;
	cursor: pointer;
	padding: 8px;
	overflow: hidden;
	.menu-item {
		width: 100%;
		border-radius: 4px;
		min-height: 36px !important;
		.q-focus-helper {
			opacity: 0 !important;
		}
	}
}
</style>
