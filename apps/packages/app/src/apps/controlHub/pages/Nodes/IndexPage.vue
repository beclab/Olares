<template>
	<MyTree
		:data="list.items"
		:default-active="defaultActive"
		:menu-options="menuOptions"
		:default-openeds="defaultOpeneds"
		:loading="loading"
		:accordion="true"
		ref="myTree"
		@open="openHandler"
		@lazy-load="onLazyLoad"
	>
		<MenuHeader></MenuHeader>
	</MyTree>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { getNodesList } from '@apps/control-hub/src/network';
import MyTree from '@apps/control-panel-common/src/components/Menu/MyTree.vue';
import indication from '@apps/control-panel-common/src/assets/indication.svg';
import MenuHeader from '@apps/control-hub/src/layouts/MenuHeader.vue';
import { getPodsList } from '@apps/control-panel-common/src/network';
import podIcon from '@apps/control-panel-common/src/assets/pod.svg';
import { useSplitMenu } from '@apps/control-panel-common/src/stores/menu';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';

const splitMenu = useSplitMenu();
const router = useRouter();
const shouldExecuteResponseHandler = ref(true);
const menuOptions = {
	title: 'title',
	code: 'id',
	icon: 'logo',
	subTitle: 'subTitle',
	router: true
};

const defaultOpeneds = ref();
const defaultActive = ref();
const myTree = ref();

let list = reactive<any>({
	items: [],
	totalItems: 0
});

const openHandler = (id: string, target: any) => {
	const data = target.children[0];
	defaultActive.value = data.id;
	router.push(data.route);
};

let loading = ref(false);

const fetchData = () => {
	const params = {
		sortBy: 'createTime'
		// labelSelector: '!node-role.kubernetes.io/edge',
	};
	loading.value = true;
	getNodesList(params)
		.then((res) => {
			list.items = res.data.items.map((item: any, index: number) => ({
				title: item.metadata.name,
				id: `${item.metadata.name}`,
				selectable: true,
				img: indication,
				imgActiveDisabled: true,
				originData: item,
				lazy: true,
				route: { path: `/nodes/detail/${item.metadata.name}` }
			}));
			const originData = list.items[0].originData;
			if (!defaultActive.value && shouldExecuteResponseHandler.value) {
				router.push({
					path: `/nodes/detail/${originData.metadata.name}`
				});
				defaultActive.value = `${originData.metadata.name}`;
			}
			nextTick(() => {
				myTree.value.setExpanded(originData.metadata.name, true);
			});
		})
		.finally(() => {
			loading.value = false;
		});
};

const idFormat = (namespace: any, name: any) =>
	namespace && name ? `${namespace}-${name}` : undefined;

const fetchPods = (nodeName: string) => {
	const params = {
		nodeName,
		sortBy: 'startTime'
	};
	return getPodsList(params).then((res) => {
		const result = res.data.items.map((item) => ObjectMapper.pods(item));
		return result.map((item: any) => ({
			title: item.name,
			id: idFormat(item.namespace, item.name),
			img: podIcon,
			originData: item,
			selectToExpend: true,
			status: item.podStatus.type,
			route: {
				path: `/${splitMenu.type}/pods/overview/${item.spec.nodeName}/${item.namespace}/${item.name}`
			}
		}));
	});
};

const onLazyLoad = async ({ node, key, done, fail }: any) => {
	const result = await fetchPods(key);
	done(result);
};

onMounted(() => {
	fetchData();
});

onBeforeUnmount(() => {
	shouldExecuteResponseHandler.value = false;
	defaultActive.value = undefined;
});
</script>

<style lang="scss" scoped>
.my-nodes-container {
	padding: 0 16px;
	line-height: 44px;
	font-size: 12px;
	.my-tab-wrapper {
		font-family: 'Roboto';
		font-style: normal;
		font-weight: 500;
		font-size: 12px;
		line-height: 16px;
	}
}
</style>
