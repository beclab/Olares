<template>
	<page-title-component
		:show-back="false"
		:title="t(`home_menus.${MENU_TYPE.VPN.toLowerCase()}`)"
	>
		<template v-slot:end v-if="deviceStore.isMobile">
			<div
				class="add-btn row justify-center items-center q-px-md"
				@click="onSubmit"
				v-if="commitEnable"
			>
				<q-icon size="20px" name="sym_r_check" color="ink-1" />
			</div>
		</template>
	</page-title-component>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<app-menu-feature :menu-type="MENU_TYPE.VPN" />
		<bt-list>
			<bt-form-item
				v-if="adminStore.isAdmin"
				:title="t('allow_ssh_via_vpn')"
				:description="
					t('Enable remote SSH connections to the cluster through a VPN')
				"
			>
				<bt-switch
					truthy-track-color="blue-default"
					:model-value="aclStore.allow_ssh"
					:disable="aclStore.state === ActionType.APPLYING"
					@update:model-value="setAclToggle"
				/>
			</bt-form-item>
			<bt-form-item
				v-if="adminStore.isAdmin"
				:title="t('Subnet Routes')"
				:description="
					t(
						'Allow VPN users to access other devices on the local network (e.g., printers, file servers)'
					)
				"
			>
				<bt-switch
					truthy-track-color="blue-default"
					:model-value="aclStore.allow_subroutes"
					@update:model-value="setSubroutesToggle"
				/>
			</bt-form-item>

			<bt-form-item
				v-if="adminStore.isAdmin"
				:title="t('forcing_vpn_access_to_private_network')"
				:description="
					t(
						'When enabled, Olares only accepts connections made through the LarePass VPN. All other access methods will be blocked.'
					)
				"
			>
				<bt-switch
					truthy-track-color="blue-default"
					:model-value="headScaleStore.headScaleStatus"
					@update:model-value="setHeadScaleToggle"
				/>
			</bt-form-item>
			<bt-form-item
				:title="t('view_the_headScale_connection_status')"
				:description="
					t('Check the real-time status of all connected VPN clients.')
				"
				@click="gotoPage('/vpn/active_headscale')"
				:chevron-right="true"
				:width-separator="false"
			/>
		</bt-list>

		<AdaptiveLayout v-if="adminStore.isAdmin">
			<template v-slot:pc>
				<q-list class="q-py-md q-list-class q-mt-lg">
					<div
						class="row items-center justify-between item-margin-left item-margin-right"
					>
						<div class="text-body1 text-ink-1 row justify-start items-center">
							<span>{{ t('ACL DST port') }}</span>
							<settings-tooltip
								:description="
									t(
										'Define fine-grained access policies to control which users or apps can connect to specific network ports via VPN.'
									)
								"
							/>
						</div>
						<div
							class="add-btn row justify-center items-center"
							@click="addACL"
						>
							<q-icon size="20px" name="sym_r_add" color="ink-1" />
						</div>
					</div>
					<div
						v-if="aclStore.displayPort && aclStore.displayPort.length > 0"
						class="column item-margin-left item-margin-right q-mt-md"
					>
						<q-table
							tableHeaderStyle="height: 32px;"
							table-header-class="text-body3 text-ink-3"
							flat
							:bordered="false"
							:rows="aclStore.displayPort"
							:columns="columns"
							row-key="id"
							hide-pagination
							hide-selected-banner
							hide-bottom
							:rowsPerPageOptions="[0]"
						>
							<template v-slot:body-cell-actions="props">
								<q-td :props="props" class="text-ink-2">
									<q-icon
										v-if="canRemove(props.row)"
										name="sym_r_delete"
										size="16px"
										color="ink-2"
										@click.stop="removeACL(props.row)"
									/>
								</q-td>
							</template>
							<template v-slot:body-cell-appOwner="props">
								<q-td :props="props">
									<div class="row items-center">
										<setting-avatar :size="24" style="margin-right: 4px" />
										<div>
											{{ props.row.appOwner }}
										</div>
									</div>
								</q-td>
							</template>
							<template v-slot:body-cell-appTitle="props">
								<q-td :props="props">
									<div class="row items-center">
										<q-img
											no-spinner
											width="24px"
											height="24px"
											:src="getAppIcon(props.row)"
											style="border-radius: 6px; margin-right: 4px"
										/>
										<div>
											{{ props.row.appTitle }}
										</div>
									</div>
								</q-td>
							</template>
						</q-table>
					</div>
					<empty-component
						class="q-pb-xl"
						v-else
						:info="t('no_acl_dst_port_added')"
						:empty-image-top="40"
					/>
				</q-list>
			</template>
			<template v-slot:mobile>
				<module-title class="q-mt-xl">
					<div class="row items-center justify-between">
						<div class="row justify-start items-center">
							<span>{{ t('ACL DST port') }}</span>
							<settings-tooltip
								:description="
									t(
										'Define fine-grained access policies to control which users or apps can connect to specific network ports via VPN.'
									)
								"
							/>
						</div>
						<div
							class="add-btn row justify-center items-center"
							@click="addACL"
						>
							<q-icon size="20px" name="sym_r_add" color="ink-1" />
						</div>
					</div>
				</module-title>
				<div v-if="aclStore.displayPort.length > 0">
					<bt-grid
						class="mobile-items-list"
						:repeat-count="2"
						v-for="(port, index) in aclStore.displayPort"
						:key="index"
						:paddingY="12"
					>
						<template v-slot:title>
							<div
								class="text-subtitle3-m row justify-between items-center clickable-view q-mb-md"
							>
								<div>
									{{ port.port }}
								</div>
								<q-icon
									v-if="canRemove(port)"
									name="sym_r_delete"
									color="ink-2"
									size="20px"
									@click.stop="removeACL(port)"
								/>
							</div>
						</template>
						<template v-slot:grid>
							<bt-grid-item
								:label="t('Application')"
								mobileTitleClasses="text-body3-m"
								:value="port.appTitle"
							/>
							<bt-grid-item
								:label="t('User')"
								mobileTitleClasses="text-body3-m"
								:value="port.appOwner"
							/>
						</template>
					</bt-grid>
				</div>
				<empty-component
					class="q-pb-xl"
					v-else
					:info="t('no_acl_dst_port_added')"
					:empty-image-top="40"
				/>
			</template>
		</AdaptiveLayout>

		<div
			class="row justify-end"
			v-if="!deviceStore.isMobile && adminStore.isAdmin"
		>
			<q-btn
				dense
				:disable="!commitEnable"
				flat
				class="confirm-btn q-px-md q-my-lg"
				:label="t('apply')"
				@click="onSubmit"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import SettingAvatar from 'src/components/settings/base/SettingAvatar.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import AppMenuFeature from 'src/components/settings/AppMenuFeature.vue';
