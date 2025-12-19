<template>
	<div class="workflow-panel-root">
		<div class="workflow-panel-card bg-background-1">
			<div
				class="workflow-panel-tab-background row justify-between items-center"
			>
				<div class="workflow-panel-tab row justify-start items-center">
					<tab-item
						:title="t('recommendation.summary')"
						:index="1"
						:cur-index="index"
						@on-item-click="tab = 'summary'"
					/>
					<tab-item
						v-if="container"
						:title="t('recommendation.containers')"
						:index="2"
						:cur-index="index"
						@on-item-click="tab = 'containers'"
					/>
				</div>

				<div
					class="workflow-panel-close row justify-center items-center cursor-pointer"
					@click="emit('onClose')"
				>
					<q-icon size="16px" name="sym_r_clear" color="ink-3" />
				</div>
			</div>

			<q-separator class="bg-separator" />

			<q-tab-panels v-model="tab" animated>
				<q-tab-panel name="summary" class="workflow-tab-panel">
					<workflow-summary :nodeStatus="nodeStatus!" :workflow="workflow" />
				</q-tab-panel>

				<q-tab-panel
					v-if="container"
					name="containers"
					class="workflow-tab-panel"
				>
					<workflow-container :container="container" />
				</q-tab-panel>
			</q-tab-panels>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed, ref, watch, PropType } from 'vue';
import { WorkflowDetail, NodeStatus, Container } from 'src/stores/argo';
import WorkflowSummary from './WorkflowSummary.vue';
import WorkflowContainer from './WorkflowContainer.vue';
import TabItem from 'src/components/rss/TabItem.vue';
import { useI18n } from 'vue-i18n';

const emit = defineEmits(['onClose']);
const { t } = useI18n();

const props = defineProps({
	workflow: {
		type: Object as PropType<WorkflowDetail>,
		required: true
	},
	nodeStatus: {
		type: Object as PropType<NodeStatus>,
		required: false
	}
});

const container = ref<Container | undefined>(undefined);

const tab = ref('summary');

watch(
	() => props.nodeStatus,
	() => {
		container.value = props.workflow.spec.templates.find(
			(template) => template.name == props.nodeStatus?.templateName
		)?.container;
		if (!container.value) {
			tab.value = 'summary';
		}
	},
	{
		immediate: true
	}
);

const index = computed(() => {
	return tab.value === 'summary' ? 1 : 2;
});
</script>

<style lang="scss">
.workflow-panel-root {
	width: 100%;
	height: 100%;
	padding: 44px;

	.workflow-panel-card {
		box-shadow: 0 4px 10px 0 #0000001a;
		border-radius: 12px;
		overflow: hidden;

		.workflow-panel-tab-background {
			width: 100%;
			height: 56px;
			padding-left: 32px;

			.workflow-panel-tab {
				height: 55px;
			}

			.workflow-panel-close {
				height: 56px;
				width: 56px;
			}
		}

		.workflow-tab-panel {
			padding: 0;
		}
	}
}
</style>
