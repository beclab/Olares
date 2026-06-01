<template>
	<bt-custom-dialog ref="CustomRef" :cancel="false" :ok="false" size="medium">
		<div
			style="
				height: 300px;
				border-radius: 6px;
				overflow: hidden;
				position: relative;
			"
		>
			<div class="yaml-tool-container">
				<q-btn
					flat
					round
					dense
					color="white"
					icon="content_copy"
					size="xs"
					@click="copyCommand"
				>
				</q-btn>
			</div>

			<v-ace-editor
				v-model:value="code"
				lang="terminal"
				theme="chaos"
				:readonly="readonly"
				style="height: calc(100%)"
				:options="{
					showGutter: true,
					showPrintMargin: false,
					fontSize: 14,
					wrap: false,
					readOnly: true,
					highlightActiveLine: true
				}"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { VAceEditor } from 'vue3-ace-editor';
import ace from 'ace-builds';
import 'ace-builds/src-noconflict/theme-textmate';
import 'ace-builds/src-noconflict/mode-groovy';
import 'ace-builds/src-noconflict/theme-terminal';
// import 'ace-builds/src-noconflict/ext-searchbox';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

interface Props {
	readonly?: boolean;
	command?: string;
}

const props = withDefaults(defineProps<Props>(), {
	readonly: true,
	command: ''
});

const { t } = useI18n();

function extractShellCommands(shellStr) {
	return shellStr;
}

const code = ref<string>(extractShellCommands(props.command));

const copyCommand = async () => {
	try {
		if (navigator.clipboard && window.isSecureContext) {
			await navigator.clipboard.writeText(code.value);
		} else {
			const textArea = document.createElement('textarea');
			textArea.value = code.value;
			textArea.style.position = 'fixed';
			textArea.style.left = '-999999px';
			textArea.style.top = '-999999px';
			document.body.appendChild(textArea);
			textArea.focus();
			textArea.select();
			document.execCommand('copy');
			document.body.removeChild(textArea);
		}

		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('COPIED_SUCCESSFUL')
		});
	} catch (error) {
		//
	}
};
</script>

<style lang="scss" scoped>
.yaml-container {
	font-family: 'Roboto';
}
.yaml-tool-container {
	position: absolute;
	bottom: 2px;
	right: 2px;
	z-index: 99;
}
</style>
