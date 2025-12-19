<template>
	<component :is="drawer ? MyContentPage2 : MyContentPage">
		<template #extra>
			<div class="col-auto">
				<QButtonStyle v-for="item in options3" :key="item.value">
					<q-btn dense flat :icon="item.icon" @click="item.onClick">
						<q-tooltip>
							<div style="white-space: nowrap">
								{{ item.label }}
							</div>
						</q-tooltip>
					</q-btn>
				</QButtonStyle>
				<MoreSelection
					:options="options"
					size="md"
					v-permission
				></MoreSelection>
			</div>
		</template>
		<MyCard
			no-content-gap
			square
			flat
			animated
			class="q-mx-lg q-mt-lg"
			v-if="isStudio"
		>
			<template #title>
				<div class="row items-center no-wrap" style="cursor: pointer">
					<MyCardHeader
						class="no-wrap ellipsis"
						:isStudio="isStudio"
						:title="$route.params.pods_name"
						:img="selectedNodes?.img"
					/>
					<div class="q-ml-sm" v-permission>
						<q-icon
							v-for="item in options4"
							:key="item.value"
							:name="item.icon"
							size="16px"
							color="blue-default"
							@click="item.onClick"
						>
							<q-tooltip>
								<div style="white-space: nowrap">
									{{ item.label }}
								</div>
							</q-tooltip>
						</q-icon>
					</div>
				</div>
			</template>
			<template #extra>
				<div class="full-width row justify-end">
					<q-resize-observer @resize="onResize" />

					<transition enter-active-class="animated fadeInRight">
						<div
							v-show="!isMini"
							v-permission
							class="row items-center q-pr-sm"
							style="gap: 8px"
						>
							<MyButton
								v-for="item in options2"
								:key="item.value"
								:icon="item.icon"
								:label="item.label"
								@click="item.onClick"
							>
							</MyButton>
						</div>
					</transition>
					<transition enter-active-class="animated fadeIn slower">
						<MoreSelection
							:options="options"
							size="md"
							v-permission
							v-show="isMini"
						></MoreSelection>
					</transition>
				</div>
			</template>
			<div class="q-pt-sm">
				<DetailData ref="detaiDatalRef"></DetailData>
			</div>
		</MyCard>
		<EnvironmentVariables ref="envRef"> </EnvironmentVariables>
		<Yaml
			ref="yamlRef"
			:name="t('EDIT_YAML')"
			name-key="pods_name"
			:readonly="isStudio2"
			@change="fetchList"
		></Yaml>
		<Dialog2 :title="$t('DETAILS')" ref="dialog2Ref" full-width full-height>
			<PodsMonitoring></PodsMonitoring>
		</Dialog2>
		<DeleteDialog
			:title="`${$t('STOP_POD')} ${$route.params.pods_name}`"
			:desc="$t('POD')"
			:name="$route.params.pods_name"
			ref="deleteDialogRef"
			@submit="confirmHandler"
		></DeleteDialog>
		<DeleteDialog
			:title="`${$t('RESTART_POD')} ${$route.params.pods_name}`"
			:desc="$t('POD')"
			:name="$route.params.pods_name"
			ref="deleteDialogRef2"
			@submit="confirmHandler2"
		></DeleteDialog>
	</component>
</template>

