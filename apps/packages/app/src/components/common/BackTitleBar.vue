<template>
	<div v-show="hideBackBar">
		<div class="back-title-bar2 row items-center">
			<q-btn
				:class="showBack ? 'back-show-btn' : 'back-hidden-btn'"
				flat
				@click="backFunc"
			>
				<q-icon name="sym_r_chevron_left" size="24px" color="grey-8"></q-icon>
			</q-btn>
			<q-space />

			<div class="column justify-center content-center items-center title-bg">
				<div
					class="text-color-title text-subtitle1 title"
					v-show="title.length > 0"
				>
					{{ title }}
				</div>
			</div>

			<q-space />
			<q-btn
				:class="rightImagePath.length ? 'right-show-btn' : 'right-hidden-btn'"
				flat
				rounded
				dense
				@click="rightAction"
			>
				<q-img
					v-show="rightImagePath.length"
					:src="getRequireImage(rightImagePath)"
					width="24px"
					height="24px"
				>
				</q-img>
			</q-btn>

			<q-img
				class="trashcan"
				v-if="showTrashcan"
				src="../../assets/contact/trashcan.svg"
				@click="onTrashcan"
			>
			</q-img>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { ref } from 'vue';
import { getRequireImage } from '../../utils/imageUtils';

defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	topTitle: {
		type: String,
		required: false,
		default: ''
	},
	showBack: {
		type: Boolean,
		required: false,
		default: true
	},
	rightImagePath: {
		type: String,
		required: false,
		default: ''
	},
	showTrashcan: {
		type: Boolean,
		required: false,
		default: false
	}
});

const route = useRouter();

const emit = defineEmits(['leftAction', 'rightAction', 'onTrashcan']);

const rightAction = () => {
	emit('rightAction');
};

const backFunc = () => {
	if (window.history.length <= 1) {
		return;
	}
	route.go(-1);
};

const onTrashcan = () => {
	emit('onTrashcan');
};

const hideBackBar = ref(true);
</script>

<style lang="scss" scoped>
.back-title-bar2 {
	height: 44px;
	width: 100%;
	padding: 0 12px;

	.base-btn {
		width: 40px;
		height: 100%;
	}

	.back-show-btn {
		@extend .base-btn;
	}

	.back-hidden-btn {
		@extend .base-btn;
		visibility: hidden;
	}

	.title-bg {
		min-width: 150px;

		.title {
			text-align: center;
		}
	}

	.right-hidden-btn {
		@extend .base-btn;
		visibility: hidden;
	}

	.right-show-btn {
		@extend .base-btn;
		width: 40px;
	}

	.trashcan {
		width: 14px;
		height: 16px;
	}
}
</style>
