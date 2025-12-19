<template>
	<ul style="height: 100%" v-if="items && items.length > 0">
		<q-virtual-scroll
			style="height: 100%"
			:items="items"
			separator
			:virtual-scroll-item-size="itemSize"
			v-slot="{ item, index }"
		>
			<li
				class="item-li"
				:key="item.date ? item.date + '_' + index : item"
				@touchstart="startPress"
				@touchend="endPress"
				@touchmove="onMove"
			>
				<div
					class="check-box row items-center justify-center"
					:class="{ isSelect: deviceStore.isInEditor }"
				>
					<img
						v-if="isSelected(item)"
						:src="activeImage"
						@click="toggleSelect(item)"
					/>
					<img
						v-else-if="!$q.dark.isActive && !isSelected(item)"
						:src="normalImage"
						@click="toggleSelect(item)"
					/>
					<img v-else :src="normalDarkImage" @click="toggleSelect(item)" />
				</div>

				<div
					class="check-content"
					:class="{ isSelect: deviceStore.isInEditor }"
				>
					<slot :file="item" />
				</div>
			</li>
		</q-virtual-scroll>
	</ul>

	<file-transfer-no-data v-else />
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { useQuasar } from 'quasar';
import { isValidDate } from './../../utils/utils';
import FileTransferNoData from './FileTransferNoData.vue';
import { useDeviceStore } from '../../stores/device';
import { busOff, busOn } from '../../utils/bus';

type IdsType = number | string;

interface DateMarker {
	date: string;
	ids: IdsType[];
}

interface EnableSelect {
	id: IdsType;
	selectedEnable: (id: IdsType) => boolean;
}

const props = defineProps({
	items: {
		type: Array,
		required: true
	},
	lockEvent: {
		type: Boolean,
		required: true
	},
	lockTouch: {
		type: Boolean,
		required: false,
		default: false
	},
	hasDate: {
		type: Boolean,
		required: false,
		default: false
	},
	itemSize: {
		type: Number,
		required: false,
		default: 72
	}
});

const emits = defineEmits(['showSelectMode', 'itemOnUnableSelect']);

const $q = useQuasar();

const activeImage = `${
	$q.platform.is.electron ? '.' : ''
}/img/checkbox/check_box.svg`;
const normalImage = `${
	$q.platform.is.electron ? '.' : ''
}/img/checkbox/uncheck_box_light.svg`;
const normalDarkImage = `${
	$q.platform.is.electron ? '.' : ''
}/img/checkbox/uncheck_box_dark.svg`;

const selected = ref<Set<IdsType>>(new Set<IdsType>());
const isAllSelected = ref(false);
// const isSelectMode = ref(false);
const timeout = ref();
const longPressDuration = ref(500);
const deviceStore = useDeviceStore();

const arraysEqual = (arr1, arr2): boolean => {
	if (arr1.length !== arr2.length) return false;

	const cur_arr1: any[] = [];
	for (let i = 0; i < arr1.length; i++) {
		let arrItem = arr1[i];
		if (arrItem.date) {
			cur_arr1.push(arrItem.date);
		} else {
			cur_arr1.push(arrItem);
		}
	}

	const set1 = new Set(cur_arr1);
	const set2 = new Set(arr2);

	if ([...set2].every((item) => set1.has(item))) {
		return true;
	} else {
		return false;
	}
};

const isSelected = (item: IdsType | DateMarker | EnableSelect) => {
	let cur_item: IdsType;
	if (item && item.date) {
		cur_item = item.date;
	} else if (item && item.id) {
		cur_item = item.id;
	} else {
		cur_item = item;
	}
	if (selected.value.has(cur_item)) {
		return true;
	} else {
		return false;
	}
};

const updateSelectMode = () => {
	const selectedIds = [...selected.value].filter((item) => !isValidDate(item));

	emits('showSelectMode', selectedIds);
};

