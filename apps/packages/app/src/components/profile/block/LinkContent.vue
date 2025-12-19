<template>
	<div class="link-root row">
		<bio-image :img-url="imgUrl" @update:img-url="imageUpdate" />
		<div class="column justify-between link-text">
			<edit-view
				:placeholder="t('blocks.link_title')"
				:model-value="title"
				@update:model-value="titleUpdate"
			/>
			<edit-view
				:placeholder="t('base.url')"
				:model-value="link"
				@update:model-value="linkUpdate"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import EditView from '@apps/profile/src/components/profile/base/EditView.vue';
import BioImage from '@apps/profile/src/components/profile/base/BioImage.vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const props = defineProps({
	imgUrl: String,
	title: String,
	link: String
});

console.log(props.imgUrl);
console.log(props.title);
console.log(props.link);

const emit = defineEmits(['update:imgUrl', 'update:title', 'update:link']);

const titleUpdate = (data: string) => {
	emit('update:title', data);
};

const linkUpdate = (data: string) => {
	emit('update:link', data);
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
		height: 72px;
	}
}
</style>