import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtGridItem from 'src/components/settings/base/BtGridItem.vue';
import ModuleTitle from 'src/components/settings/ModuleTitle.vue';
import BtGrid from 'src/components/settings/base/BtGrid.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import EditAppAclPortDialog from './EditAppAclPortDialog.vue';
import { useApplicationStore } from 'src/stores/settings/application';
import { useHeadScaleStore } from 'src/stores/settings/headscale';
import { useDeviceStore } from 'src/stores/settings/device';
import { getRequireImage } from 'src/utils/settings/helper';
import { useAdminStore } from 'src/stores/settings/admin';
import { computed, onMounted } from 'vue';
import { MENU_TYPE } from 'src/constant';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import {
	useAclStore,
	ActionType,
	PortAclInfo,
	PortStatus
} from 'src/stores/settings/acl';
import SettingsTooltip from 'src/components/settings/base/SettingsTooltip.vue';
import { notifyFailed } from 'src/utils/settings/btNotify';

const applicationStore = useApplicationStore();
const headScaleStore = useHeadScaleStore();
const deviceStore = useDeviceStore();
const adminStore = useAdminStore();
const aclStore = useAclStore();
const router = useRouter();
const { t } = useI18n();
const $q = useQuasar();

const setAclToggle = async () => {
	await aclStore.toggleAclStatus(!aclStore.allow_ssh);
	await aclStore.getAclStatus();
	await aclStore.getAllApplicationAcls();
};

