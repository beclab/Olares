<template>
	<div
		class="iframe-box-mobile bg-background-1"
		:class="isFull ? 'animationClass-mobile' : ''"
		:style="{
			width: `${value.width}px`,
			height: `${value.height}px`,
			position: 'relative',
			left: `${value.left}px`,
			top: `${value.top}px`
		}"
		@click="onTop"
	>
		<div
			class="iframeDiv-mobile"
			v-if="!value.active || isDrag || isResize"
		></div>
		<iframe
			class="iframe-window"
			allow="web-share; clipboard-read; clipboard-write"
			:src="value.pathto ? value.url + '?pathto=' + value.pathto : value.url"
			ref="iframeRef"
		/>

		<div class="title" style="width: 100%; height: 40px; padding: 0 12px">
			<q-img class="app_icon" :src="value.icon" />

			<span class="app_title"> {{ value.title }} {{ value.isResizable }}</span>

			<q-space />

			<q-btn
				flat
				round
				size="xs"
				class="window_btn"
				@click="onMini"
				@mouseover="h_min = true"
				@mouseleave="h_min = false"
			>
				<q-img v-show="h_min" src="../../../assets/desktop/window_min_h.svg" />
				<q-img v-show="!h_min" src="../../../assets/desktop/window_min_n.svg" />
			</q-btn>

			<q-btn
				flat
				round
				size="xs"
				class="window_btn"
				@click="onFull"
				@mouseover="h_full = true"
				@mouseleave="h_full = false"
			>
				<q-img
					v-show="h_full"
					src="../../../assets/desktop/window_full_h.svg"
				/>
				<q-img
					v-show="!h_full"
					src="../../../assets/desktop/window_full_n.svg"
				/>
			</q-btn>

			<q-btn
				flat
				round
				size="xs"
				class="window_btn"
				@click="onNew"
				@mouseover="h_new = true"
				@mouseleave="h_new = false"
			>
				<q-img
					v-show="h_new"
					style="width: 13px; height: 13px"
					src="../../../assets/desktop/window_new_h.svg"
				/>
				<q-img
					v-show="!h_new"
					style="width: 13px; height: 13px"
					src="../../../assets/desktop/window_new_n.svg"
				/>
			</q-btn>

			<q-btn
				flat
				round
				size="xs"
				@click.once="onClose"
				class="window_btn"
				@mouseover="h_close = true"
				@mouseleave="h_close = false"
			>
				<q-img
					v-show="h_close"
					src="../../../assets/desktop/window_close_h.svg"
				/>
				<q-img
					v-show="!h_close"
					src="../../../assets/desktop/window_close_n.svg"
				/>
			</q-btn>
		</div>

		<BtLoading
			v-if="iframeLoad"
			:show="true"
			textColor="#4999ff"
			color="#4999ff"
			text=""
			backgroundColor="#ffffff"
		>
		</BtLoading>
	</div>
</template>

<script lang="ts">
import { ref, defineComponent, onMounted, nextTick } from 'vue';

interface WindowRect {
	left: number;
	top: number;
	width: number;
	height: number;
}

