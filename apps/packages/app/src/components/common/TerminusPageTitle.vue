<template>
	<div
		class="page-title"
		:class="center ? ' column justify-center items-center' : ''"
	>
		<div
			class="text-ink-1 text-h5"
			:style="center ? 'text-align : center' : ''"
			v-show="label"
		>
			<q-img v-if="imgSrc" :src="imgSrc" width="32px" height="32px" />
			{{ label }}
		</div>
		<div
			class="login-sub-title page-title__desc text-ink-2"
			:style="center ? 'text-align : center' : 'color : $prompt-message'"
			v-show="desc"
		>
			{{ desc }}
		</div>
		<div
			class="login-sub-title page-title__desc"
			v-show="descHtml"
			v-html="descHtml"
		/>
	</div>
</template>

<script lang="ts" setup>
defineProps({
	label: {
		type: String,
		default: '',
		required: false
	},
	center: {
		type: Boolean,
		default: false
	},
	desc: {
		type: String,
		default: '',
		required: false
	},
	descHtml: {
		type: String,
		default: '',
		required: false
	},
	imgSrc: {
		type: String,
		default: '',
		required: false
	}
});
</script>

<style scoped lang="scss">
.page-title {
	margin-top: 32px;
	margin-bottom: 12px;

	&__desc {
		margin-top: 8px;
	}
}
</style>
