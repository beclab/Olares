<template>
	<bt-custom-dialog
		ref="customRef"
		size="large"
		:title="t('recommendation.logs')"
		:ok="false"
		:cancel="false"
	>
		<div class="my-scroll-container" ref="logContainerRef">
			<div class="log-container">
				<bt-scroll-area class="log-scroll-wrapper" ref="scrollAreaRef">
					<log-empty v-if="!logsData && !loading" style="color: #bbb" center />
					<div
						v-else
						class="log-content"
						v-html="converter.ansi_to_html(logsData)"
					/>
				</bt-scroll-area>
			</div>
			<div class="logs-tool-container row">
				<q-btn flat dense @click="handleRealtime">
					<q-icon
						size="24px"
						:color="themeVar.toolIconColor"
						name="stop"
						v-if="isRealtime"
					/>
					<q-icon
						size="24px"
						:color="themeVar.toolIconColor"
						name="play_arrow"
						v-else
					/>
				</q-btn>
				<q-separator spaced inset vertical :color="themeVar.splitColor" />

				<q-btn flat dense @click="refresh" :disable="isRealtime">
					<q-icon size="24px" :color="themeVar.toolIconColor" name="refresh" />
				</q-btn>
				<q-separator spaced inset vertical :color="themeVar.splitColor" />
				<q-btn flat dense @click="handleDownload">
					<q-icon size="24px" :color="themeVar.toolIconColor" name="download" />
				</q-btn>
				<template v-if="isIframe">
					<q-separator spaced inset vertical :color="themeVar.splitColor" />
					<q-btn flat dense @click="openTab">
						<q-icon size="24px" :color="themeVar.toolIconColor" name="launch" />
					</q-btn>
				</template>
				<template v-if="fullscreen">
					<q-separator spaced inset vertical :color="themeVar.splitColor" />
					<q-btn
						:color="themeVar.toolIconColor"
						flat
						dense
						@click="toggle"
						:icon="$q.fullscreen.isActive ? 'fullscreen_exit' : 'fullscreen'"
					/>
				</template>
			</div>
			<q-inner-loading
				:dark="themeVar.loadingDark"
				:color="themeVar.loadingColor"
				:showing="loading"
			>
			</q-inner-loading>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { saveAs } from 'file-saver';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { AnsiUp } from 'ansi_up';
import { PropType } from 'vue/dist/vue';
import { NodeStatus, useArgoStore, WorkflowDetail } from 'src/stores/argo';
import LogEmpty from './LogEmpty.vue';

const props = defineProps({
	workflow: {
		type: Object as PropType<WorkflowDetail>,
		required: true
	},
	nodeStatus: {
		type: Object as PropType<NodeStatus>,
		required: true
	},
	theme: {
		type: String,
		default: 'dark'
	},
	fullscreen: {
		type: Boolean,
		default: false
	}
});

const { t } = useI18n();
const customRef = ref();
const converter = new AnsiUp();
const logsData = ref<any[]>([]);
const loading = ref(false);
const logContainerRef = ref();
const scrollAreaRef = ref();
const scrollAreaWidth = ref(0);
const isRealtime = ref(false);
// const perPageCount = 1000;
// const tailLines = ref(perPageCount);
let element: any = null;
const isIframe = ref(self.top !== self);
const argoStore = useArgoStore();
// const onLoadMore = () => {
// 	tailLines.value = tailLines.value + perPageCount;
// 	fetchData(true, true);
// };

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const fetchData = async (showLoading = true, _loadMore = false) => {
	if (showLoading) {
		loading.value = true;
	}

	argoStore
		.getArchivedLog(
			argoStore.namespace,
			props.workflow.metadata.uid,
			props.nodeStatus.id
		)
		.then((data) => {
			logsData.value = data;
			loading.value = false;
			// const array = data
			// 	.split('\n')
			// 	.filter((block) => block.trim() !== '')
			// 	.map((block) => {
			// 		try {
			// 			return JSON.parse(block) as ArtifactLogEntry;
			// 		} catch (e) {
			// 			console.log(block);
			// 		}
			// 	});
		});
};

onMounted(() => {
	element = document.getElementById('box');
});

const scrollToBottom = () => {
	element && element.scrollIntoView();
};

const handleDownload = async () => {
	const result = await argoStore.getArchivedLog(
		argoStore.namespace,
		props.workflow.metadata.uid,
		props.nodeStatus.id
	);
	const blob = new Blob([result], { type: 'text/plain;charset=utf-8' });
	saveAs(
		blob,
		`${argoStore.namespace}-${props.workflow.metadata.uid}-${props.nodeStatus.id}.log`
	);
};

const openTab = () => {
	window.open(
		argoStore.getArchivedLogUrl(
			argoStore.namespace,
			props.workflow.metadata.uid,
			props.nodeStatus.id
		)
	);
};

const refresh = () => {
	fetchData();
};

const handleRealtime = () => {
	isRealtime.value = !isRealtime.value;
	scrollToBottom();
	fetchData(false);
	if (isRealtime.value) {
		setLock();
	} else {
		clearLock();
	}
};

const locker = ref();
const clearLock = () => {
	locker.value && clearTimeout(locker.value);
};
const setLock = () => {
	clearLock();
	locker.value = setTimeout(() => {
		fetchData(false);
	}, 5 * 1000);
};

const $q = useQuasar();
const toggle = () => {
	const target = logContainerRef.value;
	$q.fullscreen.toggle(target);
};

onMounted(() => {
	scrollAreaWidth.value = window.innerWidth;

	fetchData(true, false);
});

onBeforeUnmount(() => {
	clearLock();
});

const themeStyle = {
	light: {
		loadingDark: false,
		loadingColor: 'dark',
		fontWeight: 500,
		fontColor: '#303133',
		contentBG: '#f8f8f8',
		toolBG: '#ffffff',
		toolBorderColor: 'rgba(0,0,0,0.1)',
		splitColor: '',
		toolIconColor: 'grey-6'
	},
	dark: {
		loadingDark: true,
		loadingColor: 'white',
		fontWeight: 600,
		fontColor: '#b7c4d1',
		contentBG: '#242e42',
		toolBG: '#36435c',
		toolBorderColor: '',
		splitColor: 'grey-6',
		toolIconColor: 'grey-5'
	}
};

const themeVar = reactive(themeStyle[props.theme]);
</script>

<style lang="scss" scoped>
.my-scroll-container {
	--fontWeight: v-bind(themeVar.fontWeight);
	--fontColor: v-bind(themeVar.fontColor);
	--contentBG: v-bind(themeVar.contentBG);
	--toolBG: v-bind(themeVar.toolBG);
	--toolBorderColor: v-bind(themeVar.toolBorderColor);
	height: 520px;
	position: relative;

	.logs-tool-container {
		position: absolute;
		right: 12px;
		top: 8px;
		z-index: 2;
		padding: 0px 2px;
		border-radius: 8px;
		background: var(--toolBG);
		border: 1px solid var(--toolBorderColor);
	}

	.log-container {
		height: 100%;
		background: var(--contentBG);

		.log-view-more {
			font-size: 12px;
			color: var(--fontColor);
			cursor: pointer;
			height: 20px;
			translate: transformX(-50%);
		}

		.log-content {
			position: relative;
			padding: 40px 20px 20px;
			font-size: 12px;
			min-height: 120px;
			color: var(--fontColor);
			font-weight: var(--fontWeight);
			line-height: 20px;
			white-space: pre;
		}
	}

	::v-deep(.log-scroll-wrapper) {
		height: 100%;
	}
}
</style>
