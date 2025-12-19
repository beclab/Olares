<template>
	<q-dialog class="d-creatVault text-center" ref="dialogRef">
		<q-card class="q-dialog-plugin">
			<v-c-card-container :item="item" :show-details="true" />
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { useDialogPluginComponent } from 'quasar';
import { VCCardItem } from '../../../utils/vc';
import VCCardContainer from './VCCardContainer.vue';

defineProps({
	item: {
		type: Object as PropType<VCCardItem>,
		required: true
	}
});

const { dialogRef } = useDialogPluginComponent();
</script>

<style scoped lang="scss">
.d-creatVault {
	.q-dialog-plugin {
		width: 500px;
		border-radius: 20px;
	}
}
</style>
