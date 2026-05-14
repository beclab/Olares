<template>
	<div class="editor">
		<div class="files-right-header row items-center justify-between">
			<div class="row items-center justify-start">
				<span>{{ fileInfo.name }}</span>
				<span
					class="statusIcon q-ml-sm"
					:style="{
						background: isEditing ? '#FFC46D' : 'rgba(41, 204, 95, 1)'
					}"
				></span>
			</div>
			<div class="toolbar-actions row items-center" v-if="fileInfo.name">
				<q-btn
					flat
					dense
					round
					size="sm"
					icon="sym_r_code"
					color="ink-2"
					@click="formatDocument"
				>
					<q-tooltip>{{ $t('editor.format') }}</q-tooltip>
				</q-btn>
				<q-btn
					flat
					dense
					round
					size="sm"
					icon="sym_r_search"
					color="ink-2"
					@click="openFind"
				>
					<q-tooltip>{{ $t('editor.find') }}</q-tooltip>
				</q-btn>
				<q-btn
					flat
					dense
					round
					size="sm"
					icon="sym_r_unfold_less"
					color="ink-2"
					@click="foldAll"
				>
					<q-tooltip>{{ $t('editor.foldAll') }}</q-tooltip>
				</q-btn>
				<q-btn
					flat
					dense
					round
					size="sm"
					icon="sym_r_unfold_more"
					color="ink-2"
					@click="unfoldAll"
				>
					<q-tooltip>{{ $t('editor.unfoldAll') }}</q-tooltip>
				</q-btn>
				<q-separator vertical inset style="height: 12px" />
				<q-btn
					flat
					dense
					round
					size="sm"
					icon="sym_r_undo"
					color="ink-2"
					@click="undo"
					:disable="!canUndo"
				>
					<q-tooltip>{{ $t('editor.undo') }}</q-tooltip>
				</q-btn>
				<q-btn
					flat
					dense
					round
					size="sm"
					icon="sym_r_redo"
					color="ink-2"
					@click="redo"
					:disable="!canRedo"
				>
					<q-tooltip>{{ $t('editor.redo') }}</q-tooltip>
				</q-btn>
				<q-separator vertical inset style="height: 12px" />

				<!-- <q-btn
					flat
					dense
					round
					size="sm"
					icon="settings"
					color="ink-2"
					@click="showSettings = true"
				>
					<q-tooltip>editor settings</q-tooltip>
				</q-btn> -->
				<q-btn
					flat
					dense
					round
					size="sm"
					icon="sym_r_save"
					color="ink-2"
					@click="onSaveFile"
				>
					<q-tooltip>{{ $t('btn_save') }}</q-tooltip>
				</q-btn>
			</div>
		</div>

		<!-- 编辑器设置对话框 -->
		<q-dialog v-model="showSettings">
			<q-card style="min-width: 400px">
				<q-card-section>
					<div class="text-h6">编辑器设置</div>
				</q-card-section>

				<q-card-section class="q-pt-none">
					<div class="q-gutter-md">
						<div>
							<div class="text-caption q-mb-xs">字体大小</div>
							<q-slider
								v-model="editorSettings.fontSize"
								:min="10"
								:max="24"
								:step="1"
								label
								label-always
							/>
						</div>

						<div>
							<div class="text-caption q-mb-xs">缩进大小</div>
							<q-slider
								v-model="editorSettings.tabSize"
								:min="2"
								:max="8"
								:step="1"
								label
								label-always
							/>
						</div>

						<q-select
							v-model="editorSettings.wordWrap"
							:options="wordWrapOptions"
							label="自动换行"
							dense
							outlined
						/>

						<q-select
							v-model="editorSettings.renderWhitespace"
							:options="whitespaceOptions"
							label="空格显示"
							dense
							outlined
						/>

						<q-toggle v-model="editorSettings.minimap" label="显示代码缩略图" />

						<q-toggle v-model="editorSettings.lineNumbers" label="显示行号" />

						<q-toggle v-model="editorSettings.folding" label="启用代码折叠" />

						<q-separator />

						<q-toggle
							v-model="editorSettings.formatOnSave"
							label="保存时自动格式化"
						/>
					</div>
				</q-card-section>

				<q-card-actions align="right">
					<q-btn
						flat
						label="重置"
						color="negative"
						@click="resetEditorSettings"
					/>
					<q-btn flat label="关闭" color="primary" v-close-popup />
				</q-card-actions>
			</q-card>
		</q-dialog>
		<div class="files-right-content">
			<vue-monaco-editor
				class="files-monaco"
				:theme="$q.dark.isActive ? 'olares-dark' : 'olares-light'"
				:language="fileInfo.lang"
				:options="editorOptions"
				@change="changeCode"
				@mount="handleEditorMount"
				:value="fileInfo.code"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { PropType, ref, computed } from 'vue';
