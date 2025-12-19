<template>
	<div class="editor">
		<div class="files-right-header row items-center justify-between">
			<div class="row items-center justify-start">
				<!-- <img
					class="q-mr-sm"
					src="../../assets/icon-txt.svg"
					style="width: 12px"
				/> -->
				<span>{{ fileInfo.name }}</span>
				<span
					class="statusIcon q-ml-sm"
					:style="{
						background: isEditing ? '#FFC46D' : 'rgba(41, 204, 95, 1)'
					}"
				></span>
			</div>
			<div class="save-icon row items-center justify-center">
				<q-icon
					class="cursor-pointer"
					name="sym_r_save"
					size="20px"
					color="ink-2"
					@click="onSaveFile"
				/>
			</div>
		</div>
		<div class="files-right-content">
			<vue-monaco-editor
				class="files-monaco"
				:tabSize="4"
				:theme="$q.dark.isActive ? 'vs-dark' : 'vs'"
				:language="fileInfo.lang"
				:contentHeight="10"
				@change="changeCode"
				:value="fileInfo.code"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { useQuasar } from 'quasar';
import { FilesCodeType } from '@apps/studio/src/types/core';

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

const onSaveFile = () => {
	emits('onSaveFile');
};

const changeCode = (value) => {
	emits('changeCode', value);
};

const editorMount = (value) => {
	emits('editorMount');
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

	.save-icon {
		width: 32px;
	}
}
.files-right-content {
	height: calc(100% - 56px);
	padding: 12px 0;
	background: $background-1;

	.files-monaco {
		height: 100%;
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
	width: 22px !important;

	color: $ink-3 !important;
	&.active-line-number {
		color: $ink-2 !important;
	}
}

// ::v-deep(.monaco-editor .margin) {
// 	width: 30px !important;
// }
// ::v-deep(.monaco-editor .margin-view-overlays) {
// 	width: 30px !important;
// }

::v-deep(.monaco-scrollable-element) {
	width: calc(100% - 32px) !important;
}

::v-deep(.monaco-editor) {
	outline-width: 0 !important;
}

::v-deep(.monaco-editor .monaco-scrollable-element) {
	left: 42px !important;
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
