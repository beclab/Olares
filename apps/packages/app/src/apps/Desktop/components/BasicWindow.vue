<template>
	<VueDragResize
		:isActive="true"
		:key="value"
		:w="value.width"
		:h="value.height"
		:parentW="value.max_width"
		:parentH="value.max_height"
		:minw="value.min_width"
		:minh="value.min_height"
		:x="value.left"
		:y="value.top"
		:isResizable="value.isResizable"
		v-on:dragging="dragging"
		v-on:dragstop="dragStop"
		v-on:resizing="resizing"
		v-on:resizestop="resizeStop"
		@click="onTop"
		:stickSize="24"
		class="iframe-box bg-background-1"
		:class="
			isFull ? (value.isResizable ? 'animationClass' : 'animationClass2') : ''
		"
		:parentLimitation="true"
		ref="dragRef"
	>
		<div class="iframeDiv" v-if="!value.active || isDrag || isResize"></div>
		<iframe
			class="iframe-window"
			allow="web-share; clipboard-read; clipboard-write"
			:src="value.pathto ? value.url + '?pathto=' + value.pathto : value.url"
			ref="iframeRef"
		/>

		<div class="title" style="width: 100%; height: 40px; padding: 0 12px">
			<q-img
				class="app_icon"
				:src="value.icon"
				style="width: 20px; height: 20px"
			/>

			<span class="app_title ink-1">
				{{
					['Files', 'Market', 'Dashboard', 'Settings', 'Control Hub'].includes(
						value.title
					)
						? t(`system_app.${value.title}`)
						: value.title
				}}
			</span>

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
				<img v-show="h_min" src="../../../assets/desktop/window_min_h.svg" />
				<img v-show="!h_min" src="../../../assets/desktop/window_min_n.svg" />
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
				<img v-show="h_full" src="../../../assets/desktop/window_full_h.svg" />
				<img v-show="!h_full" src="../../../assets/desktop/window_full_n.svg" />
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
				<img
					v-show="h_new"
					style="width: 13px; height: 13px"
					src="../../../assets/desktop/window_new_h.svg"
				/>
				<img
					v-show="!h_new"
					style="width: 13px; height: 13px"
					src="../../../assets/desktop/window_new_n.svg"
				/>
			</q-btn>

			<q-btn
				flat
				round
				size="xs"
				@click="onClose"
				class="window_btn"
				@mouseover="h_close = true"
				@mouseleave="h_close = false"
			>
				<img
					v-show="h_close"
					src="../../../assets/desktop/window_close_h.svg"
				/>
				<img
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
			:backgroundColor="$q.dark.isActive ? 'rgba(31, 31, 31, 1)' : '#ffffff'"
		>
		</BtLoading>
	</VueDragResize>
</template>

<script lang="ts">
import { ref, defineComponent, onMounted, nextTick, onUnmounted } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';

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
		const { t } = useI18n();
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
		const $q = useQuasar();

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

		let dragStop = () => {
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

		let resizeStop = () => {
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

			const href = value.value.redirectUrl || value.value.url;

			window.open(href);
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

		const message = (event: any) => {
			if (event.data.message === 'theme_update') {
				if (event.data.info.theme.toString() === '1') {
					$q.dark.set(false);
				} else {
					$q.dark.set(true);
				}
				iframeRef.value.contentWindow.postMessage(
					{
						message: 'theme_apps_update',
						info: event.data.info
					},
					'*'
				);
			}
		};

		onMounted(() => {
			window.addEventListener('message', message);

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

		onUnmounted(() => {
			window.removeEventListener('message', message);
		});

		return {
			$q,
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
			dragStop,
			resizing,
			resizeStop,
			activated,
			deactivated,
			onClose,
			onMini,
			onFull,
			onTop,
			onNew,
			t
		};
	}
});
</script>

<style lang="scss">
.animationClass {
	top: 0 !important;
	left: calc(100vw + 108px) !important;
	transition: all 0.3s;

	.content-container {
		width: calc(100vw - 108px) !important;
		height: 100vh !important;
		transition: all 0.3s;
	}
}

.animationClass2 {
	top: 0 !important;
	transition: all 0.3s;

	.content-container {
		height: 100vh !important;
		transition: all 0.3s;
	}
}
</style>

<style lang="scss">
.iframe-box {
	box-shadow: 0px 0px 2px 0px rgba(0, 0, 0, 0.4),
		0px 0px 40px 0px rgba(0, 0, 0, 0.2);
	border-radius: 10px;

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
	}

	.title {
		display: flex;
		align-items: center;
		justify-content: space-around;
		border-bottom: 1px solid $separator;
		background-color: $background-1;

		.app_title {
			font-family: 'Inter';
			font-style: normal;
			font-weight: 600;
			font-size: 16px;
			line-height: 20px;
			margin-left: 12px;
		}

		.app_icon {
			width: 20px;
			height: 20px;
			padding: 0px;
			border-radius: 4px;
		}

		.window_btn {
			margin-left: 8px;
			// margin-top: 4px;
			.q-focus-helper {
				display: none;
			}
		}
	}
}

.iframeDiv {
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