import { useQuasar } from 'quasar';
import { FilesCodeType } from '@apps/studio/src/types/core';
import * as monaco from 'monaco-editor';
import { useEditorSettings } from 'src/composables/useEditorSettings';

const { settings: editorSettings, resetSettings: resetEditorSettings } =
	useEditorSettings();

const showSettings = ref(false);

const wordWrapOptions = [
	{ label: '自动换行', value: 'on' },
	{ label: '不换行', value: 'off' },
	{ label: '按列换行', value: 'wordWrapColumn' },
	{ label: '限制换行', value: 'bounded' }
];

const whitespaceOptions = [
	{ label: '不显示', value: 'none' },
	{ label: '边界', value: 'boundary' },
	{ label: '选中时', value: 'selection' },
	{ label: '全部显示', value: 'all' }
];

defineProps({
	isEditing: {
		type: Boolean,
		default: false,
		required: false
	},
	fileInfo: {
		type: Object as PropType<FilesCodeType>,
		required: true
	}
});

const emits = defineEmits(['onSaveFile', 'changeCode', 'editorMount']);

const $q = useQuasar();
const editorInstance = ref<any>(null);
const canUndo = ref(false);
const canRedo = ref(false);

const editorOptions = computed(() => ({
	tabSize: editorSettings.value.tabSize,
	insertSpaces: true,
	detectIndentation: true,
	fontSize: editorSettings.value.fontSize,

	folding: editorSettings.value.folding,
	foldingStrategy: 'indentation' as const,
	showFoldingControls: 'always' as const,
	foldingHighlight: true,
	glyphMargin: false,
	foldingMaximumRegions: 5000,
	unfoldOnClickAfterEndOfLine: true,

	minimap: {
		enabled: editorSettings.value.minimap,
		maxColumn: 80,
		renderCharacters: false,
		showSlider: 'mouseover' as const
	},

	quickSuggestions: {
		other: true,
		comments: false,
		strings: true
	},
	suggestOnTriggerCharacters: true,
	acceptSuggestionOnEnter: 'on' as const,
	tabCompletion: 'on' as const,
	wordBasedSuggestions: 'matchingDocuments' as const,
	suggestSelection: 'first' as const,

	formatOnPaste: true,
	formatOnType: true,

	scrollBeyondLastLine: false,
	automaticLayout: true,
	wordWrap: editorSettings.value.wordWrap,
	wrappingIndent: 'indent' as const,
	smoothScrolling: true,
	cursorBlinking: 'smooth' as const,
	cursorSmoothCaretAnimation: 'on' as const,
	lineNumbers: editorSettings.value.lineNumbers
		? ('on' as const)
		: ('off' as const),

	matchBrackets: 'always' as const,
	bracketPairColorization: {
		enabled: true
	},

	lineHeight: 22,
	fontFamily: 'Consolas, "Courier New", monospace',
	fontLigatures: true,
	renderLineHighlight: 'all' as const,
	renderWhitespace: editorSettings.value.renderWhitespace,

	scrollbar: {
		vertical: 'auto' as const,
		horizontal: 'auto' as const,
		verticalScrollbarSize: 10,
		horizontalScrollbarSize: 10
	},

	find: {
		addExtraSpaceOnTop: false,
		autoFindInSelection: 'never' as const,
		seedSearchStringFromSelection: 'always' as const
	},

	hover: {
		enabled: true,
		delay: 300
	},
	links: true,
	colorDecorators: true,
	contextmenu: true,
	mouseWheelZoom: true
}));

const onSaveFile = async () => {
	if (editorSettings.value.formatOnSave && editorInstance.value) {
		await editorInstance.value.getAction('editor.action.formatDocument')?.run();
	}
	emits('onSaveFile');
};

