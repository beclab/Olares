<template>
	<div class="lp-app-tile">
		<div class="lp-app-tile__drag-mask" aria-hidden="true" />
		<div
			v-show="showDeleteButton"
			class="delete_launch"
			@click.stop="emit('delete-click', $event)"
		/>
		<div
			v-if="showInstallLoading"
			class="install_loading_status"
			:style="loadingOverlayStyle"
		>
			<svg viewBox="0 0 32 32">
				<circle r="16" cx="16" cy="16" />
			</svg>
		</div>
		<div class="lp-app-tile__icon" :style="iconBoxStyle">
			<img :src="icon" :alt="label" :style="imgStyle" />
			<BookmarkBadge v-if="kind === 'bookmark'" />
		</div>
		<div class="launch_pad_apps_name" :data-index="appId">
			<span v-if="showStoppedState" class="app_state q-mr-sm suspend_color" />
			<span v-if="showCrashState" class="app_state q-mr-sm crash_color" />
			{{ label }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { borderRadiusFormat } from 'src/utils/desktop/utils';
import BookmarkBadge from './components/BookmarkBadge.vue';

const props = defineProps<{
	appId: string;
	icon: string;
	label: string;
	tileSizePx: number;
	showDeleteButton: boolean;
	showInstallLoading: boolean;
	showStoppedState: boolean;
	showCrashState: boolean;
	iconIsGrayscale: boolean;
	kind?: 'app' | 'bookmark';
}>();

const emit = defineEmits<{ 'delete-click': [e: MouseEvent] }>();

const borderRadiusPx = computed(() =>
	borderRadiusFormat(props.tileSizePx, props.tileSizePx)
);

const imgStyle = computed(() => {
	const r = borderRadiusPx.value;
	const filter = props.iconIsGrayscale
		? 'grayscale(100%) brightness(0.8)'
		: 'grayscale(0%)';
	return `width:${props.tileSizePx}px;height:${props.tileSizePx}px;border-radius:${r}px;filter:${filter}`;
});

// Icon-sized relative box so the bookmark badge's percentage offsets align
// to the icon itself (the tile root is taller because of the label below).
const iconBoxStyle = computed(
	() =>
		`position:relative;display:block;width:${props.tileSizePx}px;height:${props.tileSizePx}px;`
);

const loadingOverlayStyle = computed(() => {
	const r = borderRadiusPx.value;
	return `width:${props.tileSizePx}px;height:${props.tileSizePx}px;border-radius:${r}px;`;
});
</script>

<style lang="scss">
@mixin launchpad-app-tile-shell {
	border-radius: 16px;
	z-index: 1;
	box-sizing: border-box;

	> img {
		display: block;
	}
}

@mixin launchpad-app-tile-children {
	.launch_pad_apps_name {
		width: 140%;
		transform: translateX(-15%);
		text-align: center;
		font-size: 14px;
		font-family: Roboto-Medium, Roboto;
		font-weight: 500;
		color: #ffffff;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		margin-top: 12px;

		.app_state {
			display: inline-block;
			width: 8px;
			height: 8px;
			border-radius: 4px;

			&.suspend_color {
				background-color: $warning;
			}

			&.crash_color {
				background-color: $negative;
			}

			&.running_color {
				background-color: $positive;
			}

			&.upgrade_error_color {
				background-color: $blue;
			}
		}
	}

	.delete_launch {
		width: 20px;
		height: 20px;
		background-image: url('../../assets/desktop/delete_app_icon.svg');
		background-repeat: no-repeat;
		background-position: center;
		background-size: contain;
		position: absolute;
		top: -5px;
		left: -5px;
		z-index: 100;
		cursor: pointer;
	}

	.install_loading_status {
		position: absolute;
		top: 0px;
		left: 0px;
		z-index: 99;
		background-image: url('../../assets/desktop/installing.svg');
		background-position: center;
		background-size: 100% 100%;
		background-repeat: no-repeat;
		display: flex;
		align-items: center;
		justify-content: center;

		svg {
			transform: rotate(-90deg);
			border-radius: 50%;
			height: 40%;
		}

		circle {
			fill: rgba(0, 0, 0, 0.5);
			stroke: rgba(255, 255, 255, 1);
			stroke-width: 32;
			stroke-dasharray: 0 100;
			animation: fillup 5s linear infinite;
		}

		@keyframes fillup {
			to {
				stroke-dasharray: 158 158;
				opacity: 0;
			}
		}
	}
}

@mixin launchpad-drag-preview-host-shell {
	position: fixed;
	left: var(--lp-preview-left, 0);
	top: var(--lp-preview-top, 0);
	width: var(--lp-preview-width, auto);
	height: var(--lp-preview-height, auto);
	margin: 0;
	padding: 0;
	pointer-events: none;
	z-index: 2147483647;
	overflow: visible;
	box-sizing: border-box;
	background: transparent;

	.lp-app-tile {
		@include launchpad-app-tile-shell;
		position: relative !important;
		top: 0 !important;
		left: 0 !important;
		width: var(--lp-preview-width, auto) !important;
		margin: 0;
		transition: none !important;
		animation: none !important;
		-webkit-animation: none !important;
		transform: none !important;
		opacity: 1 !important;
		height: auto;
	}
}

.lp-app-tile {
	position: relative;
	@include launchpad-app-tile-shell;
	@include launchpad-app-tile-children;

	.lp-app-tile__drag-mask {
		width: 140%;
		height: 120%;
		transform: translate(-13%, -10%);
		position: absolute;
		top: 0;
		left: 0;
		z-index: 2;
		overflow: hidden;
		cursor: pointer;
	}
}

.lp-drag-preview-host {
	@include launchpad-drag-preview-host-shell;
}
</style>
