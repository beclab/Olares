<template>
	<div v-show="!deviceStore.isScaning" class="itemView bg-background-1">
		<div class="header">
			<div
				class="row items-center justify-between q-px-md"
				style="height: 56px"
			>
				<div
					:class="['view-hearder', 'q-px-sm', { isedit: editing_t1 }]"
					:style="{
						width: editing_t1 ? 'calc(100% - 40px)' : 'calc(100% - 120px)',
						display: 'flex',
						alignItems: 'center',
						justifyContent: 'center'
					}"
				>
					<q-icon
						v-if="isMobile && !editing_t1"
						name="sym_r_chevron_left"
						size="24px"
						@click="goBack"
					/>
					<q-icon
						v-if="!isMobile && item && item.icon"
						:name="showItemIcon(item.icon)"
						color="ink-1"
						size="24px"
					/>

					<div class="hearder-input q-ml-sm">
						<div v-if="editing_t1">
							<q-input
								borderless
								v-model="name"
								dense
								ref="nameRef"
								:placeholder="t('vault_t.enter_item_name')"
								input-class="text-body3 text-ink-2"
							>
							</q-input>
						</div>
						<div v-else class="full-height column items-start justify-center">
							<template v-if="!name">
								{{ t('new_item') }}
							</template>
							<template v-else>
								<span class="text-ink-3 text-overline">
									{{ vault?.label }}
								</span>
								<span class="vault-name text-ink-1 text-subtitle2">
									{{ name }}
								</span>
							</template>
						</div>
					</div>
				</div>
				<div
					class="row view-option items-center justify-end"
					:style="{ width: editing_t1 ? '40px' : '120px' }"
				>
					<div class="optionItem" v-if="!editing_t1">
						<q-btn
							class="btn-size-sm btn-no-text btn-no-border"
							:class="isFavorite ? 'text-yellow-default' : 'text-ink-2'"
							:icon="isFavorite ? 'star' : 'sym_r_star'"
							@click="setFavorite(!isFavorite)"
						>
							<q-tooltip>{{ t('favorite') }}</q-tooltip>
						</q-btn>
					</div>
					<div
						class="optionItem q-mt-xs q-mx-xs"
						@click="onEdit"
						v-if="!editing_t1 && isEditable"
					>
						<q-btn
							class="btn-size-sm btn-no-text btn-no-border"
							icon="sym_r_edit_note"
							text-color="ink-2"
						>
							<q-tooltip>{{ t('buttons.edit') }}</q-tooltip>
						</q-btn>
					</div>

					<div
						class="optionItem q-mt-xs q-mx-xs"
						v-if="!editing_t1 && !isEditable"
						disabled
					>
						<q-btn
							class="btn-size-sm btn-no-text btn-no-border"
							icon="sym_r_edit_note"
							text-color="ink-2"
						>
							<q-tooltip>{{ t('buttons.edit') }}</q-tooltip>
						</q-btn>
					</div>

					<div class="optionItem" v-if="editing_t1">
						<q-btn
							class="btn-size-sm btn-no-text btn-no-border"
							icon="sym_r_add"
							text-color="ink-2"
							:color="showAddField ? 'grey-1' : ''"
						>
							<q-tooltip>{{ t('vault_t.add_field') }}</q-tooltip>
							<q-menu
								class="popup-menu bg-background-2"
								flat
								style="overflow-y: scroll"
								v-model="showAddField"
							>
								<q-list class="q-py-sm" style="min-width: 200px">
									<template
										v-for="(filed, index) in [...Object.values(FIELD_DEFS)]"
										:key="index"
									>
										<q-item
											dense
											class="popup-item row items-center justify-center text-ink-2 q-px-sm q-mt-xs"
											clickable
											v-close-popup
											@click="addFieldClick(filed)"
										>
											<span class="optionIcon">
												<q-icon :name="filed.icon" size="24px" />
											</span>
											<q-item-section class="q-ml-sm">{{
												translate(`${filed.name}`)
											}}</q-item-section>
										</q-item>
									</template>
								</q-list>
							</q-menu>
						</q-btn>
					</div>

					<div class="optionItem" v-if="!editing_t1 && isEditable">
						<q-btn
							class="btn-size-sm btn-no-text btn-no-border"
							icon="sym_r_more_horiz"
							text-color="ink-2"
						>
							<q-tooltip>{{ t('buttons.more') }}</q-tooltip>
							<q-menu class="popup-menu bg-background-2">
								<q-list class="q-py-sm" dense padding>
									<q-item
										class="popup-item row items-center justify-start text-body3 text-ink-2"
										clickable
										dense
										v-close-popup
										@click="moveItem"
										style="white-space: nowrap"
									>
										<q-icon size="16px" name="sym_r_move_up" class="q-mr-sm" />
										{{ t('vault_t.move_to_vault') }}
									</q-item>
									<q-item
										class="popup-item row items-center justify-start text-body3 text-ink-2"
										clickable
										dense
										v-close-popup
										@click="deleteItem"
									>
										<q-icon size="16px" name="sym_r_delete" class="q-mr-sm" />
										{{ t('vault_t.delete_item') }}
									</q-item>
								</q-list>
							</q-menu>
						</q-btn>
					</div>
				</div>
			</div>
		</div>
		<div class="container2">
			<q-scroll-area
				style="height: 100%"
				:thumb-style="scrollBarStyle.thumbStyle"
			>
				<div>
					<div class="tags listRow">
						<div
							class="row items-center justify-start q-px-lg q-py-sm text-ink-2"
						>
							<q-icon name="sym_r_more" size="24px" />
							<span class="text-body3 q-ml-xs">
								{{ t('tags') }}
							</span>
						</div>
					</div>

					<div class="tags listRow q-py-xs q-px-lg">
						<q-select
							:readonly="!isEditable"
							class="tagSelect q-pl-xs"
							popup-content-class="options_selected_Account tags"
							v-model="tags"
							behavior="menu"
							use-input
							use-chips
							multiple
							dense
							borderless
							stack-label
							hide-bottom-space
							hide-dropdown-icon
							:placeholder="t('vault_t.add_tags_placeholder')"
							:options="filterTagOptions"
							option-label="name"
							option-value="name"
							emit-value
							map-options
							@new-value="createTagValue"
							@filter="filterTagFn"
							@focus="focusTagFn"
						>
							<template v-slot:selected-item="scope">
								<q-chip
									:removable="
										chipShowRemoveIcon &&
										chipShowRemoveIcon.index === scope.index
											? true
											: false
									"
									square
									icon="sym_r_sell"
									@remove="scope.removeAtIndex(scope.index)"
									@mouseover="chipMouseOver(scope)"
									@mouseleave="chipMouseLeave(scope)"
									:tabindex="scope.tabindex"
									class="q-ma-none tagChip text-overline"
								>
									{{ scope.opt.name }}
								</q-chip>
							</template>

							<template v-slot:option="scope">
								<q-item
									class="row items-center justify-between text-ink-1"
									dense
									v-bind="scope.itemProps"
								>
									<span class="row items-center">
										<q-icon class="q-mr-sm" name="sym_r_sell" />
										<q-item-label>{{ scope.opt.name }}</q-item-label>
									</span>
									<q-item-label>{{ scope.opt.count }}</q-item-label>
								</q-item>
							</template>
						</q-select>
					</div>

					<div
						class="listRow row items-center justify-start text-body3 q-py-md q-px-lg q-py-sm text-ink-2"
					>
						<q-icon name="sym_r_article" size="20px" />
						<span class="q-ml-xs">{{ t('fields') }}</span>
					</div>

					<div class="fileds column">
						<div
							v-for="(field, index) in item?.fields"
							:key="itemID + 'fa' + index"
							style="overflow: hidden"
							:style="{ height: editing_t1 ? '85px' : 'auto' }"
						>
							<FiledComponent2
								:field="field"
								:index="index"
								:editing="editing_t1"
								:isEditable="isEditable"
								:masked="false"
								@fieldUpdate="updateFiled"
								:canMoveUp="!!index"
								:canMoveDown="index < (item?.fields.length || 0) - 1"
								@remove="removeField(index)"
								@moveup="moveField(index, 'up')"
								@movedown="moveField(index, 'down')"
								@onEdit="onEdit"
								@startScan="startScan"
							/>
						</div>
					</div>

					<div
						class="listRow cursor-pointer text-ink-3 text-body3 q-py-md"
						@click="openMenu"
					>
						<div v-if="isEditable" class="row items-center justify-center">
							<q-icon name="sym_r_add" size="20px" />
							<span class="q-ml-xs">{{ t('vault_t.add_field') }}</span>
						</div>
						<div v-else disabled class="row items-center justify-center">
							<q-icon name="sym_r_add" size="20px" />
							<span class="q-ml-xs">{{ t('vault_t.add_field') }}</span>
						</div>
					</div>

					<template v-if="!isBex">
						<div
							class="listRow q-pa-md q-pl-lg row items-center text-ink-2 text-body3"
						>
							<q-icon name="sym_r_lab_profile" size="20px" />
							<span class="text-li-title">
								{{ t('attachments') }}
							</span>
						</div>

						<div
							class="listRow attach q-pa-md row justify-start q-px-lg"
							v-for="(attach, index3) in attachments"
							:key="itemID + 'aa' + index3"
							@click="
								() => {
									if (!editing_t1) {
										openAttachment(attach);
									}
								}
							"
						>
							<div
								v-if="editing_t1"
								class="reduce row items-center justify-left q-mr-xs"
								@click.stop="onDelete(attach)"
							>
								<q-icon
									name="sym_r_do_not_disturb_on"
									size="20px"
									style="padding-left: 2px"
								/>
							</div>
							<div class="attachment row justify-start">
								<q-icon
									name="sym_r_lab_profile"
									size="16px"
									class="text-blue q-mb-xs"
									style="margin-left: 2px"
								/>
								<AttachmentComponent
									:itemID="item?.id!"
									:attach="attach"
									:index="index3"
									:editing="editing_t1"
									@remove="removeAttach(attach)"
								/>
							</div>
						</div>

						<div class="listRow q-pa-md">
							<div
								v-if="isEditable && isLarePassActive"
								class="row items-center justify-center uploadFile text-ink-3 text-body3"
							>
								<TerminusSelectLocalFile @on-success="chooseAttachment">
									<div class="text-body3 row items-center">
										<q-icon name="sym_r_add" size="20px" />
										<span class="text-li-title">
											{{
												t(
													'vault_t.click_or_drag_files_here_to_add_an_attachment'
												)
											}}
										</span>
									</div>
								</TerminusSelectLocalFile>
							</div>
							<div
								v-else
								disabled
								class="row items-center justify-center uploadFile text-ink-3 text-body3"
							>
								<q-icon name="sym_r_add" size="20px" />
								<span class="text-li-title">
									{{
										t('vault_t.click_or_drag_files_here_to_add_an_attachment')
									}}
								</span>
							</div>
						</div>
					</template>

					<div
						class="listRow q-pa-md q-pl-lg row items-center text-ink-2 text-body3"
					>
						<q-icon name="sym_r_today" size="20px" />
						<span class="text-li-title">
							{{ t('expiration') }}
						</span>
					</div>

					<div
						class="listRow q-pa-md text-ink-3 text-body3 cursor-pointer"
						v-if="!isEditExpir && !expiresAfter_t1"
						@click="handleEditExpir(1)"
					>
						<!-- <div class="row items-center justify-center" v-if="editena">

						</div> -->
						<div v-if="isEditable" class="row items-center justify-center">
							<q-icon name="sym_r_add" size="20px" />
							<span class="text-li-title">
								{{ t('vault_t.add_expiration') }}
							</span>
						</div>
						<div v-else disabled class="row items-center justify-center">
							<q-icon name="sym_r_add" size="20px" />
							<span class="text-li-title">
								{{ t('vault_t.add_expiration') }}
							</span>
						</div>
					</div>

					<div
						class="listRow q-pa-md row items-center justify-center"
						v-if="!isEditExpir && expiresAfter_t1"
					>
						<span class="text-li-title">
							<!-- {{ item?.expiresAt }} -->
							<!-- {{ date.formatDate(item?.expiresAt, 'YYYY-MM-DD') }} -->
							<!-- {{ now }} -->
							<span class="text-ink-1 text-subtitle3">
								{{
									item?.expiresAt && item.expiresAt > now
										? t('expires') + ' '
										: t('expired') + ' '
								}}
							</span>
							<span class="text-ink-1 text-subtitle3">
								{{
									item?.expiresAt
										? formatDateFromNow(item.expiresAt, false)
										: ''
								}}.
							</span>
						</span>
					</div>

					<div
						class="listRow q-pa-md row items-center justify-center"
						v-if="isEditExpir"
					>
						<span class="text-ink-1 text-body3">{{ t('expire') }}</span>
						<q-input
							class="q-mx-sm expireInput"
							inputClass="q-pl-sm"
							borderless
							dense
							type="number"
							v-model="expiresAfter_t1"
							@update:model-value="onUpdateExpiresAfter"
							onkeyup="this.value=this.value.replace(/\D|^0/g,'')"
							input-style="width: 100%; height: 32px;"
						/>
					</div>

					<div
						class="listRow q-pa-md row items-center justify-center text-ink-3 text-body3"
						v-if="isEditExpir"
						@click="handleEditExpir(0)"
					>
						<q-icon name="sym_r_do_not_disturb_on" size="20px" />
						<span class="text-li-title">
							{{ t('vault_t.remove_expiratio') }}
						</span>
					</div>

					<div
						class="listRow q-pa-md q-pl-lg row items-center text-ink-2 text-body3"
					>
						<q-icon name="sym_r_history" size="20px" />
						<span class="text-li-title">
							{{ t('history') }}
						</span>
					</div>

					<div
						class="listRow q-pa-md q-pl-lg row items-center justify-between"
						style="border-bottom: 0"
					>
						<div class="hisData text-body3" v-if="item">
							<q-icon class="q-mr-sm" name="sym_r_schedule" size="20px" />
							<span class="q-mr-xs text-ink-2">{{
								formatDateTime(item.updated)
							}}</span>
							<span class="text-ink-3"
								>({{ formatDateFromNow(item.updated) }})</span
							>
						</div>

						<div
							class="currenetVersion q-pa-xs q-mr-md text-overline text-ink-2"
						>
							{{ t('vault_t.current_version') }}
						</div>
					</div>

					<div v-if="item?.history && item.history.length > 0">
						<template
							v-for="(history, index) in item?.history"
							:key="'hs' + index"
						>
							<div
								class="listRow q-pa-md q-pl-lg row items-center justify-between history"
								@click="showHistoryEntry(index)"
								style="border-bottom: 0"
							>
								<div class="hisData text-body3">
									<q-icon class="q-mr-sm" name="sym_r_schedule" size="20px" />
									<span class="q-mr-xs text-ink-2">{{
										formatDateTime(history.updated)
									}}</span>
									<span class="text-ink-3"
										>({{ formatDateFromNow(history.updated) }})</span
									>
								</div>

								<q-icon
									class="visibility q-mr-md"
									name="sym_r_visibility"
									size="24px"
									color="ink-2"
								/>
							</div>
						</template>
					</div>
				</div>
			</q-scroll-area>
		</div>
		<div
			v-if="editing_t1"
			class="footer row iterm-center justify-between bg-background-1"
		>
			<q-btn
				class="reset text-ink-1"
				:class="isMobile ? 'text-subtitle1' : 'text-body3 btn-height-web'"
				:label="t('cancel')"
				type="reset"
				flat
				dense
				no-caps
				@click="onCancel"
				unelevated
				color="ink-2"
			/>
			<q-btn
				class="confirm text-grey-10"
				:class="isMobile ? 'text-subtitle1' : 'text-body3 btn-height-web'"
				:label="t('buttons.save')"
				type="submit"
				@click="onSave"
				unelevated
				no-caps
				dense
				color="yellow-6"
				:loading="saveLoading"
			/>
		</div>
	</div>
	<ScanComponent
		v-if="deviceStore.isScaning"
		@cancel="scanCancel"
		@scan-result="scanResult"
	/>
