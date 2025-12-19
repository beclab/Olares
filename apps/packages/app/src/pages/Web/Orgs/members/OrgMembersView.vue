<template>
	<div v-if="member && org" class="orgMemberView bg-background-1">
		<div class="header">
			<div class="row justify-between q-pa-md">
				<div :class="['row', 'items-center', 'view-hearder']">
					<q-icon
						v-if="isMobile"
						name="sym_r_chevron_left"
						color="ink-1"
						size="24px"
						@click="goBack"
					/>
					<div class="hearder-input column justify-center">
						<div class="text-subtitle2 text-ink-1 text-weight-bold">
							{{ member.did }}
						</div>
						<div class="text-overline text-ink-3">
							@{{ userStore.getCurrentDomain() }}
						</div>
					</div>
				</div>

				<div class="row view-option items-center justify-center">
					<div
						class="role text-body3 row items-center justify-center"
						v-if="isOwner || isAdmin || isSuspended"
						:class="{
							owner: isOwner,
							admin: !isOwner && isAdmin,
							suspended: isSuspended
						}"
					>
						<q-icon name="sym_r_manage_accounts" />
						<span class="owner" v-if="isOwner">{{ t('owner') }}</span>
						<span class="admin" v-else-if="isAdmin">{{ t('admin') }}</span>
						<span class="suspended" v-else-if="isSuspended">
							{{ t('suspended') }}
						</span>
					</div>

					<!-- <q-item-section class="q-mx-sm" v-if="!isOwner && accountIsOwner">
						<q-icon name="sym_r_more_horiz" size="28px">
							<q-menu class="popup-menu">
								<q-list dense padding style="width: 160px"> -->
					<!-- <q-item
										class="row items-center justify-start popup-item"
										clickable
										dense
										v-close-popup
										@click="onRemoveMember"
									>
										<q-icon class="q-mr-xs" name="sym_r_delete" size="20px" />
										{{ t('remove') }}
									</q-item> -->
					<!-- <q-item
										class="row items-center justify-start popup-item"
										clickable
										dense
										v-close-popup
										@click="onSuspendedMember"
										v-if="!isSuspended"
									>
										<q-icon class="q-mr-xs" name="sym_r_block" size="20px" />
										{{ t('suspend') }}
									</q-item> -->
					<!-- <q-item
										class="row items-center justify-start popup-item"
										clickable
										dense
										v-close-popup
										@click="_unsuspendMember"
										v-if="isSuspended"
									>
										<q-icon class="q-mr-xs" name="sym_r_block" size="20px" />
										<q-icon size="22px" name="block" class="q-mr-xs" />
										{{ t('unsuspend') }}
									</q-item> -->
					<!-- <q-item
										class="row items-center justify-start popup-item"
										clickable
										dense
										v-close-popup
										@click="onMakeAdmin"
										v-if="!isAdmin"
									>
										<q-icon
											class="q-mr-xs"
											name="sym_r_manage_accounts"
											size="20px"
										/>
										{{ t('make_admin') }}
									</q-item> -->
					<!-- <q-item
										class="row items-center justify-start popup-item"
										clickable
										dense
										v-close-popup
										@click="onRemoveAdmin"
										v-if="isAdmin"
									>
										<q-icon
											class="q-mr-xs"
											name="sym_r_person_cancel"
											size="20px"
										/>
										{{ t('remove_admin') }}
									</q-item> -->
					<!-- <q-item
										class="row items-center justify-start popup-item"
										dense
										clickable
										v-close-popup
										@click="onMakeOwner"
										v-if="accountIsOwner && !isOwner"
									>
										<q-icon
											class="q-mr-xs"
											name="sym_r_person_raised_hand"
											size="20px"
										/>
										{{ t('make_owner') }}
									</q-item> -->
					<!-- </q-list>
							</q-menu>
						</q-icon>
					</q-item-section> -->
				</div>
			</div>
		</div>
		<div class="container2">
			<q-scroll-area
				style="height: 100%"
				:thumb-style="scrollBarStyle.thumbStyle"
			>
				<div>
					<div class="listRow column justify-center">
						<div class="header q-pa-md row justify-between q-mb-sm">
							<div class="items-center">
								<span class="text-ink-2 text-body3 text-li-title">
									{{ t('Vault') }}
								</span>
							</div>
							<div>
								<q-icon name="sym_r_add" size="20px" color="ink-1" />
								<q-menu
									class="popup-menu bg-background-2"
									flat
									v-if="_availableVaults.length > 0"
								>
									<q-list class="q-py-sm" dense>
										<template
											v-for="(vault, index) in _availableVaults"
											:key="'avn' + index"
										>
											<q-item
												class="row items-center justify-start popup-item"
												clickable
												v-close-popup
												@click="_addVault(vault)"
												style="width: 140px; white-space: nowrap"
											>
												{{ vault.name }}
											</q-item>
											<q-separator v-if="index < _availableVaults.length - 1" />
										</template>
									</q-list>
								</q-menu>
								<q-menu class="popup-menu bg-background-2" flat v-else>
									<q-list class="q-py-sm" dense>
										<q-item
											class="row items-center justify-center text-ink-1 text-body3"
											clickable
											v-close-popup
											style="white-space: nowrap"
										>
											{{ t('no_more_vaults_available') }}
										</q-item>
									</q-list>
								</q-menu>
							</div>
						</div>

						<div class="body q-pa-md">
							<div v-if="_vaults.length == 0 && _indirectVaults.length == 0">
								{{ t('this_member_does_not_have_access_to_any_vaults_yet') }}
							</div>
							<div v-else>
								<div v-for="(v, index) in _vaults" :key="'va' + index">
									<div
										class="vault-list bg-background-1 q-pa-md row items-center justify-between"
									>
										<div class="col-7 rowLeft">
											<div class="avator q-mr-md">
												<q-icon name="sym_r_deployed_code" size="20px" />
											</div>
											<div>
												{{
													org.vaults.find((vault) => vault.id === v.id)?.name
												}}
											</div>
										</div>

										<div class="col-5 rowRight">
											<q-select
												v-if="accountDid !== member.did"
												class="select-input"
												popup-content-class="options_selected_Account"
												dropdown-icon="sym_r_expand_more"
												:model-value="v"
												dense
												borderless
												:options="authOptions"
												option-label="auth"
												@update:model-value="
													(value) => {
														v.readonly = value === 'Readonly' ? true : false;
														v.auth = value;
														updateMember(v, value);
														// onEdit();
													}
												"
												style="width: 100px"
											>
												<template
													v-slot:option="{ itemProps, opt, toggleOption }"
												>
													<q-item v-bind="itemProps">
														<q-item-section>
															<q-item-label>{{ opt }}</q-item-label>
														</q-item-section>
														<q-item-section side>
															<q-checkbox
																v-if="opt === v.auth"
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
												@click="
													_removeVault(
														org.vaults.find((vault) => vault.id === v.id)
													)
												"
											/>
										</div>
									</div>
									<q-separator v-if="index < _vaults.length - 1" />
								</div>

								<div v-for="(v, index) in _indirectVaults" :key="'iva' + index">
									<div class="listRow q-pa-md row items-center justify-between">
										{{ org.vaults.find((vault) => vault.id === v.id)?.name }}
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</q-scroll-area>
		</div>
		<div
			v-if="accountIsAdmin && hasChanges"
			class="footer row iterm-center justify-between"
		>
			<q-btn
				class="reset"
				:label="t('buttons.cancel')"
				type="reset"
				outline
				no-caps
				@click="clearChanges"
				unelevated
				color="grey-7"
			/>
			<q-btn
				class="confirm text-grey-9"
				:label="t('buttons.save')"
				type="submit"
				@click="onSave"
				unelevated
				no-caps
				color="yellow-6"
			/>
		</div>
	</div>
	<div
		class="orgMemberView bg-background-1 row items-center justify-center"
		v-else
	>
		<BtIcon class="q-mb-lg" src="itemSelect" :width="215" :height="148" />
		{{ t('no_member_selected') }}
	</div>
