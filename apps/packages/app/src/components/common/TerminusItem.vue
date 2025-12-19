<template>
	<q-item
		dense
		:clickable="clickable"
		class="click-item-root row items-center"
		:class="showBoard ? borderClasses : ''"
		:style="{
			height: itemHeight + 'px',
			'border-radius': borderRadius + 'px',
			'--padding-left': paddingLeft + 'px'
		}"
	>
		<!-- :style="
			'height:' + itemHeight + 'px;' + 'border-radius: ' + borderRadius + 'px;'
		" -->
		<div
			class="img row items-center justify-center"
			:class="imgBgClasses"
			:style="`border-radius:${
				wholePictureSize / 2
			}px; width:${wholePictureSize}px; height:${wholePictureSize}px;`"
			v-if="iconName || imagePath"
		>
			<q-icon
				:name="iconName"
				:size="`${iconSize}px`"
				:color="iconColor"
				v-if="iconName"
			/>
			<img
				:src="getRequireImage(imagePath)"
				:style="`width:${iconSize}px;height:${iconSize}px;`"
				v-if="imagePath"
			/>
		</div>
		<div
			class="item-content row items-center justify-between"
			:style="`width:calc(100% - ${imageWidth()}px - 8px);`"
		>
			<div
				class="content-frontend column justify-center"
				:class="contentFrontendClasses"
			>
				<div
					class="title text-subtitle2"
					:class="titleClasses"
					:style="titleColor ? `color:${titleColor};` : ''"
				>
					<slot name="title"></slot>
				</div>

				<div class="detail text-body3 text-ink-2" v-if="$slots.detail">
					<slot name="detail"></slot>
				</div>
			</div>

			<div
				class="column justify-center items-end"
				v-if="$slots.side"
				:style="sideStyle"
			>
				<slot name="side"></slot>
			</div>
		</div>
	</q-item>
</template>

<script setup lang="ts">
import { getRequireImage } from '../../utils/imageUtils';

const prop = defineProps({
	titleColor: {
		type: String,
		required: false
	},

	titleClasses: {
		type: String,
		required: false,
		default: 'text-ink-1'
	},

	borderRadius: {
		type: Number,
		default: 20,
		required: false
	},

	terminusName: {
		type: String,
		required: false,
		default: ''
	},

	iconName: {
		type: String,
		default: '',
		required: false
	},
	iconColor: {
		type: String,
		default: '',
		required: false
	},

	iconSize: {
		type: Number,
		default: 20,
		required: false
	},

	imgBgClasses: {
		type: String,
		required: false,
		default: ''
	},

	wholePictureSize: {
		type: Number,
		default: 40,
		required: false
	},
	itemHeight: {
		type: Number,
		default: 64,
		required: false
	},
	clickable: {
		type: Boolean,
		required: false,
		default: true
	},
	showBoard: {
		type: Boolean,
		required: false,
		default: true
	},
	imagePath: {
		type: String,
		required: false,
		default: ''
	},
	sideStyle: {
		type: String,
		required: false,
		default: ''
	},
	borderClasses: {
		type: String,
		required: false,
		default: 'border-class'
	},
	contentFrontendClasses: {
		type: String,
		required: false,
		default: ''
	},
	paddingLeft: {
		type: Number,
		required: false,
		default: 16
	}
});

const imageWidth = () => {
	if (prop.iconName || prop.imagePath || prop.terminusName) {
		return prop.wholePictureSize + prop.paddingLeft;
	}
	return 0;
};
</script>

<style scoped lang="scss">
.click-item-root {
	width: 100%;
	padding: 0;
	padding-left: var(--padding-left);

	.img {
		margin-right: 12px;
	}

	.item-content {
		height: 100%;
	}

	.detail {
		margin-top: 4px;
	}

	.title {
		width: calc(100%);
	}
}
.border-class {
	border: 1px solid $separator;
}
</style>
