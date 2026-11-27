<template>
	<div
		class="tab-parent row justify-start items-center"
		ref="tabParentRef"
		:style="{ height: tabHeight }"
	>
		<q-tabs
			:class="[
				isMobile ? 'my-page-tab-mobile' : 'my-page-tabs',
				!isMobile && overflowSource.length === 0 ? 'my-page-tabs--auto' : ''
			]"
			:model-value="modelValue"
			dense
			outside-arrows
			active-color="primary"
			indicator-color="transparent"
			align="left"
			:narrow-indicator="false"
		>
			<q-tab
				v-for="item in visibleSource"
				:key="item.id"
				style="padding: 0"
				:name="item.id"
				@click="onTabClick(item.id)"
			>
				<template v-slot:default>
					<market-tab-item
						:selected="item.id === modelValue"
						:label="item.title ? item.title : item.name"
					/>
				</template>
			</q-tab>
		</q-tabs>
		<q-btn
			v-if="overflowSource.length > 0"
			flat
			no-caps
			dense
			:class="[
				'q-ml-sm more-source-tab',
				isMobile ? 'more-source-tab-mobile' : ''
			]"
			style="padding: 0"
		>
			<market-tab-item :selected="isMoreSelected" :label="moreLabel" />
			<q-menu
				ref="moreMenuRef"
				v-model="moreMenuVisible"
				anchor="bottom right"
				self="top right"
				:offset="isMobile ? [0, 8] : [0, 6]"
			>
				<q-list :style="{ minWidth: isMobile ? '160px' : '180px' }">
					<q-item
						v-for="item in overflowSource"
						:key="item.id"
						clickable
						v-close-popup
						@click="selectOverflowTab(item.id)"
					>
						<q-item-section>{{ item.name }}</q-item-section>
					</q-item>
				</q-list>
			</q-menu>
		</q-btn>
	</div>
	<div class="tab-measure-layer" aria-hidden="true">
		<div
			v-for="item in sources"
			:key="`measure-${item.id}`"
			:ref="(el) => setMeasureItemRef(item.id, el)"
			class="measure-item"
		>
			<market-tab-item :selected="false" :label="item.name" />
		</div>
		<div ref="measureMoreRef" class="measure-item measure-more">
			<market-tab-item :selected="false" :label="moreLabel" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import MarketTabItem from '../../../components/appintro/MarketTabItem.vue';
import {
	computed,
	nextTick,
	onBeforeUnmount,
	onMounted,
	ref,
	watch
} from 'vue';

type SourceItem = {
	id: string;
	name: string;
	title?: string;
};

const props = defineProps<{
	modelValue: string;
	sources: SourceItem[];
	isMobile: boolean;
	moreLabel: string;
	tabHeight?: string;
}>();

const emit = defineEmits<{
	(event: 'update:modelValue', value: string): void;
	(event: 'tab-click', sourceId: string): void;
}>();

const moreMenuRef = ref();
const moreMenuVisible = ref(false);
const tabParentRef = ref<HTMLElement | null>(null);
const tabParentWidth = ref(0);
const measureMoreRef = ref<HTMLElement | null>(null);
const measureItemRefs = ref<Record<string, HTMLElement>>({});
const tabWidthMap = ref<Record<string, number>>({});
const moreWidth = ref(0);
const MORE_MARGIN_LEFT = 8;

const setMeasureItemRef = (id: string, el: any) => {
	if (el && el instanceof HTMLElement) {
		measureItemRefs.value[id] = el;
	} else {
		delete measureItemRefs.value[id];
	}
};

const measureItemWidths = async () => {
	await nextTick();
	const nextWidthMap: Record<string, number> = {};
	props.sources.forEach((item) => {
		const width = measureItemRefs.value[item.id]?.offsetWidth || 0;
		if (width > 0) {
			nextWidthMap[item.id] = width;
		}
	});
	tabWidthMap.value = nextWidthMap;
	const nextMoreWidth = measureMoreRef.value?.offsetWidth || 0;
	if (nextMoreWidth > 0) {
		moreWidth.value = nextMoreWidth;
	}
};

