<template>
	<div class="error-root row justify-end items-center">
		<slot />
		<slot name="reminder" />
		<transition name="fade">
			<div
				v-if="isError && errorMessage && errorMessage.length > 0"
				class="text-overline message-parent row items-center bg-background-3 q-mb-md"
				:class="paddingX ? 'item-margin-left item-margin-right' : ''"
			>
				<q-icon
					v-if="withPopup"
					size="12px"
					name="sym_r_error"
					color="red-6"
					style="margin-right: 6px"
				>
					<q-popup-proxy
						class="error_banner"
						v-if="popupMessage && popupMessage.length > 0"
					>
						<div class="banner_message" v-html="popupMessage" />
					</q-popup-proxy>
				</q-icon>
				<span class="text-body3 text-red">{{ errorMessage }}</span>
				<span
					v-if="linkMessage"
					class="link-message"
					@click="emit('onLinkClick')"
				>
					{{ linkMessage }}
				</span>
			</div>
		</transition>
		<bt-separator v-if="widthSeparator" :offset="16" />
	</div>
</template>

<script lang="ts" setup>
import BtSeparator from '../base/BtSeparator.vue';

import {
	getCurrentInstance,
	inject,
	onMounted,
	provide,
	ref,
	watch
} from 'vue';

const props = defineProps({
	isError: {
		type: Boolean,
		default: false,
		require: true
	},
	errorMessage: {
		type: String,
		require: true
	},
	withPopup: {
		type: Boolean,
		default: false
	},
	popupMessage: {
		type: String,
		require: false
	},
	linkMessage: {
		type: String,
		require: false
	},
	widthSeparator: {
		type: Boolean,
		required: false,
		default: true
	},
	paddingX: {
		type: Boolean,
		required: false,
		default: true
	}
});
const addChildComponent = inject('addChildComponent');
const onErrorStatusChange = inject('onErrorStatusChange');
const focused = ref(false);
const blured = ref(false);

const setFocused = (focus: boolean) => {
	focused.value = focus;
};

const setBlured = (blur: boolean) => {
	blured.value = blur;
};

provide('setFocused', setFocused);
provide('setBlured', setBlured);
const selfRef = ref(getCurrentInstance());
onMounted(() => {
	if (addChildComponent) {
		addChildComponent(selfRef);
	}
});

watch(
	() => props.isError,
	() => {
		if (onErrorStatusChange) {
			onErrorStatusChange(props.isError);
		}
	}
);

const getErrorStatus = () => {
	return props.isError;
};

defineExpose({ getErrorStatus });

const emit = defineEmits(['onLinkClick']);
</script>

<style scoped lang="scss">
.error-root {
	// width: 200px;
	// height: 100%;
	width: 100%;
	.message-parent {
		text-align: left;
		// position: absolute;
		// background-color: red;
		width: 100%;
		padding: 4px 8px;
		border-radius: 4px;

		.error-message {
			color: $red;
		}

		.link-message {
			color: $blue;
			cursor: pointer;
			text-decoration: none;
			margin-left: 3px;
		}
	}
}
</style>
