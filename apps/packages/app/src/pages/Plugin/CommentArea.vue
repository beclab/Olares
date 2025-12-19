<template>
	<bt-scroll-area
		ref="scrollArea"
		class="scroll-area-wrapper"
		style="height: calc(100vh - 242px); width: 100%; overflow-anchor: none"
	>
		<div class="q-py-xl q-px-lg" style="overflow-anchor: auto">
			<search-default v-if="helloShow" />
			<div class="q-gutter-y-lg">
				<template v-for="item in list" :key="item.id">
					<QueryItem :text="item.query" />
					<MarkDownPreview
						v-show="!item.custom"
						:text="item.answer || ''"
						class="q-px-lg q-py-md"
						:class="item.isError ? 'bg-red-1' : 'bg-grey-1'"
					>
					</MarkDownPreview>
				</template>
				<div v-if="!message_end">
					<MarkDownPreview
						:text="text"
						class="q-px-lg bg-grey-1 q-py-md"
						:loading="loading"
						:ended="message_end"
					>
						<div class="q-py-md">
							<q-spinner-dots
								color="grey-5"
								size="xs"
								style="line-height: 16px; height: 16px"
							/>
						</div>
					</MarkDownPreview>
				</div>
			</div>
		</div>
	</bt-scroll-area>

	<div class="comment-area q-ma-lg q-mt-md">
		<div class="footer row items-center justify-between">
			<div class="footer-left">
				<div
					class="row items-center justify-between no-wrap text-grey-10 ai-select q-mr-xs"
					@click="dialog = true"
				>
					<div class="row items-center no-wrap ai-select-left">
						<!-- <div style="flex: 0 0 24px">
							<div
								style="width: 24px; height: 24px; border-radius: 12px"
								:style="`background:${active.icon_background}`"
								class="row items-center justify-center"
							>
								{{ active.icon }}
							</div>
						</div>

						<div style="flex: 0 0 24px">
							<q-img
								:src="active.img"
								:ratio="1"
								width="24px"
								class="q-mr-xs"
								spinner-size="xs"
							/>
						</div> -->

						<div class="ellipsis text-body3 text-grey-10 q-ml-sm">
							{{ active.name }}
						</div>
					</div>
					<q-icon name="sym_r_expand_more" size="sm" />
				</div>

				<!-- <div class="event-continar q-mr-xs">
					<q-icon name="sym_r_chrome_reader_mode" size="20px" />
				</div>
				<div class="event-continar q-mr-xs">
					<q-icon name="sym_r_add_photo_alternate" size="20px" />
				</div>
				<div class="event-continar q-mr-xs">
					<q-icon name="sym_r_note_add" size="20px" />
				</div>
				<div class="event-continar q-mr-xs">
					<q-icon name="sym_r_content_cut" size="20px" />
				</div> -->
			</div>
			<img
				class="footer-right cursor-pointer"
				:src="getPluginImageUrl('add-comment.svg')"
				@click="chatModelChange"
			/>
		</div>
		<div class="comment">
			<q-input
				type="textarea"
				borderless
				input-class="comment-input"
				v-model="query"
				input-style="line-height: 12px;"
				@keydown.enter="handleEnter"
			/>
			<div class="row items-center justify-between comment-send">
				<div class="row items-center justify-center">
					<!-- <q-icon
						name="sym_o_stacks"
						size="20px"
						class="q-ml-md cursor-pointer"
					/>
					<q-icon
						name="sym_o_alternate_email"
						size="20px"
						class="q-ml-md cursor-pointer"
					/> -->
				</div>
				<q-btn
					flat
					dense
					@click="send"
					class="q-mr-xs"
					:disabled="sendDisabled"
				>
					<q-icon name="sym_r_send" size="20px" />
				</q-btn>
			</div>
		</div>
	</div>
	<MyBottomSheet v-model="dialog" title="All Bots">
		<ChatModelOptions
			v-model="active"
			:options="options"
			@update:model-value="chatModelChange"
			@delete="deleteHandler"
		>
		</ChatModelOptions>
	</MyBottomSheet>
</template>

