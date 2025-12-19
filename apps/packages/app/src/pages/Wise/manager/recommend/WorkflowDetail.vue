<template>
	<q-splitter
		v-model="splitterModel"
		:style="{ height: '100vh', background: background }"
		separator-style="background : transparent"
	>
		<template v-slot:before>
			<div ref="flowLeft" style="height: 100%; width: 100%">
				<VueFlow
					style="height: 100%; width: 100%"
					v-model="elements"
					:snap-to-grid="true"
					:draggable="false"
					:apply-changes="false"
					:nodes-draggable="false"
					:zoom-on-scroll="false"
					:zoom-on-pinch="false"
					:zoom-on-double-click="false"
					:pan-on-drag="false"
				>
					<template #node-custom="customNodeProps">
						<CustomNode v-bind="customNodeProps" />
					</template>
				</VueFlow>
			</div>
		</template>
		<template v-slot:after>
			<WorkflowPanel
				v-if="workflow && nodeStatus"
				:nodeStatus="nodeStatus"
				:workflow="workflow"
				@on-close="onClose"
			/>
		</template>
	</q-splitter>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';
import { useArgoStore, WorkflowDetail } from 'src/stores/argo';
import { VueFlow, useVueFlow, MarkerType } from '@vue-flow/core';
import WorkflowPanel from './workflow/WorkflowPanel.vue';
// import type { Node } from '@vue-flow/core';
import CustomNode from './CustomNode.vue';
// import { CustomData, CustomEvents } from 'src/utils/nodeUtil';

// type CustomNode = Node<CustomData, CustomEvents, CustomNodeTypes>;
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';
import { useColor } from '@bytetrade/ui';

// type CustomNodeTypes = 'custom' | 'special';

const { onNodeClick } = useVueFlow();

const argoStore = useArgoStore();
const splitterModel = ref(100);
const workflow = ref<WorkflowDetail>();
const nodeStatus = ref();
let ns: any = {};
const flowLeft = ref();
const elements = ref<any>([
	// { id: '1', label: 'Node 1', position: { x: 250, y: 5 } },
	// { id: '2', label: 'Node 2', position: { x: 100, y: 100 } },
	// { id: 'e1-2', source: '1', target: '2', class: 'light' },
]);
const { color: background } = useColor('background-6');

window.addEventListener('resize', () => {
	onSplitterChange();
});

watch(
	() => splitterModel.value,
	() => {
		console.log('watch');
		nextTick(() => {
			onSplitterChange();
		});
	}
);

const getPositionX = () => {
	return flowLeft.value?.offsetWidth / 2 - 60;
};

const getPositionY = (id: number) => {
	return (id - 1) * 114 + 44;
};

const onSplitterChange = () => {
	elements.value.forEach((item: any) => {
		if (item && item.position && item.position.x) {
			item.position.x = getPositionX();
		}
	});
};

const onClose = () => {
	nodeStatus.value = null;
	splitterModel.value = 100;
	elements.value.forEach((item: any) => {
		item.data.selected = false;
	});
};

onNodeClick((e) => {
	console.log('node click');
	console.log(e);
	console.log(ns);

	if (!nodeStatus.value) {
		splitterModel.value = 50;
	}

	nodeStatus.value = ns[e.node.id];

	//update selected style
	let selected: any;
	elements.value.forEach((item: any) => {
		item.data.selected = false;
		if (item.id === e.node.id) {
			selected = item;
		}
	});

	if (selected) {
		console.log(selected);
		selected.data.selected = true;
	}
});

const updateWorkflowDetails = async () => {
	elements.value = [];
	workflow.value = await argoStore.get_workflow_detail(
		argoStore.namespace,
		argoStore.workflow_id
		// argoStore.metadata?.resourceVersion
	);
	if (!workflow.value) {
		console.error('get workflow detail errored');
		return;
	}
	console.log(workflow.value.status);

	const template = workflow.value.spec.templates.find((item) => {
		if (workflow.value) {
			return item.name == workflow.value.spec.entrypoint;
		}
	});
	if (!template) {
		console.error(
			'get template errored ' + workflow.value
				? workflow.value.spec.entrypoint
				: ''
		);
		return;
	}

	const ie = [];
	ns = {};
	let id = 1;
	for (let key in workflow.value.status.nodes) {
		if (workflow.value.status.nodes[key].type == 'Steps') {
			ie.push({
				id: '' + id,
				label: workflow.value.status.nodes[key].displayName,
				position: { x: getPositionX(), y: getPositionY(id) },
				// class: 'my-custom-node-class',
				data: {
					phase: workflow.value.status.nodes[key].phase,
					selected: false
				},
				type: 'custom'
				// You can pass an object containing CSSProperties or CSS variables
				// style: { backgroundColor: 'green', width: '100px', height: '100px',borderRadius : '10px' },
			});
			ns[id] = workflow.value.status.nodes[key];
			id++;
			break;
		}
	}

	if (template.steps) {
		for (const steps of template.steps) {
			for (const step of steps) {
				let node = null;
				for (let key in workflow.value.status.nodes) {
					if (
						workflow.value.status.nodes[key].displayName == step.name &&
						workflow.value.status.nodes[key].type == 'Pod'
					) {
						node = workflow.value.status.nodes[key];
					}
				}
				if (node) {
					ie.push({
						id: '' + id,
						label: node.displayName,
						position: { x: getPositionX(), y: getPositionY(id) },
						// class: 'my-custom-node-class',
						type: 'custom',
						data: {
							phase: node.phase,
							selected: false
						}

						// You can pass an object containing CSSProperties or CSS variables
						// style: { backgroundColor: 'green', width: '100px', height: '100px' },
					});
					ns[id] = node;
					id++;
				}
			}
		}
	}

	let edge = [];
	for (let i = 0; i < ie.length - 1; ++i) {
		console.log(ie[i]);
		edge.push({
			id: 'e' + ie[i].id + '-' + ie[i + 1].id,
			source: ie[i].id,
			target: ie[i + 1].id,
			type: 'default',
			markerEnd: MarkerType.Arrow
		});
	}
	for (let i = 0; i < edge.length; ++i) {
		ie.push(edge[i]);
	}

	elements.value = ie;
	console.log(elements.value);
};

watch(
	() => argoStore.workflow_id,
	() => {
		onClose();
		if (argoStore.workflow_id) {
			updateWorkflowDetails();
		}
	},
	{
		immediate: true
	}
);
</script>

<style></style>
