<template>
	<div class="searchCard">
		<img class="icon" src="../../assets/desktop/search-icon.svg" alt="search" />
		<input
			class="input"
			type="text"
			dense
			ref="searchRef"
			v-model.trim="searchTxt"
			:placeholder="t('search_placeholder')"
		/>
		<q-icon
			v-if="searchTxt"
			class="cursor-pointer"
			name="sym_r_close"
			size="24px"
			style="color: #adadad"
			@click="handleClear"
		/>
		<!-- <span class="btn" @click="openCommand('')">Open Command</span> -->
	</div>
	<div class="list q-pt-sm" ref="listRef">
		<template
			v-for="(items, index) in appData"
			:key="`${items.title}_${index}`"
		>
			<div class="category" v-if="items.category === 'Use'">
				{{ t('use_search', { content: searchTxt }) }}
			</div>
			<div class="category" v-if="items.category === 'Result'">
				{{ t('result') }}
			</div>
			<template v-if="items.children">
				<div
					:class="[
						activeItem === `${item.title}_${item.type}` ? 'isActive' : '',
						'item'
					]"
					ref="isActiveRef"
					v-for="item in items.children"
					:key="item.title"
					@click="openCommand(item)"
				>
					<div class="txt">
						<img :src="item.icon" />
						<span class="name">{{ item.title }}</span>
						<span
							v-if="item.type === SearchCategory.Command"
							class="name q-ml-sm text-ink-2"
							>{{ item.name }}</span
						>
						<!-- <span class="desc">{{ item.type }}</span> -->
					</div>
					<div class="desc q-mr-sm text-ink-2">
						{{
							item.type === SearchCategory.Application ? t('application') : ''
						}}
						{{ item.type === SearchCategory.Command ? t('command') : '' }}
						{{ item.type === SearchCategory.Result ? t('command') : '' }}
					</div>
				</div>
			</template>
		</template>
	</div>
</template>

<script lang="ts" setup>
import { SearchCategory } from 'src/utils/interface/search';
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	showSearchDialog: {
		type: Boolean,
		require: false,
		default: false
	},
	commandList: {
		type: Array,
		require: true
	}
});

const emits = defineEmits(['openCommand']);
const { t } = useI18n();
const searchTxt = ref('');
const searchRef = ref();

const activeItem = ref();
const isActiveRef = ref();
const listRef = ref<null | HTMLElement>(null);
const appData = ref();
const timer = ref();

const classifyData = (data: any, searchTxt?: string) => {
	const newArr: any[] = [];
	for (let i = 0; i < data.length; i++) {
		const el = data[i];

		if (searchTxt) {
			if (el.type != SearchCategory.Command) {
				if (el.title.toLowerCase().indexOf(searchTxt.toLowerCase()) === -1)
					continue;
				const resultIndex = newArr.findIndex(
					(item) => item.category === SearchCategory.Result
				);
				if (resultIndex <= -1) {
					let obj = {
						category: SearchCategory.Result,
						children: [el]
					};
					newArr.unshift(obj);
				} else {
					newArr[resultIndex].children.unshift(el);
				}
			} else {
				const commandIndex = newArr.findIndex(
					(item) => item.category === SearchCategory.Use
				);
				if (commandIndex <= -1) {
					let obj = {
						category: SearchCategory.Use,
						children: [el]
					};
					newArr.push(obj);
				} else {
					newArr[commandIndex].children.push(el);
				}
			}
		} else {
			const filterTypeIndex = newArr.findIndex(
				(item) => item.category === el.type
			);
			if (filterTypeIndex <= -1) {
				let obj = {
					category: el.type,
					children: [el]
				};
				newArr.push(obj);
			} else {
				newArr[filterTypeIndex].children.push(el);
			}
		}
	}

	activeItem.value =
		newArr[0]?.children[0]?.title + '_' + newArr[0]?.children[0]?.type;

	console.log('newArr', newArr);
	return newArr;
};

const openCommand = (item: any) => {
	if (timer.value) {
		clearTimeout(timer.value);
	}
	const searchFiles =
		searchTxt.value.toLowerCase() !== 'drive' ? searchTxt.value : '';
	item.searchFiles = searchFiles;
	emits('openCommand', item);
};

// const handleActive = (title: string, type: string) => {
// 	if (timer.value) {
// 		clearTimeout(timer.value);
// 	}
// 	timer.value = setTimeout(() => {
// 		activeItem.value = `${title}_${type}`;
// 	}, 300);
// };

const keydownEnter = (event: any) => {
	if (!appData.value) {
		return false;
	}
	const keydownData = concatData(appData.value);

	const index = keydownData.findIndex(
		(item) => item.title + '_' + item.type === activeItem.value
	);
	if (event.keyCode === 38) {
		const upIndex = index - 1;
		if (upIndex >= 0) {
			activeItem.value =
				keydownData[upIndex].title + '_' + keydownData[upIndex].type;
			refreshScroll(upIndex);
		}
	}

	if (event.keyCode === 40) {
		const downIndex = index + 1;
		if (downIndex <= keydownData.length - 1) {
			activeItem.value =
				keydownData[downIndex].title + '_' + keydownData[downIndex].type;
			refreshScroll(downIndex);
		}
	}

	if (event.keyCode === 13) {
		const aItem = keydownData.find(
			(item) => item.title + '_' + item.type === activeItem.value
		);
		openCommand(aItem);
	}
};

const concatData = (data: any) => {
	let newArr: any[] = [];
	for (let i = 0; i < data.length; i++) {
		const element = data[i];
		if (!element.children) continue;
		newArr = [...newArr, ...element.children];
	}
	return newArr;
};

const refreshScroll = (index: number) => {
	isActiveRef.value[index].scrollIntoView({
		behavior: 'auto',
		block: 'nearest'
	});
};

const handleClear = () => {
	searchTxt.value = '';
};

watch(
	() => searchTxt.value,
	(newVal) => {
		appData.value = classifyData(props.commandList, newVal);
	}
);

onMounted(() => {
	nextTick(() => {
		appData.value = props.commandList && classifyData(props.commandList);
		setTimeout(() => {
			searchRef.value && searchRef.value.focus();
		}, 0);
	});

	document.addEventListener('keydown', keydownEnter);
});

onUnmounted(() => {
	document.removeEventListener('keydown', keydownEnter);
});
</script>

<style lang="scss" scoped>
.list {
	height: calc(100% - 48px);
	border-top: 1px solid $separator-color;
	overflow: scroll;
	scrollbar-width: none;

	.category {
		width: calc(100% - 32px);
		margin: 0 16px;
		height: 42px;
		line-height: 42px;
		padding-left: 16px;
		font-weight: 500;
		font-size: 14px;
		color: #857c77;
	}
	.item {
		width: calc(100% - 32px);
		height: 40px;
		line-height: 40px;
		border-radius: 12px;
		margin: 0 16px 4px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		cursor: pointer;
		position: relative;
		.txt {
			display: flex;
			align-items: center;
			img {
				width: 24px;
				height: 24px;
				border-radius: 4px;
				margin: 0 8px;
			}
		}
		.icon {
			margin-right: 16px;
			width: 24px;
			height: 24px;
			border-radius: 4px;
			display: flex;
			align-items: center;
			justify-content: center;
			&:hover {
				background: rgba(133, 124, 119, 0.2);
			}
		}

		&:hover {
			background: $background-hover;
		}
		&.isActive {
			background: $background-selected;
		}
	}
}
</style>
