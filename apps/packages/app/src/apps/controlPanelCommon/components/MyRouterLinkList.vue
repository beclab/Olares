<template>
	<MySplitter style="height: calc(100vh)">
		<template v-slot:before="scope">
			<div class="my-menu-before-wrapper">
				<slot></slot>
			</div>
			<bt-scroll-area class="my-menu-before-scroll">
				<q-infinite-scroll @load="loadHandler" :offset="200" v-if="loadMore">
					<div :style="{ width: `${scope.value}px` }">
						<MyRouterLinkListContent
							:defaultActive="defaultActive"
							:data="data"
							v-model="active"
							:defaultOpened="defaultOpened"
						>
						</MyRouterLinkListContent>
					</div>
					<template v-slot:loading>
						<div class="row justify-center q-my-md">
							<q-spinner-dots color="primary" size="40px" />
						</div>
					</template>
				</q-infinite-scroll>

				<div :style="{ width: `${scope.value}px` }" v-else>
					<MyRouterLinkListContent
						:defaultActive="defaultActive"
						:data="data"
						v-model="active"
						:defaultOpened="defaultOpened"
					>
					</MyRouterLinkListContent>
				</div>
			</bt-scroll-area>
			<Empty v-if="isEmpty" size="mini" center></Empty>
			<q-inner-loading :showing="loading"></q-inner-loading>
		</template>
		<template v-slot:after>
			<div class="my-menu-after-wrapper row items-center q-px-md">
				<MyRouterLinkListContentItem
					:data="currentItem"
					:clickable="props.clickable"
					v-ripple="props.clickable"
				>
				</MyRouterLinkListContentItem>
			</div>
			<div style="flex: 1">
				<router-view></router-view>
			</div>
		</template>
	</MySplitter>
</template>

<script setup lang="ts">
import { isArray, isNil } from 'lodash-es';
import { computed, ref, watch } from 'vue';
import MySplitter from './MySplitter.vue';
import MyRouterLinkListContent from './MyRouterLinkListContent.vue';
import MyContentHeader from './MyContentHeader.vue';
import { find, flatMapDeep } from 'lodash';
import MyRouterLinkListContentItem from './MyRouterLinkListContentItem.vue';

interface Props {
	data: any[];
	defaultActive?: string;
	loading?: boolean;
	loadMore?: boolean;
	defaultOpened?: string[];
	clickable?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	data: [] as any,
	loading: false,
	clickable: true
});

const emits = defineEmits(['selected', 'load']);
const active = ref(props.defaultActive);

const isEmpty = computed(() => {
	const empty = isNil(props.data)
		? true
		: isArray(props.data) && props.data.length === 0
		? true
		: false;
	return !props.loading && empty;
});

const currentItem = computed(() => {
	const data = flatMapDeep(props.data, (item) => {
		if (item.children) {
			return [item, ...item.children];
		}
		return item;
	});
	const target = data.find((item) => item.id === active.value);
	return target || { title: '', icon: '' };
});

const clickHandler = (item: any) => {
	active.value = item.id;
	emits('selected', item.id, item);
};

const loadHandler = (index: number, done: (stop?: boolean) => void) => {
	emits('load', index, done);
};

watch(
	() => props.defaultActive,
	(newData) => {
		active.value = newData;
	}
);
</script>

<style lang="scss" scoped>
.avatar-wrapper {
	border: 1px solid rgba(0, 0, 0, 0);
	&.avatar-active {
		border-radius: 50%;
		box-sizing: border-box;
		border: 1px solid $primary;
	}
}
.my-menu-before-wrapper {
	height: $content-header-height;
}
.my-menu-before-scroll {
	height: calc(100vh - #{$content-header-height});
}
.my-menu-after-wrapper {
	position: absolute;
	top: 0;
	left: 0;
	right: 0;
	height: $content-header-height;
}
</style>
