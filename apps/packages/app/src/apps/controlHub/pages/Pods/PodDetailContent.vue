<template>
	<FullPageWithBack :title="$route.params.name">
		<template #extra>
			<div class="col-auto">
				<MoreSelection :options="options" size="md"></MoreSelection>
			</div>
		</template>
		<div class="q-gutter-y-lg bg-red">
			<Overview></Overview>
			<ContainerWrapper></ContainerWrapper>
			<Volumes></Volumes>
			<Metadata></Metadata>
			<Environments></Environments>
			<Events></Events>
		</div>
		<q-inner-loading :showing="loading"></q-inner-loading>
		<Yaml name="" ref="yamlRef"></Yaml>
	</FullPageWithBack>
</template>

<script setup lang="ts">
import Overview from './Overview.vue';
import Metadata from './Metadata.vue';
import Environments from './Environments.vue';
import Events from './Events.vue';
import Volumes from './Volumes.vue';
import ContainerWrapper from './ContainerWrapper.vue';
import { getPodDetail } from '@apps/control-hub/src/network';
import { useRoute } from 'vue-router';
import { UsePod } from '@apps/control-panel-common/src/stores/PodData';
import { ref, watch } from 'vue';
import MyContentPage from '../../components/MyContentPage.vue';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';
import Yaml from './Yaml.vue';
import MoreSelection from '@apps/control-panel-common/src/components/MoreSelection.vue';
import { t } from '@apps/control-hub/src/boot/i18n';
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack.vue';
import RouterViewTransition from '@apps/control-panel-common/src/components/RouterViewTransition.vue';

const options = [
	{
		key: 'edit',
		icon: 'sym_r_preview',
		label: t('VIEW_YAML'),
		action: 'edit',
		onClick: () => {
			yamlRef.value.show();
		}
	}
];

const usePod = UsePod();
const route = useRoute();
const loading = ref(false);
const yamlRef = ref();

const fetchData = () => {
	const { namespace, name }: any = route.params;
	const params = {
		namespace: namespace,
		podName: name
	};
	loading.value = true;
	getPodDetail(params)
		.then((res) => {
			console.log('getPodDetail', res.data);
			usePod.setDetail(res.data);
		})
		.finally(() => {
			loading.value = false;
		});
};

watch(() => route.params, fetchData, { immediate: true });
</script>

<style lang="scss" scoped>
.border-grey-2 {
	border: 1px solid $grey-2;
}
</style>