const toggleSelectAll = () => {
	// isSelectMode.value = false;
	if (isAllSelected.value) {
		selected.value.clear();
		isAllSelected.value = false;
		emits('showSelectMode', []);
	} else {
		selected.value.clear();
		props.items.forEach((item) => {
			if (item && item.date) {
				selected.value.add(item.date);
			} else if (item && item.id) {
				if (item.selectedEnable(item.id)) {
					selected.value.add(item.id);
				}
				return;
			} else {
				selected.value.add(item);
			}
		});
		isAllSelected.value = true;
		updateSelectMode();
	}
};

const toggleSelect = (item) => {
	if (item.date) {
		if (item.ids.every((id) => selected.value.has(id))) {
			selected.value.delete(item.date);
			for (const id of item.ids) {
				selected.value.delete(id);
			}
		} else {
			selected.value.add(item.date);
			for (const id of item.ids) {
				selected.value.add(id);
			}
		}
	} else {
		if (item.id) {
			if (!item.selectedEnable(item.id)) {
				emits('itemOnUnableSelect', item.id);
				return;
			}
			if (selected.value.has(item.id)) {
				selected.value.delete(item.id);
			} else {
				selected.value.add(item.id);
			}
		} else {
			if (selected.value.has(item)) {
				selected.value.delete(item);
			} else {
				selected.value.add(item);
			}
		}
	}

	if (props.hasDate) {
		checkSelectDate(item);
	}

	isAllSelected.value = arraysEqual(props.items, [...selected.value]);
	updateSelectMode();
};

const checkSelectDate = (item: any) => {
	if (item.date) {
		return false;
	}
	let cur_date;
	for (let i = 0; i < props.items.length; i++) {
		const cell: any = props.items[i];
		if (cell && cell.date && cell.ids.find((id: number) => id === item)) {
			cur_date = cell;
		}
	}

	if (cur_date.ids.every((item: any) => selected.value.has(item))) {
		selected.value.add(cur_date.date);
	} else {
		selected.value.delete(cur_date.date);
	}
};

const handleClose = () => {
	emits('showSelectMode', null);
	deviceStore.isInEditor = false;
	isAllSelected.value = false;
	selected.value.clear();
};

const intoCheckedMode = () => {
	deviceStore.isInEditor = true;
	emits('showSelectMode', []);
};

const handleRemove = () => {
	for (const id of [...selected.value]) {
		selected.value.delete(id);
	}
	emits('showSelectMode', []);

	if (isAllSelected.value) {
		handleClose();
	}
};

defineExpose({
	handleClose,
	toggleSelectAll,
	intoCheckedMode,
	handleRemove
});

const startPress = () => {
	if (props.lockTouch) {
		return false;
	}
	if (deviceStore.isInEditor) {
		return false;
	}

	timeout.value = setTimeout(() => {
		if (props.lockEvent) {
			return false;
		}
		deviceStore.isInEditor = true;
		emits('showSelectMode', []);
	}, longPressDuration.value);
};

const endPress = () => {
	clearTimeout(timeout.value);
};

const onMove = () => {
	clearTimeout(timeout.value);
};

onUnmounted(() => {
	deviceStore.isInEditor = false;
	busOff('exitEditMode', handleClose);
});

onMounted(() => {
	busOn('exitEditMode', handleClose);
});
</script>

<style lang="scss" scoped>
ul,
li {
	width: 100%;
	list-style-type: none;
	margin: 0;
	padding: 0;
}

.item-li {
	width: 100%;
	display: flex;
	align-items: center;
	justify-content: space-between;
}
.check-box {
	width: 0px;
	height: 32px;
	transition: width 0.5s ease;
	// margin-right: 12px;
	img {
		width: 16px;
		height: 16px;
	}
	&.isSelect {
		width: 32px;
	}
}

.check-content {
	width: calc(100%);
	transition: width 0.5s ease;
	&.isSelect {
		width: calc(100% - 44px);
	}
}
</style>
