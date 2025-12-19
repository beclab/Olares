<template>
	<div class="row items-center" style="position: relative">
		<q-img
			no-spinner
			width="32px"
			height="32px"
			:src="icon"
			style="border-radius: 6px; margin-right: 4px"
		/>
		<div>
			{{ app }}
		</div>
		<ApplicationStatus :status="state" class="q-ml-lg" :size="32" />
	</div>
</template>

<script setup lang="ts">
import ApplicationStatus from 'src/components/settings/application/ApplicationStatus.vue';
const props = defineProps({
	icon: {
		type: String,
		required: true
	},
	state: {
		type: String,
		required: true
	},
	app: {
		type: String,
		required: true
	}
});
</script>

<style scoped lang="scss"></style>