const setSubroutesToggle = async () => {
	$q.loading.show();
	try {
		await aclStore.toggleSubroutesStatus(!aclStore.allow_subroutes);
	} catch (error) {
		console.log('error ===>', error);
	} finally {
		$q.loading.hide();
	}
};

onMounted(async () => {
	await aclStore.getSubroutesStatus();
	await aclStore.getAclStatus();
	await aclStore.getAllApplicationAcls();
	await headScaleStore.getDevices();
	await headScaleStore.getHeadScaleStatus();
});

const addACL = () => {
	$q.dialog({
		component: EditAppAclPortDialog
	}).onOk(async (port: string) => {
		const index = aclStore.allPortAppAclList.findIndex(
			(e) =>
				e.appName == 'olares-app' && e.proto == '' && e.port === `*:${port}`
		);

		const addItem = {
			appName: 'olares-app',
			port: `*:${port}`,
			appOwner: adminStore.olaresId.split('@')[0],
			proto: '',
			status: PortStatus.Add,
			appTitle: 'olares-app'
		};

		if (index >= 0) {
			aclStore.allPortAppAclList.splice(index, 1, addItem);
		} else {
			aclStore.allPortAppAclList.push(addItem);
		}
	});
};

const onSubmit = () => {
	aclStore.appAclSubmit('olares-app');
};

const commitEnable = computed(() => {
	return (
		aclStore.allPortAppAclList.find((port) => {
			if (port.status != PortStatus.Normal) {
				const acl = aclStore.allAppAclList.find(
					(acl) =>
						port.appName == acl.appName &&
						port.proto == acl.proto &&
						acl.dst.find((dst) => port.port == dst)
				);

				if (!acl) {
					return port.status == PortStatus.Add;
				}
				return port.status == PortStatus.Remove;
			}
			return false;
		}) != undefined
	);
});

const removeACL = async (info: PortAclInfo) => {
	//
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('delete_item', {
				item: info.port.startsWith('*:')
					? info.port.substring(2, info.port.length)
					: info.port
			}),
			message: t('are_you_sure_you_want_to_delete_item', {
				item: info.port.startsWith('*:')
					? info.port.substring(2, info.port.length)
					: info.port
			}),
			useCancel: true
		}
	}).onOk(async () => {
		info.status = PortStatus.Remove;
	});
};

const setHeadScaleToggle = async () => {
	if (!headScaleStore.headScaleStatus) {
		let total = 0;
		for (const device of headScaleStore.devices) {
			if (device.online) {
				total += 1;
			}
		}
		if (total < 3) {
			notifyFailed(
				t('you_need_at_least_2_devices_online_to_activate_headScale')
			);
			return;
		}
	}
	await headScaleStore.toggleHeadScaleStatus();
	await headScaleStore.getHeadScaleStatus();
};

const canRemove = (info: PortAclInfo) => {
	if (info.appName != 'olares-app') {
		return false;
	}
	if (info.port === '*:22') {
		return false;
	}
	return true;
};

const getAppIcon = (info: PortAclInfo) => {
	if (info.port === '*:22') {
		return getRequireImage('system/ssh.svg');
	}
	const app = applicationStore.getApplicationById(info.appName);
	return app ? app.icon : getRequireImage('system/unknown.svg');
};

const columns: any = [
	{
		name: 'port',
		align: 'left',
		label: t('port'),
		field: 'port',
		format: (val: any) => {
			return val.startsWith('*:') ? val.substring(2, val.length) : val;
		},
		sortable: false
	},
	{
		name: 'appTitle',
		align: 'left',
		label: t('Application'),
		field: 'appTitle',
		sortable: false
	},
	{
		name: 'appOwner',
		align: 'left',
		label: t('User'),
		field: 'appOwner',
		sortable: false
	},
	{
		name: 'actions',
		align: 'right',
		label: t('action'),
		sortable: false
	}
];

const gotoPage = (path) => {
	router.push({ path });
};
</script>

<style scoped lang="scss">
.add-btn {
	border-radius: 4px;
	// height: 24px;
	width: 24px;
	background-color: $background-1;
	cursor: pointer;
	text-decoration: none;

	.add-title {
		color: $ink-2;
	}
}

.add-btn:hover {
	background-color: $btn-bg-hover;
}
</style>
