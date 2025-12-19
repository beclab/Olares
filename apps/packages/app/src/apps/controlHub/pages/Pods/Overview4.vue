<template>
	<MyContentPage :title="$route.params.name">
		<template #extra>
			<div class="col-auto">
				<QButtonStyle>
					<q-btn dense flat icon="sym_r_preview" @click="clickHandler">
						<q-tooltip>
							<div style="white-space: nowrap">
								{{ $t('VIEW_YAML') }}
							</div>
						</q-tooltip>
					</q-btn>
				</QButtonStyle>
			</div>
		</template>
		<MyPage>
			<Overview></Overview>
			<ContainerWrapper></ContainerWrapper>
			<Volumes></Volumes>
			<Metadata></Metadata>
			<Environments></Environments>
			<Events></Events>
			<q-inner-loading :showing="loading"></q-inner-loading>
		</MyPage>
		<Yaml name="" ref="yamlRef"></Yaml>
	</MyContentPage>
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
import MyContentPage from '../../components/MyContentPage2.vue';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';
import Yaml from './Yaml.vue';
import MoreSelection from '@apps/control-panel-common/src/components/MoreSelection.vue';
import { t } from '@apps/control-hub/src/boot/i18n';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';

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

const clickHandler = () => {
	yamlRef.value.show();
};

watch(() => route.params, fetchData, { immediate: true });
</script>

<style></style>
