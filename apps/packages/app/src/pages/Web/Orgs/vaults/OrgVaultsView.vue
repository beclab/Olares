<template>
	<div
		class="orgItemView bg-background-1 text-ink-1 row items-center justify-center"
		v-if="isBlank"
	>
		<img
			class="q-mb-md"
			style="margin-top: 64px"
			src="../../../../assets/layout/nodata.svg"
		/>
		<span>
			{{ t('no_vault_selected') }}
		</span>
	</div>
	<div v-else class="orgItemView bg-background-1">
		<div class="header">
			<div class="row items-center justify-between q-pa-md">
				<div
					:class="{
						'view-width': !editing_t1,
						'full-width': editing_t1,
						'view-hearder': true
					}"
					calss="row items-center"
				>
					<div class="hearder-input row items-center">
						<q-icon
							v-if="isMobile"
							name="sym_r_chevron_left"
							size="24px"
							@click="goBack"
						/>
						<q-input
							class="create-valut-input"
							v-if="editing_t1"
							v-model="name"
							dense
							borderless
							color="grey-7"
							ref="nameRef"
							:placeholder="t('vault_t.enter_item_name')"
							input-class="text-body3"
							:style="{
								width: isMobile ? 'calc(100% - 40px' : '100%'
							}"
						/>
						<div v-else class="text text-subtitle1" @click="nameClick">
							{{ name ? name : t('new_item') }}
						</div>
					</div>
				</div>
				<div
					class="row items-center justify-between view-option"
					v-if="!editing_t1"
				>
					<q-icon name="sym_r_edit_note" size="24px" @click="onEdit" />

					<q-icon name="sym_r_more_horiz" size="24px">
						<q-menu class="popup-menu">
							<q-list dense padding>
								<q-item
									class="row items-center justify-start popup-item"
									style="width: 140px"
									clickable
									v-close-popup
									@click="onDelete"
								>
									<q-icon size="22px" name="sym_r_delete" class="q-mr-sm" />
									{{ t('delete') }}
								</q-item>
							</q-list>
						</q-menu>
					</q-icon>
				</div>
			</div>
		</div>
		<div class="container2">
			<q-scroll-area
				style="height: 100%"
				:thumb-style="scrollBarStyle.thumbStyle"
			>
				<div class="listRow column justify-center q-mx-md">
					<div class="header q-pa-md row justify-between">
						<div class="items-center">
							<span class="text-ink-1 text-li-title">
								{{ t('members') }}
							</span>
						</div>
						<div>
							<q-icon name="sym_r_add" size="24px" />
							<q-menu class="popup-menu" v-if="_availableMembers.length > 0">
								<q-list dense padding>
									<template
										v-for="(vault, index) in _availableMembers"
										:key="'avn' + index"
									>
										<q-item
											class="column popup-item"
											clickable
											v-close-popup
											@click="addMember(vault)"
											style="width: 140px; white-space: nowrap"
										>
											<div class="text-subtitle2 member-title">
												{{ vault.did }}
											</div>
											<div class="text-subtitle2 member-section">
												{{ userStore.getCurrentDomain() }}
											</div>
										</q-item>
										<q-separator v-if="index < _availableMembers.length - 1" />
									</template>
								</q-list>
							</q-menu>
							<q-menu v-else>
								<q-item
									class="row items-center justify-center"
									v-close-popup
									style="white-space: nowrap"
								>
									{{ t('no_more_members_available') }}
								</q-item>
							</q-menu>
						</div>
					</div>

					<div class="body1">
						<div
							v-if="members.length == 0"
							class="row items-center justify-center"
							style="height: 160px"
						>
							{{ t('no_members_have_been_given_access_to_this_vault_yet') }}
						</div>

						<template v-else>
							<div
								class="listRow-content q-pa-md row items-center justify-between"
								:class="index < members.length - 1 ? 'borderBottom' : ''"
								v-for="(member, index) in members"
								:key="'member' + index"
							>
								<div class="col-7 rowLeft">
									<div class="avator q-mr-md">
										<TerminusAvatar
											:info="userStore.getUserTerminusInfo(member.id || '')"
											:size="28"
										/>
									</div>
									<div>
										<div class="text-body1 text-weight-bold">
											{{ member.did }}
										</div>
										<div class="text-caption text-ink-1">
											{{ userStore.getCurrentDomain() }}
										</div>
									</div>
								</div>
								<!-- v-if="accountDid !== member.did" -->

								<div class="col-5 rowRight">
									<q-select
										class="select-input"
										popup-content-class="options_selected_Account"
										:model-value="member"
										dense
										borderless
										:options="authOptions"
										option-label="auth"
										dropdown-icon="sym_r_expand_more"
										@update:model-value="
											(value) => {
												member.readonly = value === 'Readonly' ? true : false;
												member.auth = value;
												updateMember(member, value);
												onEdit();
											}
										"
										style="width: 100px"
									>
										<template v-slot:option="{ itemProps, opt, toggleOption }">
											<q-item v-bind="itemProps">
												<q-item-section>
													<q-item-label>{{ opt }}</q-item-label>
												</q-item-section>
												<q-item-section side>
													<q-checkbox
														v-if="opt === member.auth"
														:model-value="true"
														checked-icon="sym_r_check_circle"
														unchecked-icon=""
														indeterminate-icon="help"
														@update:model-value="toggleOption"
														color="ink-2"
													/>
												</q-item-section>
											</q-item>
										</template>
									</q-select>

									<q-icon
										class="clear q-mx-xs text-ink-2"
										size="20px"
										name="sym_r_delete"
										style="justify-content: stretch"
										@click="removeMember(member)"
									/>
								</div>
							</div>
						</template>
					</div>
				</div>
			</q-scroll-area>
		</div>
		<div
			v-if="editing_t1"
			class="footer row iterm-center justify-between"
			:style="{
				'margin-bottom': isMobile ? '20px' : 0
			}"
		>
			<q-btn
				class="reset"
				:label="t('cancel')"
				type="reset"
				outline
				no-caps
				@click="onCancel"
				unelevated
				color="ink-2"
			/>
			<q-btn
				class="confirm text-grey-9"
				:label="t('save')"
				type="submit"
				@click="onSave"
				unelevated
				no-caps
				color="yellow-6"
				:loading="saveLoading"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed, ref, watch, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { OrgMember } from '@didvault/sdk/src/core';