</template>

<script lang="ts" setup>
import {
	defineComponent,
	computed,
	ref,
	watch,
	onMounted,
	nextTick
} from 'vue';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { VaultItem, FIELD_DEFS, FieldDef, Field } from '@didvault/sdk/src/core';
import { formatDateTime, translate } from '@didvault/sdk/src/util';
import { formatDateFromNow } from 'src/utils/format';
import { app } from '../../globals';
import FiledComponent2 from './FiledComponent2.vue';
import { AttachmentInfo } from '@didvault/sdk/src/core';
import { auditVaults } from '../../utils/audit';
import { Dialog, useQuasar } from 'quasar';
import { useRouter } from 'vue-router';
import UploadAttachment from './UploadAttachment.vue';
import AttachmentComponent from './AttachmentComponent.vue';
import OpenAttachment from './OpenAttachment.vue';
import HistoryEntryDialog from './dialog/HistoryEntryDialog.vue';
import { useMenuStore } from '../../stores/menu';

import MoveItemsPC from './dialog/MoveItemsPC.vue';
import MoveItemsMobile from './dialog/MoveItemsMobile.vue';
import DeleteItem from './dialog/DeleteItem.vue';

import { showItemIcon } from './../../utils/utils';
import { scrollBarStyle } from '../../utils/contact';
import ScanComponent from '../../components/common/ScanComponent.vue';
import { notifyFailed } from '../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import { bexVaultUpdate } from 'src/utils/bexFront';
import { useVaultStore } from '../../stores/vault';
import { BtDialog } from '@bytetrade/ui';
import { useDeviceStore } from '../../stores/device';
import { getAppPlatform } from 'src/application/platform';
import { useTermipassStore } from 'src/stores/termipass';
import { UserStatusActive } from 'src/utils/checkTerminusState';
// import { redirectToSecretInPLugin } from 'src/utils/common-safe';
import TerminusSelectLocalFile from '../../components/common/TerminusSelectLocalFile.vue';

