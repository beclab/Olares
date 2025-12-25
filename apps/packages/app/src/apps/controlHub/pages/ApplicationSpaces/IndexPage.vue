<template>
	<MyTree
		:data="list"
		active-text-color="#1976D2"
		:menu-options="menuOptions"
		:default-openeds="defaultOpeneds"
		:default-active="defaultActive"
		:loading="loading"
		:accordion="false"
		ref="myTreeRef"
		:header-after-hide="headerHide"
	>
		<MenuHeader></MenuHeader>
	</MyTree>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { getNamespacesGroup } from '@apps/control-hub/src/network';
import MyTree from '@apps/control-panel-common/src/components/Menu/MyTree.vue';
import { get } from 'lodash-es';
import { getNamespaceIcon, customNamesapceIcon } from './config';
import MenuHeader from '@apps/control-hub/src/layouts/MenuHeader.vue';

const menuOptions = {
	title: 'title',
	code: 'id',
	icon: 'logo',
	router: true
};

const list = ref([]);
const loading = ref(false);
const route = useRoute();
const myTreeRef = ref();

const headerHide = computed(() => !!route.meta.headerHide);

const defaultActive = computed(() => {
	return route.params.namespace;
});

const defaultOpeneds = ref();

const fetchData = () => {
	const params = {
		sortBy: 'createTime',
		labelSelector: 'kubesphere.io/workspace!=kubesphere.io/devopsproject'
	};
	loading.value = true;

	getNamespacesGroup(params)
		.then((res) => {
			const result = res.data;
			const data: any = result;
			const newData = data.map((workspace: any) => ({
				title: workspace.title,
				id: workspace.title,
				selectable: false,
				icon: getNamespaceIcon('default'),
				children: workspace.data.map((item: any) => {
					return {
						title: item.metadata.name,
						id: item.metadata.name,
						img: getNamespaceIcon(item.metadata.name, workspace.title),
						route: {
							path: `/application-spaces/workloads/${item.metadata.name}`
						}
					};
				})
			}));

			list.value = newData.filter(
				(item: any) => item.children && item.children.length > 0
			);

			const users = data.map((item) => item.title);

			const target = users.find((item) =>
				get(route, 'params.namespace', '').includes(item)
			);
			const kind = get(route, 'params.kind', '');
			nextTick(() => {
				if (!kind) {
					myTreeRef.value.setExpanded(users[0], true);
				} else if (target) {
					myTreeRef.value.setExpanded(target, true);
				} else {
					myTreeRef.value.setExpanded(users[users.length - 1], true);
				}
			});
		})
		.finally(() => {
			loading.value = false;
		});
};

fetchData();
</script>

<style lang="scss" scoped></style>
