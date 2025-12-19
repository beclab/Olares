<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Transcode path')"
		:cancel="t('cancel')"
		:ok="t('confirm')"
		size="medium"
		@onSubmit="confirmAction"
		:okDisabled="!enableEdit"
	>
		<div class="text-body3 text-ink-2">
			{{
				t(
					'Browse or enter the path to use for transcode files. The folder must be writeable.'
				)
			}}
		</div>

		<terminus-edit
			class="q-mt-lg"
			v-model="folderRef"
			:label="t('Folder')"
			style="width: 100%"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';

const props = defineProps({
	folder: {
		type: String,
		required: false,
		default: ''
	}
});

const { t } = useI18n();

const CustomRef = ref();

const folderRef = ref(props.folder);

const enableEdit = computed(() => {
	return folderRef.value.length > 0 && folderRef.value != props.folder;
});

const confirmAction = () => {
	CustomRef.value.onDialogOK(folderRef.value);
};
</script>

<style scoped lang="scss">
.cpu-core {
	text-align: right;
}
</style>