const props = defineProps({
	itemID: {
		type: String,
		required: true
	},
	isNew: {
		type: Boolean,
		required: true
	}
});

const $q = useQuasar();
const vaultStore = useVaultStore();
const Router = useRouter();
const editing_t1 = ref(props.isNew);
const showAddField = ref(false);
const meunStore = useMenuStore();
const now = new Date();
const larePassStore = useTermipassStore();

const nameValue = ref();

const item = ref<VaultItem | null>();
const expiresAfter_t1 = ref<number | undefined>(undefined);
const isFavorite = ref(false);

const name = ref();
const attachments: any = ref([]);
const file = ref(null);
const isEditExpir = ref(false);
const nameRef = ref();
const fieldsForm = ref();
const field_defs_filter = ref();

const tags: any = ref([]);
let stringOptions: any[] = [];
const chipShowRemoveIcon = ref();
const saveLoading = ref(false);
const deviceStore = useDeviceStore();

const { t } = useI18n();

let filterTagOptions = ref(stringOptions);
const isMobile = ref(
	process.env.PLATFORM == 'MOBILE' ||
		process.env.PLATFORM == 'BEX' ||
		$q.platform.is.mobile
);

const isLarePass = getAppPlatform().isClient;

const isBex = ref(process.env.PLATFORM == 'BEX' ? true : false);

