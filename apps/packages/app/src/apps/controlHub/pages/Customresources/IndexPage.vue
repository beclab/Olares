<template>
	<MyRouterLinkList
		:defaultActive="defaultActive"
		:defaultOpened="defaultOpened"
		:data="data"
		:loading="loading"
		:clickable="false"
	>
		<MenuHeader></MenuHeader>
	</MyRouterLinkList>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { RouteLocation, useRoute, useRouter } from 'vue-router';
import { getCustomresources } from '@apps/control-hub/src/network';
import { CRDMapper } from '@apps/control-hub/src/utils/object.mapper';
import MyRouterLinkList from '@apps/control-panel-common/src/components/MyRouterLinkList.vue';
import MenuHeader from '@apps/control-hub/src/layouts/MenuHeader.vue';
import userSpaceIcon from '@apps/control-hub/src/assets/user-space.svg';
import { groupBy } from 'lodash';

const menuOptions = {
	title: 'title',
	code: 'id',
	icon: 'logo',
	subTitle: 'subTitle',
	router: true
};

interface IData {
	id: string;
	title: string;
	subTitle: string;
	route: RouteLocation;
}
const route = useRoute();
let list = ref();
const data = ref<any>([]);
const defaultActive = ref();
const router = useRouter();
const shouldExecuteResponseHandler = ref(true);
const page = ref(1);
const limit = ref(24);
const total = ref(0);
const defaultOpened = ref();
const loading = ref(false);

const itemClick = (data: any) => {
	router.push({
		path: `/customresources/${data.latestVersion}/${data.module}`
	});
};

const fetchData = () => {
	const params = {
		sortBy: 'createTime'
	};
	loading.value = true;
	getCustomresources(params)
		.then((res) => {
			let firstPath = '';
			let firstActive = '';
			let path = '';
			let id = '';
			total.value = res.data.totalItems;
			list.value = res.data.items.map((item) => CRDMapper(item));
			const groupedData = groupBy(list.value, 'group');
			const result = Object.keys(groupedData).map((key: string) => {
				return {
					id: key,
					title: key,
					children: groupedData[key].map((item) => {
						path = `/customresources/detail/${item.group}/${item.latestVersion}/${item.module}/${item.uid}`;
						id = `${item.uid}`;

						if (!firstPath) {
							firstPath = path;
						}
						if (!firstActive) {
							firstActive = id;
							defaultOpened.value = [key];
						}
						return {
							id,
							title: item.kind,
							img: userSpaceIcon,
							subTitle: `${item.group}/${item.latestVersion}`,
							route: {
								path
							}
						};
					})
				};
			});

			toDefaultRoute(firstPath, firstActive);

			data.value = result;
		})
		.finally(() => {
			loading.value = false;
		});
};

const toDefaultRoute = (firstPath: string, firstActive: string) => {
	if (shouldExecuteResponseHandler.value) {
		const { uid, group } = route.params;
		if (uid) {
			defaultActive.value = uid;
			defaultOpened.value = [group];
		} else {
			!route.params.module && firstPath && router.push(firstPath);
			defaultActive.value = firstActive;
		}
	}
};

onMounted(() => {
	fetchData();
});

onBeforeUnmount(() => {
	shouldExecuteResponseHandler.value = false;
});
</script>

<style lang="scss" scoped>
.my-nodes-container {
	padding: 0 16px;
	line-height: 44px;
	.my-tab-wrapper {
		font-family: 'Roboto';
		font-style: normal;
		font-weight: 500;
		font-size: 12px;
		line-height: 16px;
	}
}
</style>