</template>

<script lang="ts" setup>
import { computed, ref, onMounted, watch, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Group, OrgRole, OrgMemberStatus } from '@didvault/sdk/src/core';
import { app } from '../../../../globals';
import { useQuasar } from 'quasar';
import { useMenuStore } from '../../../../stores/menu';
import { useUserStore } from '../../../../stores/user';
import { scrollBarStyle } from '../../../../utils/contact';
import {
	notifyFailed,
	notifySuccess
} from '../../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import { BtDialog } from '@bytetrade/ui';
import { busOn, busOff } from '../../../../utils/bus';
const $q = useQuasar();

const route = useRoute();
const router = useRouter();
const meunStore = useMenuStore();
const userStore = useUserStore();

let _vaults = ref<{ id: string; readonly: boolean; auth: string }[]>([]);
let _groups = ref<string[]>([]);
const authOptions = ['Readonly', 'Editable'];
const accountDid = app.account?.did;
const org = ref();
const member = ref();
const isMobile = ref(
	process.env.PLATFORM == 'MOBILE' ||
		process.env.PLATFORM == 'BEX' ||
		$q.platform.is.mobile
);

const initOrg = () => {
	org.value = app.orgs.find((org) => org.id == meunStore.org_id);
};

const initMember = () => {
	if (org.value && route.params.org_type) {
		member.value = org.value.getMember({
			did: route.params.org_type
		});
	} else {
		member.value = null;
	}
};

