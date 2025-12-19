<template>
	<div class="title-contianer row no-wrap items-center">
		<template v-if="$slots.avatar">
			<div
				class="title-avatra"
				:class="[listItemTitleSize, titleClass]"
				:style="titleStyle"
			>
				<slot name="avatar"></slot>
			</div>
		</template>
		<div
			class="column justify-between full-width"
			:class="[ellipsis ? 'ellipsis' : '']"
		>
			<MyTitle v-bind="props" :ellipsis="ellipsis" v-if="$slots.title">
				<slot name="title"></slot>
			</MyTitle>
			<MyTitle v-bind="props" v-else :ellipsis="ellipsis" :title="title">
			</MyTitle>
			<MySubTitle class="q-mt-xs" v-bind="props" v-if="$slots.subTitle">
				<slot name="subTitle"></slot>
			</MySubTitle>
			<MySubTitle class="q-mt-xs" v-else-if="subTitle" :title="subTitle">
			</MySubTitle>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, CSSProperties } from 'vue';
import MyTitle from './MyTitle.vue';
import MySubTitle from './MySubTitle.vue';

interface Props {
	title?: string;
	subTitle?: string;
	bolder?: boolean;
	size: 'sm' | 'md';
	iconSize: 'sm' | 'md' | 'lg';
	titleClass?: string;
	titleStyle?: CSSProperties | undefined;
	ellipsis?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	bolder: true,
	size: 'md',
	iconSize: 'md',
	ellipsis: true
});

const listItemTitleSize = computed(() => `title-avatra-${props.iconSize}`);
</script>
<style lang="scss" scoped>
.title-contianer {
	display: flex;
	font-size: 14px;
	.title-avatra {
		position: relative;
		display: flex;
	}
	.title-avatra-sm {
		flex: 0 0 16px;
		width: 16px;
		line-height: 16px;
		margin-right: 4px;
	}
	.title-avatra-md {
		flex: 0 0 20px;
		width: 20px;
		line-height: 20px;
		margin-right: 8px;
	}
	.title-avatra-lg {
		flex: 0 0 32px;
		width: 32px;
		line-height: 32px;
		margin-right: 12px;
	}
}
</style>