const focusTagFn = (): void => {
	if (isEditable.value) {
		editing_t1.value = true;
	}
};

function createTagValue(val: string, done: any) {
	if (val.length > 0) {
		const hasOption = stringOptions.find((item) => item.name === val);
		if (!hasOption) {
			const obj = {
				name: val,
				count: 1,
				readonly: 0
			};
			stringOptions.push(obj);
		}
		done(val, 'toggle');
	}
}

function filterTagFn(val: string, update: any) {
	update(() => {
		if (val === '') {
			filterTagOptions.value = stringOptions;
		} else {
			const needle = val.toLowerCase();
			filterTagOptions.value = stringOptions.filter(
				(v) => v.name.toLowerCase().indexOf(needle) > -1
			);
		}
	});
}

function refreshItem(itemID: string, isField?: boolean) {
	// 	if (vaultStore.editing_item && vaultStore.editing_item.item) {
	// 	item.value = vaultStore.editing_item.item;
	// 	if (isField && item.value) {
	// 		fieldsForm.value = [...item.value.fields];
	// 	}

	// 	isFavorite.value = app.account!.favorites.has(props.itemID);

	// 	name.value = item.value?.name;
	// 	expiresAfter_t1.value = item.value?.expiresAfter || 0;
	// 	tags.value = item.value ? item.value.tags : [];
	// 	stringOptions = app.tags;
	// 	filterTagOptions.value = app.tags;
	// 	attachments.value = item.value?.attachments;
	// } else
	if (app.getItem(itemID)) {
		let item2 = app.getItem(itemID)!.item.clone();
		app.updateLastUsed(item2);
		item.value = item2;

		if (isField) {
			fieldsForm.value = [...item.value.fields];
		}

		isFavorite.value = app.account!.favorites.has(props.itemID);

		name.value = item.value.name;
		expiresAfter_t1.value = item.value.expiresAfter || 0;
		tags.value = [...item.value.tags];
		stringOptions = app.tags;
		filterTagOptions.value = app.tags;
		attachments.value = item.value.attachments;
	} else {
		item.value = null;
	}
}

