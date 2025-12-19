<template>
	<div class="chat">
		<div class="header">
			<div class="but">
				<img src="../../assets/desktop/terminus-icon.svg" alt="terminus" />
				<span>Ashia</span>
				<img
					@click="goBack"
					src="../../assets/desktop/delete-icon.svg"
					alt="delete"
				/>
			</div>
		</div>
		<q-scroll-area class="list" ref="chatRef">
			<div
				class="message"
				v-for="(item, index) in searchStore.chatList"
				:key="index"
			>
				<img
					class="icon"
					v-if="item.name === 'Terminus'"
					src="../../assets/desktop/terminus-icon.svg"
					alt="terminus"
				/>
				<div class="me" v-if="item.name === 'Me'">
					<TerminusAvatar :info="tokenStore.terminus" :size="24" />
				</div>
				<div class="askTxt" v-if="filesItem && index === 0">
					<div class="ask_files">
						<span class="as_kai">Ask AI</span>
						<span class="content">
							<img
								:src="
									filesItem?.fileType
										? `./files/${filesItem?.fileType}.svg`
										: `./files/other.svg`
								"
							/>
							<span>
								{{ getFileName(filesItem.path) }}
							</span>
						</span>
					</div>
					<div class="message">{{ item.text }}</div>
				</div>
				<div class="txt" v-else>{{ item.text }}</div>
			</div>
		</q-scroll-area>

		<div class="submit">
			<div class="fileAsk" v-if="showAskAi">
				<span class="txt">Ask AI</span>
				<img
					:src="
						filesItemActive?.fileType
							? `./files/${filesItemActive?.fileType}.svg`
							: `./files/other.svg`
					"
				/>
				<span class="content">{{ getFileName(filesItemActive.path) }}</span>
				<span class="close" @click="closeFilesSearch">
					<img src="../../assets/desktop/close.svg" />
				</span>
			</div>
			<div class="submitMessage">
				<input
					class="input"
					type="text"
					autofocus
					ref="inputRef"
					v-model.trim="askValue"
					placeholder="Ask Ashia something..."
				/>
				<span class="btn" @click="submit" :disable="!searchStore.can_input"
					>Submit Message</span
				>
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useSearchStore } from 'src/stores/search';
import { v4 as uuidv4 } from 'uuid';
import { useTokenStore } from '../../stores/desktop/token';

export default {
	name: 'TerminusAi',
	props: {
		showSearchDialog: {
			type: Boolean,
			require: false,
			default: false
		},
		filesItem: {
			type: Object,
			require: false
		},
		isDoc: {
			type: Boolean,
			require: true
		}
	},
	components: {},
	defineProps: [],

	setup(props: any, context: any) {
		const askValue = ref('');
		const searchStore = useSearchStore();
		const tokenStore = useTokenStore();

		const uid = ref();
		const chatRef = ref();
		const inputRef = ref();
		const verticalSize = ref();
		const filesItemActive = ref(props.filesItem);
		const showAskAi = ref(false);

		const submit = async () => {
			if (!askValue.value) return false;
			showAskAi.value = false;
			let path = filesItemActive.value?.path || '';
			let type = 'basic';
			if (props.filesItem) {
				type = 'single_doc';
			}
			if (props.isDoc) {
				type = 'full_doc';
				path = 'full_doc';
			}
			await searchStore.sendChat(uid.value, askValue.value, path, type);
			askValue.value = '';
			filesItemActive.value = null;
		};

		const goBack = () => {
			context.emit('openCommand', '');
		};

		const keydownEnter = (event: any) => {
			if (event.keyCode === 13) {
				if (searchStore.can_input) {
					submit();
				}
			}
		};

		const scrollToBottom = () => {
			nextTick(() => {
				if (verticalSize.value === chatRef.value.getScroll().verticalSize) {
					return false;
				}
				chatRef.value.setScrollPosition(
					'vertical',
					chatRef.value.getScroll().verticalSize,
					300
				);
				verticalSize.value = chatRef.value.getScroll().verticalSize;
			});
		};

		const inputFocus = () => {
			nextTick(() => {
				inputRef.value.focus();
			});
		};

		const checkFileAction = (item: any) => {
			if (filesItemActive.value && !filesItemActive.value.message) {
				showAskAi.value = true;
			} else {
				showAskAi.value = false;
			}
			if (item && item.from === 'summarize') {
				askValue.value = item.message;
			}
			submit();
		};

		const getFileName = (path: string) => {
			const pathArr = path.split('/');
			return pathArr[pathArr.length - 1];
		};

		const closeFilesSearch = () => {
			filesItemActive.value = null;
			showAskAi.value = false;
		};

		watch(
			() => searchStore.chatList,
			(newVal) => {
				if (newVal) {
					scrollToBottom();
				}
			},
			{ deep: true }
		);

		onMounted(async () => {
			uid.value = uuidv4();

			inputFocus();
			scrollToBottom();
			searchStore.initChat();
			checkFileAction(filesItemActive.value);
			document.addEventListener('keydown', keydownEnter);
		});

		onUnmounted(() => {
			searchStore.initChat();
			document.removeEventListener('keydown', keydownEnter);
		});

		return {
			askValue,
			chatRef,
			searchStore,
			inputRef,
			tokenStore,
			filesItemActive,
			showAskAi,
			submit,
			goBack,
			closeFilesSearch,
			getFileName
		};
	}
};
</script>

