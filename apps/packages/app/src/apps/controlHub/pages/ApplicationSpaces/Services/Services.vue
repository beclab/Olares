<template>
	<MyTree
		:data="list"
		active-text-color="#1976D2"
		:default-active="defaultActive"
		:default-openeds="defaultOpeneds"
		:menu-options="menuOptions"
		hiddenEmpty
	/>
	<q-inner-loading :showing="loading"> </q-inner-loading>
</template>

<script setup lang="ts">
import { getNameSpaceServicesList } from '@apps/control-hub/src/network';
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import MyTree from '@apps/control-panel-common/src/components/Menu/MyTree.vue';

const data = ['Deployments', 'Statefulsets', 'Daemonsets'];

const menuOptions = {
	title: 'title',
	code: 'id',
	icon: 'logo',
	router: true,
	collapse: false
};

const router = useRouter();
const route = useRoute();
const list = ref<any[]>([]);
const loading = ref(false);
const defaultActive = ref();
const shouldExecuteResponseHandler = ref(true);
const defaultOpeneds = ref('service');

const fetchData = () => {
	const { namespace }: { [key: string]: any } = route.params;

	loading.value = true;
	list.value = [];
	const params = {
		sortBy: 'createTime',
		namespace
	};
	getNameSpaceServicesList(params)
		.then((res) => {
			let firstPath = '';
			let firstActive = '';
			let path = '';
			let id = '';

			const data = res.data.items.map((item: any) => {
				path = `/application-spaces/services/${namespace}/detail/${item.metadata.name}`;
				id = item.metadata.name;
				if (!firstPath) {
					firstPath = path;
				}
				if (!firstActive) {
					firstActive = id;
				}
				return {
					id,
					title: item.metadata.name,
					route: {
						path,
						key: item.metadata.name,
						id: item.metadata.name
					}
				};
			});

			list.value = data.length
				? [
						{
							title: 'Service',
							id: 'service',
							selectable: false,
							children: data
						}
				  ]
				: [];
			toDefaultRoute(firstPath, firstActive);
		})
		.finally(() => {
			loading.value = false;
		});
};

const toDefaultRoute = (firstPath: string, firstActive: string) => {
	if (shouldExecuteResponseHandler.value) {
		// firstPath && router.push(firstPath);
		// defaultActive.value = firstActive;
	}
};

watch(
	() => route.params.namespace,
	async (newId) => {
		fetchData();
	}
);

onBeforeUnmount(() => {
	shouldExecuteResponseHandler.value = false;
});

onMounted(() => {
	fetchData();
});
</script>

<style></style>
