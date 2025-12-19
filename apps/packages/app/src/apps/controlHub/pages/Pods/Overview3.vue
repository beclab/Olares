<template>
	<FullPageWithBack :title="namespace">
		<template #extra>
			<Yaml name="" ref="yamlRef"></Yaml>
			<QButtonStyle>
				<q-btn dense flat icon="sym_r_preview" @click="clickHandler">
					<q-tooltip>
						<div style="white-space: nowrap">
							{{ $t('VIEW_YAML') }}
						</div>
					</q-tooltip>
				</q-btn>
			</QButtonStyle>
		</template>
		<div
			style="margin: 0 -44px; margin-bottom: -44px; width: calc(100% + 88px)"
		>
			<div class="q-gutter-y-lg q-pa-lg bg-background-6">
				<Overview></Overview>
				<div class="full-width" style="overflow: hidden">
					<ContainerWrapper></ContainerWrapper>
				</div>
				<Volumes></Volumes>
				<Metadata></Metadata>
				<Environments></Environments>
				<div style="position: relative">
					<Events></Events>
				</div>
				<q-inner-loading :showing="loading"></q-inner-loading>
			</div>
		</div>
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
import { computed, ref, watch } from 'vue';
import MyContentPage from '../../components/MyContentPage.vue';
import Yaml from './Yaml.vue';
import { t } from '@apps/control-hub/src/boot/i18n';
import FullPageWithBack from '@apps/control-panel-common/src/components/FullPageWithBack.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';

const usePod = UsePod();
const route = useRoute();
const loading = ref(false);
const yamlRef = ref();

const namespace = computed(() => route.params.namespace as string);

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

const clickHandler = () => {
	yamlRef.value.show();
};

watch(() => route.params, fetchData, { immediate: true });
</script>

<style></style>
