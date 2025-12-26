<template>
	<div class="relative-position terminal-content">
		<div class="terminal-content-wrapper">
			<div style="height: calc(100%)" ref="terminalRef"></div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import ReconnectingWebSocket from 'reconnecting-websocket';
import { debounce } from 'lodash';
import { onMounted } from 'vue';
import { onBeforeUnmount } from 'vue';
import { watch } from 'vue';
import { nextTick } from 'vue';
import 'xterm/css/xterm.css';

interface Props {
	node: string;
	visible: boolean;
}
const props = withDefaults(defineProps<Props>(), {});

const aceVisileb = ref(false);
let messageQueue: string[] = [];
let fitAddon: FitAddon | undefined = undefined;
let termTemp: Terminal | undefined = undefined;
let inited = false;

const termA = ref<any>(null);
const ws = ref();
const initTimer = ref();
const first = ref(true);

const terminalRef = ref<any>();
let locker: NodeJS.Timeout | undefined = undefined;
const isWsOpen = () => {
	return ws.value && ws.value.readyState === ws.value.OPEN;
};

const packResize = (col: number, row: number) =>
	JSON.stringify({
		Op: 'resize',
		Cols: col,
		Rows: row
	});

const resizeRemoteTerminal = () => {
	const { cols, rows } = termA.value;
	if (isWsOpen()) {
		ws.value.send(packResize(cols, rows));
	}
};

const disableTermStdin = (disabled = true) => {
	const { textarea = {} } = termA.value;
	textarea.disabled = disabled;
};

const unpackStdout = (data: any) => data.Data;

const onWSReceive = (MessageEvent: any) => {
	const data = JSON.parse(MessageEvent.data);
	initTimer.value && clearInterval(initTimer.value);

	setHeartBeat();

	if (first.value) {
		first.value = false;
		disableTermStdin(false);
		termA.value.reset();
		termA.value.element && termA.value.focus();
		resizeRemoteTerminal();
	}

	const stdout = unpackStdout(data);
	termA.value.write(stdout);
};

const disconnect = () => {
	if (termA.value) {
		disableTermStdin(true);
	}

	if (ws.value) {
		ws.value.close(1000);
	}
};

const fatal = (message: any) => {
	const isEdgeNode = false;
	if (!message && first)
		message = `Could not connect to the ${
			isEdgeNode ? 'node' : 'container'
		}. Do you have sufficient privileges?`;
	if (!message) message = 'disconnected';
	if (!first.value) message = `\r\n${message}`;
	if (first.value) termA.value.reset();

	termA.value.write(`\x1b[31m${message}\x1b[m\r\n`);
};

const onWSError = (ex: any) => {
	initTimer.value && clearInterval(initTimer.value);
	fatal(ex.message);
};

const createWS = () => {
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const websocketUrl = `${protocol}//${window.location.host}/kapis/terminal.kubesphere.io/v1alpha2/nodes/${props.node}/exec`;
	return new ReconnectingWebSocket(websocketUrl);
};

const fitTerm = () => {
	if (props.visible) {
		fitAddon && fitAddon.fit();
		resizeRemoteTerminal();
	}
};

const onResize = debounce(fitTerm, 300);

const onTerminalResize = () => {
	window.addEventListener('resize', onResize);
};

const packStdin = (data: any) =>
	JSON.stringify({
		Op: 'stdin',
		Data: data
	});

const sendTerminalInput = (data: any) => {
	if (isWsOpen()) {
		ws.value.send(packStdin(data));
	} else {
		messageQueue.push(packStdin(data));
	}
};

const onTerminalKeyPress = () => {
	termA.value.onData(sendTerminalInput);
};

const focus = () => {
	if (isWsOpen()) {
		termA.value.focus();
	}
};

const show = () => {
	aceVisileb.value = true;

	termA.value = initTerm();
	ws.value = createWS();
	ws.value.addEventListener('open', () => {
		if (inited) {
			termTemp && termTemp.write('\r\n');
		} else {
			inited = true;
		}
		messageQueue.forEach((msg) => ws.value.send(msg));
		messageQueue = [];
	});

	ws.value.addEventListener('message', onWSReceive);

	ws.value.addEventListener('error', onWSError);
	ws.value.addEventListener('close', (err: any) => console.log(err));

	onTerminalResize();
	onTerminalKeyPress();

	disableTermStdin();
};

const hide = () => {
	clearHeartBeat();
	termA.value && termA.value.destroy && termA.value.destroy();
	ws.value && ws.value.close(1000);
	first.value = true;
	disconnect();
	removeResizeListener();
	initTimer.value && clearInterval(initTimer.value);

	aceVisileb.value = false;
};

const DEFAULT_TERMINAL_OPTS = {
	backgroundOpacity: 0.9,
	cursorBlink: true,
	fontFamily: '"Fira Code", "DejaVu Sans Mono", monospace',
	fontSize: 14,
	letterSpacing: 0.8,
	lineHeight: 1.2,
	theme: {
		background: '#181d28'
	}
};

const getTerminalOpts = () => {
	return DEFAULT_TERMINAL_OPTS;
};

const renderConnecting = (term: any, initText: any) => {
	let count = 0;
	const timer = setInterval(() => {
		term.reset();
		term.write(`${initText}${'.'.repeat(++count)}`);
		if (count > 2) {
			count = 0;
		}
	}, 500);
	return timer;
};

const removeResizeListener = () => {
	window.removeEventListener('resize', onResize);
};

const clearHeartBeat = () => {
	locker && clearInterval(locker);
};
const setHeartBeat = () => {
	clearHeartBeat();
	locker = setInterval(() => {
		ws.value.send(JSON.stringify({ op: 'stdin' }));
	}, 20 * 1000);
};
const initTerm = () => {
	const initText = 'connecting';
	const terminalOpts = getTerminalOpts();

	termTemp = new Terminal(terminalOpts);
	fitAddon = new FitAddon();
	termTemp.loadAddon(fitAddon);
	termTemp.loadAddon(new WebLinksAddon());
	if (terminalRef.value) {
		termTemp.open(terminalRef.value);
		fitAddon.fit();
	}

	initTimer.value = renderConnecting(termTemp, initText);

	return termTemp;
};

watch(
	() => props.visible,
	(status) => {
		if (status) {
			nextTick(() => {
				fitTerm();
				focus();
			});
		}
	}
);

onMounted(() => {
	show();
});

onBeforeUnmount(() => {
	hide();
});

defineExpose({ focus: focus });
</script>

<style lang="scss" scoped>
.terminal-content {
	overflow: hidden;
	height: 100%;
	background: #181d28;
	.terminal-content-wrapper {
		position: absolute;
		inset: 20px;
		top: 12px;
		bottom: 12px;
	}
}
</style>
