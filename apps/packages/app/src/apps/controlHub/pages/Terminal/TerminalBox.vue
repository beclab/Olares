<template>
	<div
		class="relative-position terminal-box-container"
		:style="{
			height: terminalStore.terminal_show ? 'calc(100vh - 96px)' : '0px'
		}"
		v-if="terminalStore.terminal_init && terminalStore.current_node"
	>
		<div class="node-terminal-content">
			<Terminal
				class="node-terminal-wrapper"
				ref="terminalRef"
				flat
				:node="node"
				:visible="terminalStore.terminal_show"
				:key="terminalStore.current_node"
			></Terminal>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import Terminal from '@apps/control-panel-common/src/containers/TerminalNode.vue';
import { useTerminalStore } from '@apps/control-hub/src/stores/TerminalStore';
const terminalStore = useTerminalStore();
const terminalRef = ref();
const node = computed(() => terminalStore.current_node);
</script>

<style lang="scss" scoped>
.terminal-box-container {
	position: absolute;
	inset: 20px;
	top: 76px;
	background: #fff;
	border-radius: 12px;
}
.node-terminal-content {
	position: absolute;
	inset: 0;
}
.node-terminal-wrapper {
	width: 100%;
	height: 100%;
	border-radius: 12px;
	overflow: hidden;
}
</style>
