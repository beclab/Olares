<template>
	<MyContentPage>
		<template #extra>
			<div class="col-auto">
				<PreviewButton :tooltipText="$t('VIEW_YAML')" @click="clickHandler" />
			</div>
		</template>
		<MyPage>
			<Overview>
				<template #extra>
					<PreviewButton
						v-if="isStudio"
						:tooltipText="$t('VIEW_YAML')"
						@click="clickHandler"
					/>
				</template>
			</Overview>
			<ContainerWrapper></ContainerWrapper>
			<Volumes></Volumes>
			<Metadata></Metadata>
			<Environments></Environments>
			<Events></Events>
			<q-inner-loading :showing="loading"></q-inner-loading>
		</MyPage>
		<Yaml :name="name" ref="yamlRef"></Yaml>
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
import { computed, h, ref, watch } from 'vue';
import MyContentPage from '../../components/MyContentPage.vue';
import MyPage from '@apps/control-panel-common/src/containers/MyPage.vue';
import Yaml from './Yaml.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import { get } from 'lodash';
import { QBtn, QTooltip } from 'quasar'; // 确保已导入 Quasar 组件
import { useIsStudio } from '@apps/control-hub/src/stores/hook';
const isStudio = useIsStudio();
const usePod = UsePod();
const route = useRoute();
const loading = ref(false);
const yamlRef = ref();
const name = computed(() => get(usePod, 'data.name'));
const fetchData = () => {
	const { namespace, name }: any = route.params;
	const params = {
		namespace: namespace,
		podName: name
	};
	loading.value = true;
	usePod.setDetail({});
	getPodDetail(params)
		.then((res) => {
			usePod.setDetail(res.data);
		})
		.finally(() => {
			loading.value = false;
		})
		.catch(() => {
			usePod.setDetail({});
		});
};

const PreviewButton = (props, { emit }) => {
	return h(
		QButtonStyle,
		{},
		h(
			QBtn,
			{
				dense: true,
				flat: true,
				icon: props.icon || 'sym_r_preview',
				onClick: () => emit('click')
			},
			[
				h(
					QTooltip,
					{},
					h('div', { style: 'white-space: nowrap' }, props.tooltipText)
				)
			]
		)
	);
};

watch(() => route.params, fetchData, { immediate: true });
const clickHandler = () => {
	yamlRef.value.show();
};
</script>

<style></style>
