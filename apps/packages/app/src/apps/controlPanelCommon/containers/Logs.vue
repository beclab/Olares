<template>
	<div class="my-scroll-container" ref="logContainerRef">
		<div class="log-container" style="border-radius: 4px">
			<bt-scroll-area class="log-scroll-wrapper" ref="scrollAreaRef">
				<div class="log-content">
					<div
						class="log-view-more"
						:style="{ marginLeft: `${scrollAreaWidth / 2}px` }"
					>
						<span
							flat
							dense
							@click="onLoadMore"
							v-show="tailLines <= logsData.length"
						>
							{{ t('VIEW_MORE') }}
						</span>
					</div>
					<div v-for="item in logsData" :key="item">
						<div class="log-item" v-html="converter.ansi_to_html(item)"></div>
					</div>
				</div>
				<div id="box"></div>
				<Empty
					v-show="logsData.length <= 1 && !loading"
					style="color: #bbb"
					center
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
		<q-inner-loading class="logs-loading-wrapper" :showing="loading">
		</q-inner-loading>
	</div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import {
	getContainersLogs,
	getNameSpacePodsList,
	getWorkloadList
} from '../network';
import {
	nextTick,
	onBeforeUnmount,
	onMounted,
	reactive,
	ref,
	watch
} from 'vue';
import Empty from '../components/Empty.vue';
import { PATTERN_UTC_TIME } from '../utils/constants';
import { saveAs } from 'file-saver';
import { AnsiUp } from 'ansi_up';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface Props {
	theme?: 'dark' | 'light';
	fullscreen?: boolean;
	container?: string;
}

const props = withDefaults(defineProps<Props>(), {
	theme: 'dark',
	fullscreen: false
});

const converter = new AnsiUp();

const logsData = ref<any[]>([]);
const loading = ref(false);
const route = useRoute();
const logContainerRef = ref();
const scrollAreaRef = ref();
const scrollAreaWidth = ref(0);
const isRealtime = ref(false);
const perPageCount = 1000;
const tailLines = ref(perPageCount);
let element: any = null;
const isIframe = ref(self.top !== self);

const namespace = ref();
const name = ref();
const containerWrap = ref();

const onLoadMore = () => {
	tailLines.value = tailLines.value + perPageCount;
	fetchData(true, true);
};

const fetchData = (showLoading = true, loadMore = false) => {
	const params = {
		namespace: namespace.value,
		podName: name.value,
		container: containerWrap.value,
		tailLines: tailLines.value,
		timestamps: true
		// follow: isRealtime.value,
		// ...param,
	};

	if (showLoading) {
		loading.value = true;
	}
	getContainersLogs(params)
		.then((res) => {
			const data = String(res.data)
				.replace(/\\r\\n/g, '\n')
				.split('\n');

			logsData.value = data.map((text, index) => {
				const match = text.match(PATTERN_UTC_TIME);
				const key = match ? match[0] : index;
				const content = match ? text.replace(match[0], '') : text;
				return content;
			});
			if (isRealtime.value) {
				setLock();
			}
			console.log('loadMore', loadMore);
			if (!loadMore) {
				nextTick(() => {
					scrollToBottom();
				});
			}
		})
		.finally(() => {
			loading.value = false;
		});
};

const { kind }: Record<string, any> = route.params;

const kindType = {
	deployment: {
		kind: 'deployments',
		ownerKind: 'ReplicaSet'
	},
	daemonset: {
		kind: 'daemonsets',
		ownerKind: 'DaemonSet'
	},
	statefulset: {
		kind: 'statefulsets',
		ownerKind: 'StatefulSet'
	}
} as any;

const fetchDeployment = () => {
	const { deployment, container, kind }: Record<string, any> = route.params;
	const params = {
		name: deployment
	};
	loading.value = true;

	getWorkloadList(params, kindType[kind].kind)
		.then((res) => {
			const target = res.data.items[0];
			fetchPods(target);
		})
		.catch(() => {
			loading.value = false;
		});
};

const fetchPods = (data: any) => {
	const matchLabels = data.spec.selector.matchLabels;
	const labels = Object.keys(matchLabels).map(
		(key) => `${key}=${matchLabels[key]}`
	);
	const { kind }: Record<string, any> = route.params;

	const params = {
		namespace: data.metadata.namespace,
		ownerKind: kindType[kind].ownerKind,
		labelSelector: labels.join('')
	};
	getNameSpacePodsList(params)
		.then((res) => {
			const target = res.data.items[0];
			namespace.value = target.metadata.namespace;
			name.value = target.metadata.name;
			const containerTarget = target.spec.containers.find(
				(item) => item.name === route.params.container
			);
			containerWrap.value = containerTarget?.name;

			fetchData();
		})
		.catch(() => {
			loading.value = false;
		});
};

onMounted(() => {
	element = document.getElementById('box');
});

const scrollToBottom = () => {
	element && element.scrollIntoView();
};

const handleDownload = async () => {
	const params = {
		namespace: namespace.value,
		podName: name.value,
		container: containerWrap.value
	};
	const result = await getContainersLogs(params);
	const blob = new Blob([result.data], { type: 'text/plain;charset=utf-8' });
	saveAs(blob, `${containerWrap.value}.log`);
};

const openTab = () => {
	const url = `/container/logs/v2/${namespace.value}/${name.value}/${containerWrap.value}?theme=${props.theme}`;
	window.open(url);
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
	scrollAreaWidth.value =
		scrollAreaRef.value.$el.clientWidth || window.innerWidth;
});

onBeforeUnmount(() => {
	clearLock();
});

function init() {
	namespace.value = route.params.namespace;
	name.value = route.params.name;
	containerWrap.value = route.params.container || props.container;
	if (namespace.value && name && containerWrap) {
		fetchData(true, false);
	} else if (route.params.kind) {
		fetchDeployment();
	}
}

watch(() => route.params.container, init);

init();

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

const type = route.query.theme ? route.query.theme : props.theme;
const themeVar = reactive(themeStyle[type as Props['theme']]);
</script>

<style lang="scss" scoped>
.my-scroll-container {
	--fontWeight: v-bind(themeVar.fontWeight);
	--fontColor: v-bind(themeVar.fontColor);
	--contentBG: v-bind(themeVar.contentBG);
	--toolBG: v-bind(themeVar.toolBG);
	--toolBorderColor: v-bind(themeVar.toolBorderColor);

	height: 100vh;
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
			padding: 20px;
			font-size: 12px;
			min-height: 120px;
			// font-family: PT Mono, Monaco, Menlo, Consolas, Courier New, monospace;
			.log-item {
				line-height: 20px;
				color: var(--fontColor);
				font-weight: var(--fontWeight);
				line-height: 20px;
				white-space: pre;
			}
		}
	}
	::v-deep(.log-scroll-wrapper) {
		height: 100%;
	}
	::v-deep(.logs-loading-wrapper) {
		background: var(--contentBG);
		color: white;
	}
}
</style>
