<template>
	<q-dialog v-model="show" persistent>
		<q-card class="settings-card" flat>
			<div class="settings-header row items-center justify-between">
				<div class="text-ink-1 text-subtitle1">
					{{ t('desktop_settings.title') }}
				</div>
				<q-btn
					flat
					dense
					icon="close"
					size="sm"
					color="ink-3"
					@click="onClose"
				/>
			</div>

			<q-tabs
				v-model="tab"
				dense
				align="left"
				active-color="primary"
				indicator-color="primary"
				class="text-ink-2"
			>
				<q-tab name="bookmarks" :label="t('desktop_settings.bookmarks')" />
				<q-tab name="advanced" :label="t('desktop_settings.advanced')" />
			</q-tabs>

			<q-separator />

			<q-tab-panels v-model="tab" animated keep-alive>
				<q-tab-panel name="bookmarks" class="settings-panel">
					<div class="row items-center justify-between q-mb-md">
						<div class="text-ink-2">
							{{ t('desktop_settings.bookmarks_desc') }}
						</div>
						<q-btn
							flat
							dense
							icon="add"
							:label="t('desktop_settings.add_bookmark')"
							color="primary"
							@click="openCreateForm"
						/>
					</div>

					<div v-if="bookmarkStore.loading" class="text-ink-3 q-pa-md">
						{{ t('desktop_settings.loading') }}
					</div>
					<div
						v-else-if="bookmarkStore.bookmarks.length === 0"
						class="text-ink-3 q-pa-md text-center"
					>
						{{ t('desktop_settings.no_bookmarks') }}
					</div>
					<q-list v-else class="bookmark-list" bordered separator>
						<q-item
							v-for="bm in bookmarkStore.bookmarks"
							:key="bm.id"
							class="bookmark-row"
						>
							<q-item-section avatar>
								<img
									v-if="bm.icon"
									:src="bm.icon"
									class="bookmark-icon"
									:alt="bm.name"
								/>
								<q-icon v-else name="bookmark" size="24px" />
							</q-item-section>
							<q-item-section>
								<div class="text-ink-1 ellipsis">{{ bm.name }}</div>
								<div class="text-ink-3 text-caption ellipsis">
									{{ bm.url }}
								</div>
							</q-item-section>
							<q-item-section side>
								<div class="row no-wrap items-center">
									<q-chip
										dense
										color="grey-3"
										text-color="ink-2"
										:label="bm.openMode"
									/>
									<q-btn
										flat
										dense
										round
										icon="edit"
										size="sm"
										color="ink-2"
										@click="openEditForm(bm)"
									/>
									<q-btn
										flat
										dense
										round
										icon="delete"
										size="sm"
										color="red"
										@click="confirmDelete(bm)"
									/>
								</div>
							</q-item-section>
						</q-item>
					</q-list>

					<q-banner
						v-if="bookmarkStore.lastError"
						class="bg-red-1 text-red q-mt-md"
						dense
					>
						{{ bookmarkStore.lastError }}
					</q-banner>
				</q-tab-panel>
			</q-tab-panels>
		</q-card>

		<q-dialog v-model="formOpen" persistent>
			<q-card class="form-card" flat>
				<div class="text-ink-1 text-subtitle1 q-mb-md">
					{{
						editingId
							? t('desktop_settings.edit_bookmark')
							: t('desktop_settings.add_bookmark')
					}}
				</div>
				<q-input
					v-model="form.name"
					:label="t('desktop_settings.name')"
					filled
					dense
					maxlength="128"
				/>
				<q-input
					v-model="form.url"
					:label="t('desktop_settings.url')"
					filled
					dense
					class="q-mt-md"
				/>
				<q-input
					v-model="form.icon"
					:label="t('desktop_settings.icon_optional')"
					filled
					dense
					class="q-mt-md"
				/>
				<q-select
					v-model="form.openMode"
					:options="openModeOptions"
					:label="t('desktop_settings.open_mode')"
					filled
					dense
					emit-value
					map-options
					class="q-mt-md"
				/>
				<div class="row justify-end q-mt-lg">
					<q-btn
						flat
						:label="t('buttons.cancel')"
						color="ink-2"
						@click="formOpen = false"
					/>
					<q-btn
						unelevated
						color="primary"
						:label="t('buttons.save')"
						:loading="formSaving"
						@click="submitForm"
					/>
				</div>
			</q-card>
		</q-dialog>
	</q-dialog>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { Bookmark, BookmarkOpenMode } from 'src/types/desktop-layout';
import { useBookmarkStore } from 'src/stores/desktop/bookmark';
import { useApplicationStore } from 'src/stores/desktop/app';
import ConfirmDialog from './ConfirmDialog.vue';
import ResetLayoutAction from './ResetLayoutAction.vue';