<script lang="ts" setup>
import { computed, nextTick, onBeforeMount, ref, watch } from 'vue';
import { ai as api } from '../../api';
import MarkDownPreview from './components/MarkDownPreview.vue';
import QueryItem from './components/QueryItem.vue';
import { Typewriter } from './components/typewriter';
import SearchDefault from './SearchDefault.vue';
import MyBottomSheet from './components/MyBottomSheet.vue';
import ChatModelOptions from './components/ChatModelOptions.vue';
import { createAppsList, modelImg } from './config';
import { getPluginImageUrl } from '../../stores/bex-url';

const query = ref();
const helloShow = ref(true);
const dialog = ref(false);

// interface AppItem {
// 	api_base_url: string;
// 	app_url: string;
// 	created_at: number;
// 	enable_api: boolean;
// 	enable_site: boolean;
// 	icon: string;
// 	icon_background: string;
// 	id: string;
// 	is_demo: boolean;
// 	mode: string;
// 	model_config: any;
// 	name: string;
// }

const text = ref('');
const textCache = ref('');
const loading = ref(false);
const apps = ref([{ name: '', id: '', img: '' }]);
// const app = ref<any>();
const task_id = ref();
const scrollArea = ref();
const list = ref();
const message_end = ref(true);
const conversation_id = ref('');
// const tempChatId = 'test10';
const CHAT_MESSAGE_ID = 'chat-message-id';
const options = computed(() => apps.value);
const active = ref<any>(options.value[0]);
const modeChat = computed(() => true); //active.value.mode === 'chat'

const typewriter = new Typewriter((str: string) => {
	text.value += str || '';
});

const getInputs = (app: { mode: any }, value: any) => {
	let params = {};
	switch (app.mode) {
		case 'completion':
			params = {
				target_language: '中文',
				query: value
			};
	}
	return params;
};

const getParams = (app: { model_config: any }, value: any) => {
	const inputs = getInputs(app as any, value);

	let params: any = {
		inputs,
		model_config: app.model_config,
		response_mode: 'streaming'
	};
	// if (app.mode === 'chat') {
	params = {
		...params,
		conversation_id: conversation_id.value,
		query: value
	};
	// }
	return params;
};

const receiveMessageRef = ref(false);

const onmessage = async (ev: { data: string }) => {
	receiveMessageRef.value = true;
	const data = JSON.parse(ev.data);
	if (data.event === 'agent_message' || data.event === 'message') {
		typewriter.add(data.answer);
		textCache.value += data.answer;
		task_id.value = data.task_id;
		conversation_id.value = data.conversation_id;
		loading.value = false;
	} else if (data.event === 'message_end') {
		if (modeChat.value) {
			const data2 = await api.getMessageById(
				active.value.id,
				'chat',
				conversation_id.value
			);
			list.value = data2.data.data;
		} else {
			list.value = list.value.map((item: any, index: number) => {
				if (list.value.length - 1 === index) {
					return { ...item, answer: textCache.value, custom: false };
				}
				return { ...item, custom: false };
			});
		}
		message_end.value = true;
		typewriter.done();
		reset();
	}
};

const reset = () => {
	text.value = '';
	textCache.value = '';
};

const onerror = () => {
	loading.value = false;
	message_end.value = true;
};

const onclose = () => {
	if (!receiveMessageRef.value) {
		loading.value = false;
		message_end.value = true;
	}
};

const onopen = (res: Response) => {
	if (res.status != 200) {
		const clonedResponse = res.clone();
		const reader = clonedResponse.body?.getReader();
		const decoder = new TextDecoder();

		let totalErrorJsonData = '';

		reader?.read().then(function processText({ value, done }) {
			if (done) {
				if (totalErrorJsonData.length) {
					const errorObj = JSON.parse(totalErrorJsonData);
					if (list.value.length > 0) {
						const lastItem = list.value[list.value.length - 1];
						list.value[list.value.length - 1] = {
							...lastItem,
							answer: errorObj.message,
							custom: false,
							isError: true
						};
					}
					scrollToBottom();
				}
				return;
			}
			const text = decoder.decode(value);
			totalErrorJsonData = totalErrorJsonData + text;

			return reader.read().then(processText);
		});

		// message_end.value = true;
		// controller.abort();
	}
};