import { OrgMemberStatus } from '@didvault/sdk/src/core';
import { app } from '../../../../globals';
import { useQuasar, Dialog } from 'quasar';
import { useMenuStore } from '../../../../stores/menu';
import { scrollBarStyle } from '../../../../utils/contact';
import { useUserStore } from '../../../../stores/user';
import {
	notifyFailed,
	notifyWarning,
	notifySuccess
} from '../../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import DeleteVault from './DeleteVault.vue';
const $q = useQuasar();
const route = useRoute();
const router = useRouter();
let editing_t1 = ref(false);
const meunStore = useMenuStore();
const nameRef = ref();
const userStore = useUserStore();
const org = ref();
const isMobile = ref(
	process.env.PLATFORM == 'MOBILE' ||
		process.env.PLATFORM == 'BEX' ||
		$q.platform.is.mobile
);

const initOrg = () => {
	org.value = app.orgs.find((org) => org.id == meunStore.org_id);
};

const _vault = computed(function () {
	if (!route.params.org_type) {
		return null;
	}
	if (!org.value) {
		return;
	}
	return org.value.vaults.find((e) => e.id == route.params.org_type);
	// return app.getVault(route.params.org_type as string);
});

const groups = ref<any>([]);
const members = ref<any>([]);
const name = ref();
const isEditExpir = ref(false);
const authOptions = ['Readonly', 'Editable'];
const saveLoading = ref(false);

const _getCurrentGroups = function () {
	if (!org.value) {
		return [];
	}

	const groups: { name: string; readonly: boolean }[] = [];

	for (const group of org.value.groups) {
		const vault = group.vaults.find((v) => v.id === route.params.org_type);
		if (vault) {
			groups.push({ name: group.name, readonly: vault.readonly });
		}
	}

	return groups;
};

