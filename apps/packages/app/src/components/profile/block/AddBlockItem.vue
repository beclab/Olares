<template>
	<div
		class="block-add row justify-between items-center q-pa-sm cursor-pointer"
		@click="onItemClick"
	>
		<div class="row justify-start items-center">
			<q-img :src="`profile/block/${type}.svg`" class="block-type q-ml-md" />
			<div class="text-body1 text-ink-1 q-ml-sm">{{ nickName }}</div>
		</div>
		<q-icon style="margin: 6px" color="ink-2" size="20px" name="sym_r_add" />
	</div>
</template>

<script setup lang="ts">
import { computed, PropType } from 'vue';
import {
	ALIGNMENT_TYPE,
	BLOCK_TYPE,
	ImageBlock,
	LinkBlock,
	SIZE_TYPE,
	RATIO_TYPE,
	TextBlock,
	generateUniqueId
} from '@apps/profile/src/types/User';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const userStore = useUserStore();
const emit = defineEmits(['onItemClick']);
const props = defineProps({
	type: {
		type: Object as PropType<BLOCK_TYPE>,
		required: true
	},
	selected: {
		type: Boolean,
		required: false
	}
});

const nickName = computed(() => {
	switch (props.type) {
		case BLOCK_TYPE.TEXT:
			return t('blocks.text');
		case BLOCK_TYPE.IMAGE:
			return t('blocks.image');
		case BLOCK_TYPE.LINK:
			return t('blocks.link');
		default:
			return '';
	}
});

const onItemClick = () => {
	if (userStore.user) {
		let block;
		const idList = userStore.user.block.data.filter((item) => item.id);
		switch (props.type) {
			case BLOCK_TYPE.TEXT:
				block = {
					id: generateUniqueId(idList),
					type: BLOCK_TYPE.TEXT,
					nickName: t('blocks.text'),
					enabled: true,
					title: '',
					description: '',
					textAlignment: ALIGNMENT_TYPE.CENTER,
					transparent: false
				} as TextBlock;
				break;
			case BLOCK_TYPE.LINK:
				block = {
					id: generateUniqueId(idList),
					type: BLOCK_TYPE.LINK,
					nickName: t('blocks.link'),
					enabled: true,
					url: '',
					title: '',
					subTitle: '',
					size: SIZE_TYPE.LARGER,
					textAlignment: ALIGNMENT_TYPE.CENTER,
					outline: true,
					shadow: false,
					icon: ''
				} as LinkBlock;
				break;
			case BLOCK_TYPE.IMAGE:
				block = {
					id: generateUniqueId(idList),
					type: BLOCK_TYPE.IMAGE,
					nickName: t('blocks.image'),
					enabled: true,
					title: '',
					description: '',
					img: '',
					ratio: RATIO_TYPE.RATIO_16_9
				} as ImageBlock;
				break;
		}
		console.log(block);
		if (block) {
			userStore.user.block.data.push(block);
			emit('onItemClick', block.id);
		}
	}
};
</script>

<style scoped lang="scss">
.block-add {
	width: 100%;
	height: 56px;
	border-radius: 8px;
	border: 1px solid $input-stroke;

	.block-type {
		width: 24px;
		height: 24px;
	}
}
</style>
