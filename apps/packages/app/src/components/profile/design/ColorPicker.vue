<template>
	<div class="color-picker column justify-start items-start">
		<span v-if="title" class="color-picker__title">{{ title }}</span>
		<div class="color-picker__edit flex">
			<div
				class="color-picker__edit__selected"
				:style="{ background: hexColor }"
			>
				<q-popup-proxy cover transition-show="scale" transition-hide="scale">
					<q-color
						v-model="selectedColor"
						no-header
						no-footer
						@update:model-value="selectedColorUpdate"
						default-view="palette"
					/>
				</q-popup-proxy>
			</div>
			<div class="color-picker__edit__input_background">
				<q-input
					borderless
					type="text"
					input-class="color-input"
					class="color-input"
					:model-value="hexColor"
					@update:model-value="inputColorUpdate"
				/>
			</div>
			<color-opacity-edit-view
				:input-disabled="disableRef"
				@on-update="emitHexColor"
				ref="opacityEdit"
				v-model:model-value="opacityNumber"
				:hex-color="hexColor"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import ColorOpacityEditView from '@apps/profile/src/components/profile/design/ColorOpacityEditView.vue';
import { onMounted, ref } from 'vue';
import { colors } from 'quasar';
import { colorsRgba } from 'quasar/dist/types/utils/colors';
import hexToRgb = colors.hexToRgb;
import rgbToHex = colors.rgbToHex;

const props = defineProps({
	title: String,
	hexColor: String,
	alphaAble: {
		type: Boolean,
		default: true
	}
});

const disableRef = ref(false);
const selectedColor = ref();
const opacityEdit = ref();
const opacityNumber = ref(100);
const emit = defineEmits(['update:hexColor']);

const selectedColorUpdate = () => {
	const rgba: colorsRgba = hexToRgb(selectedColor.value);
	rgba.a = opacityNumber.value;
	const hex = rgbToHex(rgba);
	emit('update:hexColor', hex);
};

const emitHexColor = (alpha: number) => {
	const rgba: colorsRgba = hexToRgb(selectedColor.value);
	rgba.a = alpha;
	const hex = rgbToHex(rgba);
	emit('update:hexColor', hex);
};

const inputColorUpdate = (value: string) => {
	try {
		const rgba: colorsRgba = hexToRgb(value);
		console.log(rgba);
		opacityEdit.value.setHexColor(value);
		emit('update:hexColor', value);
	} catch (e) {
		console.log(e);
	}
};

onMounted(() => {
	selectedColor.value = props.hexColor;
	disableRef.value = !props.alphaAble;
	console.log(disableRef.value);
});
</script>

<style scoped lang="scss">
.color-picker {
	width: 100%;
	height: auto;
	padding-top: 12px;

	&__title {
		color: $ink-3;
		font-family: Roboto;
		font-size: 12px;
		font-weight: 400;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: left;
		margin-bottom: 4px;
	}

	&__edit {
		width: 100%;
		height: 32px;

		&__selected {
			width: 32px;
			height: 32px;
			border-radius: 8px;
			border: 1px solid $input-stroke;
		}

		&__input_background {
			width: 160px;
			height: 100%;
			border-radius: 8px;
			border: 1px solid $input-stroke;
			margin-left: 8px;
			margin-right: 8px;
			padding-bottom: -20px;
			position: relative;

			.color-input {
				margin-top: -13px;
				padding-left: 10px;
				padding-right: 10px;
				position: absolute;
				color: $ink-2 !important;
			}
		}
	}
}
</style>
