<template>
	<MyQDialog2
		v-bind="$attrs"
		v-model="visible"
		@before-hide="beforeHide"
		:title="title"
		@onSubmit="confirmHandler"
		:loading="loading"
		:ok-disabled="disabled"
	>
		<div class="q-gutter-y-md">
			<div class="text-ink-2">
				<div
					v-html="
						$t('REMOVE_CONFIRM_DSC', { name: lowerCase(desc), resource: name })
					"
				></div>
			</div>
			<q-input
				v-model="deleteId"
				type="text"
				:placeholder="name"
				outlined
				dense
			/>
		</div>
	</MyQDialog2>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import MyQDialog2 from './MyQDialog2.vue';
import { lowerCase } from 'lodash';

interface Props {
	name: string;
	desc: string;
	loading: boolean;
	title: string;
}

const props = withDefaults(defineProps<Props>(), {});
const emits = defineEmits(['submit']);

const visible = ref(false);
const deleteId = ref();

const disabled = computed(
	() => !props.name || !deleteId.value || props.name !== deleteId.value
);

const dialogShow = () => {
	visible.value = true;
};

const dialogHide = () => {
	visible.value = false;
};

const confirmHandler = () => {
	emits('submit');
};

const beforeHide = () => {
	deleteId.value = undefined;
};

defineExpose({ show: dialogShow, hide: dialogHide });
</script>

<style></style>
