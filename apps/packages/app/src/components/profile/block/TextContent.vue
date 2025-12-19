<template>
	<div class="link-root row">
		<edit-view
			height="96px"
			:placeholder="t('base.desc')"
			:model-value="title"
			@update:model-value="titleUpdate"
		/>
	</div>
</template>

<script setup lang="ts">
import EditView from '@apps/profile/src/components/profile/base/EditView.vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
const props = defineProps({
	title: String
});

console.log(props.title);

const emit = defineEmits(['update:title']);

const titleUpdate = (data: string) => {
	emit('update:title', data);
};
</script>

<style scoped lang="scss">
.link-root {
	width: 100%;
	height: 100%;
}
</style>
