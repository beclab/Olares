<template>
	<q-dialog ref="dialogRef">
		<q-card class="add-block-dialog">
			<div
				class="add-block-header row justify-between items-center full-width q-pa-lg"
			>
				<div class="text-ink-1 text-subtitle1">{{ t('blocks.add_block') }}</div>
				<q-icon
					size="20px"
					color="ink-3"
					name="sym_r_close"
					class="cursor-pointer"
					@click="onDialogCancel"
				/>
			</div>
			<div class="add-block-grid column full-width q-px-lg q-pb-lg">
				<template v-for="item in blockArray" :key="item">
					<add-block-item :type="item" @on-item-click="onBlockAdd" />
				</template>
			</div>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import AddBlockItem from '@apps/profile/src/components/profile/block/AddBlockItem.vue';
import { useDialogPluginComponent } from 'quasar';
import { BLOCK_TYPE } from '@apps/profile/src/types/User';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

const blockArray = [BLOCK_TYPE.TEXT, BLOCK_TYPE.LINK, BLOCK_TYPE.IMAGE];
const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();
const { t } = useI18n();
const router = useRouter();

const onBlockAdd = (id: string) => {
	router.push({
		name: 'blockEditor',
		params: {
			id
		}
	});
	onDialogOK();
};
</script>
<style lang="scss">
.add-block-dialog {
	background: $background-1;
	width: 400px;
	height: auto;
	position: relative;
	border-radius: 12px !important;

	.add-block-header {
		height: 64px;
	}

	.add-block-grid {
		display: grid;
		align-items: center;
		justify-items: center;
		justify-content: center;
		grid-row-gap: 12px;
		grid-template-columns: repeat(1, minmax(0, 1fr));
	}
}
</style>
