<template>
	<bt-scroll-area loading style="height: 100%">
		<MyLoading :showing="loading">
			<q-card flat :bordered="false">
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
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue';
import indicationIcon from '@apps/control-hub/assets/indication.svg';
import arrowRightIcon from '@apps/control-hub/assets/arrow-right.svg';
import { useRoute, useRouter } from 'vue-router';
import {
	getNameSpacePodsList,
	getPodsList
} from '@apps/control-hub/src/network';
import {
	PodItem,
	ResourcesResponse
} from '@apps/control-hub/src/types/network';
import MyLoading from '@apps/control-hub/src/components/MyLoading.vue';
import ContainersList from '../Containers/IndexPage.vue';
import MyQTabs from '@apps/control-hub/src/components/MyQTabs.vue';
import { get, isEmpty } from 'lodash';
import { t } from '@apps/control-hub/src/boot/i18n';
import DetailPage from '@apps/control-hub/src/containers/DetailPage.vue';
import { getLocalTime } from '@apps/control-hub/src/utils/index.js';
import Metadata from '@apps/control-hub/src/containers/Metadata.vue';
import { PodsMapper } from '@apps/control-hub/src/utils/object.mapper';

let list = reactive<ResourcesResponse>({
	items: [],
	totalItems: 0
});

let loading = ref(false);
const visible = ref(false);
const route = useRoute();
const podData = ref();
const itemClick = (data: any) => {
	podData.value = data;
	const detail = PodsMapper(data);
	visible.value = true;
};

const fetchData = () => {
	const { namespace }: any = route.params;

	const params = {
		sortBy: 'startTime',
		namespace
	};

	loading.value = true;
	getNameSpacePodsList(params)
		.then((res) => {
			list.items = res.data.items;
			list.totalItems = res.data.totalItems;
		})
		.finally(() => {
			loading.value = false;
		});
};

watch(
	() => route.params.namespace,
	async (newId) => {
		fetchData();
	}
);
onMounted(() => {
	fetchData();
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
