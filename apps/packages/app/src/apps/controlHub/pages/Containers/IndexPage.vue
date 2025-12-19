<template>
	<MySplitter>
		<template v-slot:before>
			<MyScrollArea>
				<MyTree
					:data="list.items"
					active-text-color="#1976D2"
					:menu-options="menuOptions"
					unique-opened
					autoView
					:default-active="defaultActive"
					:loading="loading"
				>
					<template #menu-item="{ data }">
						<q-item-section>
							<q-item-label lines="1">
								{{ data.title }}
							</q-item-label>
							<q-item-label caption lines="1">
								<div></div>
								{{ data.metadata.namespace }}&nbsp;
								{{ data.metadata.name }}
							</q-item-label>
						</q-item-section>
					</template>
				</MyTree>
			</MyScrollArea>
		</template>
		<template v-slot:after>
			<div style="flex: 1; height: 100%">
				<router-view></router-view>
			</div>
		</template>
	</MySplitter>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { getPodDetail } from '@apps/control-hub/src/network';
import MyTree from '@apps/control-panel-common/src/components/Menu/MyTree.vue';
import MySplitter from '@apps/control-panel-common/src/components/MySplitter.vue';
import MyScrollArea from '@apps/control-panel-common/src/containers/MyScrollArea.vue';
import { useSplitMenu } from '@apps/control-panel-common/src/stores/menu';

const route = useRoute();
const router = useRouter();
const splitMenu = useSplitMenu();

const menuOptions = {
	title: 'title',
	code: 'id',
	icon: 'logo',
	subTitle: 'subTitle',
	router: true
};

const tab = ref('1');
const tabList = (item: any, data: any) => {
	const path = `${data.spec.nodeName}/${data.metadata.namespace}/${data.metadata.name}/${item.name}`;
	const idTemp = `${data.metadata.namespace}-${data.metadata.name}-${item.name}`;
	const params = {
		namespace: data.metadata.namespace,
		name: data.metadata.namespace,
		container: name
	};

	console.log('name-route', route.params);
	return [
		{
			title: 'Overview',
			id: `overview-${idTemp}`,
			route: { path: `/${splitMenu.type}/pods/containers/overview/${path}` }
		},
		{
			title: 'Environments',
			id: `environments-${idTemp}`,
			route: {
				path: `/${splitMenu.type}/pods/containers/environments/${path}`
			}
		},
		{
			title: 'Logs',
			id: `logs-${idTemp}`,
			route: {
				path: `/${splitMenu.type}/pods/containers/logs/${path}`
			}
		}
	];
};

const statusList = ref();
const visible = ref(false);
const containerData = ref();
const containerPath = ref({
	namespace: '',
	podName: '',
	container: ''
});

const idFormat = (container: any) => container;
const type = route.fullPath.split('/')[4];
const defaultActive = ref(route.params.container);
let list = reactive<any>({
	items: [],
	totalItems: 0,
	containers: []
});

let loading = ref(false);

const itemClick = (id: string, target: any) => {
	const data = target.originData;
	defaultActive.value = data.container.name;
	router.push({
		path: `/${splitMenu.type}/pods/containers/overview/${data.spec.nodeName}/${data.metadata.namespace}/${data.metadata.name}/${data.container.name}`
	});
};

const fetchData = () => {
	const { namespace, name }: Record<string, any> = route.params;
	const params = {
		sortBy: 'startTime',
		namespace,
		podName: name
	};
	loading.value = true;
	getPodDetail(params)
		.then((res) => {
			const data = res.data;
			list.items = res.data.spec.containers.map((item) => {
				const path = `${data.spec.nodeName}/${data.metadata.namespace}/${data.metadata.name}/${item.name}`;

				return {
					title: item.name,
					metadata: res.data.metadata,
					id: item.name,
					originData: { ...res.data, container: item },
					route: {
						path: `/${splitMenu.type}/pods/containers/overview/${path}`
					}
				};
			});

			if (defaultActive.value) {
				const target = list.items.find(
					(item: any) => item.id === idFormat(route.params.container)
				);
				target && itemClick('', target);
			} else {
				itemClick('', list.items[0]);
			}
		})
		.finally(() => {
			loading.value = false;
		});
};
watch(
	() => route.params.name,
	() => {
		fetchData();
	}
);
fetchData();
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