const _getCurrentMembers = function () {
	if (!org.value) {
		return [];
	}

	const members: {
		did: string;
		name: string;
		readonly: boolean;
		auth: string;
	}[] = [];

	for (const member of org.value.members) {
		const vault = member.vaults.find((v) => v.id === route.params.org_type);

		if (vault) {
			members.push({
				did: member.did,
				name: member.name,
				readonly: vault.readonly,
				// role: member.role,
				auth: vault.readonly ? 'Readonly' : 'Editable'
			});
		}
	}

	return members;
};

const _availableMembers = computed(function () {
	if (!org.value || !org.value.members) {
		return [];
	}
	return (
		(org.value &&
			org.value.members.filter(
				(member) =>
					!members.value.some((m) => m.did === member.did) &&
					member.status === OrgMemberStatus.Active
			)) ||
		[]
	);
});

async function clearChanges(): Promise<void> {
	await initOrg();
	groups.value = await _getCurrentGroups();
	members.value = [];
	let membertSelf = await _getCurrentMembers();
	for (let i = 0; i < membertSelf.length; i++) {
		const element = membertSelf[i];
		const obj = {
			...element,
			auth: element.readonly ? 'Readonly' : 'Editable'
		};
		members.value.push(obj);
	}

	name.value = (_vault.value && _vault.value.name) || '';
}

watch(
	() => route.params.org_type,
	async (newVaule, oldVaule) => {
		if (oldVaule == newVaule) {
			return;
		}

		name.value = '';
		groups.value = [];
		members.value = [];

		if (route.params.org_type) {
			if (route.params.org_type == 'new') {
				editing_t1.value = true;

				let member = org.value?.getMember(app.account!);
				if (member) {
					addMember(member);
				}
				setTimeout(() => {
					nameRef.value.focus();
				}, 0);
			} else {
				// name.value = app.getVault(route.params.org_type)!.name;
				editing_t1.value = false;
				await clearChanges();
			}
		}
	}
);

const isBlank = computed(function () {
	if (app.state.locked) {
		return true;
	}
	if (!route.params.org_type) {
		return true;
	}
	return false;
});

const updateMember = (member, value) => {
	const memberself = JSON.parse(JSON.stringify(members.value));
	for (let i = 0; i < memberself.length; i++) {
		const element = memberself[i];
		if (member.did === element.did) {
			element.readonly = value === 'Readonly' ? true : false;
			element.auth = value;
		}
	}
	members.value = memberself;
};

function onEdit() {
	if (!editing_t1.value) {
		editing_t1.value = true;
	}

	let item2 = app.getItem(route.params.org_type as string)?.item.clone();
	isEditExpir.value = item2?.expiresAfter ? true : false;
}

function onCancel() {
	editing_t1.value = false;
	clearChanges();
	if (route.params.org_type === 'new') {
		router.push({
			path: '/org/Vaults/'
		});
	}
}

async function onDelete() {
	const confirmed = await new Promise((resolve) =>
		Dialog.create({
			title: t('delete_vault'),
			component: DeleteVault
		})
			.onOk((data) => {
				if (data.toLowerCase() === 'delete') {
					resolve(true);
				} else {
					notifyWarning(t('please_re_enter'));
				}
			})
			.onCancel(() => {
				resolve(false);
			})
	);

	if (confirmed) {
		await app.deleteVault(route.params.org_type as string);
		notifySuccess(t('delete_vault_success'));
		// route.params.org_type = '';
		router.push({
			path: '/org/Vaults/'
		});
	}
}

const getItems = () => {
	if (!org.value) {
		return [];
	}
	return org.value.vaults.filter((vault) => vault.org?.id == meunStore.org_id);
};

