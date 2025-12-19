<template>
	<q-icon
		color="blue-default"
		name="sym_r_terminal"
		flat
		dense
		size="16px"
		@click="yamlShow"
	>
		<slot></slot>
	</q-icon>
	<Dialog :title="data.container" v-model="visible2" @show="show" @hide="hide">
		<div class="fit relative-position">
			<div class="absolute-full" style="border-radius: 4px; overflow: hidden">
				<div class="relative-position terminal-content">
					<div class="terminal-content-wrapper">
						<div style="height: calc(100%)" ref="terminalRef"></div>
					</div>
				</div>
			</div>
		</div>
	</Dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import '@xterm/xterm/css/xterm.css';
import ReconnectingWebSocket from 'reconnecting-websocket';
import { debounce } from 'lodash';
import Dialog from '../components/Dialog/Dialog.vue';

interface Props {
	data: {
		namespace: string;
		podName: string;
		container: string;
	};
}
const props = withDefaults(defineProps<Props>(), {});

const visible2 = ref(false);
const aceVisileb = ref(false);
let messageQueue: string[] = [];

const yamlShow = () => {
	visible2.value = true;
};
const termA = ref<any>(null);
const ws = ref();
const initTimer = ref();
const first = ref(true);

const terminalRef = ref<any>();
let locker: NodeJS.Timeout | undefined = undefined;
const isWsOpen = () => {
	return ws.value && ws.value.readyState === ws.value.OPEN;
};

const packResize = (col: any, row: any) =>
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
		ws.value.close(true);
	}
};

const fatal = (message: any) => {
	// const { isEdgeNode } = this.props;
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
	const websocketUrl = `${protocol}//${window.location.host}/kapis/terminal.kubesphere.io/v1alpha2/namespaces/${props.data.namespace}/pods/${props.data.podName}/exec?container=${props.data.container}&shell=sh`;
	return new ReconnectingWebSocket(websocketUrl);
};

const fitTerm = () => termA.value.fit();

const onResize = debounce(fitTerm, 800);

const onTerminalResize = () => {
	window.addEventListener('resize', onResize);
	termA.value.onResize(resizeRemoteTerminal);
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

const show = () => {
	aceVisileb.value = true;

	termA.value = initTerm();
	ws.value = createWS();
	ws.value.addEventListener('open', () => {
		messageQueue.forEach((msg) => ws.value.send(msg));
		messageQueue = [];
	});

	ws.value.addEventListener('message', onWSReceive);
	ws.value.addEventListener('error', onWSError);

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
	lineHeight: 1.2,
	cursorBlink: true,
	// cursorStyle: 'underline',
	fontSize: 12,
	fontFamily: "Monaco, Menlo, Consolas, 'Courier New', monospace",
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

	const termTemp: any = new Terminal(terminalOpts);
	const fitAddon = new FitAddon();
	termTemp.loadAddon(fitAddon);
	termTemp.loadAddon(new WebLinksAddon());
	if (terminalRef.value) {
		termTemp.open(terminalRef.value);
		fitAddon.fit();
	}

	initTimer.value = renderConnecting(termTemp, initText);

	return termTemp;
};
</script>

<style lang="scss" scoped>
.terminal-content {
	background: rgb(24, 29, 40);
	border-radius: 4px;
	overflow: hidden;
	height: 100%;
	.terminal-content-wrapper {
		position: absolute;
		top: 8px;
		left: 8px;
		right: 8px;
		bottom: 8px;
	}
}
</style>
