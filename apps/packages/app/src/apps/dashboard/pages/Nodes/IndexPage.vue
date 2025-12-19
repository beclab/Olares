<template>
	<q-card flat :bordered="false" square class="my-nodes-container">
		Nodes
	</q-card>
	<div>
		<bt-scroll-area
			loading
			style="height: calc(100vh - 44px)"
			class="my-scroll-container"
		>
			<MyLoading :showing="loading">
				<q-card flat :bordered="false" style="margin: 8px">
					<q-list>
						<template v-for="(item, index) in list.items" :key="index">
							<div @click="itemClick(item)">
								<q-item>
									<q-item-section avatar>
										<img :src="indicationIcon" alt="" />
									</q-item-section>
									<q-item-section>
										<q-item-label lines="1">{{
											item.metadata.name
										}}</q-item-label>
									</q-item-section>
									<q-item-section side>
										<img :src="arrowRightIcon" alt="" />
									</q-item-section>
								</q-item>
							</div>
							<q-separator spaced inset v-if="index < list.items.length - 1" />
						</template>
					</q-list>
				</q-card>
			</MyLoading>
		</bt-scroll-area>
	</div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';

import { useRouter } from 'vue-router';
import { getNodesList } from '@apps/control-panel-common/src/network/index.ts';
import MyLoading from '@apps/control-panel-common/src/components/MyLoading.vue';
let list = reactive<any>({
	items: [],
	totalItems: 0
});

let loading = ref(false);
const router = useRouter();
const itemClick = (data: any) => {
	router.push({
		path: `nodes/${data.metadata.name}/${data.metadata.creationTimestamp}`
	});
};
const params = {
	sortBy: 'createTime'
	// labelSelector: '!node-role.kubernetes.io/edge',
};

onMounted(() => {
	loading.value = true;
	getNodesList(params)
		.then((res) => {
			console.log(res.data.items);
			list.items = res.data.items;
		})
		.finally(() => {
			loading.value = false;
		});
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
