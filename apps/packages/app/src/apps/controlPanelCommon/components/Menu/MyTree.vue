<template>
	<MySplitter>
		<template v-slot:before="scope">
			<div class="my-menu-before-wrapper" v-if="!headerBeforeHide">
				<slot v-bind:value="scope.value"></slot>
				<q-resize-observer @resize="onResize" />
			</div>
			<div v-else class="my-menu-before-wrapper-no-header"></div>
			<bt-scroll-area class="my-menu-before-scroll">
				<q-tree
					ref="treeRef"
					dense
					:nodes="data"
					node-key="id"
					label-key="title"
					:selected-color="selectedColor"
					:accordion="true"
					v-model:selected="selected"
					v-model:expanded="expanded"
					@update:selected="onSelected"
					@update:expanded="onExpanded"
					no-selection-unset
					v-bind="$attrs"
					no-results-label=" "
					no-nodes-label=" "
					text-color="ink-2"
					class="my-tree-wrapper"
				>
					<template v-slot:default-body="prop">
						<div v-if="prop.node.story" class="q-ml-sm">
							{{ prop.node.story }}
						</div>
					</template>
					<template v-slot:default-header="prop">
						<div
							:id="prop.key"
							:class="{
								'tree-img-wrapper': !prop.node.children
							}"
							class="my-tree-header-container"
						>
							<div
								:class="{
									'my-tree-header-wrapper': !prop.node.children,
									'selected-node': isSelected(prop.key, prop.node.children)
								}"
								class="row no-wrap items-center my-tree-header-wrapper"
							>
								<q-inner-loading :showing="prop.node.loading">
									<q-spinner-dots size="sm" color="primary" />
								</q-inner-loading>
								<div class="header-top relative-position">
									<q-img
										class="my-tree-avatar"
										v-if="prop.node.img"
										:src="imgFilter(prop)"
										:placeholder-src="defaultIcon"
										:class="{
											'expanded-node-img': isImgActive(prop)
										}"
										:ratio="1"
										width="20px"
										height="20px"
										fit="scale-down"
										no-spinner
										loading="lazy"
										style="border-radius: 50%"
									>
										<template v-slot:error>
											<div style="padding: 0px" class="absolute-full">
												<img :src="defaultIcon" alt="" />
											</div>
										</template>
									</q-img>
									<MyBadge
										class="my-tree-badge-wrapper"
										v-if="prop.node.status"
										:type="prop.node.status"
										outlined
										size="sm"
									></MyBadge>
								</div>

								<span
									:class="{
										'expanded-node-text': isExpanded(prop),
										'q-ml-xs': !prop.node.img
									}"
									style="
										flex: 1;
										white-space: nowrap;
										overflow: hidden;
										text-overflow: ellipsis;
									"
									>{{ prop.node.title }}</span
								>
							</div>
						</div>
					</template>
				</q-tree>
			</bt-scroll-area>
			<Empty v-if="isEmpty && !hiddenEmpty" size="mini" center></Empty>
			<q-inner-loading :showing="loading"></q-inner-loading>
		</template>
		<template v-slot:after>
			<div
				class="my-menu-after-wrapper"
				v-if="!headerAfterHide"
				:style="{
					zIndex: zIndex,
					width: size.widths[index + 1] ? `${size.widths[index + 1]}px` : 'auto'
				}"
			>
				<MyContentHeader
					:title="selectedNodes.title"
					:img="selectedNodes.img"
				></MyContentHeader>
			</div>
			<div style="flex: 1" class="my-menu-after-content">
				<router-view v-slot="{ Component }">
					<component :is="Component" />
					<template v-if="!Component">
						<slot name="after-default"></slot>
					</template>
				</router-view>
			</div>
		</template>
	</MySplitter>
</template>

