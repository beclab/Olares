<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="edit ? t('dialog.edit_tag_name') : t('dialog.create_a_new_view')"
		@onSubmit="onOK"
		:okLoading="isLoading"
		:okDisabled="okDisabled"
		:ok="t('base.confirm')"
		:cancel="t('base.cancel')"
	>
		<div class="column">
			<div class="prompt-name q-mb-xs text-body3">
				{{ t('tag_name') }}
			</div>
			<q-input
				class="prompt-input text-body3"
				v-model="tagName"
				borderless
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				no-error-icon
				placeholder=""
			/>

			<div v-if="edit" class="prompt-name q-mb-xs text-body3 q-mt-lg">
				{{ t('base.tag_id') }}
			</div>
			<q-input
				v-if="edit"
				class="prompt-input text-body3"
				v-model="tagId"
				borderless
				readonly
				no-error-icon
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				placeholder=""
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { Label } from '../../../utils/rss-types';
import { PropType } from 'vue/dist/vue';
import { useI18n } from 'vue-i18n';
import { computed, onMounted, ref } from 'vue';
import { useRssStore } from '../../../stores/rss';

const props = defineProps({
	data: {
		type: Object as PropType<Label>,
		require: false
	}
});

const { t } = useI18n();
const tagId = ref('');
const tagName = ref('');
const edit = ref(!!props.data);
const isLoading = ref(false);
const rssStore = useRssStore();
const customRef = ref();

onMounted(() => {
	if (edit.value) {
		tagName.value = props.data.name;
		tagId.value = props.data.id;
	}
});

const okDisabled = computed(() => {
	if (!tagName.value) {
		return true;
	}

	if (edit.value) {
		return tagName.value === props.data.name;
	}

	return false;
});

const onOK = async () => {
	isLoading.value = true;

	if (edit.value) {
		await rssStore
			.updateLabel({
				...props.data,
				name: tagName.value
			})
			.finally(() => {
				isLoading.value = false;
			});
	} else {
		await rssStore.addLabel(tagName.value).finally(() => {
			isLoading.value = false;
		});
	}

	customRef.value.onDialogOK();
};
</script>

<style scoped lang="scss">
.prompt-name {
	color: $ink-3;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.prompt-style {
	border: 1px solid $input-stroke;
	border-radius: 8px;
	color: $ink-3;
}

.prompt-input {
	padding-left: 7px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	color: $ink-3;
	height: 32px;
}
</style>
