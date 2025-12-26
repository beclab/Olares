<template>
	<div
		class="row items-center justify-left q-mx-lg mosaic-header"
		v-if="store.user.viewMode === 'mosaic'"
	>
		<div class="cursor-pointer">
			<q-icon
				:class="showPopMenu ? 'iconAnticlockwise' : 'iconClockwise'"
				name="sym_r_expand_less"
				size="24px"
				color="ink-2"
			></q-icon>
			{{ t(`files.file_${sortedActive}`) }}
			<popup-menu @handleEvent="handleEvent" @popupState="updatePopupState" />
		</div>
		<q-icon
			arrow_upward
			:name="sortAsc ? 'sym_r_arrow_upward' : 'sym_r_arrow_downward_alt'"
			size="24px"
			color="ink-2 cursor-pointer"
			@click="sort(sortedActive)"
		/>
	</div>

	<div
		class="common-div"
		:style="
			origin_id === FilesIdType.PAGEID || origin_id === FilesIdType.SHARE
				? 'padding: 0 20px 0px 20px'
				: ''
		"
	>
		<div
			class="item header"
			:class="
				origin_id === FilesIdType.PAGEID || origin_id === FilesIdType.SHARE
					? ''
					: 'text-body3'
			"
			:style="{
				padding:
					origin_id === FilesIdType.PAGEID || origin_id === FilesIdType.SHARE
						? '0px 20px 3px 0'
						: '0 14px'
			}"
		>
			<div></div>
			<div>
				<p
					class="q-pl-xs"
					:class="{
						active: nameSorted,
						name: true
					}"
					tabindex="0"
				>
					<span
						v-if="isPad"
						class="select-common"
						:class="
							filesStore.currentFileList[origin_id] &&
							filesStore.currentFileList[origin_id]?.items &&
							filesStore.currentFileList![origin_id]!.items.length > 0 &&
							selectFileCount ==
								filesStore.currentFileList[origin_id]?.items.length &&
							!store.showPadPopup
								? 'selected'
								: 'unselect'
						"
						@click.stop="selectAll"
					>
						<q-icon
							class="icon text-ink-on-brand"
							name="check"
							v-if="
								selectFileCount ==
									filesStore.currentFileList[origin_id]?.items.length &&
								!store.showPadPopup
							"
						></q-icon>
					</span>
					<span
						@click="sort('name')"
						:title="$t('files.sortByName')"
						:aria-label="$t('files.sortByName')"
						>{{ $t('files.name') }}</span
					>
					<i
						class="material-icons"
						@click="sort('name')"
						:title="$t('files.sortByName')"
						:aria-label="$t('files.sortByName')"
						>{{ nameIcon }}</i
					>
				</p>

				<p
					v-if="isPad && (selectFileCount == 0 || store.showPadPopup)"
					:class="{ active: modifiedSorted }"
					class="action1"
					role="button"
					tabindex="0"
				>
					<span>{{ 'Action' }}</span>
				</p>

				<p
					:class="{ active: modifiedSorted }"
					class="modified"
					role="button"
					tabindex="0"
					@click="sort('modified')"
					:title="$t('files.sortByLastModified')"
					:aria-label="$t('files.sortByLastModified')"
				>
					<i class="material-icons">{{ modifiedIcon }}</i>
					<span>{{ $t('files.lastModified') }}</span>
				</p>

				<p
					:class="{ active: typeSorted }"
					class="type"
					role="button"
					tabindex="0"
					@click="sort('type')"
					:title="$t('files.typeBySize')"
					:aria-label="$t('files.typeBySize')"
				>
					<span>{{ $t('files.type') }}</span>
					<i class="material-icons">{{ typeIcon }}</i>
				</p>

				<p
					:class="{ active: sizeSorted }"
					class="size"
					role="button"
					tabindex="0"
					@click="sort('size')"
					:title="$t('files.sortBySize')"
					:aria-label="$t('files.sortBySize')"
				>
					<span>{{ $t('files.size') }}</span>
					<i class="material-icons">{{ sizeIcon }}</i>
				</p>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDataStore } from '../../stores/data';
import { getAppPlatform } from '../../application/platform';
import { useFilesStore, FilesIdType } from './../../stores/files';
import { FilesSortType } from './../../utils/contact';
import PopupMenu from './PopupMenu.vue';
const props = defineProps({
	origin_id: {
		type: Number,
		required: true
	}
});
const { t } = useI18n();
const store = useDataStore();
const isPad = ref(getAppPlatform() && getAppPlatform().isPad);
const filesStore = useFilesStore();
const selectFileCount = ref(0);
if (isPad.value) {
	watch(
		() => filesStore.selected[props.origin_id],
		async () => {
			selectFileCount.value = filesStore.selected[props.origin_id].length;
		},
		{
			deep: true
		}
	);
}

