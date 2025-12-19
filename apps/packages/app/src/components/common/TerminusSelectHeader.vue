<template>
	<div
		class="row justify-between"
		style="width: 100%; height: 56px; padding: 0 20px"
	>
		<div class="row items-center">
			<q-btn
				class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
				icon="sym_r_chevron_left"
				text-color="ink-2"
				@click="handleClose"
			>
				<q-tooltip>{{ t('return') }}</q-tooltip>
			</q-btn>

			<q-btn
				class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
				icon="sym_r_done_all"
				text-color="ink-2"
				@click="handleSelectAll"
			>
				<q-tooltip>{{ selectAll ? t('cancel') : t('select_all') }}</q-tooltip>
			</q-btn>
		</div>
		<div class="row items-center text-h7">
			{{
				t('vault_t.count_items_selected', {
					count: selectIds.length
				})
			}}
		</div>
		<div class="row items-center">
			<span class="q-mr-xs cursor-pointer checkOperate" v-if="showMove">
				<q-btn
					v-if="selectIds.length"
					class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
					icon="sym_r_low_priority"
					text-color="ink-2"
					@click="handleMove"
				>
					<q-tooltip>{{ t('move_to') }}</q-tooltip>
				</q-btn>

				<q-btn
					v-else
					class="text-grey-6 btn-size-sm btn-no-text btn-no-border"
					icon="sym_r_low_priority"
					text-color="ink-2"
					disabled
				>
					<q-tooltip>{{ t('move_to') }}</q-tooltip>
				</q-btn>
			</span>

			<span class="cursor-pointer checkOperate">
				<q-btn
					v-if="selectIds.length"
					class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
					icon="sym_r_delete"
					text-color="ink-2"
					@click="handleRemove"
				>
					<q-tooltip>{{ t('delete') }}</q-tooltip>
				</q-btn>
				<q-btn
					v-else
					class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
					icon="sym_r_delete"
					text-color="ink-2"
					disabled
				>
					<q-tooltip>{{ t('delete') }}</q-tooltip>
				</q-btn>
			</span>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

defineProps({
	selectIds: {
		type: Array,
		required: true,
		default: null
	},
	showMove: {
		type: Boolean,
		required: false,
		default: false
	}
});

const { t } = useI18n();
const selectAll = ref(false);

const emits = defineEmits([
	'handleClose',
	'handleSelectAll',
	'handleMove',
	'handleRemove'
]);

const handleClose = () => {
	emits('handleClose');
};

const handleSelectAll = () => {
	emits('handleSelectAll');
};

const handleMove = () => {
	emits('handleMove');
};

const handleRemove = () => {
	emits('handleRemove');
};
</script>

<style scoped lang="scss">
.user-header_bg {
	width: 100%;

	.user-header {
		width: 100%;
		height: 56px;
		padding-left: 20px;
		padding-right: 20px;
		position: relative;

		&__title {
			display: flex;
			align-items: center;
		}
	}
}
</style>