watch(
	() => props.itemID,
	(newVaule, oldVaule) => {
		if (oldVaule == newVaule) {
			return;
		}
		if (!newVaule) {
			return;
		}
		editing_t1.value = false;
		refreshItem(newVaule, true);
	}
);

watch(
	() => props.isNew,
	(newVal) => {
		if (newVal) {
			setTimeout(() => {
				editing_t1.value = newVal;
				nameRef.value && nameRef.value.focus();
			}, 500);
		}
	}
);

onMounted(() => {
	field_defs_filter.value = FIELD_DEFS;
	delete field_defs_filter.value.note;
	delete field_defs_filter.value.apiSecret;

	refreshItem(props.itemID, true);
	if (props.isNew) {
		nextTick(() => {
			nameRef.value && nameRef.value.focus();
		});
	}
});

const vault = computed(() => {
	if (!props.itemID) {
		return null;
	}
	return app.getItem(props.itemID)?.vault;
});

const isEditable = computed(() => {
	const enable = vault.value ? app.isEditable(vault.value) : true;
	return enable;
});

const isLarePassActive = computed(() => {
	if (!isLarePass) {
		return true;
	}
	return larePassStore.totalStatus?.isError === UserStatusActive.active;
});

const setFavorite = async (favorite: boolean) => {
	isFavorite.value = favorite;
	await app.toggleFavorite(props.itemID, favorite);
};

const isBlank = computed(function () {
	if (app.state.locked || !item.value || !vault.value) {
		return true;
	}
	return false;
});

function onEdit() {
	if (!isEditable.value) {
		return;
	}

	if (!editing_t1.value) {
		editing_t1.value = true;
	}

	let item2 = app.getItem(props.itemID)!.item.clone();
	isEditExpir.value = item2.expiresAfter ? true : false;
}