// onopen?: (response: Response) => Promise<void>;

const sendDisabled = computed(() => !message_end.value);

const send = () => {
	const queryValue = query.value;
	query.value = '';

	helloShow.value = false;
	if (sendDisabled.value) {
		return;
	}
	if (conversation_id.value) {
		controller.abort();
	}
	message_end.value = false;
	const data = list.value || [];
	list.value = data.concat([
		{ id: CHAT_MESSAGE_ID, query: queryValue, custom: true }
	]);
	scrollToBottom();
	search(queryValue);
};

const getApps = async () => {
	const res = await api.getApps();
	let data = res.data.data;
	// let target: any = undefined;
	if (!data) {
		const arr = createAppsList.map((item) => api.createApp(item.params));
		const res = await Promise.all(arr);
		data = res.map((item) => item.data);
	}
	data.sort(
		(a: { icon: string }, b: { icon: string }) =>
			parseInt(a.icon.split('-')[1]) - parseInt(b.icon.split('-')[1])
	);
	apps.value = data.map((item: { icon: string | number }) => ({
		...item,
		img: modelImg[item.icon]
	}));
	active.value = apps.value[0];
};

let controller: any = undefined;
const search = async (value: any) => {
	typewriter.start();
	loading.value = true;
	controller = new AbortController();

	if (active.value) {
		const res = await api.getAppById(active.value.id);
		const params = getParams(res.data, value);
		const callbacks = {
			onmessage,
			onerror,
			onclose,
			onopen,
			signal: controller.signal
		};
		receiveMessageRef.value = false;
		// active.value.mode
		api.getMessage2(params, active.value.id, 'chat', callbacks);
	}
};

const deleteHandler = async (data: { id: any }) => {
	await api.deleteApp(data.id);
	getApps();
};

const resetAll = () => {
	list.value = undefined;
	task_id.value = '';
	conversation_id.value = '';
	message_end.value = true;
	helloShow.value = true;
	query.value = '';
	reset();
};

const chatModelChange = () => {
	dialog.value = false;
	resetAll();
	controller.abort();
};

const scrollToBottom = () => {
	nextTick(() => {
		const scrollTarget = scrollArea.value?.$refs?.scrollRef?.getScrollTarget();
		const duration = 0;

		scrollArea.value &&
			scrollArea.value?.$refs?.scrollRef?.setScrollPosition(
				'vertical',
				scrollTarget.scrollHeight,
				duration
			);
	});
};

watch(
	() => text.value.length,
	() => {
		scrollToBottom();
	}
);

const handleEnter = (event: KeyboardEvent) => {
	if (event.ctrlKey || event.shiftKey || event.altKey) {
		event.preventDefault();
		send();
	} else {
		query.value = query.value + '\n';
	}
};

onBeforeMount(async () => {
	getApps();
});
</script>

<style lang="scss">
.ai-select {
	height: 40px;
	border-radius: 20px;
	padding: 0 8px;
	background: rgba(246, 246, 246, 1);
	position: relative;
	cursor: pointer;
	.ai-select-left {
		max-width: 120px;
		margin-right: 14px;
	}
}
.comment-input {
	height: 100px !important;
	padding: 8px 12px !important;
	resize: none !important;
}
</style>

<style lang="scss" scoped>
.comment-area {
	width: calc(100% - 40px);
	position: absolute;
	bottom: 0;
}

.footer {
	margin-bottom: 8px;
	.footer-left {
		display: flex;
		align-items: center;
		justify-content: center;

		.event-continar {
			width: 24px;
			height: 24px;
			border-radius: 4px;
			display: flex;
			align-items: center;
			justify-content: center;
			cursor: pointer;
			&:hover {
				background-color: rgba(0, 0, 0, 0.05);
			}
		}
	}
	.footer-right {
		width: 32px;
		height: 32px;
	}
}

.comment {
	width: 100%;
	height: 140px;
	border-radius: 12px;
	padding-top: 4px;
	background: $grey-1;

	.comment-send {
	}
}
.scroll-area-wrapper {
	& > ::v-deep(.q-scrollarea__container) {
		min-width: 0px;
		overflow-x: hidden;
	}
}
</style>
