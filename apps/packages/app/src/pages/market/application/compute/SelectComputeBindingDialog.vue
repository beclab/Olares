<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('compute.launch_title', { title: appTitle })"
		size="medium"
		:ok="okLabel"
		:cancel="t('cancel')"
		:ok-disabled="!requirementMet"
		@onSubmit="submitBinding"
	>
		<template #footerMore>
			<div
				class="manage-accelerator-link text-body2 text-blue-default cursor-pointer"
				@click="goManageAccelerator"
			>
				{{ t('compute.manage_accelerator') }}
			</div>
		</template>

		<div class="binding-dialog column no-wrap">
			<div class="binding-dialog__header">
				<compute-validation-notice
					v-if="topValidationMessage"
					class="q-mb-md"
					:message="topValidationMessage"
				/>

				<!-- app summary -->
				<div class="binding-summary row items-center justify-between">
					<div class="row items-center no-wrap" style="gap: 12px">
						<q-img
							v-if="appIcon"
							no-spinner
							:src="appIcon"
							class="binding-summary__icon"
						/>
						<div class="column">
							<div class="text-body1 text-ink-1">{{ appTitle }}</div>
							<div class="row items-center q-mt-xs" style="gap: 4px">
								<compute-resource-badge
									v-for="tag in requirementTags"
									:key="tag"
									:label="tag"
									tone="neutral"
								/>
							</div>
						</div>
					</div>

					<div
						class="binding-summary__usage row items-center"
						:class="`binding-summary__usage--${usageState}`"
					>
						<div class="column items-center binding-summary__usage-col">
							<span class="text-body2">
								<span :class="usageValueClass">{{ allocatedText }}</span>
								<span class="text-ink-3 q-ml-xs">{{ requiredText }}</span>
							</span>
							<span class="text-caption" :class="usageLabelClass">
								{{ usageLabel }}
							</span>
						</div>
					</div>
				</div>
			</div>

			<div class="binding-dialog__scroll col">
				<!-- section header -->
				<div class="row items-center justify-between q-mb-sm">
					<span class="text-body2 text-ink-1">{{ sectionTitle }}</span>
					<q-icon
						name="sym_r_refresh"
						size="20px"
						class="text-ink-3 cursor-pointer"
						:class="{ 'rotate-anim': refreshing }"
						@click="refresh"
					>
						<q-tooltip class="text-body3">{{ t('compute.refresh') }}</q-tooltip>
					</q-icon>
				</div>

				<!-- refreshing placeholder -->
				<div
					v-if="refreshing"
					class="binding-loading column items-center justify-center"
				>
					<bt-loading :loading="true" size="48px" />
					<div
						class="text-body2 text-ink-2 q-mt-md text-center"
						style="white-space: pre-line"
					>
						{{ t('compute.getting_status') }}
					</div>
				</div>

				<!-- node list -->
				<div v-else class="column" style="gap: 12px">
					<compute-node-card
						v-for="node in nodeViews"
						:key="node.nodeName"
						:node-name="node.nodeName"
						:type-tags="node.typeTags"
						:status-label="nodeStatusBadgeLabel(node)"
						:status-tone="nodeStatusBadgeTone(node)"
						:status-tip="nodeStatusTip(node)"
						:validation-message="nodeValidationMessage(node)"
						:expanded="isNodeExpanded(node)"
						@toggle="toggleNode(node.nodeName)"
					>
						<compute-device-scope-list
							:node="node"
							:binding="binding"
							:parsed-validation="parsedValidation"
							:busy-app-name="busyAppName"
							@unbind="onUnbind"
						/>
					</compute-node-card>
				</div>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import ComputeNodeCard from 'src/components/compute/ComputeNodeCard.vue';