function onCancel() {
	meunStore.isEdit = false;
	isEditExpir.value = false;
	vaultStore.editing_item = null;
	if (editing_t1.value) {
		editing_t1.value = false;
	}

	if (app.getItem(props.itemID)) {
		let item2 = app.getItem(props.itemID)!.item.clone();
		expiresAfter_t1.value = item2.expiresAfter || 0;
		item.value = item2;
	}

	clearChanges();
}

const clearChanges = async () => {
	if (props.isNew) {
		await app.deleteItems([item.value!]);
		bexVaultUpdate();
		// goBack();
		if (isMobile.value) {
			goBack();
		} else {
			Router.push({
				path: '/items/'
			});
		}
	}
};

async function deleteItem() {
	$q.dialog({
		component: DeleteItem,
		componentProps: {
			title: t('delete_vault'),
			content: t('delete_vault_message')
		}
	}).onOk(async () => {
		await app.deleteItems([item.value!]);
		bexVaultUpdate();
		editing_t1.value = false;
		Router.push({
			path: '/items/'
		});
	});
}

async function onSave() {
	if (!name.value) {
		notifyFailed(t('vault_t.item_name_is_null'));
		return;
	}

	if (item.value!.fields.find((cell) => cell.type === 'totp' && !cell.value)) {
		notifyFailed(t('vault_t.one_time_password_is_required'));
		return;
	}

	isEditExpir.value = false;
	meunStore.isEdit = false;
	saveLoading.value = true;

	// if (vaultStore.editing_item) {
	// 	vaultStore.editing_item = await addItem(
	// 		name.value,
	// 		vaultStore.editing_item.item.icon,
	// 		vaultStore.editing_item.item.fields,
	// 		vaultStore.editing_item.item.tags,
	// 		vaultStore.editing_item.vault,
	// 		[],
	// 		undefined,
	// 		expiresAfter_t1.value as any,
	// 		attachments.value,
	// 		true,
	// 		props.itemID
	// 	);
	// } else {
	await app.updateItem(item.value!, {
		name: name.value,
		fields: fieldsForm.value,
		tags: [...tags.value],
		auditResults: [],
		lastAudited: undefined,
		expiresAfter: expiresAfter_t1.value,
		attachments: attachments.value
	});
	auditVaults([vault.value!], {
		updateOnlyItemWithId: item.value!.id
	});
	// }

	saveLoading.value = false;
	refreshItem(props.itemID);
	bexVaultUpdate();
	vaultStore.editing_item = null;

	if (editing_t1.value) {
		editing_t1.value = false;
	}
	// redirectToSecretInPLugin();
}

async function _addField(fieldDef: FieldDef) {
	const fileObj = new Field({
		name: fieldDef.name,
		value: '',
		type: fieldDef.type
	});
	item.value!.fields.push(fileObj);
	fieldsForm.value.push(fileObj);
}

const openMenu = () => {
	if (!isEditable.value) {
		return;
	}
	showAddField.value = true;
	editing_t1.value = true;
};

async function addFieldClick(chooseField) {
	if (!chooseField) {
		return;
	}
	_addField(chooseField);
}

function updateFiled(ob: any) {
	fieldsForm.value[ob.index].value = ob.value;
}

async function removeField(index: number) {
	if (!isEditable.value) {
		return;
	}
	BtDialog.show({
		title: t('vault_t.remove_field'),
		message: t('vault_t.remove_field_message'),
		okStyle: {
			background: 'yellow-default',
			color: '#1F1F1F'
		},
		cancel: true,
		okText: t('base.confirm'),
		cancelText: t('base.cancel')
	})
		.then(async (res: any) => {
			if (res) {
				item.value!.fields = item.value!.fields.filter((_, i) => i !== index);
				fieldsForm.value = [...(item.value as any).fields];
			}
		})
		.catch((err: Error) => {
			console.log('click cancel', err);
		});
}

function moveField(index: number, target: 'up' | 'down' | number) {
	const field = item.value!.fields[index];
	item.value!.fields.splice(index, 1);
	const targetIndex =
		target === 'up' ? index - 1 : target === 'down' ? index + 1 : target;

	item.value!.fields.splice(targetIndex, 0, field);

	fieldsForm.value = [...(item.value as any).fields];
}

function openAttachment(attach: AttachmentInfo) {
	if (!isEditable.value) {
		return;
	}
	if (editing_t1.value) {
		return;
	}

	$q.dialog({
		component: OpenAttachment,
		componentProps: {
			itemID: props.itemID,
			info: attach
		}
	}).onOk((isModify) => {
		if (isModify) {
			refreshItem(props.itemID);
		}
	});
}

