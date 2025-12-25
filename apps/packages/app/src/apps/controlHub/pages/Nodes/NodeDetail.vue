<template>
	<MyContentPage>
		<MyPage>
			<Detail :detail="detail"></Detail>
			<Metadata v-if="data" :data="data.metadata"> </Metadata>
			<Events :detail="detail"></Events>
			<q-inner-loading :showing="loading"> </q-inner-loading>
		</MyPage>
	</MyContentPage>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import { ref, watch } from 'vue';
import { getNodeDetail } from '@apps/control-hub/src/network';
import Detail from './Detail.vue';
import Events from './Events.vue';
import Metadata from '@apps/control-panel-common/src/containers/Metadata.vue';
import MyContentPage from '@apps/control-hub/src/components/MyContentPage.vue';
import { ObjectMapper } from '@apps/control-panel-common/src/utils/object.mapper';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';

const loading = ref(false);
const data = ref();
const detail = ref();
const route = useRoute();

const fetchList = () => {
	const { node }: Record<string, any> = route.params;
	loading.value = true;
	getNodeDetail(node)
		.then((res) => {
			data.value = res.data;
			detail.value = ObjectMapper.nodes(res.data);
		})
		.finally(() => {
			loading.value = false;
		});
};

watch(
	() => route.params.node,
	() => {
		fetchList();
	},
	{
		immediate: true
	}
);
</script>

<style lang="scss" scoped></style>