const getMaxVisibleCountByWidth = (
	sources: SourceItem[],
	containerWidth: number
) => {
	if (containerWidth <= 0 || sources.length === 0) {
		return sources.length;
	}

	const fallbackTabWidth =
		Object.values(tabWidthMap.value)[0] || (props.isMobile ? 96 : 100);
	const sourceWidths = sources.map(
		(item) => tabWidthMap.value[item.id] || fallbackTabWidth
	);

	let used = 0;
	let maxWithoutMore = 0;
	for (let i = 0; i < sourceWidths.length; i++) {
		if (used + sourceWidths[i] > containerWidth) {
			break;
		}
		used += sourceWidths[i];
		maxWithoutMore++;
	}
	if (maxWithoutMore >= sources.length) {
		return sources.length;
	}

	const reserveMore =
		(moreWidth.value || (props.isMobile ? 88 : 108)) + MORE_MARGIN_LEFT;
	let visibleWithMore = 0;
	let usedWithMore = 0;
	for (let i = 0; i < sourceWidths.length; i++) {
		if (usedWithMore + sourceWidths[i] + reserveMore > containerWidth) {
			break;
		}
		usedWithMore += sourceWidths[i];
		visibleWithMore++;
	}
	return Math.max(1, visibleWithMore);
};

const visibleSource = computed(() => {
	const sources = props.sources;
	const maxCount = getMaxVisibleCountByWidth(sources, tabParentWidth.value);

	if (sources.length <= maxCount) {
		return sources;
	}

	const selectedId = props.modelValue;
	const selectedIndex = sources.findIndex((item) => item.id === selectedId);
	if (selectedIndex < 0 || selectedIndex < maxCount) {
		return sources.slice(0, maxCount);
	}

	return sources.slice(0, maxCount - 1).concat(sources[selectedIndex]);
});

const overflowSource = computed(() => {
	const visibleIds = new Set(visibleSource.value.map((item) => item.id));
	return props.sources.filter((item) => !visibleIds.has(item.id));
});

const isMoreSelected = computed(() =>
	overflowSource.value.some((item) => item.id === props.modelValue)
);

const updateTabParentWidth = () => {
	tabParentWidth.value = tabParentRef.value?.clientWidth || 0;
};

const onTabClick = (sourceId: string) => {
	emit('update:modelValue', sourceId);
	emit('tab-click', sourceId);
};

const selectOverflowTab = (sourceId: string) => {
	if (!sourceId) {
		return;
	}
	emit('update:modelValue', sourceId);
	emit('tab-click', sourceId);
	moreMenuRef.value?.hide();
};

onMounted(() => {
	updateTabParentWidth();
	void measureItemWidths();
	window.addEventListener('resize', updateTabParentWidth);
	window.addEventListener('resize', measureItemWidths);
});

onBeforeUnmount(() => {
	window.removeEventListener('resize', updateTabParentWidth);
	window.removeEventListener('resize', measureItemWidths);
});

watch(
	() => [props.sources, props.isMobile, props.moreLabel],
	() => {
		void measureItemWidths();
	},
	{ deep: true, immediate: true }
);
</script>

<style scoped lang="scss">
.tab-parent {
	width: 100%;
	max-width: 100%;
	overflow: hidden;
	flex-wrap: nowrap;
}

.my-page-tabs {
	flex: 1;
	min-width: 0;
	border-radius: 8px;
	border: 1px solid $separator;
}

.my-page-tabs--auto {
	flex: 0 0 auto;
	width: auto;
	max-width: 100%;
}

.my-page-tab-mobile {
	flex: 0 0 auto;
	width: auto;
	max-width: 100%;
	padding: 0 8px;
	box-shadow: 0 2px 8px -2px rgba(0, 0, 0, 0.1);
}

.more-source-tab {
	flex-shrink: 0;
	min-width: fit-content;
}

.more-source-tab-mobile {
	box-shadow: 0 2px 8px -2px rgba(0, 0, 0, 0.1);
}

.more-source-tab :deep(.q-btn__content) {
	padding: 0;
}

.q-tabs--dense .q-tab {
	min-height: 32px;
}

.tab-measure-layer {
	position: absolute;
	left: -9999px;
	top: -9999px;
	visibility: hidden;
	pointer-events: none;
	display: flex;
}

.measure-item {
	display: inline-flex;
}

.measure-more {
	margin-left: 8px;
}
</style>
