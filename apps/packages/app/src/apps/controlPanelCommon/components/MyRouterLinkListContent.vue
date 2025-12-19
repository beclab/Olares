<template>
	<q-list class="q-px-md full-height" dense>
		<div v-for="(child, index) in data" :key="index">
			<QExpansionItemStyle>
				<q-expansion-item
					dense-toggle
					switch-toggle-side
					group="expand"
					expand-icon-class="text-ink-1"
					:default-opened="defaultOpened?.includes(child.id)"
					v-if="child.children"
				>
					<template #header>
						<div class="ellipsis text-ink-1" style="align-self: center">
							{{ child.title }}
						</div>
					</template>
					<div class="q-py-xs" v-for="item in child.children" :key="item.uid">
						<div @click="clickHandler(item)">
							<MyRouterLinkListContentItem
								:active="item.id === active"
								:data="item"
								class="my-menu-link q-py-none q-px-sm rounded-borders"
								clickable
								v-ripple
							>
							</MyRouterLinkListContentItem>
						</div>
					</div>
				</q-expansion-item>
			</QExpansionItemStyle>
		</div>
	</q-list>
</template>

<script setup lang="ts">
import { isArray, isNil } from 'lodash-es';
import { computed, ref, watch } from 'vue';
import { RouteLocationPathRaw, RouteLocationRaw, useRouter } from 'vue-router';
import MyBadge from './MyBadge.vue';
import MyTitle from './MyListItem/MyTitle.vue';
import MySubTitle from './MyListItem/MySubTitle.vue';
import QExpansionItemStyle from './QExpansionItemStyle.vue';
import MyRouterLinkListContentItem from './MyRouterLinkListContentItem.vue';

interface Props {
	data: any[];
	modelValue: any;
	defaultOpened?: string[];
}

const props = withDefaults(defineProps<Props>(), {
	data: [] as any,
	loading: false
});
const router = useRouter();
const emits = defineEmits(['update:modelValue']);
const isEmpty = computed(() => {
	const empty = isNil(props.data)
		? true
		: isArray(props.data) && props.data.length === 0
		? true
		: false;
	return !props.loading && empty;
});

const active = computed(() => props.modelValue);

const routePush = (route: RouteLocationRaw) => {
	router.push(route);
};

const clickHandler = (item: any) => {
	routePush(item.route);
	emits('update:modelValue', item.id, item);
};
</script>
