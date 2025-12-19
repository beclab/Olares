<template>
	<div
		class="google-search-container"
		:style="{ opacity: transparent ? 0 : 1 }"
	>
		<q-card flat bordered class="my-card bg-white">
			<q-card-section class="q-pa-none">
				<div class="row items-center no-wrap">
					<q-img
						:src="vaultPluginIcon"
						:ratio="1"
						width="24px"
						spinner-size="xs"
					/>
					<div class="col q-ml-sm">
						<div class="text-subtitle1 text-grey-10">LarePass</div>
					</div>

					<div class="col-auto">
						<q-img
							:src="vaultPluginOperation"
							height="20px"
							width="132px"
							spinner-color="primary"
							spinner-size="xs"
							style="cursor: pointer"
							@click="refresh"
						/>
					</div>
				</div>
			</q-card-section>
			<!-- <q-btn
				color="primary"
				icon="check"
				label="chat"
				@click="search('你是谁')"
			/>
			<q-btn
				color="primary"
				icon="check"
				label="stop"
				@click="stopChat"
			/>
      <q-btn color="primary" icon="check" label="cancel" @click="cancelClick" /> -->

			<!-- {{ userStore.current_user }} -->

			<!-- {{ task_id }} -->
			<MyTabs
				v-model="tab"
				dense
				:options="options"
				no-caps
				class="q-mt-md"
				active-class="my-tabs-active-class"
				content-class="my-tabs-content-class"
			>
				<q-tab label="Default" name="one"> </q-tab>
				<q-tab label="Search Copilot" name="two" />
			</MyTabs>
			<q-tab-panels v-model="tab" animated class="my-tab-panels-wrapper">
				<q-tab-panel name="one" class="my-tab-panel-wrapper">
					<SearchDefault
						:text="text"
						:loading="loading"
						:ended="message_end"
					></SearchDefault>
				</q-tab-panel>

				<q-tab-panel name="two" class="my-tab-panel-wrapper">
					<div v-show="loading" class="panel-loading row justify-center">
						<q-spinner-dots color="grey-5" size="xs" />
					</div>
					<SearchCopilot> </SearchCopilot>
				</q-tab-panel>
			</q-tab-panels>
		</q-card>
	</div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
// import { useQuasar } from 'quasar';
import SearchDefault from './components/SearchDefault.vue';
import SearchCopilot from './components/SearchCopilot.vue';
import vaultPluginIcon from 'assets/images/vault-plugin.svg';
import vaultPluginOperation from 'assets/images/vault-plugin-operation.svg';
import Image65 from 'assets/images/image65.png';
import Image68 from 'assets/images/image68.png';
import MyTabs from './components/MyTabs.vue';
import { ai as api } from '../../api';
import { Typewriter } from './components/typewriter';
import { createAppsGoogleSearch } from './config';
import urlUtils from '../../utils/url';

// const $q: any = useQuasar();
const transparent = ref(true);
interface AppItem {
	api_base_url: string;
	app_url: string;
	created_at: number;
	enable_api: boolean;
	enable_site: boolean;
	icon: string;
	icon_background: string;
	id: string;
	is_demo: boolean;
	mode: string;
	model_config: any;
	name: string;
}

const options = [
	{ label: 'Default', value: 'one', img: Image65 },
	{ label: 'Search Copilot', value: 'two', img: Image68 }
];

// const tempChatId = 'test6';

const tab = ref('one');
const text = ref('');
const name = ref();
const loading = ref(false);
const task_id = ref();
const message_end = ref(true);
const app = ref<AppItem>();

let controller: any = undefined;

const typewriter = new Typewriter((str: string) => {
	text.value += str || '';
});

const onmessage = (ev) => {
	const data = JSON.parse(ev.data);
	if (data.event === 'message') {
		loading.value = false;
		typewriter.add(data.answer);
		task_id.value = data.task_id;
	} else if (data.event === 'message_end') {
		message_end.value = true;
		typewriter.done();
		controller && controller.abort();
	}
};

const onerror = () => {
	loading.value = false;
};

const onclose = () => {
	message_end.value = true;
};

// const stopChat = async () => {
// 	const data = api.stopChat(task_id.value);
// };

const getParams = (app, value) => {
	const model_config = {
		pre_prompt: '',
		prompt_type: 'simple',
		chat_prompt_config: {},
		completion_prompt_config: {},
		user_input_form: [],
		dataset_query_variable: '',
		opening_statement: '',
		more_like_this: {
			enabled: false
		},
		suggested_questions_after_answer: {
			enabled: false
		},
		speech_to_text: {
			enabled: false
		},
		retriever_resource: {
			enabled: false
		},
		sensitive_word_avoidance: {
			enabled: false,
			type: '',
			configs: []
		},
		external_data_tools: [],
		agent_mode: {
			enabled: true,
			tools: []
		},
		dataset_configs: {
			retrieval_model: 'single'
		},
		file_upload: {
			image: {
				enabled: false,
				number_limits: 3,
				detail: 'high',
				transfer_methods: ['remote_url', 'local_file']
			}
		},
		annotation_reply: {
			enabled: false
		}
	};
	return {
		conversation_id: '',
		model_config: { ...model_config, ...app.model_config },
		query: value,
		// conversation_id: '570fc26a-80d7-4d07-988d-84a1a3191b44',
		inputs: {},
		response_mode: 'streaming'
	};
};

const search = async (value) => {
	message_end.value = false;
	typewriter.start();
	controller = new AbortController();

	text.value = '';
	loading.value = true;

	const data = await api.getApps();
	app.value = data.data.data?.find(
		(item) => item.name === createAppsGoogleSearch.name
	);
	if (!app.value) {
		const params2 = createAppsGoogleSearch.params;
		const data2 = await api.createApp(params2);
		app.value = data2.data;
	}

	if (app.value) {
		const params = getParams(app.value, value);
		const callbacks = {
			onmessage,
			onerror,
			onclose,
			signal: controller.signal
		};
		api.getMessage2(params, app.value.id, app.value.mode, callbacks);
	}
};

const refresh = () => {
	controller && controller.abort();
	search(name.value);
};
// const cancelClick = () => {
// 	controller.abort();
// };

onMounted(() => {
	const sarchParams = urlUtils.getSearchParamsObj();
	search(sarchParams.q);
	name.value = sarchParams.q;

	setTimeout(() => {
		transparent.value = false;
	}, 300);
});
</script>

<style lang="scss" scoped>
.google-search-container {
}
.my-card {
	width: 369px;
	padding: 16px 16px 20px 16px;
	border-radius: 12px;
	border: 1px solid $separator;
	.my-tabs-active-class {
		color: $grey-10;
		background: white;
	}
	.my-tab-panels-wrapper {
		border-radius: 0px;
	}
	.my-tab-panel-wrapper {
		padding: 20px 0 0 0;
		.panel-loading {
			position: absolute;
			left: 50%;
			transform: translateX(-50%);
			top: 20px;
		}
	}
}
</style>