<style lang="scss" scoped>
.chat {
	width: 100%;
	height: 100%;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: space-between;

	.header {
		width: 100%;
		height: 42px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 12px;
		.but {
			width: 143px;
			height: 42px;
			display: flex;
			align-items: center;
			justify-content: space-around;
			background: rgba(26, 19, 15, 0.06);
			border-radius: 4px;
			margin-left: 24px;
			span {
				color: #1a130f;
			}
			img {
				cursor: pointer;
			}
		}

		.path {
			margin-right: 20px;
			max-width: 450px;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
			color: #857c77;
			font-size: 14px;
		}
	}

	.list {
		width: 100%;
		flex: 1;
		padding: 8px 24px;
		border-top: 1px solid #d8d8d8;
		.message {
			display: flex;
			align-items: start;
			justify-content: start;
			margin: 12px 0;
			.icon {
				margin-right: 12px;
				width: 24px;
				height: 24px;
				margin-top: 1px;
			}
			.me {
				background: #5348cc;
				border-radius: 4px;
				margin-right: 12px;
				width: 24px;
				height: 24px;
				line-height: 24px;
				text-align: center;
				margin-top: 1px;
				overflow: hidden;
			}
			.txt {
				background: rgba(255, 255, 255, 0.6);
				border: 1px solid #d8d8d8;
				box-shadow: inset 0px 0px 1px rgba(255, 255, 255, 0.2);
				backdrop-filter: blur(15px);
				border-radius: 4px;
				line-height: 20px;
				color: #1a130f;
				font-size: 14px;
				padding: 12px;
				min-width: 37px;
				box-sizing: border-box;
			}
			.askTxt {
				background: rgba(255, 255, 255, 0.6);
				border: 1px solid #d8d8d8;
				box-shadow: inset 0px 0px 1px rgba(255, 255, 255, 0.2);
				backdrop-filter: blur(15px);
				border-radius: 4px;
				padding: 12px;
				min-width: 37px;
				box-sizing: border-box;
				.ask_files {
					display: flex;
					align-items: center;
					.as_kai {
						font-size: 12px;
						line-height: 14px;
						color: #857c77;
						padding: 0 8px;
						border-left: 1px solid #d8d8d8;
					}
					.content {
						padding: 8px;
						background: rgba(26, 19, 15, 0.06);
						border-radius: 6px;
						display: flex;
						align-items: center;
						span {
							display: inline-block;
							margin-left: 8px;
							max-width: 560px;
							overflow: hidden;
							white-space: nowrap;
							text-overflow: ellipsis;
							color: #1a130f;
							font-size: 12px;
						}
					}
				}
				.message {
					margin-top: 8px;
				}
			}
		}
	}

	.submit {
		width: 96%;
		background: rgba(255, 255, 255, 0.5);
		border: 1px solid #d8d8d8;
		border-radius: 12px;
		margin: 0 auto 16px;
		padding: 0 16px;

		.fileAsk {
			width: 100%;
			height: 30px;
			line-height: 30px;
			background-color: #857c77;
			margin-top: 12px;
			background: rgba(26, 19, 15, 0.06);
			border-radius: 6px;
			display: flex;
			align-items: center;
			position: relative;
			.txt {
				width: 50px;
				line-height: 14px;
				font-size: 12px;
				text-align: center;
				display: inline-block;
				border-right: 1px solid #d8d8d8;
				color: #857c77;
				margin-right: 8px;
			}
			.content {
				color: #1a130f;
				margin-left: 8px;
				font-size: 12px;
				line-height: 14px;
				max-width: 70%;
				overflow: hidden;
				white-space: nowrap;
				text-overflow: ellipsis;
				flex: 1;
			}
			.close {
				width: 15px;
				height: 15px;
				line-height: 15px;
				text-align: center;
				position: absolute;
				right: 8px;
				top: 7px;
				cursor: pointer;
				img {
					width: 7px;
					height: 7px;
				}
				&:hover {
					background: rgba(26, 19, 15, 0.06);
					border-radius: 4px;
				}
			}
		}
		.submitMessage {
			display: flex;
			align-items: center;
			justify-content: space-between;
			.input {
				width: 100%;
				flex: 1;
				height: 54px;
				flex: 1;
				border: none;
				outline: none;
				margin-right: 16px;
				font-weight: 400;
				font-size: 16px;
				line-height: 20px;
				color: #857c77;
				background: rgba(255, 255, 255, 0);
				&::placeholder {
					color: #bdbdbd;
				}
			}
			.btn {
				padding: 8px 12px;
				font-weight: 400;
				font-size: 14px;
				line-height: 14px;
				color: #857c77;
				background: rgba(26, 19, 15, 0.06);
				border-radius: 4px;
				cursor: pointer;
				&:hover {
					background: rgba(26, 19, 15, 0.1);
				}
			}
		}
	}
}
</style>