const accountIsOwner = computed(function () {
	if (!app.account) {
		return false;
	}
	return org.value!.isOwner(app.account);
});

const accountIsAdmin = computed(function () {
	if (!app.account) {
		return false;
	}
	return org.value!.isAdmin(app.account);
});

const isAdmin = computed(function () {
	if (!member.value) {
		return false;
	}
	return org.value!.isAdmin(member.value);
});

const isOwner = computed(function () {
	if (!member.value) {
		return false;
	}
	return org.value!.isOwner(member.value);
});

const isSuspended = computed(function () {
	if (!member.value) {
		return false;
	}
	return org.value!.isSuspended(member.value);
});

const init = async () => {
	await initOrg();
	await initMember();
	await clearChanges();
};

watch(
	() => route.params.org_type,
	async (newVaule: any, oldVaule: any) => {
		if (oldVaule == newVaule) {
			return;
		}
		init();
	}
);

onMounted(async () => {
	busOn('orgSubscribe', init);

	init();
});

onUnmounted(() => {
	busOff('orgSubscribe', init);
});

const _getCurrentGroups = function () {
	if (!org.value || !member.value) {
		return [];
	}

	return org.value.getGroupsForMember(member.value).map((g) => g.name) || [];
};

const _getCurrentVaults = function () {
	if (!org.value || !member.value) {
		return [];
	}

	return member.value.vaults.map((v) => ({ ...v })) || [];
};

const hasChanges = computed(function () {
	if (!org.value || !member.value) {
		return false;
	}

	const currentVaults = _getCurrentVaults();
	const hasVaultsChanged =
		_vaults.value.length !== currentVaults.length ||
		_vaults.value.some((vault) => {
			const other = currentVaults.find((v) => v.id === vault.id);
			return !other || other.readonly !== vault.readonly;
		});

	const currentGroups = _getCurrentGroups();
	const hasGroupsChanged =
		_groups.value.length !== currentGroups.length ||
		_groups.value.some((g) => !currentGroups.includes(g));

	return hasVaultsChanged || hasGroupsChanged ? true : false;
});

async function clearChanges(): Promise<void> {
	org.value = app.orgs.find((org) => org.id == meunStore.org_id)!;
	_vaults.value = [];
	_groups.value = _getCurrentGroups();
	let vaultSelf = _getCurrentVaults();
	for (let i = 0; i < vaultSelf.length; i++) {
		const element = vaultSelf[i];
		const obj = {
			...element,
			auth: element.readonly ? 'Readonly' : 'Editable'
		};
		_vaults.value.push(obj);
	}
}

async function _addGroup(group: Group) {
	_groups.value.push(group.name);
}

async function _addVault(vault: { id: string; name: string }) {
	_vaults.value.push({
		id: vault.id,
		readonly: false,
		auth: 'Editable'
	});
}

async function _removeVault(vault: { id: string } | undefined) {
	if (!vault) return;
	_vaults.value = _vaults.value.filter((v) => v.id !== vault.id);
}

const _indirectVaults = computed(function () {
	if (!org.value || !member.value) {
		return [];
	}

	let vaults: { id: string; readonly: boolean; groups: string[] }[] = [];

	for (const groupName of _groups.value) {
		const group = org.value.groups.find((g) => g.name === groupName)!;
		for (const vault of group.vaults) {
			if (_vaults.value.some((v) => v.id === vault.id)) {
				continue;
			}

			const existing = vaults.find((v) => v.id === vault.id);
			if (existing) {
				existing.groups.push(group.name);
				existing.readonly = existing.readonly && vault.readonly;
			} else {
				vaults.push({
					id: vault.id,
					readonly: vault.readonly,
					groups: [group.name]
				});
			}
		}
	}

	return vaults;
});

