<template>
	<q-card flat :bordered="false" square class="my-networdks-container">
		<div class="row justify-between">
			<q-tabs
				v-model="tab"
				align="left"
				class="text-grey"
				active-color="primary"
				indicator-color="primary"
			>
				<q-tab
					v-for="item in tabList"
					:key="item.name"
					:name="item.name"
					style="min-height: 44px"
					no-caps
				>
					<span class="my-tab-wrapper">
						{{ item.label }}
					</span>
				</q-tab>
			</q-tabs>
		</div>
	</q-card>
	<bt-scroll-area
		style="height: calc(100vh - 44px)"
		class="my-scroll-container"
	>
		<MyLoading :showing="loading">
			<q-card flat :bordered="false" style="margin: 8px">
				<q-list>
					<template v-for="(child, childIndex) in listTab" :key="childIndex">
						<template
							v-for="(item, index) in child.spec.containers"
							:key="`child-${index}`"
						>
							<div @click="itemClick(child, item)">
								<q-item>
									<q-item-section avatar>
										<img :src="indicationIcon" alt="" />
									</q-item-section>
									<q-item-section>
										<q-item-label lines="1">
											{{ item.name }}
										</q-item-label>
										<q-item-label caption lines="1">
											<div></div>
											{{ child.metadata.namespace }}&nbsp;
											{{ child.metadata.name }}
										</q-item-label>
									</q-item-section>
									<q-item-section side>
										<img :src="arrowRightIcon" alt="" />
									</q-item-section>
								</q-item>
							</div>
							<q-separator
								spaced
								inset
								v-if="
									index < child.spec.containers.length &&
									childIndex < listTab.length - 1
								"
							/>
						</template>
					</template>
				</q-list>
			</q-card>
		</MyLoading>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue';
import indicationIcon from '@apps/dashboard/assets/indication.svg';
import arrowRightIcon from '@apps/dashboard/assets/arrow-right.svg';
import { useRoute, useRouter } from 'vue-router';
import { getPodsList, getUsers } from '@apps/dashboard/src/network';
import {
	Container,
	PodItem,
	ResourcesResponse
} from '@apps/dashboard/src/types/network';
import MyLoading from '@apps/dashboard/src/components/MyLoading.vue';

const UnderlayerNamespaces = [
	'default',
	'kube-node-lease',
	'kube-public',
	'kube-system',
	'kubekey-system',
	'kubesphere-controls-system',
	'kubesphere-monitoring-federated',
	'kubesphere-monitoring-system',
	'kubesphere-system',
	'os-system',
	'os-gpu',
	'os-network',
	'os-framework',
	'os-platform'
];

const tab = ref('1');
const tabList = [
	{
		name: '1',
		label: 'Sys'
	},
	{
		name: '2',
		label: 'User'
	}
];

let list = reactive<ResourcesResponse>({
	items: [],
	totalItems: 0
});

const loading = ref(false);
const users = ref<any>([]);

const router = useRouter();
const itemClick = (data: PodItem, contianer: Container) => {
	router.push({
		path: `/logs/${data.metadata.namespace}/${data.metadata.name}/${contianer.name}`
	});
};

const fetchData = () => {
	const params = {
		sortBy: 'startTime'
	};

	loading.value = true;

	Promise.all([getUsers(), getPodsList(params)])
		.then(([usersRes, res]) => {
			list.items = res.data.items;
			list.totalItems = res.data.totalItems;
			users.value = usersRes.data.items;
		})
		.finally(() => {
			loading.value = false;
		});
};

const listTab = computed(() => {
	const user = users.value[0]?.metadata?.name ?? '';

	if (tab.value === '1') {
		return list.items.filter(
			(item) =>
				!item.metadata.namespace.includes(user) &&
				UnderlayerNamespaces.includes(item.metadata.namespace)
		);
	} else {
		return list.items.filter((item) => item.metadata.namespace.includes(user));
	}
});

fetchData();
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