async function _addFileAttachment(file: File) {
	if (!isEditable.value) {
		return;
	}
	if (!file) {
		return;
	}

	if ($q.platform.is.nativeMobile && file.size > 5 * 1024 * 1024) {
		notifyFailed(
			t(
				'vault_t.the_selected_file_is_too_large_only_files_of_up_to_5m_are_supported'
			)
		);
		return;
	} else if (file.size > 1024 * 1024 * 1024) {
		notifyFailed(
			t(
				'vault_t.the_selected_file_is_too_large_only_files_of_up_to_1t_are_supported'
			)
		);
		return;
	}

	$q.dialog({
		component: UploadAttachment,
		componentProps: {
			itemID: props.itemID,
			file: file
		}
	}).onOk(() => {
		BtNotify.show({
			type: NotifyDefinedType.SUCCESS,
			message: t('vault_t.upload_complete_successfully')
		});

		if (app.getItem(props.itemID)) {
			let item2 = app.getItem(props.itemID)!.item;
			attachments.value = item2.attachments;
			item.value!.updated = item2.updated;
		}
	});
}

function chooseAttachment(filelist: FileList) {
	if (filelist.length != 1) {
		return;
	}

	let f: File = filelist[0];

	_addFileAttachment(f);
}

async function removeAttach(att: AttachmentInfo) {
	const confirmed = await new Promise((resolve) =>
		Dialog.create({
			title: t('vault_t.delete_attachment'),
			message: t('vault_t.delete_attachment_message'),
			cancel: true,
			persistent: true
		})
			.onOk(() => {
				resolve(true);
			})
			.onCancel(() => {
				resolve(false);
			})
	);
	if (confirmed) {
		await app.deleteAttachment(props.itemID, att);
		refreshItem(props.itemID);
	}
}

const handleEditExpir = (type: number) => {
	if (!isEditable.value) {
		return;
	}
	if (!type) {
		isEditExpir.value = false;
		expiresAfter_t1.value = 0;
	} else {
		isEditExpir.value = true;
		editing_t1.value = true;
	}
};

const onDelete = (attach) => {
	meunStore.dialogShow = true;
	BtDialog.show({
		title: t('vault_t.delete_attachment'),
		message: t('vault_t.delete_attachment_message'),
		okStyle: {
			background: 'yellow-default',
			color: '#1F1F1F'
		},
		cancel: true,
		okText: t('base.confirm'),
		cancelText: t('base.cancel')
	})
		.then(async (res: any) => {
			if (res) {
				await app.deleteAttachment(props.itemID, attach);
				await refreshItem(props.itemID);
			}
		})
		.catch((err: Error) => {
			console.log('click cancel', err);
		});
};

const toggleDrawer = () => {
	meunStore.rightDrawerOpen = false;
};

const showHistoryEntry = (historyIndex: number) => {
	if (!isEditable.value) {
		return false;
	}

	$q.dialog({
		component: HistoryEntryDialog,
		componentProps: {
			item,
			vault,
			historyIndex
		}
	}).onOk(async () => {
		console.log('sssres');

		restoreHistoryEntry(item.value, historyIndex);
	});
};

const restoreHistoryEntry = (item, historyIndex) => {
	BtDialog.show({
		title: t('vault_t.restore_version'),
		message: t('vault_t.restore_version_message'),
		okStyle: {
			background: 'yellow-default',
			color: '#1F1F1F'
		},
		platform: $q.platform.is.mobile ? 'mobile' : 'web',
		cancel: true,
		okText: t('base.confirm'),
		cancelText: t('base.cancel')
	})
		.then(async (res: any) => {
			if (res) {
				const historyEntry = item!.history[historyIndex];
				app.updateItem(item!, {
					name: historyEntry.name,
					fields: historyEntry.fields,
					tags: historyEntry.tags,
					auditResults: [],
					lastAudited: undefined,
					attachments: attachments.value
				});

				refreshItem(props.itemID);
				auditVaults([vault.value!], {
					updateOnlyItemWithId: item.value!.id
				});
				Router.push({
					path: '/items/' + (props.itemID ? props.itemID : '')
				});
			}
		})
		.catch((err: Error) => {
			console.log('click cancel', err);
		});
};

const chipMouseOver = (item) => {
	chipShowRemoveIcon.value = item;
};

const chipMouseLeave = () => {
	chipShowRemoveIcon.value = null;
};

const moveItem = async () => {
	if (item.value?.attachments && item.value.attachments.length > 0) {
		$q.dialog({
			title: t('confirm'),
			message: t('vault_t.can_not_move_item_message'),
			cancel: true,
			persistent: true
		});
	} else {
		await showMoveItemsDialog();
	}
};

