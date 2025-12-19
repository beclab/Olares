<template>
	<q-menu :auto-close="false" max-height="500px" style="border-radius: 12px">
		<div class="bio-popup-menu">
			<div class="bio-popup-title-bar row justify-between items-center">
				<div class="row justify-start">
					<q-icon
						class="bio-clickable q-mr-xs"
						v-if="back"
						size="20px"
						color="ink-2"
						@click="emit('onBackClick')"
						name="sym_r_chevron_left"
					/>
					<span class="bio-popup-title">{{ title }}</span>
				</div>
				<q-icon
					class="bio-clickable"
					v-close-popup
					size="20px"
					color="grey-8"
					name="sym_r_close"
				/>
			</div>
			<div class="bio-popup-label" v-if="label">{{ label }}</div>
			<div style="padding-bottom: 12px">
				<slot />
			</div>
		</div>
	</q-menu>
</template>

<script lang="ts" setup>
defineProps({
	title: {
		type: String,
		required: true
	},
	label: {
		type: String,
		required: false
	},
	back: {
		type: Boolean,
		default: false
	}
});

const emit = defineEmits(['onBackClick']);
</script>

<style scoped lang="scss">
.bio-popup-menu {
	width: 320px;
	max-width: 320px !important;
	box-shadow: 0 4px 10px 0 #00000033;
	border-radius: 12px !important;
	padding-left: 8px;
	padding-right: 8px;

	.bio-popup-title-bar {
		width: 100%;
		height: 44px;
		padding-left: 12px;
		padding-right: 12px;

		.bio-popup-title {
			font-family: Roboto;
			font-size: 14px;
			font-weight: 500;
			line-height: 20px;
			letter-spacing: 0em;
			text-align: left;
			color: $ink-1;
		}
	}

	.bio-popup-label {
		padding-left: 12px;
		padding-right: 12px;
		padding-bottom: 12px;
		font-family: Roboto;
		font-size: 12px;
		font-weight: 400;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: left;
		color: $ink-3;
	}
}
</style>