import ComputeResourceBadge from 'src/components/compute/ComputeResourceBadge.vue';
import ComputeDeviceScopeList from 'src/components/compute/ComputeDeviceScopeList.vue';
import ComputeValidationNotice from 'src/components/compute/ComputeValidationNotice.vue';
import UnbindAppDialog from 'src/components/compute/UnbindAppDialog.vue';
import BtLoading from 'src/components/base/BtLoading.vue';
import { useComputeBinding } from 'src/components/compute/useComputeBinding';
import { useComputeAvailability } from 'src/components/compute/useComputeAvailability';
import { useComputeSummary } from 'src/components/compute/useComputeSummary';
import { ComputeNodeView } from 'src/components/compute/types';
import {
	ComputeAvailabilityRefetch,
	ComputeAvailabilityResponse,
	ComputeBindingItem,
	ComputeBindingValidation,
	COMPUTE_VALIDATION_PREFIX,
	computePressureResourceLabelKey,
	formatComputePressureAmount,
	isNodeOperable,
	nodeStatusTone,
	pressuredDimensions
} from 'src/constant/compute';
import { getApplication } from 'src/application/base';
import {
	AcceleratorAppView,
	useAcceleratorStore
} from 'src/stores/settings/accelerator';
import { useUserStore } from 'src/stores/user';
import { openUrl } from 'src/utils/bex/tabs';
import { replaceOriginDomain } from 'src/utils/url2';
import { computed, PropType, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';

const props = defineProps({
	appTitle: {
		type: String,
		required: true
	},
	appIcon: {
		type: String,
		default: ''
	},
	availability: {
		type: Object as PropType<ComputeAvailabilityResponse>,
		required: true
	},
	validation: {
		type: Object as PropType<ComputeBindingValidation | undefined>,
		default: undefined
	},
	previousBinding: {
		type: Array as PropType<ComputeBindingItem[] | undefined>,
		default: undefined
	},
	// Dry-run re-fetch of the prompt (validate-only, no submit) used by refresh.
	refetch: {
		type: Function as PropType<ComputeAvailabilityRefetch>,
		default: undefined
	}
});

const { t } = useI18n();
const router = useRouter();
const $q = useQuasar();
const store = useAcceleratorStore();
const CustomRef = ref();

/* ---- data + selection + summary ---- */
const data = useComputeAvailability({
	availability: props.availability,
	validation: props.validation,
	refetch: props.refetch,
	resolveApp: store.resolveApp
});
const { nodeViews, refreshing, refresh, parsedValidation } = data;

const binding = useComputeBinding(data.availability);
binding.initSelection(props.previousBinding);

const summary = useComputeSummary({
	binding,
	nodeViews: data.nodeViews,
	requirement: data.requirement
});
const {
	requirementTags,
	requirementMet,
	usageState,
	allocatedText,
	requiredText,
	usageValueClass,
	usageLabelClass,
	usageLabel,
	sectionTitle,
	okLabel
} = summary;

/* ---- node card helpers ---- */
// A `not_enough` node that still exposes operable cards (multi-node apps whose
// total across nodes meets the requirement): keep the "insufficient" marker but
// in a softer warning tone, and let the cards stay selectable.
const insufficientButSelectable = (node: ComputeNodeView): boolean =>
	node.status === 'not_enough' && isNodeOperable(node.raw);

const nodeStatusBadgeLabel = (node: ComputeNodeView): string =>
	insufficientButSelectable(node)
		? t('compute.insufficient_vram')
		: t(`compute.node_status.${node.status}`);

const nodeStatusBadgeTone = (node: ComputeNodeView) =>
	insufficientButSelectable(node) ? 'warn' : nodeStatusTone(node.status);

const nodeStatusTip = (node: ComputeNodeView): string => {
	// Selectable insufficient node: the "cannot run" tip would contradict the
	// selectable cards, so drop it.
	if (node.status === 'available' || insufficientButSelectable(node)) return '';
	return t(`compute.node_status_tip.${node.status}`);
};

// Detailed node-pressure copy: a base sentence followed by one line per
// pressured resource dimension (Memory / CPU / Disk). Empty when the node has
// no pressure breakdown.
const nodePressureMessage = (nodeName: string): string => {
	const np = data.validation.value?.nodePressure;
	if (!np || np.nodeName !== nodeName) return '';
	const dims = pressuredDimensions(np);
	if (dims.length === 0) return '';
	const lines = dims.map((d) =>
		t('compute.node_pressure_dimension', {
			resource: t(computePressureResourceLabelKey(d.resource)),
			total: formatComputePressureAmount(d.resource, d.capacity),
			used: formatComputePressureAmount(d.resource, d.used),
			needed: formatComputePressureAmount(d.resource, d.required)
		})
	);
	return [t('compute.validation_code.node_pressure'), ...lines].join('\n');
};

// Node-scoped validation (e.g. node-pressure) surfaces on its own node card,
// without hijacking the static status badge.
const nodeValidationMessage = (node: ComputeNodeView): string => {
	const v = parsedValidation.value;
	if (v && !v.isDeviceScoped && v.i18nKey && v.targetId === node.nodeName) {
		if (v.prefix === COMPUTE_VALIDATION_PREFIX.NODE_PRESSURE) {
			return nodePressureMessage(node.nodeName) || t(v.i18nKey);
		}
		return t(v.i18nKey);
	}
	return '';
};

const topValidationMessage = computed((): string => {
	const v = parsedValidation.value;
	if (!v) return '';
	const attachedToNode =
		!v.isDeviceScoped &&
		!!v.targetId &&
		nodeViews.value.some((n) => n.nodeName === v.targetId);
	const attachedToDevice =
		v.isDeviceScoped &&
		!!v.targetId &&
		nodeViews.value.some((n) =>
			n.devices.some((d) => d.deviceId === v.targetId)
		);
	if (attachedToNode || attachedToDevice) return '';
	if (v.prefix === COMPUTE_VALIDATION_PREFIX.NODE_PRESSURE) {
		const detailed = nodePressureMessage(v.targetId);
		if (detailed) return detailed;
	}
	if (v.i18nKey) return t(v.i18nKey);
	return data.validation.value?.reason || '';
});

/* ---- expand / collapse ---- */
const expandedNodes = reactive<Record<string, boolean>>({});
// Default expansion follows operability: nodes with selectable cards (available
// or a multi-node `not_enough` whose cards remain operable) open; the rest stay
// collapsed.
const defaultExpanded = (node: ComputeNodeView) => isNodeOperable(node.raw);
const isNodeExpanded = (node: ComputeNodeView) =>
	expandedNodes[node.nodeName] ?? defaultExpanded(node);
const toggleNode = (nodeName: string) => {
	const view = nodeViews.value.find((n) => n.nodeName === nodeName);
	const current =
		expandedNodes[nodeName] ?? (view ? defaultExpanded(view) : true);
	expandedNodes[nodeName] = !current;
};

/* ---- unbind ---- */
const busyAppName = ref('');
const onUnbind = (app: AcceleratorAppView, cardName: string) => {
	// Multi-card / cross-node apps release every linked card at once.
	const scope = data.availability.value.scope;
	const releasesAll =
		app.multiCard ||
		scope === 'single-node-cards' ||
		scope === 'cross-node-cards';
	$q.dialog({
		component: UnbindAppDialog,
		componentProps: {
			title: t('compute.unbind_title', { app: app.title }),
			messageKeypath: releasesAll
				? 'compute.unbind_desc_all'
				: 'compute.unbind_desc',
			appTitle: app.title,
			cardName,
			okText: t('compute.unbind')
		}
	}).onOk(async () => {
		busyAppName.value = app.appName;
		try {
			await store.releaseApp(app.appName);
			// Freeing a card invalidates the prompt — pull a fresh one from validate.
			await refresh();
		} finally {
			busyAppName.value = '';
		}
	});
};

/* ---- actions ---- */
const goManageAccelerator = () => {
	if (process.env.APPLICATION === 'SETTINGS') {
		CustomRef.value?.onDialogCancel();
		router.push('/gpu').catch(() => undefined);
		return;
	}
	const userStore = useUserStore();
	const url = getApplication().platform?.isClient
		? userStore.getModuleSever('settings', 'https:', '/gpu')
		: `${replaceOriginDomain(location.origin, 'settings', true)}/gpu`;
	openUrl(url);
};

const submitBinding = () => {
	if (!requirementMet.value) return;
	const computeBinding = binding.buildBinding();
	if (computeBinding.length === 0) return;
	CustomRef.value?.onDialogOK({ computeBinding });
};
</script>

<style scoped lang="scss">
.manage-accelerator-link {
	position: absolute;
	left: 20px;
}

.binding-dialog {
	max-height: calc(100dvh * 0.75 - 160px);
	min-height: 0;

	&__header {
		flex-shrink: 0;
		padding-bottom: 16px;
	}

	&__scroll {
		flex: 1 1 auto;
		min-height: 0;
		overflow-y: auto;

		&::-webkit-scrollbar {
			width: 4px;
		}

		&::-webkit-scrollbar-thumb {
			border-radius: 4px;
			background: $ink-3;
		}
	}
}

.binding-summary {
	&__icon {
		width: 40px;
		height: 40px;
		border-radius: 10px;
	}

	&__usage {
		padding: 8px 16px;
		border-radius: 8px;
		background: $background-3;
		gap: 16px;

		&--none {
			background: $background-3;
			border: 1px solid $separator;
		}

		&--sufficient {
			background: $green-soft;
			border: 1px solid $separator;
		}

		&--insufficient {
			background: $orange-alpha;
			border: 1px solid $separator;
		}
	}

	&__usage-col {
		min-width: 64px;
	}

	&__divider {
		width: 1px;
		height: 28px;
		background: $separator;
	}
}

.binding-loading {
	min-height: 240px;
}

.rotate-anim {
	animation: binding-refresh-spin 0.8s linear infinite;
}

@keyframes binding-refresh-spin {
	from {
		transform: rotate(0deg);
	}
	to {
		transform: rotate(360deg);
	}
}
</style>
