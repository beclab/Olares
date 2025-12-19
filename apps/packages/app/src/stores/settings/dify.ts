import { defineStore } from 'pinia';
import {
	delete_model,
	get_model_info,
	get_model_list,
	install_model,
	load_model,
	stop_model
} from 'src/api/dify';
import { bus, BUS_EVENT } from 'src/utils/bus';
import { DifyModelInfo, MODEL_STATUS } from 'src/constant';

export type DifyState = {
	difyModels: DifyModelInfo[];
	runningModels: DifyModelInfo[];
	timer: any;
};

export const useDifyStore = defineStore('dify', {
	state: () => {
		return {
			difyModels: [],
			runningModels: [],
			timer: null
		} as DifyState;
	},
	actions: {
		async get_models() {
			this.difyModels = await get_model_list();

			const list = this.difyModels.filter((item) => {
				return item.status === MODEL_STATUS.installing;
			});

			list.forEach((item) => {
				this.addTask(item);
			});
		},
		isWorking(): boolean {
			const array = this.difyModels.filter((item) => {
				return (
					item.status === MODEL_STATUS.loading ||
					item.status === MODEL_STATUS.stopping
				);
			});
			return array.length > 0;
		},
		frontend_preflight(model: DifyModelInfo) {
			model.status = MODEL_STATUS.installable;
		},
		async install_model(model: DifyModelInfo) {
			try {
				model.status = MODEL_STATUS.pending;
				await install_model(model.id);
			} catch (e: any) {
				model.status = MODEL_STATUS.noInstalled;
				bus.emit(
					BUS_EVENT.APP_BACKEND_ERROR,
					e && e.message ? e.message : 'model install error'
				);
			}
			this.addTask(model);
		},
		addTask(model: DifyModelInfo) {
			const index = this.runningModels.findIndex(
				(item) => item.id === model.id
			);
			if (index > -1) {
				this.runningModels.splice(index, 1, model);
			} else {
				this.runningModels.push(model);
			}

			if (!this.timer && this.runningModels.length > 0) {
				this.timer = setInterval(() => {
					this.runningModels = this.runningModels.filter((item) => {
						return item.status === MODEL_STATUS.installing;
					});

					if (this.runningModels.length > 0) {
						this.runningModels.forEach((item) => {
							this.query_status(item);
						});
					} else {
						clearInterval(this.timer);
						this.timer = null;
					}
				}, 3000);
			}
		},
		async query_status(model: DifyModelInfo) {
			const info = await get_model_info(model.id);
			if (info && model && info?.id === model.id) {
				model.status = info.status;
				model.progress = info.progress;
			}
		},
		async uninstall_model(model: DifyModelInfo) {
			const oldStatus = model.status;
			try {
				model.status = MODEL_STATUS.uninstalling;
				await delete_model(model.id);
				await this.query_status(model);
			} catch (e) {
				model.status = oldStatus;
			}
		},
		async stop_model(model: DifyModelInfo) {
			const oldStatus = model.status;
			try {
				model.status = MODEL_STATUS.stopping;
				await stop_model(model.id);
				await this.query_status(model);
			} catch (e) {
				model.status = oldStatus;
			}
		},
		async load_model(model: DifyModelInfo) {
			const oldStatus = model.status;
			try {
				model.status = MODEL_STATUS.loading;
				await load_model(model.id);
				this.difyModels.forEach((item) => {
					if (item.status === MODEL_STATUS.running) {
						item.status = MODEL_STATUS.installed;
					}
				});
				await this.query_status(model);
			} catch (e: any) {
				model.status = oldStatus;
				console.log(e);
				bus.emit(
					BUS_EVENT.APP_BACKEND_ERROR,
					e && e.message ? e.message : 'model load error'
				);
			}
		},
		get_mock_data() {
			this.difyModels = JSON.parse(
				'[\n' +
					'{\n' +
					'"id": "test_agent1",\n' +
					'"file_name": "stealth-v1.3.Q4_K_M.gguf",\n' +
					'"progress": 100,\n' +
					'"status": "no_installed",\n' +
					'"folder_path": "model/stealth-v1.2-7b",\n' +
					'"model": {\n' +
					'"description": "This is an application that provides the ability to convert code snippets in multiple programming languages. You can input the code you wish to convert, select the target programming language, and get the desired output.",\n' +
					'"engine": "nitro",\n' +
					'"format": "gguf",\n' +
					'"id": "stealth-v1.2-7b",\n' +
					'"metadata": {\n' +
					'"author": "Jan",\n' +
					'"size": 4370000000,\n' +
					'"tags": ["Programming"]\n' +
					'},\n' +
					'"name": "Code translator",\n' +
					'"object": "model",\n' +
					'"parameters": {\n' +
					'"frequency_penalty": 0,\n' +
					'"max_tokens": 4096,\n' +
					'"presence_penalty": 0,\n' +
					'"stop": [],\n' +
					'"stream": true,\n' +
					'"temperature": 0.7,\n' +
					'"top_p": 0.95\n' +
					'},\n' +
					'"settings": {\n' +
					'"ctx_len": 4096,\n' +
					'"prompt_template": "<|im_start|>system\\n{system_message}<|im_end|>\\n<|im_start|>user\\n{prompt}<|im_end|>\\n<|im_start|>assistant"\n' +
					'},\n' +
					'"source_url": "https://huggingface.co/janhq/stealth-v1.3-GGUF/resolve/main/stealth-v1.3.Q4_K_M.gguf",\n' +
					'"version": "1.0"\n' +
					'},\n' +
					'"type": "default"\n' +
					'},\n' +
					'{\n' +
					'"id": "test_agent2",\n' +
					'"file_name": "yarn-mistral-7b-128k.Q4_K_M.gguf",\n' +
					'"progress": 100,\n' +
					'"status": "no_installed",\n' +
					'"folder_path": "model/yarn-mistral-7b",\n' +
					'"model": {\n' +
					'"description": "A multilingual translator that provides translation capabilities in multiple languages. Input the text you need to translate and select the target language.",\n' +
					'"engine": "nitro",\n' +
					'"format": "gguf",\n' +
					'"id": "yarn-mistral-7b",\n' +
					'"metadata": {\n' +
					'"author": "NousResearch, The Bloke",\n' +
					'"size": 4370000000,\n' +
					'"tags": ["Translate"]\n' +
					'},\n' +
					'"name": "Translation assistant",\n' +
					'"object": "model",\n' +
					'"parameters": {\n' +
					'"frequency_penalty": 0,\n' +
					'"max_tokens": 4096,\n' +
					'"presence_penalty": 0,\n' +
					'"stop": [],\n' +
					'"stream": true,\n' +
					'"temperature": 0.7,\n' +
					'"top_p": 0.95\n' +
					'},\n' +
					'"settings": {\n' +
					'"ctx_len": 4096,\n' +
					'"prompt_template": "{prompt}"\n' +
					'},\n' +
					'"source_url": "https://huggingface.co/TheBloke/Yarn-Mistral-7B-128k-GGUF/resolve/main/yarn-mistral-7b-128k.Q4_K_M.gguf",\n' +
					'"version": "1.0"\n' +
					'},\n' +
					'"type": "default"\n' +
					'},\n' +
					'{\n' +
					'"id": "test_agent3",\n' +
					'"file_name": "codeninja-1.0-openchat-7b.Q4_K_M.gguf",\n' +
					'"progress": 100,\n' +
					'"status": "no_installed",\n' +
					'"folder_path": "model/codeninja-1.0-7b",\n' +
					'"model": {\n' +
					'"description": "When you explain a concept, evaluate whether it is concise, complete, and easy to understand in order to avoid falling into the expert mindset trap.",\n' +
					'"engine": "nitro",\n' +
					'"format": "gguf",\n' +
					'"id": "codeninja-1.0-7b",\n' +
					'"metadata": {\n' +
					'"author": "Beowolx",\n' +
					'"size": 4370000000,\n' +
					'"tags": ["Learning"]\n' +
					'},\n' +
					'"name": "Feynman method tutor",\n' +
					'"object": "model",\n' +
					'"parameters": {\n' +
					'"frequency_penalty": 0,\n' +
					'"max_tokens": 4096,\n' +
					'"presence_penalty": 0,\n' +
					'"stop": [\n' +
					'"<|end_of_turn|>"\n' +
					'],\n' +
					'"stream": true,\n' +
					'"temperature": 0.7,\n' +
					'"top_p": 0.95\n' +
					'},\n' +
					'"settings": {\n' +
					'"ctx_len": 4096,\n' +
					'"prompt_template": "GPT4 Correct User: {prompt}<|end_of_turn|>GPT4 Correct Assistant:"\n' +
					'},\n' +
					'"source_url": "https://huggingface.co/beowolx/CodeNinja-1.0-OpenChat-7B-GGUF/resolve/main/codeninja-1.0-openchat-7b.Q4_K_M.gguf",\n' +
					'"version": "1.0"\n' +
					'},\n' +
					'"type": "default"\n' +
					'},\n' +
					'{\n' +
					'"id": "test_agent4",\n' +
					'"file_name": "llama-2-7b-chat.Q4_K_M.gguf",\n' +
					'"progress": 100,\n' +
					'"status": "no_installed",\n' +
					'"folder_path": "model/llama2-chat-7b-q4",\n' +
					'"model": {\n' +
					'"description": "A simulated front-end interviewer that tests the skill level of front-end development through questioning.",\n' +
					'"engine": "nitro",\n' +
					'"format": "gguf",\n' +
					'"id": "llama2-chat-7b-q4",\n' +
					'"metadata": {\n' +
					'"author": "MetaAI, The Bloke",\n' +
					'"size": 4080000000,\n' +
					'"tags": ["HR"]\n' +
					'},\n' +
					'"name": "AI Front-end interviewer",\n' +
					'"object": "model",\n' +
					'"parameters": {\n' +
					'"frequency_penalty": 0,\n' +
					'"max_tokens": 4096,\n' +
					'"presence_penalty": 0,\n' +
					'"stop": [],\n' +
					'"stream": true,\n' +
					'"temperature": 0.7,\n' +
					'"top_p": 0.95\n' +
					'},\n' +
					'"settings": {\n' +
					'"ctx_len": 4096,\n' +
					'"prompt_template": "[INST] <<SYS>>\\n{system_message}<</SYS>>\\n{prompt}[/INST]"\n' +
					'},\n' +
					'"source_url": "https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGUF/resolve/main/llama-2-7b-chat.Q4_K_M.gguf",\n' +
					'"version": "1.0"\n' +
					'},\n' +
					'"type": "default"\n' +
					'},\n' +
					'{\n' +
					'"id": "test_agent5",\n' +
					'"file_name": "mistral-7b-instruct-v0.2.Q4_K_M.gguf",\n' +
					'"progress": 100,\n' +
					'"status": "no_installed",\n' +
					'"folder_path": "model/mistral-ins-7b-q4",\n' +
					'"model": {\n' +
					'"description": "As your legal advisor, I will answer your legal questions for you.",\n' +
					'"engine": "nitro",\n' +
					'"format": "gguf",\n' +
					'"id": "mistral-ins-7b-q4",\n' +
					'"metadata": {\n' +
					'"author": "MistralAI, The Bloke",\n' +
					'"cover": "https://raw.githubusercontent.com/janhq/jan/main/models/mistral-ins-7b-q4/cover.png",\n' +
					'"size": 4370000000,\n' +
					'"tags": ["Advice"]\n' +
					'},\n' +
					'"name": "Legal Advisor",\n' +
					'"object": "model",\n' +
					'"parameters": {\n' +
					'"frequency_penalty": 0,\n' +
					'"max_tokens": 4096,\n' +
					'"presence_penalty": 0,\n' +
					'"stop": [],\n' +
					'"stream": true,\n' +
					'"temperature": 0.7,\n' +
					'"top_p": 0.95\n' +
					'},\n' +
					'"settings": {\n' +
					'"ctx_len": 4096,\n' +
					'"prompt_template": "[INST] {prompt} [/INST]"\n' +
					'},\n' +
					'"source_url": "https://huggingface.co/TheBloke/Mistral-7B-Instruct-v0.2-GGUF/resolve/main/mistral-7b-instruct-v0.2.Q4_K_M.gguf",\n' +
					'"version": "1.0"\n' +
					'},\n' +
					'"type": "default"\n' +
					'},\n' +
					'{\n' +
					'"id": "test_agent6",\n' +
					'"file_name": "Noromaid-7b-v0.1.1.q5_k_m.gguf",\n' +
					'"progress": 100,\n' +
					'"status": "no_installed",\n' +
					'"folder_path": "model/noromaid-7b",\n' +
					'"model": {\n' +
					'"description": "Meeting minutes generator",\n' +
					'"engine": "nitro",\n' +
					'"format": "gguf",\n' +
					'"id": "noromaid-7b",\n' +
					'"metadata": {\n' +
					'"author": "NeverSleep",\n' +
					'"size": 4370000000,\n' +
					'"tags": ["Writing"]\n' +
					'},\n' +
					'"name": "Meeting Minutes and Summary",\n' +
					'"object": "model",\n' +
					'"parameters": {\n' +
					'"frequency_penalty": 0,\n' +
					'"max_tokens": 4096,\n' +
					'"presence_penalty": 0,\n' +
					'"stop": [],\n' +
					'"stream": true,\n' +
					'"temperature": 0.7,\n' +
					'"top_p": 0.95\n' +
					'},\n' +
					'"settings": {\n' +
					'"ctx_len": 4096,\n' +
					'"prompt_template": "### Instruction:{prompt}\\n### Response:"\n' +
					'},\n' +
					'"source_url": "https://huggingface.co/NeverSleep/Noromaid-7b-v0.1.1-GGUF/resolve/main/Noromaid-7b-v0.1.1.q5_k_m.gguf",\n' +
					'"version": "1.0"\n' +
					'},\n' +
					'"type": "default"\n' +
					'},\n' +
					'{\n' +
					'"id": "test_agent7",\n' +
					'"file_name": "openchat-3.5-1210.Q4_K_M.gguf",\n' +
					'"progress": 100,\n' +
					'"status": "no_installed",\n' +
					'"folder_path": "model/openchat-3.5-7b",\n' +
					'"model": {\n' +
					'"description": "Health and Nutrition Expert.",\n' +
					'"engine": "nitro",\n' +
					'"format": "gguf",\n' +
					'"id": "openchat-3.5-7b",\n' +
					'"metadata": {\n' +
					'"author": "Openchat",\n' +
					'"size": 4370000000,\n' +
					'"tags": ["Advice"]\n' +
					'},\n' +
					'"name": "Diet and exercise plan",\n' +
					'"object": "model",\n' +
					'"parameters": {\n' +
					'"frequency_penalty": 0,\n' +
					'"max_tokens": 4096,\n' +
					'"presence_penalty": 0,\n' +
					'"stop": [\n' +
					'"<|end_of_turn|>"\n' +
					'],\n' +
					'"stream": true,\n' +
					'"temperature": 0.7,\n' +
					'"top_p": 0.95\n' +
					'},\n' +
					'"settings": {\n' +
					'"ctx_len": 4096,\n' +
					'"prompt_template": "GPT4 Correct User: {prompt}<|end_of_turn|>GPT4 Correct Assistant:"\n' +
					'},\n' +
					'"source_url": "https://huggingface.co/TheBloke/openchat-3.5-1210-GGUF/resolve/main/openchat-3.5-1210.Q4_K_M.gguf",\n' +
					'"version": "1.0"\n' +
					'},\n' +
					'"type": "default"\n' +
					'},\n' +
					'{\n' +
					'"id": "test_agent8",\n' +
					'"file_name": "openchat-3.5-1210.Q4_K_M.gguf",\n' +
					'"progress": 100,\n' +
					'"status": "no_installed",\n' +
					'"folder_path": "model/openchat-3.5-7b",\n' +
					'"model": {\n' +
					'"description": "Write SQL from natural language by pasting in your schema with the request.Please describe your query requirements in natural language and select the target database type.",\n' +
					'"engine": "nitro",\n' +
					'"format": "gguf",\n' +
					'"id": "openchat-3.5-7b",\n' +
					'"metadata": {\n' +
					'"author": "Openchat",\n' +
					'"size": 4370000000,\n' +
					'"tags": ["Programming"]\n' +
					'},\n' +
					'"name": "SQL Creator",\n' +
					'"object": "model",\n' +
					'"parameters": {\n' +
					'"frequency_penalty": 0,\n' +
					'"max_tokens": 4096,\n' +
					'"presence_penalty": 0,\n' +
					'"stop": [\n' +
					'"<|end_of_turn|>"\n' +
					'],\n' +
					'"stream": true,\n' +
					'"temperature": 0.7,\n' +
					'"top_p": 0.95\n' +
					'},\n' +
					'"settings": {\n' +
					'"ctx_len": 4096,\n' +
					'"prompt_template": "GPT4 Correct User: {prompt}<|end_of_turn|>GPT4 Correct Assistant:"\n' +
					'},\n' +
					'"source_url": "https://huggingface.co/TheBloke/openchat-3.5-1210-GGUF/resolve/main/openchat-3.5-1210.Q4_K_M.gguf",\n' +
					'"version": "1.0"\n' +
					'},\n' +
					'"type": "default"\n' +
					'},\n' +
					'{\n' +
					'"id": "test_agent9",\n' +
					'"file_name": "openchat-3.5-1210.Q4_K_M.gguf",\n' +
					'"progress": 100,\n' +
					'"status": "no_installed",\n' +
					'"folder_path": "model/openchat-3.5-7b",\n' +
					'"model": {\n' +
					'"description": "Alice in Borderland Based ChatGPT Game",\n' +
					'"engine": "nitro",\n' +
					'"format": "gguf",\n' +
					'"id": "openchat-3.5-7b",\n' +
					'"metadata": {\n' +
					'"author": "Openchat",\n' +
					'"size": 4370000000,\n' +
					'"tags": ["Entertainment"]\n' +
					'},\n' +
					'"name": "Game：Borderland",\n' +
					'"object": "model",\n' +
					'"parameters": {\n' +
					'"frequency_penalty": 0,\n' +
					'"max_tokens": 4096,\n' +
					'"presence_penalty": 0,\n' +
					'"stop": [\n' +
					'"<|end_of_turn|>"\n' +
					'],\n' +
					'"stream": true,\n' +
					'"temperature": 0.7,\n' +
					'"top_p": 0.95\n' +
					'},\n' +
					'"settings": {\n' +
					'"ctx_len": 4096,\n' +
					'"prompt_template": "GPT4 Correct User: {prompt}<|end_of_turn|>GPT4 Correct Assistant:"\n' +
					'},\n' +
					'"source_url": "https://huggingface.co/TheBloke/openchat-3.5-1210-GGUF/resolve/main/openchat-3.5-1210.Q4_K_M.gguf",\n' +
					'"version": "1.0"\n' +
					'},\n' +
					'"type": "default"\n' +
					'},\n' +
					'{\n' +
					'"id": "test_agent10",\n' +
					'"file_name": "phi-2.Q8_0.gguf",\n' +
					'"progress": 100,\n' +
					'"status": "no_installed",\n' +
					'"folder_path": "model/phi-2-3b",\n' +
					'"model": {\n' +
					'"description": "When using voice input, people often habitually use filler words and phrases. Voice input optimization can convert these into written language.",\n' +
					'"engine": "nitro",\n' +
					'"format": "gguf",\n' +
					'"id": "phi-2-3b",\n' +
					'"metadata": {\n' +
					'"author": "Microsoft",\n' +
					'"size": 2960000000,\n' +
					'"tags": ["Tools"]\n' +
					'},\n' +
					'"name": "Voice input optimization",\n' +
					'"object": "model",\n' +
					'"parameters": {\n' +
					'"frequency_penalty": 0,\n' +
					'"max_tokens": 4096,\n' +
					'"presence_penalty": 0,\n' +
					'"stop": [],\n' +
					'"stream": true,\n' +
					'"temperature": 0.7,\n' +
					'"top_p": 0.95\n' +
					'},\n' +
					'"settings": {\n' +
					'"ctx_len": 4096,\n' +
					'"prompt_template": "Intruct:\\n{prompt}\\nOutput:"\n' +
					'},\n' +
					'"source_url": "https://huggingface.co/TheBloke/phi-2-GGUF/resolve/main/phi-2.Q8_0.gguf",\n' +
					'"version": "1.0"\n' +
					'},\n' +
					'"type": "default"\n' +
					'}\n' +
					']'
			);
		}
	}
});
