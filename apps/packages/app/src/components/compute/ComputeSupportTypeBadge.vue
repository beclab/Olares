<template>
	<compute-resource-badge
		:label="t(acceleratorSupportTypeLabelKey(supportType))"
		:tone="acceleratorSupportTypeTone(supportType)"
	/>
</template>

<script setup lang="ts">
import ComputeResourceBadge from './ComputeResourceBadge.vue';
import {
	acceleratorSupportTypeLabelKey,
	acceleratorSupportTypeTone
} from 'src/constant/compute';
import { useI18n } from 'vue-i18n';

defineProps({
	/** Device support type, e.g. 'TimeSlice' | 'MemorySlice' | 'Exclusive'. */
	supportType: {
		type: String,
		required: true
	}
});

const { t } = useI18n();
</script>