export default defineComponent({
	name: 'BasicWindow',
	props: {
		modelValue: {
			type: Object,
			default: () => {
				return {
					id: '',
					width: 400,
					height: 300,
					max_width: 400,
					max_height: 300,
					min_width: 400,
					min_height: 300,
					top: 0,
					left: 0,
					title: '',
					url: '',
					icon: '',
					active: true
				};
			}
		}
	},

	setup(props, context) {
		const value = ref(props.modelValue);
		const h_close = ref(false);
		const h_min = ref(false);
		const h_full = ref(false);
		const h_new = ref(false);
		const isActive = ref(false);
		const isDrag = ref(false);
		const isResize = ref(false);
		const iframeLoad = ref(true);
		const iframeRef = ref();
		const dragRef = ref();
		const isFull = ref(false);
		const clickNew = ref(false);

		const app_icon: string = value.value.icon;

		let dragging = (rect: WindowRect) => {
			if (!value.value.active) {
				context.emit('onTop', value.value);
			}
			isDrag.value = true;
			let callback_obj = {};
			Object.assign(callback_obj, value.value, rect);
			context.emit('update:modelValue', callback_obj);
		};

		let dragstop = () => {
			isDrag.value = false;
		};

		let resizing = (rect: WindowRect) => {
			if (!value.value.active) {
				context.emit('onTop', value.value);
			}
			isResize.value = true;
			let callback_obj = {};
			Object.assign(callback_obj, value.value, rect);
			context.emit('update:modelValue', callback_obj);
		};

		let resizestop = () => {
			isResize.value = false;
		};

		let activated = () => {
			isActive.value = true;
		};

		let deactivated = () => {
			isActive.value = false;
		};

		let onClose = () => {
			context.emit('close', value.value);
		};

		let onMini = () => {
			context.emit('mini', value.value);
		};

		let onFull = () => {
			isFull.value = !isFull.value;
		};

		let onNew = () => {
			clickNew.value = true;
			setTimeout(() => {
				clickNew.value = false;
			}, 500);
			window.open(value.value.url);
		};

		let onTop = () => {
			if (clickNew.value) {
				return false;
			}
			if (checkDoubleClick()) {
				isFull.value = !isFull.value;
			}
			context.emit('onTop', value.value);
		};

		let lastTapTimeFunc: NodeJS.Timeout;
		let lastTapDiffTime = 0;
		const checkDoubleClick = function () {
			let curT = new Date().getTime();
			let lastT = lastTapDiffTime;
			lastTapDiffTime = curT;
			let diff = curT - lastT;
			if (diff < 300) {
				clearTimeout(lastTapTimeFunc);
				return true;
			} else {
				lastTapTimeFunc = setTimeout(function () {
					return false;
				}, 300);
			}
		};

		onMounted(() => {
			nextTick(() => {
				new MutationObserver(function (mutations) {
					mutations.some(function (mutation) {
						if (
							mutation.type === 'attributes' &&
							mutation.attributeName === 'src'
						) {
							return true;
						}

						return false;
					});
				}).observe(document.body, {
					attributes: true,
					attributeFilter: ['src'],
					attributeOldValue: true,
					characterData: false,
					characterDataOldValue: false,
					childList: false,
					subtree: true
				});

				if (iframeRef.value.attachEvent) {
					// IE
					iframeRef.value.attachEvent('onload', () => {
						iframeLoad.value = false;
					});
				} else {
					// not IE
					iframeRef.value.onload = function () {
						iframeLoad.value = false;
					};
				}
			});
		});

		return {
			value,
			isActive,
			isDrag,
			isResize,
			h_close,
			h_min,
			h_full,
			h_new,
			app_icon,
			iframeLoad,
			iframeRef,
			dragRef,
			isFull,
			dragging,
			dragstop,
			resizing,
			resizestop,
			activated,
			deactivated,
			onClose,
			onMini,
			onFull,
			onTop,
			onNew
		};
	}
});
</script>

<style lang="scss">
.animationClass-mobile {
	top: 0 !important;
	left: 0vw !important;
	transition: all 0.3s;
	width: 100vw !important;
	height: 100vh !important;
	z-index: 10 !important;
}
</style>

<style lang="scss">
.iframe-box-mobile {
	box-shadow: 0 1px 5px rgb(0 0 0 / 20%), 0 2px 2px rgb(0 0 0 / 14%),
		0 3px 1px -2px rgb(0 0 0 / 12%);
	border-radius: 10px;
	overflow: hidden;

	.iframe-window {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 0;
		right: 0;
		width: 100%;
		height: 100%;
		padding-top: 40px;
		border-radius: 10px;
		border-width: 0px;
		z-index: -1;
	}

	.title {
		display: flex;
		align-items: center;
		justify-content: space-around;
		border-bottom: 1px solid #e0e0e0;
		position: absolute;
		top: 0;
		left: 0;

		.app_title {
			font-family: 'Source Han Sans CN';
			font-style: normal;
			font-weight: 500;
			font-size: 16px;
			line-height: 20px;
			/* identical to box height, or 100% */

			color: #414141;
			margin-left: 15px;
		}

		.app_icon {
			width: 20px;
			height: 20px;
			padding: 0px;
			border-radius: 4px;
		}

		.window_btn {
			width: 16px;
			height: 16px;
			min-width: 16px;
			min-height: 16px;
			margin-left: 20px;
			padding: 2px;
			// margin-top: 4px;
			.q-focus-helper {
				display: none;
			}
		}
	}
}

.iframeDiv-mobile {
	width: 100%;
	height: 100%;
	position: absolute;
	z-index: 1111;
	filter: alpha(opacity=0);
	opacity: 0;
	background: transparent;
	margin-top: 30px;
	/*display: none;*/
}
</style>