const props = defineProps<{
	modelValue: boolean;
}>();
const emit = defineEmits(['update:modelValue']);
const { t } = useI18n();
const $q = useQuasar();
const bookmarkStore = useBookmarkStore();
const appStore = useApplicationStore();

const show = computed({
	get: () => props.modelValue,
	set: (v: boolean) => emit('update:modelValue', v)
});

const tab = ref<'bookmarks' | 'advanced'>('bookmarks');
const formOpen = ref(false);
const editingId = ref<string | null>(null);
const formSaving = ref(false);
const form = reactive<{
	name: string;
	url: string;
	icon: string;
	openMode: BookmarkOpenMode;
}>({
	name: '',
	url: '',
	icon: '',
	openMode: 'iframe'
});

const openModeOptions = computed(() => [
	{ value: 'iframe', label: t('desktop_settings.open_iframe') },
	{ value: 'window', label: t('desktop_settings.open_window') }
]);

watch(show, (v) => {
	if (v) {
		void bookmarkStore.fetchAll();
	}
});

function resetForm() {
	editingId.value = null;
	form.name = '';
	form.url = '';
	form.icon = '';
	form.openMode = 'iframe';
}

function openCreateForm() {
	resetForm();
	formOpen.value = true;
}

function openEditForm(bm: Bookmark) {
	editingId.value = bm.id;
	form.name = bm.name;
	form.url = bm.url;
	form.icon = bm.icon || '';
	form.openMode = bm.openMode;
	formOpen.value = true;
}

async function submitForm() {
	formSaving.value = true;
	try {
		const payload = {
			name: form.name.trim(),
			url: form.url.trim(),
			openMode: form.openMode,
			icon: form.icon.trim() ? form.icon.trim() : undefined
		};
		if (!payload.name || !payload.url) {
			$q.notify({
				type: 'warning',
				message: t('desktop_settings.required_fields')
			});
			return;
		}
		const normalizedName = payload.name.toLowerCase();
		const dupBookmark = bookmarkStore.bookmarks.some(
			(b) =>
				b.id !== editingId.value &&
				b.name.trim().toLowerCase() === normalizedName
		);
		const dupApp = appStore.myApps.some(
			(a) => (a.title || a.name || '').trim().toLowerCase() === normalizedName
		);
		if (dupBookmark || dupApp) {
			$q.notify({
				type: 'warning',
				message: t('desktop_settings.name_exists')
			});
			return;
		}
		if (editingId.value) {
			const updated = await bookmarkStore.update(editingId.value, payload);
			if (updated) {
				// Reflect name/icon/url edits on the live desktop grid + dock.
				appStore.refreshDesktopForBookmarks();
				formOpen.value = false;
			} else {
				$q.notify({
					type: 'negative',
					message: bookmarkStore.lastError || t('desktop_settings.save_failed')
				});
			}
		} else {
			const created = await bookmarkStore.create(payload);
			if (created) {
				// Auto-place onto both LaunchPad surfaces (M2 step m2-fe-autoplace).
				await appStore.autoPlaceBookmarkOnLaunchpads(created.id);
				formOpen.value = false;
			} else {
				$q.notify({
					type: 'negative',
					message: bookmarkStore.lastError || t('desktop_settings.save_failed')
				});
			}
		}
	} finally {
		formSaving.value = false;
	}
}

function confirmDelete(bm: Bookmark) {
	$q.dialog({
		component: ConfirmDialog,
		componentProps: {
			title: t('desktop_settings.delete_bookmark_title'),
			message: t('desktop_settings.delete_bookmark_msg', { name: bm.name }),
			showCancel: true
		}
	}).onOk(async () => {
		// Route through the app store so the bookmark is also removed from the
		// launchpad/dock layouts and the live desktop, not just the CRUD list.
		await appStore.deleteBookmarkFromDesktop(bm.id);
	});
}

function onClose() {
	show.value = false;
}
</script>

<style lang="scss" scoped>
.settings-card {
	width: 640px;
	max-width: 90vw;
	border-radius: 12px;
	padding: 20px;
	background: white;
}
.settings-header {
	margin-bottom: 12px;
}
.settings-panel {
	min-height: 320px;
}
.bookmark-list {
	border-radius: 8px;
	max-height: 360px;
	overflow-y: auto;
}
.bookmark-row {
	min-height: 56px;
}
.bookmark-icon {
	width: 28px;
	height: 28px;
	border-radius: 6px;
	object-fit: contain;
}
.form-card {
	width: 480px;
	max-width: 90vw;
	border-radius: 12px;
	padding: 20px;
	background: white;
}
</style>
