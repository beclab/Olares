<template>
	<div
		class="social-btn-background-normal row justify-center items-center"
		@click="onButtonClick"
	>
		<q-icon
			size="24px"
			color="ink-2"
			:name="expand ? 'sym_r_expand_less' : 'sym_r_expand_more'"
		/>
	</div>
</template>

<script lang="ts" setup>
const props = defineProps({
	expand: {
		type: Boolean,
		required: true
	}
});

const emit = defineEmits(['update:expand']);

const onButtonClick = () => {
	emit('update:expand', !props.expand);
};
</script>

<style scoped lang="scss">
.social-btn-background-normal {
	width: 40px;
	height: 40px;
	border-radius: 12px;
	border: 0.5px solid var(--border, #ebebeb);
	background: white;
}
</style>
