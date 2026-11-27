<template>
	<div class="bookmark-badge" aria-hidden="true">
		<svg viewBox="0 0 24 24" class="bookmark-badge__glyph">
			<path
				d="M6 2.5h12a1 1 0 0 1 1 1V21l-7-4-7 4V3.5a1 1 0 0 1 1-1z"
				fill="none"
				stroke="#3d7fff"
				stroke-width="2"
				stroke-linejoin="round"
				stroke-linecap="round"
			/>
		</svg>
	</div>
</template>

<script lang="ts" setup></script>

<style lang="scss" scoped>
// Mimics an app's baked-in corner badge (e.g. OpenClaw CLI's "</>" mark): a
// solid circular plate at the bottom-right of the icon holding a white glyph,
// with a light ring + shadow so it reads on any icon. Percentage-based so it
// scales with the icon; must be placed inside a container sized exactly to the
// icon (see LaunchpadAppTile / the mobile `.relative-position` wrapper).
.bookmark-badge {
	position: absolute;
	right: 8%;
	bottom: 8%;
	width: 38%;
	height: 38%;
	border-radius: 50%;
	background: #fff;
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0px 1px 3px rgba(0, 0, 0, 0.3);
	box-sizing: border-box;
	z-index: 101;
	pointer-events: none;

	.bookmark-badge__glyph {
		width: 56%;
		height: 56%;
	}
}
</style>
