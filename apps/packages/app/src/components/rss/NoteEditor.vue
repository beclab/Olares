<template>
	<div class="full-width column cursor-pointer" @click="onEdit">
		<div class="row full-width justify-between items-center q-mb-xs">
			<div class="text-body3 drawer-title" style="height: 24px">
				{{ t('base.note') }}
			</div>

			<div v-if="false" class="note-operate">
				<q-btn
					class="btn-size-xs btn-no-text btn-no-border btn-circle-border bg-white"
					color="ink-2"
					outline
					no-caps
					icon="sym_r_more_horiz"
				>
					<bt-popup>
						<bt-popup-item
							v-close-popup
							:title="t('base.edit')"
							icon="sym_r_edit_square"
							@on-item-click="onEdit"
						/>
						<bt-popup-item
							v-close-popup
							:title="t('base.copy')"
							icon="sym_r_content_copy"
							@on-item-click="onCopy"
						/>
						<bt-popup-item
							v-close-popup
							:title="t('base.delete')"
							icon="sym_r_delete"
							@on-item-click="onDelete"
						/>
					</bt-popup>
				</q-btn>
			</div>
		</div>
		<div class="note-root">
			<!-- edit -->
			<div v-if="edit" class="edit-note column">
				<q-input
					dense
					borderless
					:placeholder="t('main.add_a_document_note_here')"
					input-class="text-body3 text-ink-2"
					style="height: 50px; overflow: scroll; scrollbar-width: none"
					:input-style="{ resize: 'none' }"
					type="textarea"
					v-model="noteContent"
				/>
				<div class="edit-button-group row justify-end">
					<q-btn
						class="q-mr-sm btn-size-xs"
						:label="t('base.cancel')"
						color="orange-6"
						outline
						no-caps
						@click.stop="onCancel"
					/>
					<q-btn
						class="q-mr-md btn-size-xs"
						:label="t('base.save')"
						color="orange-6"
						no-caps
						@click.stop="onSave"
					/>
				</div>
			</div>
			<!-- empty -->
			<div
				v-else-if="
					!entryNote ||
					(entryNote && entryNote.deleted) ||
					(entryNote && !entryNote.deleted && !entryNote.content)
				"
				class="empty-note text-body3 cursor-pointer"
				@click="onEdit"
			>
				{{ t('main.add_a_document_note_here') }}
			</div>
			<!-- display -->
			<div v-else class="display-note row justify-start">
				<div class="display-note-text text-body2">
					{{ noteContent }}
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useRssStore } from '../../stores/rss';
import { CreateNote, Note } from '../../utils/rss-types';
import { ref, watch } from 'vue';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import BtPopupItem from 'src/components/base/BtPopupItem.vue';
import BtPopup from 'src/components/base/BtPopup.vue';
import { useReaderStore } from '../../stores/rss-reader';
import { getApplication } from 'src/application/base';

const { t } = useI18n();
const rssStore = useRssStore();
const readerStore = useReaderStore();
const edit = ref(false);
const entryNote = ref<Note | null | undefined>(null);
const noteContent = ref<string>('');

watch(
	() => [readerStore.readingEntry],
	() => {
		if (readerStore.readingEntry) {
			edit.value = false;
			entryNote.value = rssStore.getEntryNote(readerStore.readingEntry);
			noteContent.value = entryNote.value ? entryNote.value!.content : '';
		}
	},
	{
		deep: true,
		immediate: true
	}
);

const onSave = async () => {
	if (
		!readerStore.readingEntry ||
		(entryNote.value && noteContent.value === entryNote.value.context)
	) {
		return;
	}

	if (entryNote.value) {
		entryNote.value.content = noteContent.value;
		entryNote.value.deleted = false;
		rssStore.updateNote(entryNote.value!);
	} else {
		const note: CreateNote = {
			entry_id: readerStore.readingEntry!.id,
			content: noteContent.value,
			start: 0,
			length: 0,
			highlight: ''
		};
		entryNote.value = await rssStore.addNote(note);
	}
	edit.value = false;
};

const onEdit = () => {
	edit.value = true;
	noteContent.value = entryNote.value ? entryNote.value!.content : '';
};

const onCancel = () => {
	noteContent.value = entryNote.value ? entryNote.value!.content : '';
	edit.value = false;
};

const onCopy = () => {
	if (entryNote.value) {
		getApplication()
			.copyToClipboard(entryNote.value!.content)
			.then(() => {
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('copy_success')
				});
			})
			.catch((e) => {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: t('copy_failure_message', e.message)
				});
			});
	}
};

const onDelete = () => {
	if (entryNote.value) {
		rssStore.removeNote(entryNote.value!.id);
	}
};
</script>

<style lang="scss" scoped>
.note-root {
	width: 100%;

	.edit-note {
		width: 100%;
		height: 93px;
		padding: 0 8px;
		position: relative;
		border-radius: 8px;
		border: 1px solid $orange-default;

		.edit-button-group {
			position: absolute;
			right: 0;
			bottom: 8px;
		}
	}

	.empty-note {
		width: 100%;
		height: 32px;
		border-radius: 8px;
		padding: 8px;
		border: 1px solid $input-stroke;
		background: $background-6;
	}

	.display-note {
		width: 100%;
		height: auto;

		.display-note-text {
			width: calc(100% - 50px);
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			white-space: pre-wrap;
			-webkit-line-clamp: 5;
			-webkit-box-orient: vertical;
			word-wrap: break-word;
		}
	}
}
</style>