const _availableVaults = computed(function () {
	if (!org.value || !member.value) {
		return [];
	}

	return (
		org.value.vaults.filter(
			(vault) => !_vaults.value.some((v) => v.id === vault.id)
		) || []
	);
});

const updateMember = (member, value) => {
	const vaultself = JSON.parse(JSON.stringify(_vaults.value));
	for (let i = 0; i < vaultself.length; i++) {
		const element = vaultself[i];
		if (member.id === element.id) {
			element.readonly = value === 'Readonly' ? true : false;
			element.auth = value;
		}
	}
	_vaults.value = vaultself;
};

async function onSave() {
	if (!org.value || !member.value) {
		return;
	}
	try {
		await app.updateMember(org.value, member.value, {
			vaults: [..._vaults.value],
			groups: [..._groups.value]
		});
		clearChanges();
	} catch (e) {
		notifyFailed(
			e.message || t('something_went_wrong_request_try_again_later')
		);
	}
}

async function onMakeAdmin() {
	if (!org.value || !member.value) {
		return;
	}

	BtDialog.show({
		platform: 'web',
		cancel: true,
		okStyle: {
			background: '#FFEB3B',
			color: '#1F1814'
		},
		okText: t('base.confirm'),
		cancelText: t('base.cancel'),
		title: t('make_admin'),
		message: t('make_admin_message')
	})
		.then(async (res) => {
			if (res) {
				try {
					await app.updateMember(org.value, member.value, {
						role: OrgRole.Admin
					});
					meunStore.org_mode_id = '';
				} catch (e) {
					notifyFailed(
						e.message || t('something_went_wrong_please_try_again_later')
					);
				}
			}
		})
		.catch((err) => {
			console.error(err);
		});
}

async function onRemoveAdmin() {
	if (!org.value || !member.value) {
		return;
	}

	BtDialog.show({
		platform: 'web',
		cancel: true,
		okStyle: {
			background: '#FFEB3B',
			color: '#1F1814'
		},
		okText: t('base.confirm'),
		cancelText: t('base.cancel'),
		title: t('remove_admin'),
		message: t('remove_admin_message')
	})
		.then(async (res) => {
			if (res) {
				try {
					await app.updateMember(org.value, member.value, {
						role: OrgRole.Member
					});
					meunStore.org_mode_id = '';
				} catch (e) {
					notifyFailed(
						e.message || t('something_went_wrong_please_try_again_later')
					);
				}
			}
		})
		.catch((err) => {
			console.error(err);
		});
}

const goBack = () => {
	router.go(-1);
};

const { t } = useI18n();
</script>

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
.orgMemberView {
	height: 100%;
	display: flex;
	flex-direction: column;

	.view-hearder {
		border-radius: 10px;

		.hearder-input {
			height: 44px;
		}

		&:focus-within {
			border: 1px solid $blue;
		}
	}

	.isedit {
		border: 1px solid $separator;
	}

	.view-option {
		.role {
			height: 24px;
			padding: 0 4px;
			border-radius: 4px;
			margin-left: 10px;

			&.owner {
				border: 1px solid $deep-orange-5;
				color: $deep-orange-5;
			}

			&.admin {
				border: 1px solid $blue-4;
				color: $blue-4;
			}

			&.provisioned {
				border: 1px solid $grey-8;
				color: $grey-8;
			}

			&.suspended {
				border: 1px solid $grey-8;
				color: $grey-8;
			}
		}

		.optionItem {
			text-align: center;
			cursor: pointer;

			span.optionIcon {
				padding: 6px;
				display: inline-block;

				&:hover {
					border-radius: 4px;
				}
			}
		}

		.startItem {
			position: absolute;
			left: -50px;
			top: 8px;
		}
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
	width: 90%;
	border: 1px solid $separator;
	border-radius: 8px;
	margin: 20px auto;
	overflow: hidden;

	.header {
		background-color: $background-3;
		border-bottom: 1px solid $separator;
	}

	.body {
		.vault-list {
			width: 100%;
			cursor: pointer;

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
	}
}

.attach {
	overflow: hidden;

	.reduce {
		width: 38px;
		margin-left: -38px;
		cursor: pointer;
	}

	.attach {
		flex: 1;
	}

	&:hover {
		.reduce {
			animation: aniUp 0.5s 1 ease-in-out forwards;
		}
	}

	@keyframes aniUp {
		from {
			margin-left: -38px;
		}

		to {
			margin-left: 0px;
		}
	}
}

.q-field__control {
	height: 34px !important;
}

.text-li-title {
	margin-left: 5px;
}
</style>
