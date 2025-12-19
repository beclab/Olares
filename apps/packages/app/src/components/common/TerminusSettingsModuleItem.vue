<template>
	<div class="settings-module-root">
		<div
			class="settings-module-root__header row items-center justify-between"
			:style="
				showContent
					? ''
					: 'border-bottom-right-radius: 8px;border-bottom-left-radius: 8px;'
			"
		>
			<div class="settings-module-root___header__title text-subtitle2 q-ml-md">
				{{ title }}
			</div>
			<div class="side">
				<slot name="side"></slot>
			</div>
		</div>
		<q-separator scolor="grey-2" v-if="showContent" />
		<div class="settings-module-root__content text-body2">
			<slot name="content"></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
defineProps({
	title: {
		required: false,
		type: String,
		default: ''
	},
	showContent: {
		required: false,
		type: Boolean,
		default: true
	}
});
</script>

<style scoped lang="scss">
.settings-module-root {
	width: 100%;
	border: 1px solid $separator;
	background-color: $background-1;
	border-radius: 8px;

	&__header {
		height: 36px;
		background-color: $background-3;
		border-top-right-radius: 8px;
		border-top-left-radius: 8px;

		&__title {
			color: $ink-1;
		}
	}

	&__content {
		width: 100%;
		line-height: 44px;
	}
}
</style>
