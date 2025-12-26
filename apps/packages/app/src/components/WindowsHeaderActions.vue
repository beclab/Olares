<template>
	<div class="windows-actions-root row items-center justify-end q-pr-xs">
		<button
			class="button-item row items-center justify-center"
			@click="action('minimize')"
		>
			<img src="../assets/images/windows/minimize.svg" width="20" height="20" />
		</button>
		<button
			class="button-item row items-center justify-center"
			@click="action('maximize')"
		>
			<img
				v-if="maximized"
				src="../assets/images/windows/unmaximize.svg"
				width="20"
				height="20"
			/>
			<img
				v-else
				src="../assets/images/windows/maximize.svg"
				width="20"
				height="20"
			/>
		</button>
		<button
			class="button-item row items-center justify-center"
			@click="action('close')"
		>
			<img src="../assets/images/windows/closed.svg" width="20" height="20" />
		</button>
	</div>
</template>

<script setup lang="ts">
import { useQuasar } from 'quasar';
import { IPCWindowsHeaderMenusEventName } from 'src/platform/interface/electron/interface';
import { onMounted, ref } from 'vue';

const $q = useQuasar();

const maximized = ref(false);

onMounted(async () => {
	updateMaximized();

	window.addEventListener('resize', () => {
		updateMaximized();
	});
});

const updateMaximized = async () => {
	if ($q.platform.is.electron && $q.platform.is.win) {
		maximized.value = await window.electron.api.windows.isMaximized();
	}
};

const action = async (name: IPCWindowsHeaderMenusEventName) => {
	if (!$q.platform.is.electron) {
		return;
	}
	if (name == 'close') {
		await window.electron.api.windows.close();
	} else if (name == 'maximize') {
		await window.electron.api.windows.maximize();
	} else if (name == 'minimize') {
		await window.electron.api.windows.minimize();
	}
};
</script>

<style lang="scss" scoped>
.windows-actions-root {
	width: 130px;
	height: 100%;
	-webkit-app-region: no-drag;

	.button-item {
		width: 40px;
		height: 100%;
		cursor: pointer;
		transition: 0.2s ease all;
		border: 0;
		color: $ink-2;
		background: transparent;
		box-shadow: none;
	}

	.close-item {
		@extend .button-item;
	}

	.button-item:hover {
		background-color: $grey-12;
	}
}
</style>
