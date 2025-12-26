<template>
	<div class="relative-position" style="overflow: hidden">
		<MenuHeader
			style="position: absolute; top: 0"
			class="text-left"
		></MenuHeader>
		<MyContentPage>
			<template #extra>
				<QSectionStyle2>
					<q-select
						v-model="terminalStore.current_node"
						:options="terminalStore.node_options"
						dense
						:disable="terminalStore.data.length <= 1"
						outlined
						style="width: 160px"
						@update:model-value="fetchData"
					/>
				</QSectionStyle2>
			</template>
		</MyContentPage>
	</div>
	<q-dialog v-model="confirm" persistent>
		<q-card>
			<div class="q-pa-lg" style="width: 520px">
				<div class="row justify-between">
					<span class="text-h6">{{ $t('SELECT_NODE') }}</span>
					<q-icon name="close" size="20px" v-close-popup />
				</div>
				<q-option-group
					class="terminal-group-wrapper"
					v-model="group"
					:options="options2"
					color="primary"
					type="radio"
				>
					<template #label="scope">
						<span class="text-subtitle1"> {{ scope.value }} </span>
					</template>
				</q-option-group>
				<q-card-actions align="right" class="q-pa-none q-pt-xl">
					<q-btn
						outline
						unelevated
						color="background-1"
						no-caps
						v-close-popup
						padding="12px lg"
					>
						<span class="text-body3 text-ink-2">{{ $t('CANCEL') }}</span>
					</q-btn>
					<q-btn
						unelevated
						no-caps
						:label="$t('CONFIRM')"
						color="primary"
						v-close-popup
						padding="sm lg"
						@click="confirmHandler"
					/>
				</q-card-actions>
			</div>
		</q-card>
	</q-dialog>
</template>

<script setup lang="ts">
import { computed, onBeforeMount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import QSectionStyle2 from '@apps/control-panel-common/src/components/QSectionStyle2.vue';
import { componentName } from '@apps/control-hub/src/router/const';
import { useTerminalStore } from '@apps/control-hub/src/stores/TerminalStore';
import { ref } from 'vue';

import MenuHeader from '@apps/control-hub/src/layouts/MenuHeader.vue';
import MyContentPage from '@apps/control-hub/src/components/MyContentPage.vue';
import { onBeforeUnmount } from 'vue';

const group = ref();

const options2 = computed(() =>
	terminalStore.node_options.map((item) => ({ label: item, value: item }))
);

const route = useRoute();
const terminalStore = useTerminalStore();
const dialog2Ref = ref();
const confirm = ref();
const val = ref();
const router = useRouter();
const fetchData = (value) => {
	terminalStore.setCurrentNode(value);
	router.replace({
		name: componentName.TERMINAL_APP,
		params: {
			node: value
		}
	});
};

const confirmHandler = () => {
	updateCurrentNodeAndRouteTo(group.value);
};

const updateCurrentNodeAndRouteTo = (value: string) => {
	router.replace({
		name: componentName.TERMINAL_APP,
		params: {
			node: value
		}
	});
	terminalStore.setCurrentNode(value);
};

onBeforeMount(() => {
	const target = terminalStore.data.find(
		(item) => item.metadata.name === route.params.node
	);
	const isNotAvailableNode = !!route.params.node && !target;

	if (terminalStore.data.length === 1) {
		updateCurrentNodeAndRouteTo(terminalStore.data[0].metadata.name);
	} else if (route.name === componentName.TERMINAL_APP) {
		if (
			(!terminalStore.current_node && terminalStore.data.length > 1) ||
			isNotAvailableNode
		) {
			confirm.value = true;
		}
	}
});

onBeforeUnmount(() => {
	confirm.value = false;
});

const onClick = () => {
	dialog2Ref.value.show();
};
</script>
<style lang="scss" scoped>
.node-selection-container {
	display: none;
	position: absolute;
	right: 0;
	top: 0;
	z-index: 99999;
	background: #fff;
}
.terminal-group-wrapper {
	width: 100%;
	margin-left: 0px;
	margin-top: 24px;
	::v-deep(> div) {
		margin: 0;
		margin-right: -8px;
	}

	::v-deep(.q-radio) {
		display: flex;
		flex-direction: row-reverse;
		justify-content: space-between;
	}
}
</style>
