import chatDefault from 'assets/images/chat-model-2.svg';

import chatModel1 from 'assets/images/chat-model-1.svg';
import chatModel2 from 'assets/images/chat-model-5.svg';
import chatModel3 from 'assets/images/chat-model-2.svg';
import chatModel4 from 'assets/images/chat-model-3.svg';
import chatModel5 from 'assets/images/chat-model-4.svg';
import chatModel6 from 'assets/images/chat-model-7.svg';

export const modelImg = {
	'img-1': chatModel1,
	'img-2': chatModel2,
	'img-3': chatModel3,
	'img-4': chatModel4,
	'img-5': chatModel5,
	'img-6': chatModel6,
	'img-999': chatDefault
};

export const getCompletionParams = (name, icon = 'chatDefault') => ({
	name,
	icon,
	icon_background: '#FFEAD5',
	mode: 'completion',
	model_config: {
		opening_statement: '',
		suggested_questions: [],
		suggested_questions_after_answer: {
			enabled: false
		},
		speech_to_text: {
			enabled: false
		},
		retriever_resource: {
			enabled: false
		},
		annotation_reply: {
			enabled: false
		},
		more_like_this: {
			enabled: false
		},
		sensitive_word_avoidance: {
			enabled: false,
			type: '',
			configs: []
		},
		external_data_tools: [],
		model: {
			provider: 'openai',
			name: 'gpt-3.5-turbo-instruct',
			mode: 'completion',
			completion_params: {
				max_tokens: 1000,
				temperature: 0,
				top_p: 0,
				presence_penalty: 0.1,
				frequency_penalty: 0.1
			}
		},
		user_input_form: [
			{
				select: {
					label: '目标语言',
					variable: 'target_language',
					description: '翻译的目标语言',
					default: '中文',
					required: true,
					options: [
						'中文',
						'英文',
						'日语',
						'法语',
						'俄语',
						'德语',
						'西班牙语',
						'韩语',
						'意大利语'
					]
				}
			},
			{
				paragraph: {
					label: '文本内容',
					variable: 'query',
					required: true,
					default: ''
				}
			}
		],
		dataset_query_variable: null,
		pre_prompt: '请将以下文本翻译为{{target_language}}:\n{{query}}\n翻译:',
		agent_mode: {
			enabled: false,
			strategy: null,
			tools: []
		},
		prompt_type: null,
		chat_prompt_config: {},
		completion_prompt_config: {},
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
		}
	}
});

const getChatParams = (name, icon = 'default') => ({
	name,
	icon,
	icon_background: '#FFEAD5',
	mode: 'chat'
});

export const createAppsGoogleSearch = {
	name: 'Noromaid 7B',
	params: getChatParams('Noromaid 7B', 'img-6')
};

export const createAppsList = [
	{
		name: 'Ashia',
		params: getChatParams('Ashia', 'img-1')
	},
	{
		name: 'Translation assistant',
		params: getCompletionParams('Translation assistant', 'img-2')
	},
	{
		name: 'CodeNinja 7B',
		params: getChatParams('CodeNinja 7B', 'img-3')
	},
	{
		name: 'Openchat-3.5 7B',
		params: getChatParams('Openchat-3.5 7B', 'img-4')
	},
	{
		name: 'YaRN Mistral',
		params: getChatParams('YaRN Mistral', 'img-5')
	}
];