<script lang="ts" setup>
import {
	computed,
	defineProps,
	nextTick,
	onMounted,
	onUnmounted,
	PropType,
	ref,
	watch,
	onUpdated
} from 'vue';
import { defaultMenuOptions, MenuItem, MenuOptions } from '../../types/menu';
import { useRoute, useRouter } from 'vue-router';
import _, { isArray, isNil, isFunction } from 'lodash';
import Empty from '../Empty.vue';
import MyBadge from '../MyBadge.vue';
import MySplitter from '../MySplitter.vue';
import MyContentHeader from '../MyContentHeader.vue';
import { updateWidths, size, initSize } from './config';
import defaultIcon from '../../assets/default.svg';

const router = useRouter();

interface Props {
	data: any;
	menuOptions?: MenuOptions;
	autoView?: boolean;
	defaultOpeneds?: string | string[];
	defaultActive?: any;
	loading?: boolean;
	hiddenEmpty?: boolean;
	/**
	 * tree header z-index
	 */
	zIndex?: 1;
	/**
	 * tree header width
	 */
	index: 0;
	selectedColor: string;
	selectedBackgroundColor?: string;
	headerAfterHide?: boolean;
	headerBeforeHide?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	data: [],
	defaultActive: [],
	hiddenEmpty: false,
	index: 0,
	selectedColor: 'blue-default',
	selectedBackgroundColor: 'rgba(34, 111, 255, 0.12)'
});

const options = ref({});
const selected = ref(props.defaultActive);
const expanded = ref(props.defaultOpeneds);
const extraWidth = ref(32);

const extraWidthPx = computed(() => extraWidth.value + 'px');

const route = useRoute();

watch(
	() => props.defaultActive,
	(newData, oldData) => {
		selected.value = newData;
	}
);

watch(
	() => props.defaultOpeneds,
	(newData, oldData) => {
		expanded.value = newData;
	}
);

onMounted(() => {
	options.value = {
		...defaultMenuOptions,
		...props.menuOptions
	};
});

onUpdated(() => {
	extraWidth.value =
		document.querySelector('.my-content-page-header-extra')?.clientWidth ?? 0;
});

const resetSelected = () => {
	selected.value = [];
};

const isEmpty = computed(() => {
	const empty = isNil(props.data)
		? true
		: isArray(props.data) && props.data.length === 0
		? true
		: false;
	return !props.loading && empty;
});

const emits = defineEmits(['open', 'selected']);

const open = (id: string) => {
	const target = props.data.find((item: any) => item.id === id);
	emits('open', id, target);
};

const autoScroll = () => {
	if (props.autoView) {
		router.afterEach(() => {
			const element = document.querySelector(
				'#my-scroll-wrapper > .q-scrollarea__container '
			);
			nextTick(() => {
				element &&
					element.scrollTo({
						left: 5000,
						top: 0,
						behavior: 'smooth'
					});
			});
		});
	}
};

const changeHander = (node: any) => {
	router.push(node.route);
};

const selectedHandler = (target: any) => {
	console.log('selectedHandler', target);
};
const treeRef = ref();
const selectedNodes = computed(
	() =>
		(treeRef.value && treeRef.value.getNodeByKey(selected.value)) || {
			title: '',
			icon: ''
		}
);
const onSelected = (id: string) => {
	const selectedNodes = treeRef.value.getNodeByKey(id);
	console.log('onSelected', selectedNodes);
	if (selectedNodes) {
		router.push(selectedNodes.route);
		emits('selected', id, selectedNodes);
	}
};

const setExpanded = (id: string, state: boolean) => {
	treeRef.value.setExpanded(id, state);
};

const getExpandedNodes = () => treeRef.value.getExpandedNodes();

const onExpanded = (values: readonly string[]) => {
	const id = _.last(values);
	const selectedNodes = treeRef.value.getNodeByKey(id);
	if (selectedNodes) {
		emits('open', id, selectedNodes);
	}
};

const isSelected = (key: string, hasChildren: boolean) => {
	return selected.value === key;
};

