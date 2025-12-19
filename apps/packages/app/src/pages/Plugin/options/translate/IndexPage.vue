<template>
	<div>
		<div>sidPanel</div>
		<q-btn
			:color="stateA ? 'primary' : 'grey'"
			icon="check"
			:label="stateA ? '已启用' : '未启用'"
			@click="toggleState"
		/>
	</div>
</template>

<script setup lang="ts">
import { useStateService } from 'src/composables/mobile/useAccount';

const { stateA, updateStateA } = useStateService();

const toggleState = () => {
	updateStateA(!stateA.value);
};
</script>