<script setup lang="ts">
import { ref, withDefaults, computed, watch, readonly } from 'vue';
import { t } from '@apps/control-hub/src/boot/i18n';
import EnvironmentVariables from '@apps/control-hub/src/containers/EnvironmentVariables.vue';
import Yaml from '@apps/control-hub/src/pages/NamespacePods/Yaml.vue';
import MyContentPage from '@apps/control-hub/src/components/MyContentPage.vue';
import MyContentPage2 from '@apps/control-hub/src/components/MyContentPage2.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import Dialog2 from '@apps/control-panel-common/src/components/Dialog/Dialog.vue';
import PodsMonitoring from './PodsMonitoring/PodsMonitoring.vue';
import MoreSelection from '@apps/control-panel-common/src/components/MoreSelection.vue';
import { replicaStatus } from '@apps/control-hub/src/containers/Replica';
import { restartPods } from '@apps/control-hub/src/network';
import { useRoute } from 'vue-router';
import { usePodList } from '@apps/control-panel-common/src/stores/podList';
import DeleteDialog from '@apps/control-panel-common/src/components/DeleteDialog.vue';
import DetailData from '@apps/control-hub/src/pages/ApplicationSpaces/Workloads/DetailData.vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import { useIsStudio, useIsStudio2 } from '../../../stores/hook';
import MyButton from '@apps/control-hub/src/components/MyButton.vue';
import { selectedNodes } from '../treeStore';
import MyCardHeader from '@apps/control-hub/src/components/MyCardHeader.vue';
const isStudio = useIsStudio();
const isStudio2 = useIsStudio2();
const podList = usePodList();
interface Props {
	drawer: boolean;
}
const deleteDialogRef = ref();
const deleteDialogRef2 = ref();
const route = useRoute();
const isMini = ref(false);

const options = computed(() =>
	optionsChildren3.value.concat(optionsChildren2.value)
);
const options2 = computed(() =>
	optionsChildren3.value.concat(optionsChildren2.value)
);
const options3 = computed(() => optionsChildren3.value);
const options4 = computed(() => optionsChildren1.value);

const optionsChildren1 = computed(() => [
	{
		label: isStudio2.value ? t('VIEW_YAML') : t('EDIT_YAML'),
		value: 'edit',
		icon: isStudio2.value ? 'sym_r_preview' : 'sym_r_edit_square',
		onClick: () => {
			yamlRef.value.show();
		}
	}
]);

const optionsChildren2 = computed(() => [
	{
		label: t(replicaStatus.value.desire ? 'STOP_POD' : 'START_POD'),
		value: 'edit',
		icon: replicaStatus.value.desire
			? 'sym_r_stop_circle'
			: 'sym_r_play_circle',
		onClick: () => {
			console.log('STOP_POD');
			if (replicaStatus.value.desire) {
				deleteDialogRef.value && deleteDialogRef.value.show();
			} else {
				envRef.value.replicaChange(1);
			}
		}
	},
	{
		label: t('RESTART_POD'),
		value: 'edit',
		icon: 'sym_r_restart_alt',
		disable: !replicaStatus.value.desire,
		onClick: () => {
			deleteDialogRef2.value && deleteDialogRef2.value.show();
		}
	}
]);

const optionsChildren3 = computed(() => [
	{
		label: t('DETAILS'),
		value: 'detail',
		icon: 'sym_r_insert_chart',
		onClick: clickHandler
	}
]);

const confirmHandler = () => {
	envRef.value.replicaChange(0);
	deleteDialogRef.value && deleteDialogRef.value.hide();
};

const confirmHandler2 = () => {
	Promise.all(
		podList.data.map((item) => restartPods(item.namespace, item.name))
	).then(() => {
		deleteDialogRef2.value && deleteDialogRef2.value.hide();
	});
};

withDefaults(defineProps<Props>(), {
	drawer: false
});
const dialog2Ref = ref();
const yamlRef = ref();
const detaiDatalRef = ref();
const envRef = ref();

const clickHandler = () => {
	dialog2Ref.value.show();
};

const clickHandler2 = () => {
	yamlRef.value.show();
};

const fetchList = () => {
	detaiDatalRef.value && detaiDatalRef.value.update();
};

const onResize = (size: { width: number; height: number }) => {
	// When the width of the "extra" cannot accommodate all the buttons, they will fold.
	const buttonWidth = 276;
	isMini.value = size.width < buttonWidth;
};

watch(
	() => route.params.pods_name,
	() => {
		isMini.value = true;
	}
);
</script>
