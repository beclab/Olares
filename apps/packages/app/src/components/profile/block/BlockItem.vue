<template>
	<div
		class="block-item row justify-between items-center q-pa-sm q-mt-lg cursor-pointer"
		@click="edit"
	>
		<div class="row justify-start items-center">
			<q-icon color="ink-2" size="20px" name="sym_r_drag_indicator" />
			<q-img
				:src="`profile/block/${block.type}.svg`"
				class="block-type q-ml-md"
			/>
			<div
				class="text-body1 q-ml-md"
				:class="block.enabled ? 'text-ink-1' : 'text-ink-3'"
			>
				{{ block.nickName }}
			</div>
		</div>
		<div class="row justify-end items-center">
			<bt-switch
				size="xs"
				truthy-track-color="light-green-default"
				:model-value="block.enabled"
				@update:model-value="updateEnabled"
				color="light-green-default"
			/>
			<q-icon color="ink-2" size="20px" name="sym_r_more_horiz" @click.stop>
				<base-popup self="top right">
					<block-opt-item
						v-close-popup
						icon="sym_r_edit_square"
						:label="t('blocks.edit')"
						@click="edit"
					/>
					<block-opt-item
						v-close-popup
						icon="sym_r_content_copy"
						:label="t('blocks.copy')"
						@click="copy"
					/>
					<block-opt-item
						v-close-popup
						icon="sym_r_article_shortcut"
						:label="t('blocks.rename')"
						@click="rename"
					/>
					<block-opt-item
						v-close-popup
						icon="sym_r_delete"
						:label="t('blocks.delete')"
						@click="remove"
					/>
				</base-popup>
			</q-icon>
		</div>
	</div>
</template>

<script setup lang="ts">
import { PropType } from 'vue';
import { Block, generateUniqueId } from '@apps/profile/src/types/User';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import BasePopup from '@apps/profile/src/components/profile/base/BasePopup.vue';
import BlockOptItem from '@apps/profile/src/components/profile/block/BlockOptItem.vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { BtDialog, useColor } from '@bytetrade/ui';
import _ from 'lodash';

const props = defineProps({
	block: {
		type: Object as PropType<Block>,
		required: true
	}
});
const userStore = useUserStore();
const { t } = useI18n();
const router = useRouter();
const blackBrand = useColor('ink-on-brand-black');

const updateEnabled = (status: boolean) => {
	if (userStore.user) {
		const block = userStore.user.block.data.find(
			(item) => item.id === props.block.id
		);
		if (block) {
			block.enabled = status;
		}
	}
};

const edit = () => {
	if (userStore.user) {
		const block = userStore.user.block.data.find(
			(item) => item.id === props.block.id
		);
		if (block) {
			router.push({
				name: 'blockEditor',
				params: {
					id: block.id
				}
			});
		}
	}
};

const copy = () => {
	if (userStore.user) {
		const block = userStore.user.block.data.find(
			(item) => item.id === props.block.id
		);
		if (block) {
			const newBlock = _.cloneDeep(block);
			const idList = userStore.user.block.data.filter((item) => item.id);
			block.id = generateUniqueId(idList);
			userStore.user.block.data.push(newBlock);
		}
	}
};

const rename = () => {
	if (userStore.user) {
		const block = userStore.user.block.data.find(
			(item) => item.id === props.block.id
		);
		if (block) {
			BtDialog.show({
				title: t('blocks.rename_block'),
				okStyle: {
					background: 'linear-gradient(90deg, #8ce3ff -2.75%, #7fff93 102.75%)',
					color: blackBrand.color.value
				},
				cancel: true,
				prompt: {
					model: block.nickName,
					type: 'text', // optional
					name: t('blocks.block_name'),
					placeholder: t('blocks.enter_the_block_name')
				}
			})
				.then(async (res) => {
					if (res) {
						block.nickName = res as string;
					}
				})
				.catch((err) => {
					console.log('click ok', err);
				});
		}
	}
};

const remove = () => {
	BtDialog.show({
		title: t('blocks.are_you_sure_delete_this_block'),
		message: t('base.this_action_cannot_be_undone'),
		okStyle: {
			background: 'linear-gradient(90deg, #8ce3ff -2.75%, #7fff93 102.75%)',
			color: blackBrand.color.value
		},
		okText: t('base.confirm'),
		cancel: true
	})
		.then((res) => {
			if (res) {
				console.log('click ok');
				if (userStore.user) {
					const blocks = userStore.user.block.data.filter(
						(item) => item.id !== props.block.id
					);
					userStore.user.block.data = blocks;
				}
			} else {
				console.log('click cancel');
			}
		})
		.catch((err) => {
			console.log('click error', err);
		});
};
</script>

<style scoped lang="scss">
.block-item {
	width: 100%;
	height: auto;
	border-radius: 8px;
	border: 1px solid $input-stroke;

	.block-type {
		width: 24px;
		height: 24px;
	}
}
</style>
