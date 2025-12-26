<template>
	<div class="buttons-slider-root row justify-between items-center">
		<q-btn
			v-for="option in options"
			:key="option.value"
			:label="option.label"
			:class="
				option.value === accountType
					? 'bg-white text-blue text-subtitle2 select-btn'
					: 'text-color-sub-detail text-subtitle2 normal-btn'
			"
			flat
			@click="changeOption(option)"
		>
		</q-btn>
	</div>
</template>

<script>
import { ref } from 'vue';
import { defineComponent } from 'vue';
export default defineComponent({
	props: {
		options: []
	},
	setup(props, ctx) {
		const accountType = ref(props.options[0].value);
		const percent = 100 / props.options.length + '%';
		window.document.body.style.setProperty('--item-width-percent', percent);
		const changeOption = (option) => {
			if (option.value === accountType.value) {
				return;
			}
			accountType.value = option.value;
			ctx.emit('on-change', option.value);
		};
		return { accountType, changeOption };
	}
});
</script>

<style lang="scss" scoped>
:root {
	--item-width-percent: 50%;
}

.buttons-slider-root {
	height: 57px;
	border-radius: 22px;
	max-width: 100%;

	.base-btn {
		height: 56px;
		line-height: 46px;
		margin-top: 1px;
		border-radius: 10px;
		text-transform: capitalize;
		margin-left: 2px;
		margin-right: 2px;
		min-width: calc(var(--item-width-percent) - 10px);
		border: 1px solid $separator;
		box-sizing: border-box;
	}

	.normal-btn {
		@extend .base-btn;
		background-color: transparent;
	}

	.select-btn {
		@extend .base-btn;
		border: 1px solid $blue;
	}
}
</style>
