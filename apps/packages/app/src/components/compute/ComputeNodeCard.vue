<template>
	<div class="compute-node-card">
		<!-- header -->
		<div class="compute-node-card__header row items-center justify-between">
			<div class="row items-center no-wrap" style="gap: 8px">
				<div class="compute-node-card__avatar row items-center justify-center">
					<q-icon name="sym_r_memory" size="20px" class="text-ink-2" />
				</div>
				<span class="text-h6 text-ink-1">{{ nodeName }}</span>
				<compute-resource-badge
					v-for="tag in typeTags"
					:key="tag"
					:label="tag"
					tone="neutral"
				/>
			</div>

			<div class="row items-center no-wrap" style="gap: 12px">
				<compute-resource-badge :label="statusLabel" :tone="statusTone" dot>
					<q-tooltip v-if="statusTip" class="text-body3" max-width="240px">
						{{ statusTip }}
					</q-tooltip>
				</compute-resource-badge>
				<q-icon
					v-if="collapsible"
					:name="
						expanded ? 'sym_r_keyboard_arrow_up' : 'sym_r_keyboard_arrow_down'
					"
					size="20px"
					class="text-ink-3 cursor-pointer"
					@click="$emit('toggle')"
				/>
			</div>
		</div>

		<!-- node-level validation (kept separate from the static status badge) -->
		<div v-if="validationMessage" class="compute-node-card__validation">
			<compute-validation-notice :message="validationMessage" />
		</div>

		<!-- body -->
		<div v-if="!collapsible || expanded" class="compute-node-card__body">
			<slot />
		</div>
	</div>
</template>

<script setup lang="ts">
import ComputeResourceBadge from './ComputeResourceBadge.vue';
import ComputeValidationNotice from './ComputeValidationNotice.vue';
import { PropType } from 'vue';
import { ComputeTone } from 'src/constant/compute';

defineProps({
	nodeName: {
		type: String,
		required: true
	},
	typeTags: {
		type: Array as PropType<string[]>,
		default: () => []
	},
	statusLabel: {
		type: String,
		default: ''
	},
	statusTone: {
		type: String as PropType<ComputeTone>,
		default: 'neutral'
	},
	/** Optional hover hint shown on the status badge (e.g. why a node is unusable). */
	statusTip: {
		type: String,
		default: ''
	},
	/** Node-level validation message, rendered separately from the status badge. */
	validationMessage: {
		type: String,
		default: ''
	},
	collapsible: {
		type: Boolean,
		default: true
	},
	expanded: {
		type: Boolean,
		default: true
	}
});

defineEmits<{ (e: 'toggle'): void }>();
</script>

<style scoped lang="scss">
.compute-node-card {
	border: 1px solid $separator;
	border-radius: 12px;
	background: $background-1;
	transition: border-color 0.15s ease;

	&__header {
		padding: 16px;
		min-height: 56px;
	}

	&__validation {
		padding: 0 16px 16px;
	}

	&__avatar {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: $background-3;
	}

	&__body {
		padding: 0 16px 16px;
		border-top: 1px solid $separator;
		padding-top: 16px;
	}
}
</style>
