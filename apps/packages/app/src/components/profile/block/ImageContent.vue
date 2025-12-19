<template>
	<div class="link-root row">
		<bio-image :img-url="imgUrl" @update:img-url="imageUpdate" />
		<edit-view
			class="link-text"
			height="72px"
			:placeholder="t('base.desc')"
			:model-value="title"
			@update:model-value="titleUpdate"
		/>
	</div>
</template>

<script setup lang="ts">
import EditView from '@apps/profile/src/components/profile/base/EditView.vue';
import BioImage from '@apps/profile/src/components/profile/base/BioImage.vue';
import { useI18n } from 'vue-i18n';

defineProps({
	imgUrl: String,
	title: String
});

const { t } = useI18n();
const emit = defineEmits(['update:imgUrl', 'update:title']);

const titleUpdate = (data: string) => {
	emit('update:title', data);
};

const imageUpdate = (data: string) => {
	emit('update:imgUrl', data);
};
</script>

<style scoped lang="scss">
.link-root {
	width: 100%;
	height: 100%;

	.link-text {
		margin-left: 12px;
		width: calc(100% - 84px);
	}
}
</style>
