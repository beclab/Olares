<template>
	<div class="block-button-root column justify-center" :title="descRef">
		<div class="block-button-background row justify-center items-center">
			<q-icon size="24px" color="ink-1" :name="imageRef" />
		</div>
		<span class="block-button-text">{{ titleRef }}</span>
	</div>
</template>

<script setup lang="ts">
import { onMounted, PropType, ref } from 'vue';
import { BLOCK_TYPE } from '@apps/profile/src/types/User';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	type: Object as PropType<BLOCK_TYPE>
});

const { t } = useI18n();
const imageRef = ref();
const titleRef = ref();
const descRef = ref();

onMounted(() => {
	switch (props.type) {
		case BLOCK_TYPE.link:
			imageRef.value = 'sym_r_captive_portal';
			titleRef.value = t('blocks.link');
			descRef.value = t('blocks.add_link');
			break;
		case BLOCK_TYPE.text:
			imageRef.value = 'sym_r_article';
			titleRef.value = t('blocks.text');
			descRef.value = t('blocks.add_text');
			break;
		case BLOCK_TYPE.image:
			imageRef.value = 'sym_r_imagesmode';
			titleRef.value = t('blocks.image');
			descRef.value = t('blocks.add_image');
			break;
	}
});
</script>

<style scoped lang="scss">
.block-button-root {
	height: 60px;
	width: 40px;
	margin-right: 12px;

	.block-button-background {
		width: 40px;
		height: 40px;
		border-radius: 12px;
		border: 0.5px solid $separator-2;
		background: $background-1;

		&:hover {
			color: $ink-on-brand;
			background: var(
				--profile,
				linear-gradient(90deg, #8ce3ff -2.75%, #7fff93 102.75%)
			);
		}

		&:active {
			color: $ink-on-brand;
			background: var(
				--profile,
				linear-gradient(90deg, #8ce3ff -2.75%, #7fff93 102.75%)
			);
			filter: brightness(0.9);
		}
	}

	.block-button-text {
		width: 100%;
		height: 16px;
		margin-top: 4px;
		font-family: Roboto;
		font-size: 12px;
		font-weight: 400;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: center;
		overflow: hidden;
		color: $ink-2;
		text-align: center;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}
</style>
