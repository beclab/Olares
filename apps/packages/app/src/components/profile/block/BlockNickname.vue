<template>
	<div class="column">
		<div class="block-edit-board row justify-start items-center">
			<q-img :src="`profile/block/${type}.svg`" class="block-type q-ma-xs" />
			<div class="nick-edit-area row justify-start items-center">
				<q-input
					borderless
					type="text"
					class="username-input"
					@blur="isEdit = false"
					@focus="isEdit = true"
					input-class="platform-name"
					:model-value="text"
					@keyup.enter="onTextUpdate"
					@update:model-value="onTextUpdate"
				/>
			</div>
			<q-icon
				color="ink-2 q-ma-sm cursor-pointer"
				size="24px"
				name="sym_r_delete"
				@click="removeBlock"
			/>
		</div>
		<q-separator class="full-width q-mt-lg" color="separator" />
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { BtDialog, useColor } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
defineProps({
	type: {
		type: String,
		require: true
	},
	text: {
		type: String,
		require: true
	}
});

const isEdit = ref(false);
const route = useRoute();
const userStore = useUserStore();
const router = useRouter();
const emit = defineEmits(['update:text']);
const { t } = useI18n();
const blackBrand = useColor('ink-on-brand-black');

const onTextUpdate = (value: string) => {
	emit('update:text', value);
};

const removeBlock = () => {
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
				if (route.params.id && userStore.user) {
					const blocks = userStore.user.block.data.filter((item) => {
						return item.id !== route.params.id;
					});
					userStore.user.block.data = blocks;
					router.back();
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
.block-edit-board {
	width: 100%;
	height: 40px;
	background: $background-1;

	.block-type {
		width: 32px;
		height: 32px;
	}

	.nick-edit-area {
		height: 100%;
		border-radius: 8px;
		border: 1px solid $separator;
		width: calc(100% - 80px);

		.username-input {
			height: 100%;
			margin-left: 8px;
			width: calc(100% - 20px);
		}
	}
}
</style>