const nameSorted = computed(function () {
	return (
		filesStore.activeSort[props.origin_id] &&
		filesStore.activeSort[props.origin_id].by === 1
	);
});
const sizeSorted = computed(function () {
	return (
		filesStore.activeSort[props.origin_id] &&
		filesStore.activeSort[props.origin_id].by === 2
	);
});
const typeSorted = computed(function () {
	return (
		filesStore.activeSort[props.origin_id] &&
		filesStore.activeSort[props.origin_id].by === 3
	);
});
const modifiedSorted = computed(function () {
	return (
		filesStore.activeSort[props.origin_id] &&
		filesStore.activeSort[props.origin_id].by === 4
	);
});
const ascOrdered = computed(function () {
	return (
		filesStore.activeSort[props.origin_id] &&
		filesStore.activeSort[props.origin_id].asc
	);
});
const nameIcon = computed(function () {
	if (nameSorted.value && ascOrdered.value) {
		return 'arrow_upward';
	}
	return 'arrow_downward';
});
const sizeIcon = computed(function () {
	if (sizeSorted.value && ascOrdered.value) {
		return 'arrow_upward';
	}
	return 'arrow_downward';
});
const modifiedIcon = computed(function () {
	if (modifiedSorted.value && ascOrdered.value) {
		return 'arrow_upward';
	}
	return 'arrow_downward';
});
const typeIcon = computed(function () {
	if (typeSorted.value && ascOrdered.value) {
		return 'arrow_upward';
	}
	return 'arrow_downward';
});
const sort = async (by: string) => {
	let asc = true;
	let selfBy = 0;
	if (by === 'name') {
		selfBy = 1;
		if (nameIcon.value === 'arrow_upward') {
			asc = false;
		}
	} else if (by === 'size') {
		selfBy = 2;
		if (sizeIcon.value === 'arrow_upward') {
			asc = false;
		}
	} else if (by === 'type') {
		selfBy = 3;
		if (typeIcon.value === 'arrow_upward') {
			asc = false;
		}
	} else if (by === 'modified') {
		selfBy = 4;
		if (modifiedIcon.value === 'arrow_upward') {
			asc = false;
		}
	}
	filesStore.updateActiveSort(selfBy, asc, props.origin_id);
};
const selectAll = () => {
	if (
		selectFileCount.value !=
		filesStore.currentFileList[props.origin_id]?.items.length
	) {
		const items =
			filesStore.currentFileList[props.origin_id]?.items.map((e) => e.index) ||
			[];
		filesStore.selected[props.origin_id] = items;
	} else {
		filesStore.resetSelected(props.origin_id);
	}
};
const sortAsc = ref(false);
const sortedActive = ref('modified');
const showPopMenu = ref(false);
const updatePopupState = (value: boolean) => {
	showPopMenu.value = value;
};
const handleEvent = (value) => {
	sortedActive.value = value.name;
	sortAsc.value = false;
	switch (value.action) {
		case 'name':
			fileSort(FilesSortType.NAME);
			break;
		case 'type':
			fileSort(FilesSortType.TYPE);
			break;
		case 'modified':
			fileSort(FilesSortType.Modified);
			break;
		case 'size':
			fileSort(FilesSortType.SIZE);
			break;
		default:
			break;
	}
};
const fileSort = (sort: FilesSortType) => {
	if (filesStore.activeSort[props.origin_id].by == sort) {
		filesStore.updateActiveSort(
			sort,
			!filesStore.activeSort[props.origin_id].asc
		);
	} else {
		filesStore.updateActiveSort(sort, true);
	}
};
</script>

<style scoped lang="scss">
.common-div {
	width: 100%;
}
.mosaic-header {
	// width: 100%;
	height: 73px;
	border-bottom: 1px solid $separator;
}
.empty {
	img {
		width: 226px;
		height: 170px;
		margin-bottom: 20px;
	}
}
#listing {
	position: relative;
	overflow-y: scroll;
	&.hovered::-webkit-scrollbar-thumb {
		background-color: rgba(0, 0, 0, 0.2);
	}
}
.files-body {
	width: 100%;
	height: 100%;
}
.iconClockwise {
	animation: rotate 0.2s linear forwards;
}
@keyframes rotate {
	from {
		transform: rotate(0deg);
	}
	to {
		transform: rotate(180deg);
	}
}
.iconAnticlockwise {
	animation: reverse-rotate 0.2s linear forwards;
}
@keyframes reverse-rotate {
	0% {
		transform: rotate(180deg);
	}
	100% {
		transform: rotate(0deg);
	}
}
</style>