const showMoveItemsDialog = () => {
	const hasCheckBox = [{ item: item.value, vault: vault }];

	$q.dialog({
		component: isMobile.value ? MoveItemsMobile : MoveItemsPC,
		componentProps: {
			selected: hasCheckBox,
			leftText: t('cancel'),
			rightText: t('vault_t.move_item')
		}
	});
};

const goBack = () => {
	Router.go(-1);
};

const scanCancel = () => {
	deviceStore.isScaning = false;
};

const scanResult = (result: string) => {
	let url = new URL(result);
	let params = new URLSearchParams(url.search);
	let secret = params.get('secret');

	if (!result.startsWith('otpauth') || !secret || !item.value) {
		notifyFailed(t('errors.invalid_code_please_try_again'));
		return false;
	}

	fieldsForm.value[scanIndex.value].value = secret;
	item.value.fields[scanIndex.value].value = secret;
	deviceStore.isScaning = false;
	scanIndex.value = undefined;
};

const startScan = (index: any) => {
	scanIndex.value = index;
	deviceStore.isScaning = true;
};

// const scanIng = ref(false);
const scanIndex = ref();

const onUpdateExpiresAfter = (expires: number) => {
	if (expires < 999999) {
		return;
	}
	expiresAfter_t1.value = 999999;
};
</script>

<style lang="scss" scoped>
.itemView {
	width: 100%;
	height: 100%;
	display: flex;
	flex-direction: column;
	overflow: hidden;
	float: right;

	.view-hearder {
		border-radius: 8px;

		.hearder-input {
			width: calc(100% - 32px);
			display: inline-block;
			height: 32px;

			.text {
				height: 40px;
				line-height: 40px;
				overflow: hidden;
				white-space: nowrap;
				text-overflow: ellipsis;
			}
		}

		&:focus-within {
			border: 1px solid $light-blue-default;
		}
	}

	.vault-name {
		display: inline-block;
		width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.isedit {
		height: 32px;
		border: 1px solid $input-stroke;
		::v-deep(.q-field--dense .q-field__control) {
			height: 32px;
		}
	}

	.view-option {
		.optionItem {
			text-align: center;
			display: flex;
			align-items: center;
			justify-content: center;
			cursor: pointer;

			span.optionIcon {
				display: inline-block;
			}
		}
	}

	.container2 {
		flex: 1 1 auto;

		.tagSelect {
			width: 90%;

			.tagChip {
				height: 20px;
				line-height: 20px;
				color: $ink-2;
				margin-right: 5px;
				border: 1px solid $ink-2;
				background-color: $background-1;
			}
		}
	}
}

.header {
	flex: 0 0 auto;
}

.popup-menu {
	max-width: auto;
}

.footer {
	width: 100%;
	padding: 10px 20px;
	.confirm {
		width: 48%;
		height: 48px;
		&.btn-height-web {
			height: 32px;
		}
	}
	.reset {
		width: 48%;
		height: 48px;
		border: 1px solid $separator;

		&.btn-height-web {
			height: 32px;
		}
	}
}

.history {
	.visibility {
		opacity: 0;
	}

	position: relative;

	.guide {
		position: absolute;
		left: 22px;
		top: -16px;
		width: 0px;
		height: 32px;
		border-left: 1px solid $separator;
	}

	&:hover {
		cursor: pointer;
		background-color: $background-hover;

		.visibility {
			opacity: 1;
		}
	}
}

.listRow {
	.uploadFile {
		position: relative;
		overflow: hidden;

		.uploadInput {
			position: absolute;
			width: 100%;
			height: 100%;
			top: 0;
			left: 0;
			outline: none;
			filter: alpha(opacity=0);
			-moz-opacity: 0;
			-khtml-opacity: 0;
			opacity: 0;
		}
	}

	.expireInput {
		width: 60px;
		height: 32px;
		font-size: map-get($map: $body2, $key: size);
		border: 1px solid $input-stroke;
		border-radius: 8px;
		::v-deep(.q-field__inner) {
			height: 32px;
		}
		::v-deep(.q-field--dense .q-field__control) {
			height: 32px;
		}

		::v-deep(.q-field--dense .q-field__marginal) {
			height: 32px;
		}
	}

	.attachment {
		width: 80%;
		word-break: break-all;
	}

	.hisData {
		width: calc(100% - 104px);
		line-height: 100%;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	.currenetVersion {
		border-radius: 4px;
		border: 1px solid $yellow-default;
		background: $yellow-alpha;
	}
}

.attach {
	overflow: hidden;

	.reduce {
		width: 22px;
		cursor: pointer;
	}

	.attach {
		flex: 1;
	}
}

.q-field__control {
	height: 34px !important;
}

.text-li-title {
	margin-left: 5px;
}
</style>