const changeCode = (value) => {
	emits('changeCode', value);
};

const handleEditorMount = (editor) => {
	editorInstance.value = editor;
	emits('editorMount', editor);

	editor.onDidChangeModelContent(() => {
		const model = editor.getModel();
		if (model) {
			canUndo.value = model.canUndo();
			canRedo.value = model.canRedo();
		}
	});

	editor.addCommand(
		(window.navigator.platform.match('Mac')
			? monaco.KeyMod.CtrlCmd
			: monaco.KeyMod.CtrlCmd) |
			monaco.KeyMod.Shift |
			monaco.KeyCode.KeyF,
		() => {
			formatDocument();
		}
	);
};

const formatDocument = () => {
	if (editorInstance.value) {
		editorInstance.value.getAction('editor.action.formatDocument')?.run();
	}
};

const openFind = () => {
	if (editorInstance.value) {
		editorInstance.value.getAction('actions.find')?.run();
	}
};

const foldAll = () => {
	if (editorInstance.value) {
		editorInstance.value.getAction('editor.foldAll')?.run();
	}
};

const unfoldAll = () => {
	if (editorInstance.value) {
		editorInstance.value.getAction('editor.unfoldAll')?.run();
	}
};

const undo = () => {
	if (editorInstance.value) {
		editorInstance.value.trigger('keyboard', 'undo', null);
	}
};

const redo = () => {
	if (editorInstance.value) {
		editorInstance.value.trigger('keyboard', 'redo', null);
	}
};
</script>

<style lang="scss" scoped>
.files-right-header {
	width: 100%;
	height: 48px;
	line-height: 48px;
	border-bottom: 1px solid $separator;
	background: $background-1;
	.statusIcon {
		width: 6px;
		height: 6px;
		border-radius: 3px;
		display: inline-block;
	}

	.toolbar-actions {
		padding-right: 8px;
		gap: 4px;
	}
}
.files-right-content {
	height: calc(100% - 56px);
	padding: 12px 0;
	background: $background-1;
	display: flex;
	flex-direction: column;

	.files-monaco {
		flex: 1;
		border-radius: 12px;
	}
}
</style>

<style lang="scss" scoped>
.editor {
	width: calc(100% - 40px);
	height: 100%;
	background: $background-1;
}
.my-code-link {
	background: $background-hover;
	color: $ink-1;
}
::-webkit-scrollbar {
	width: 0px !important;
	height: 0px !important;
}

::-webkit-scrollbar-thumb {
	border-radius: 10px;
	width: 1px;
	background: rgba(255, 255, 255, 0.5);
}

::-webkit-scrollbar-track {
	box-shadow: inset 0 0 5px rgba(0, 0, 0, 0.2);
	border-radius: 10px;
	background: rgba(57, 177, 255, 0.16);
}
.monaco-editor .margin {
	background-color: $background-2 !important;
	::-webkit-scrollbar {
		display: none !important;
	}
}

.lines-content.monaco-editor-background {
	background-color: $background-2 !important;
}

.minimap.slider-mouseover {
	background-color: $background-2 !important;
}
.minimap-decorations-layer {
	background-color: $background-2 !important;
}
.decorationsOverviewRuler {
	width: 0px !important;
}

.inputarea.monaco-mouse-cursor-text {
	background-color: $ink-1 !important;
	caret-color: red !important;
}

::v-deep(.monaco-editor .line-numbers) {
	color: $ink-3 !important;
	&.active-line-number {
		color: $ink-2 !important;
	}
}

::v-deep(.monaco-editor) {
	outline-width: 0 !important;
}

::v-deep(.monaco-editor .folding) {
	cursor: pointer;
}

::v-deep(.monaco-editor) {
	.margin,
	.glyph-margin {
		background-color: transparent !important;
	}
}

::v-deep(.monaco-scrollable-element > .scrollbar > .slider) {
	border-radius: 4px;
}

.monaco-editor .inputarea {
	background-color: $ink-1 !important;
	z-index: 1 !important;
	caret-color: red !important;
}

.view-lines .view-line {
	border: 4px solid red !important;
	span {
		color: $ink-1 !important;
	}
	.mtk1 {
		color: $ink-1 !important;
	}
}

::v-deep(.decorationsOverviewRuler) {
	width: 0px !important;
}
</style>
