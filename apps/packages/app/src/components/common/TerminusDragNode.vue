<template>
	<div
		@dblclick="barDBClick"
		@mouseenter.self="mouseEnter"
		@mouseleave.self="mouseLeave"
		@mousedown.self="mouseDown"
		@mouseup.self="mouseUp"
	>
		<slot></slot>
	</div>
</template>

<script setup lang="ts">
import { useQuasar } from 'quasar';

const $q = useQuasar();

var enterFlag = false;
var mouseDownFlag = false;
var moveTimer: NodeJS.Timer | null = null;

const barDBClick = async () => {
	if (!$q.platform.is.electron) {
		return;
	}
	window.electron.api.windows.maximize();
	setTimeout(() => {
		window.electron.api.windows.winMove(false);
		mouseDownFlag = false;
	}, 10);
};

const mouseDown = async () => {
	if (!$q.platform.is.electron) {
		return;
	}

	const isMaximized = await window.electron.api.windows.isMaximized();
	if (isMaximized) {
		return;
	}
	window.electron.api.windows.winMove(true);
	mouseDownFlag = true;
};

const mouseLeave = () => {
	if (!$q.platform.is.electron) {
		return;
	}
	enterFlag = false;

	if (moveTimer !== null) {
		moveTimer = setTimeout(() => {
			mouseDownFlag = false;
			window.electron.api.windows.winMove(false);
			moveTimer = null;
		}, 1000);
	}
};
const mouseEnter = () => {
	if (!$q.platform.is.electron) {
		return;
	}
	enterFlag = true;
};

const mouseUp = () => {
	if (!$q.platform.is.electron) {
		return;
	}
	if (enterFlag && mouseDownFlag) {
		window.electron.api.windows.winMove(false);
		mouseDownFlag = false;
	}
};
</script>
