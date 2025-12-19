<template>
	<MyTree
		:data="list"
		active-text-color="#1976D2"
		:default-active="defaultActive"
		:menu-options="menuOptions"
		:default-openeds="menuDefaultOpeneds"
		hiddenEmpty
	/>
	<q-inner-loading :showing="loading"> </q-inner-loading>
</template>
<script setup lang="ts">
import {
	getConfigmaps,
	getServiceaccounts,
	getSecrets
} from '@apps/control-hub/src/network';
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import MyTree from '@apps/control-panel-common/src/components/Menu/MyTree.vue';
import { t } from '@apps/control-hub/src/boot/i18n';

const menuOptions = {
	title: 'title',
	code: 'id',
	icon: 'icon',
	router: true,
	collapse: false
};

const data = [
	{
		label: t('SECRET_PL'),
		value: 'secret'
	},
	{
		label: t('CONFIGMAP_PL'),
		value: 'configmap'
	},
	{
		label: t('SERVICE_ACCOUNT_PL'),
		value: 'service-account'
	}
];

const menuDefaultOpeneds = data.map((item) => item.value);

const router = useRouter();
const route = useRoute();
const list = ref();
const loading = ref(false);
const defaultActive = ref();
const shouldExecuteResponseHandler = ref(true);

console.log('route', route);

const fetchData = () => {
	const { namespace }: { [key: string]: any } = route.params;

	loading.value = true;
	list.value = [];
	const params = {
		sortBy: 'createTime',
		namespace
	};
	Promise.all([
		getSecrets(params),
		getConfigmaps(params),
		getServiceaccounts(params)
	])
		.then((res) => {
			let firstPath = '';
			let firstActive = '';
			let path = '';
			let id = '';
			const dataTemp = data.map((child, index) => ({
				title: child.label,
				id: child.value,
				children: res[index].data.items.map((item: any) => {
					path = `/application-spaces/configurations/${namespace}/detail/${child.value}/${item.metadata.name}`;
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
				})
			}));

			list.value = dataTemp.filter(
				(item) => item.children && item.children.length > 0
			);

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