async function onSave() {
	if (!name.value) {
		notifyFailed(t('vault_name_is_null'));
		return;
	}

	const selfMember = JSON.parse(JSON.stringify(members.value));
	for (let i = 0; i < selfMember.length; i++) {
		const element = selfMember[i];
		element.id = route.params.org_type;
		delete element.auth;
		delete element.name;
	}

	saveLoading.value = true;
	if (route.params.org_type === 'new') {
		const hasItem = getItems().find((c) => c.name === name.value);

		if (hasItem) {
			notifyFailed(t('having_the_same_vault_name'));
			saveLoading.value = false;
			return false;
		}

		try {
			const vault = await app.createVault(
				name.value,
				org.value!,
				[...selfMember],
				[...groups.value]
			);
			notifySuccess(t('create_vault_success'));
			if (!isMobile.value) {
				router.push({
					path: '/org/Vaults/' + vault.id
				});
			} else {
				router.replace({
					path: '/org/Vaults/' + vault.id
				});
			}
		} catch (error) {
			notifyFailed(error.message);
			clearChanges();
		}
	} else {
		try {
			await app.updateVaultAccess(
				org.value!.id,
				route.params.org_type as string,
				name.value,
				[...selfMember],
				[...groups.value]
			);
			notifySuccess(t('update_vault_access_success'));
		} catch (error) {
			notifyFailed(error.message);
			clearChanges();
		}
	}
	editing_t1.value = false;
	saveLoading.value = false;
}

function addMember({ did, name, role }: OrgMember) {
	members.value.push({
		did,
		name,
		role,
		readonly: false,
		auth: 'Editable'
	});
	onEdit();
}

function removeMember(member: OrgMember) {
	members.value = members.value.filter((m: any) => m.did !== member.did);
	onEdit();
}

const goBack = () => {
	router.go(-1);
};

onMounted(async () => {
	name.value = '';
	groups.value = [];
	members.value = [];
	await clearChanges();

	if (route.params.org_type) {
		if (route.params.org_type == 'new') {
			editing_t1.value = true;

			let member = org.value?.getMember(app.account!);
			if (member) {
				addMember(member);
			}
			setTimeout(() => {
				nameRef.value.focus();
			}, 0);
		} else {
			// name.value = app.getVault(route.params.org_type)!.name;
			editing_t1.value = false;
		}
	}
});

const nameClick = () => {
	if (isMobile.value) {
		goBack();
	} else {
		onEdit();
	}
};

const { t } = useI18n();
</script>

<style>
.prompt-name {
	white-space: normal !important;
}
</style>

<style lang="scss" scoped>
.select-input {
	height: 32px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	overflow: hidden;

	::v-deep(.q-field__control) {
		height: 32px;
		min-height: 32px;
		background: $background-1 !important;
		color: $ink-2;
	}

	::v-deep(.q-field__marginal) {
		height: 32px;
		min-height: 32px;
	}

	::v-deep(.q-field__native) {
		color: $ink-2;
		padding-left: 8px;
		height: 32px;
		min-height: 32px;
	}
}

.orgItemView {
	width: 100%;
	height: 100%;
	display: flex;
	flex-direction: column;
	.view-hearder {
		border-radius: 10px;
		.hearder-input {
			height: 40px;
			width: 100%;
			.text {
				height: 40px;
				line-height: 40px;
			}
			.create-valut-input {
				border: 1px solid $input-stroke;
				padding-left: 10px;
				border-radius: 8px;
			}
		}
	}

	.view-width {
		width: calc(100% - 60px);
	}

	.view-option {
		width: 60px;
		cursor: pointer;
	}

	.container2 {
		flex: 1 1 auto;

		.tagSelect {
			width: 90%;
			.tagChip {
				margin-right: 5px;
			}
		}
	}
}

.footer {
	width: 100%;
	padding: 10px 20px;
	border-top: 1px solid $input-stroke;
	.confirm {
		width: 48%;
		height: 48px;
	}
	.reset {
		width: 48%;
		height: 48px;
	}
}

.listRow {
	border: 1px solid $input-stroke;
	border-radius: 8px;
	overflow: hidden;
	cursor: pointer;
	.header {
		background-color: $background-3;
		border-bottom: 1px solid $input-stroke;
	}
	.rowLeft {
		display: flex;
		align-items: center;
		justify-content: flex-start;
	}
	.rowRight {
		display: flex;
		align-items: center;
		justify-content: flex-end;
	}
}

.body1 {
	.listRow-content {
		&.borderBottom {
			border-bottom: 1px solid $separator;
		}
		.avator {
			width: 28px;
			height: 28px;
			border-radius: 14px;
			overflow: hidden;
		}
	}
}

.member-title {
	color: $ink-1;
}
.member-section {
	color: $ink-2;
}

.q-field__control {
	height: 34px !important;
}

.text-li-title {
	margin-left: 5px;
}
</style>
