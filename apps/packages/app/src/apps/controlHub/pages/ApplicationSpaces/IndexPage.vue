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
import { computed, nextTick, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { getNamespacesGroup } from '@apps/control-hub/src/network';
import MyTree from '@apps/control-panel-common/src/components/Menu/MyTree.vue';
import { get } from 'lodash-es';
import { getNamespaceIcon, customNamesapceIcon } from './config';
import MenuHeader from '@apps/control-hub/src/layouts/MenuHeader.vue';

const APP_NAME_LABEL = 'applications.app.bytetrade.io/name';
const APP_SCOPED_NAME_LABEL =
	'applications.app.bytetrade.io/need_cluster_scoped_app';
const WORKSPACE_GROUP_SHARED = 'Shared';
const WORKSPACE_GROUP_SYSTEM = 'System';

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
	const ns = route.params.namespace;
	return typeof ns === 'string' ? ns : Array.isArray(ns) ? ns[0] : '';
});

const defaultOpeneds = ref<string[]>([]);

const findWorkspaceIdForNamespace = (namespace: string) => {
	if (!namespace) {
		return undefined;
	}
	const nodes = list.value as any[];
	for (const w of nodes) {
		if (w.children?.some((c: any) => c.id === namespace)) {
			return w.id as string;
		}
	}
	return undefined;
};

const syncTreeWithRoute = () => {
	const namespace = defaultActive.value as string;
	if (!namespace || !myTreeRef.value) {
		return;
	}
	const workspaceId = findWorkspaceIdForNamespace(namespace);
	nextTick(() => {
		nextTick(() => {
			if (!myTreeRef.value) {
				return;
			}
			if (workspaceId) {
				myTreeRef.value.setExpanded(workspaceId, true);
			}
			myTreeRef.value.setSelected?.(namespace);
		});
	});
};

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

			const userNamespaceEntries = data.flatMap((workspace: any) => {
				if (
					workspace.title === WORKSPACE_GROUP_SHARED ||
					workspace.title === WORKSPACE_GROUP_SYSTEM
				) {
					return [];
				}
				return (workspace.data || []).map((item: any) => ({
					workspaceTitle: workspace.title,
					item
				}));
			});

			const resolveSharedNamespaceImg = (item: any) => {
				const sharedAppName = item?.metadata?.labels?.[APP_NAME_LABEL];
				if (!sharedAppName) {
					return null;
				}
				const matched = userNamespaceEntries.find(
					(e) =>
						e.item?.metadata?.labels?.[APP_NAME_LABEL] === sharedAppName ||
						e.item?.metadata?.labels?.[APP_SCOPED_NAME_LABEL] === sharedAppName
				);
				if (!matched) {
					return null;
				}
				return getNamespaceIcon(
					matched.item.metadata.name,
					matched.workspaceTitle
				);
			};

			const newData = data.map((workspace: any) => ({
				title: workspace.title,
				id: workspace.title,
				selectable: false,
				icon: getNamespaceIcon('default'),
				children: workspace.data.map((item: any) => {
					const img =
						workspace.title === WORKSPACE_GROUP_SHARED
							? resolveSharedNamespaceImg(item) ??
							  getNamespaceIcon(item.metadata.name, workspace.title)
							: getNamespaceIcon(item.metadata.name, workspace.title);
					return {
						title: item.metadata.name,
						id: item.metadata.name,
						img,
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
			const namespace = defaultActive.value;
			const workspaceId = findWorkspaceIdForNamespace(namespace);

			nextTick(() => {
				nextTick(() => {
					if (!myTreeRef.value) {
						return;
					}
					if (workspaceId) {
						myTreeRef.value.setExpanded(workspaceId, true);
					} else if (users.length) {
						const kind = get(route, 'params.kind', '');
						if (!kind) {
							myTreeRef.value.setExpanded(users[0], true);
						} else {
							myTreeRef.value.setExpanded(users[users.length - 1], true);
						}
					}
					myTreeRef.value.setSelected?.(namespace);
				});
			});
		})
		.finally(() => {
			loading.value = false;
		});
};

fetchData();

watch(
	() => [route.params.namespace, route.name, route.path],
	() => {
		if ((list.value as any[]).length) {
			syncTreeWithRoute();
		}
	}
);
</script>

<style lang="scss" scoped></style>
