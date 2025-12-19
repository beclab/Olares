<template>
	<Dialog
		ref="dialogRef"
		@hide="onDialogHide"
		full-width
		persistent
		:ok="true"
		:cancel="true"
		:title="''"
		@onSubmit="onOKClick"
		@onCancel="onDialogCancel"
	>
		<q-card-section style="max-height: 50vh" class="scroll">
			<Title
				:title="t('EXTERNAL_ACCESS')"
				:desc="t('SERVICE_EXTERNAL_ACCESS_DESC')"
			></Title>
			<q-card flat bordered class="q-mt-xl">
				<q-card-section>
					<q-form class="q-gutter-md">
						<div class="row">
							<my-card flat :title="t('ACCESS_MODE')" class="col-6">
								<q-select
									v-model="type"
									:options="accessModes"
									outlined
									dense
									option-value="value"
									@update:model-value="handleTypeChange"
								/>
							</my-card>
							<my-card
								flat
								:title="t('LOAD_BALANCER_PROVIDER')"
								class="col-6"
								v-if="isLoadBalancerShow"
							>
								<q-select
									v-model="provider"
									:options="loadBalancerOption"
									:option-disable="
										(opt) =>
											Object(opt) === opt ? opt.disabled === true : true
									"
									outlined
									dense
									option-value="value"
									class="col-6"
								/>
							</my-card>
						</div>
						<my-card flat :title="t('ANNOTATION_PL')" v-if="isLoadBalancerShow">
							<template #extra>
								<q-btn
									:label="t('ADD')"
									no-caps
									denses
									outline
									color="primary"
									@click="addAnnotation"
								/>
							</template>
							<AnnotationsInput
								v-model="formTemplate.metadata.annotations"
								:options="providerOptions"
								:hiddenKeys="[/^kubesphere.io\//, 'openpitrix_runtime']"
								ref="AnnotationsInputRef"
							/>
						</my-card>
					</q-form>
				</q-card-section>
			</q-card>
		</q-card-section>
	</Dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useDialogPluginComponent } from 'quasar';
import { t } from '@apps/control-hub/src/boot/i18n';
import Title from './Title.vue';
import MyCard from '@apps/control-hub/src/components/MyCard.vue';
import { CLUSTER_PROVIDERS } from '@apps/control-panel-common/src/utils/constants';
import { useRoute, useRouter } from 'vue-router';
import { api } from '@apps/control-panel-common/src/boot/axios';
import { CLUSTER_PROVIDERS_ANNOTATIONS } from './contants';
import AnnotationsInput from '@apps/control-panel-common/src/containers/AnnotationsInput.vue';
import { get, omit, set, unset } from 'lodash-es';
import { useAppDetailStore } from '@apps/control-hub/src/stores/AppDetail';
import Dialog from '@apps/control-panel-common/src/components/Dialog/Dialog.vue';

interface Props {
	detail: any;
}
defineEmits([...useDialogPluginComponent.emits]);
const { dialogRef, onDialogHide, onDialogOK, onDialogCancel } =
	useDialogPluginComponent();

function onOKClick() {
	set(
		formTemplate.value,
		'metadata.resourceVersion',
		props.detail.resourceVersion
	);
	onDialogOK(formTemplate.value);
}

const appDetail: any = useAppDetailStore();
const route = useRoute();
const AnnotationsInputRef = ref();
const props = withDefaults(defineProps<Props>(), {});
const formTemplate = ref({ ...props.detail._originData });

const defaultProvider = CLUSTER_PROVIDERS.find(
	(item) => item.value === 'QingCloud Kubernetes Engine'
);

const accessModes = [
	{ label: t('NONE'), desc: t('ACCESS_NONE_TIP'), value: 'ClusterIP' },
	{
		label: 'NodePort',
		desc: t('ACCESS_NODEPORT_TIP'),
		value: 'NodePort'
	},
	{
		label: 'LoadBalancer',
		desc: t('ACCESS_LOADBALANCER_TIP'),
		value: 'LoadBalancer'
	}
];

const type = ref(
	accessModes.find((item) => item.value === props.detail._originData.spec.type)
);
const provider = ref(defaultProvider);
const openELBOpened = ref(true);
const loadBalancerOption = computed(() => [
	...CLUSTER_PROVIDERS,
	{
		label: 'OpenELB',
		value: 'OpenELB',
		icon: 'kubernetes',
		disabled: !openELBOpened.value,
		tip: ''
	}
]);

const isLoadBalancerShow = computed(() => type.value?.value === 'LoadBalancer');

const providerOptions = computed(() => {
	if (provider.value && provider.value.value) {
		// eslint-disable-next-line @typescript-eslint/ban-ts-comment
		//@ts-ignore
		return Object.keys(CLUSTER_PROVIDERS_ANNOTATIONS[provider.value.value]);
	} else {
		return [];
	}
});

const handleTypeChange = (value: {
	value: string;
	label: string;
	desc: string;
}) => {
	const type = value.value;
	formTemplate.value.spec.type = type;
	if (type !== 'LoadBalancer') {
		Object.keys(appDetail.data.config.loadBalancerDefaultAnnotations).forEach(
			(key) => {
				unset(formTemplate.value, `metadata.annotations["${key}"]`);
			}
		);

		providerOptions.value.forEach((key) => {
			unset(formTemplate.value, `metadata.annotations["${key}"]`);
		});

		if (type === 'ClusterIP') {
			const ports = get(formTemplate.value, 'spec.ports', []).map((port: any) =>
				omit(port, ['nodePort'])
			);
			set(formTemplate.value, 'spec.ports', ports);
		}
	}
};

const checkOpenELBStatus = async () => {
	const { cluster = 'default', namespace } = route.params as Record<
		string,
		string
	>;
	const isOpened = await isActive({
		clusters: [cluster],
		namespace
	});
	if (!isOpened) {
		openELBOpened.value = isOpened;
	}
};

const getPath = ({ cluster, namespace }: Record<string, string>) => {
	let path = '';
	if (cluster) {
		path += `/klusters/${cluster}`;
	}
	if (namespace) {
		path += `/namespaces/${namespace}`;
	}
	return path;
};

const isActive = async ({
	clusters
}: {
	clusters: string[];
	namespace?: string;
}) => {
	const req: any[] = [];
	clusters.forEach((cluster) => {
		req.push(
			api
				.get(
					`apis/network.kubesphere.io/v1alpha1${getPath({
						cluster
					})}/bgpconfs`
				)
				.catch((err) => false)
		);
	});

	const res = await Promise.all(req);
	return res.every((item) => !!item);
};

const addAnnotation = () => {
	AnnotationsInputRef.value && AnnotationsInputRef.value.addItem();
};

onMounted(() => {
	formTemplate.value = { ...props.detail._originData };
	checkOpenELBStatus();
});
</script>
