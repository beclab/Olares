<template>
	<span class="row items-center">
		<img
			:src="
				require(`../../assets/icon${
					$q.dark.isActive ? '/dark' : ''
				}/${src}.svg`)
			"
			:alt="src"
			:width="width"
			:height="height"
			:style="
				'margin-left:' +
				margin_left +
				'px;' +
				'margin-right:' +
				margin_right +
				'px;'
			"
			:class="disabled ? 'img-onlyRead' : 'img-normal'"
		/>
		<slot></slot>
	</span>
</template>

<script setup lang="ts">
defineProps({
	width: {
		type: Number,
		default: 20,
		required: false
	},
	height: {
		type: Number,
		default: 20,
		required: false
	},
	margin_left: {
		type: Number,
		default: 0,
		required: false
	},
	margin_right: {
		type: Number,
		default: 0,
		required: false
	},
	src: {
		type: String,
		default: '',
		required: true
	},
	disabled: {
		type: Boolean,
		default: false,
		required: false
	},
	isActive: {
		type: Boolean,
		default: false,
		required: false
	},
	color: {
		type: String,
		default: '$title',
		required: false
	}
});
</script>

<style lang="scss" scoped>
.icon {
	width: 1rem;
	height: 1rem;
	vertical-align: -0.15rem;
	fill: currentColor;
	overflow: hidden;
}

.img-onlyRead {
	pointer-events: none;
}
.img-normal {
	pointer-events: auto;
}
</style>