const isExpanded = (prop: Record<string, any>) => {
	const key = prop.key;
	const hasIcon = prop.node.img;
	const hasStatus = prop.node.status;
	return hasIcon || hasStatus ? expanded.value?.includes(key) : false;
};

const isImgActive = (prop: Record<string, any>) => {
	return !prop.node.imgActiveDisabled && isExpanded(prop);
};

const imgFilter = (prop: Record<string, any>) => {
	const node = prop.node;
	return isSelected(prop.key, node.children) && node.activeImg
		? node.activeImg
		: node.img;
};

const onResize = (size: { height: number; width: number }) => {
	updateWidths(props.index, size.width);
};
const selectedBG = computed(() => props.selectedBackgroundColor);

onUnmounted(() => {
	initSize();
});

autoScroll();

defineExpose({ resetSelected, setExpanded, getExpandedNodes, selectedNodes });
</script>
<style lang="scss" scoped>
.my-menu-before-wrapper {
	height: $content-header-height;
}
.my-menu-before-wrapper-no-header {
	height: 20px;
}
.my-menu-before-scroll {
	height: calc(100vh - #{$content-header-height});
}
.my-menu-after-wrapper {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	padding-right: v-bind(extraWidthPx);
	height: $content-header-height;
}
</style>
<style lang="scss" scoped>
.my-tree-wrapper {
	::v-deep(.q-tree__arrow) {
		font-size: 14px;
		border: 1px solid rbga(255, 255, 255, 0);
		color: $ink-1;
	}
	::v-deep(.q-tree__spinner) {
		font-size: 14px;
		border: 1px solid rbga(255, 255, 255, 0);
		color: $ink-1;
	}
}
::v-deep(.selected-node) {
	position: relative;
	&::after {
		content: '';
		position: absolute;
		left: -4px;
		right: 4px;
		top: 0;
		bottom: 0;
		border-radius: 4px;
		z-index: -1;
		background-color: v-bind(selectedBG);
	}
}

.my-tree-header-container {
	width: 100%;
	position: relative;
	border-radius: 4px;
	.my-tree-badge-wrapper {
		margin-left: 4px;
	}
	.my-tree-avatar {
		position: relative;
		& ~ .my-tree-badge-wrapper {
			position: absolute;
			bottom: 0;
			right: 0;
		}
	}
}
::v-deep(.q-tree .q-tree__node-header) {
	border-radius: 4px;
}
::v-deep(.q-tree .q-tree__node-header .q-focus-helper) {
	left: -6px;
	right: 0;
	width: auto;
	border-radius: 4px;
}

::v-deep(.q-focus-helper:has(+ .q-icon.q-tree__arrow)) {
	left: 0;
	right: 0;
	width: auto;
}

.my-tree-header-wrapper {
	padding: 4px 0;
	border-radius: 4px;
	padding-right: 8px;
	width: 100%;
}

::v-deep(.q-tree__node-header:before) {
	border-bottom-left-radius: 7px;
	border-color: $separator-2;
}
::v-deep(
		.q-tree__node--parent > .q-tree__node-collapsible > .q-tree__node-body:after
	) {
	border-color: #e0e0e0;
}
::v-deep(.q-tree__node:after) {
	border-color: $separator-2;
}
::v-deep(.my-menu-before-scroll .q-scrollarea__container) {
	overflow-x: hidden;
}
::v-deep(.my-menu-before-scroll .q-scrollarea__content) {
	min-width: initial;
	width: 100%;
}
::v-deep(.my-menu-before-scroll .q-tree__node-collapsible) {
	overflow-x: hidden;
}
::v-deep(
		.q-tree
			> .q-tree__node--parent
			> .q-tree__node-header
			> .q-tree__node-header-content
	) {
	color: $ink-1 !important;
}
::v-deep(
		.q-tree--dense
			.q-tree__node--parent
			> .q-tree__node-collapsible
			> .q-tree__node-body
	) {
	padding-bottom: 0px;
}
</style>
<style></style>
