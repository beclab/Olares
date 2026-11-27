<template>
	<q-dialog
		ref="dialogRef"
		position="bottom"
		transition-show="jump-up"
		transition-hide="jump-down"
		transition-duration="500"
	>
		<terminus-dialog-display-content
			:title="t('files.operate')"
			:dialog-ref="dialogRef"
		>
			<template v-slot:content>
				<q-item
					clickable
					v-close-popup
					v-for="action in actions"
					:key="action.key"
					class="selection-action-item q-pb-lg"
					@click="selectAction(action)"
				>
					<div class="selection-action-div row justify-start items-center">
						<q-icon :name="action.icon" size="20px" />
						<div class="selection-action-text text-body3">
							{{ action.label }}
						</div>
					</div>
				</q-item>
				<div style="width: 100%; height: 20px"></div>
			</template>
		</terminus-dialog-display-content>
	</q-dialog>
</template>

<script setup lang="ts">
import { useDialogPluginComponent } from 'quasar';
import { PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import TerminusDialogDisplayContent from 'src/components/common/TerminusDialogDisplayContent.vue';

interface SelectionAction {
	key: string;
	icon: string;
	label: string;
	disabled?: boolean;
}

defineProps({
	actions: {
		type: Array as PropType<SelectionAction[]>,
		required: true,
		default: () => []
	}
});

const { t } = useI18n();
const { dialogRef, onDialogOK } = useDialogPluginComponent();

const selectAction = (action: SelectionAction) => {
	if (action.disabled) {
		return;
	}
	onDialogOK(action.key);
};
</script>

<style lang="scss" scoped>
.selection-action-list {
	padding: 0 8px;
}

.selection-action-item {
	height: 48px;
	min-height: 48px;
	margin: 0;
	padding-left: 0;
	padding-right: 0;

	.selection-action-div {
		width: 100%;
		height: 100%;
		padding: 8px;

		.selection-action-text {
			color: $ink-2;
			margin-left: 8px;
		}
	}
}

:deep(.card-container) {
	width: 100vw !important;
	max-width: 100vw !important;
	border-radius: 12px 12px 0 0;
}

:deep(.dialog-content) {
	margin: 0 !important;
}

:deep(.card-action) {
	display: none !important;
}
</style>
