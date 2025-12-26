<template>
	<page-title-component
		:show-back="true"
		:title="application?.title || application?.name"
	/>

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list
			first
			v-if="application?.owner && application.owner == adminStore.user.name"
		>
			<application-operate-item :app="application" />
		</bt-list>

		<bt-list v-if="secretPermission && !deviceStore.isMobile">
			<bt-form-item
				:title="t('secrets')"
				@click="gotoSecret"
				:margin-top="false"
				:chevron-right="true"
				:width-separator="false"
			/>
		</bt-list>

		<div v-if="application?.entrances && application.entrances.length">
			<module-title
				class="q-mb-sm"
				:class="{
					'q-mt-lg': !deviceStore.isMobile,
					'q-mt-xl': deviceStore.isMobile
				}"
				>{{ t('entrances') }}
			</module-title>

			<bt-list first>
				<template
					v-for="(entrance, index) in application.entrances"
					:key="index"
				>
					<application-item
						:icon="
							entrance.icon
								? entrance.icon
								: application?.icon
								? application?.icon
								: ''
						"
						:title="entrance.title"
						:status="entrance.state"
						:width-separator="index !== application.entrances.length - 1"
						:margin-top="index !== 0"
						@click="gotoEntrance(entrance)"
						:hide-status="false"
					/>
				</template>
			</bt-list>
		</div>

		<div
			v-if="application?.sharedEntrances && application.sharedEntrances.length"
		>
			<module-title
				class="q-mb-sm"
				:class="{
					'q-mt-lg': !deviceStore.isMobile,
					'q-mt-xl': deviceStore.isMobile
				}"
				>{{ t('Shared entrances') }}
			</module-title>

			<bt-list first>
				<template
					v-for="(entrance, index) in application.sharedEntrances"
					:key="index"
				>
					<application-item
						:icon="
							entrance.icon
								? entrance?.icon
								: application.icon
								? application.icon
								: ''
						"
						:title="entrance.title || entrance.name"
						:hideStatus="true"
						:width-separator="index !== application.sharedEntrances.length - 1"
						:margin-top="index !== 0"
						@click="gotoEntrance(entrance, true)"
					/>
				</template>
			</bt-list>
		</div>

		<div v-if="appRegisterProviders && appRegisterProviders.length">
			<module-title
				class="q-mb-sm"
				:class="{
					'q-mt-lg': !deviceStore.isMobile,
					'q-mt-xl': deviceStore.isMobile
				}"
				>{{ t('providers') }}
			</module-title>
			<bt-list first>
				<template
					v-for="(provider, index) in appRegisterProviders"
					:key="index"
				>
					<bt-form-item
						:title="`${provider.dataType}/${provider.group}/${provider.version}`"
						:margin-top="false"
						:width-separator="index + 1 < appRegisterProviders.length"
						:chevron-right="true"
						@click="gotoPermission(provider)"
					/>
				</template>
			</bt-list>
		</div>

		<module-title
			v-if="application?.owner && application.owner == adminStore.user.name"
			class="q-mb-sm"
			:class="{
				'q-mt-lg': !deviceStore.isMobile,
				'q-mt-xl': deviceStore.isMobile
			}"
			>{{ t('Environment Variables') }}
		</module-title>

		<bt-list
			first
			v-if="application?.owner && application.owner == adminStore.user.name"
		>
			<bt-form-item
				:title="t('Manage Environment Variables')"
				:margin-top="false"
				:width-separator="false"
				:chevron-right="true"
				@click="
					gotoManagerEnvironment({
						appName: Route.params.name
					})
				"
			/>
		</bt-list>

		<div
			v-if="
				(appPermissions &&
					appPermissions.permissions &&
					appPermissions.permissions.length > 0) ||
				(application && application.ports && application.ports.length > 0) ||
				(aclStore.appAclList && aclStore.appAclList.length > 0)
			"
		>
			<module-title
				class="q-mb-sm"
				:class="{
					'q-mt-lg': !deviceStore.isMobile,
					'q-mt-xl': deviceStore.isMobile
				}"
				>{{ t('permissions') }}
			</module-title>
			<bt-list first>
				<div v-if="appPermissions && appPermissions.permissions">
					<template
						v-for="(permission, index) in appPermissions.permissions"
						:key="index"
					>
						<bt-form-item
							:title="`${permission.dataType}/${permission.group}/${permission.version}`"
							@click="gotoPermission(permission)"
							:margin-top="false"
							:width-separator="
								index + 1 < appPermissions.permissions.length ||
								!!(
									application &&
									application.ports &&
									application.ports.length > 0
								) ||
								(aclStore.appAclList && aclStore.appAclList.length > 0)
							"
							:chevron-right="true"
						/>
					</template>
				</div>
				<bt-form-item
					v-if="aclStore.appAclList && aclStore.appAclList.length > 0"
					:title="t('acls')"
					@click="gotoAclPage"
					:margin-top="false"
					:width-separator="
						!!(application && application.ports && application.ports.length > 0)
					"
					:chevron-right="true"
				/>

				<bt-form-item
					v-if="application && application.ports.length > 0"
					:title="t('export_ports')"
					@click="gotoPorts"
					:margin-top="false"
					:width-separator="false"
					:chevron-right="true"
				/>
			</bt-list>
		</div>

		<div class="full-width q-mb-lg" />
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useApplicationStore } from 'src/stores/settings/application';
import { useSecretStore } from 'src/stores/settings/secret';
import BtList from 'src/components/settings/base/BtList.vue';
import ModuleTitle from 'src/components/settings/ModuleTitle.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ApplicationItem from 'src/components/settings/application/ApplicationItem.vue';
import ApplicationOperateItem from 'src/components/settings/application/ApplicationOperateItem.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { useAclStore } from 'src/stores/settings/acl';
import { TerminusEntrance } from '@bytetrade/core';
import { useAdminStore } from 'src/stores/settings/admin';
import { useI18n } from 'vue-i18n';
import { bus } from 'src/utils/bus';
import {
	AppPermission,
	Permission,
	PermissionProviderRegister
} from 'src/constant/global';

const applicationStore = useApplicationStore();
const secretStore = useSecretStore();
const deviceStore = useDeviceStore();
const { t } = useI18n();
const aclStore = useAclStore();
const Route = useRoute();
const router = useRouter();
const adminStore = useAdminStore();
const application = ref(
	applicationStore.getApplicationById(Route.params.name as string)
);
const secretPermission = ref(false);

const gotoSecret = () => {
	router.push('/application/secret/' + application.value?.name);
};

const gotoEntrance = (entrance: TerminusEntrance, shared = false) => {
	if (shared) {
		router.push(
			'/application/domain/' +
				application.value?.name +
				'/' +
				entrance.name +
				'/shared'
		);
		return;
	}
	router.push(
		'/application/entrance/' +
			application.value?.name +
			'/' +
			entrance.name +
			(shared ? '/shared' : '')
	);
};

const gotoPorts = () => {
	router.push('/application/ports/' + application.value?.name);
};

const gotoPermission = (
	permission: Permission | PermissionProviderRegister
) => {
	gotoPermissionDetail({
		dataType: permission.dataType,
		group: permission.group,
		version: permission.version,
		title: `${permission.dataType}/${permission.group}/${permission.version}`
	});
};

const gotoAclPage = () => {
	if (application.value) {
		router.push({
			name: 'appAcl',
			params: {
				name: application.value.name
			}
		});
	}
};

const gotoPermissionDetail = (query: any) => {
	router.push({
		path: '/application/permission/detail',
		query
	});
};

const gotoManagerEnvironment = (query: any) => {
	router.push({
		path: '/application/environment/manager',
		query
	});
};

async function checkSecretPermission() {
	const res = await secretStore.checkSecretPermission(application.value?.name);
	if (res && res.permission === true) {
		secretPermission.value = true;
	} else {
		secretPermission.value = false;
	}
}

const appPermissions = ref<AppPermission | undefined>(undefined);

const getPermissions = async () => {
	try {
		appPermissions.value = await applicationStore.getPermissions(
			application.value?.name
		);
	} catch (error) {
		console.log(error);
	}
};

const appRegisterProviders = ref<PermissionProviderRegister[] | undefined>([]);
const getProviders = async () => {
	try {
		appRegisterProviders.value = await applicationStore.getProviderRegistryList(
			application.value?.name
		);
	} catch (error) {
		console.log(error);
	}
};

const updateApplication = () => {
	application.value = applicationStore.getApplicationById(
		Route.params.name as string
	);
};

onMounted(async () => {
	bus.on('entrance_state_event', updateApplication);
	checkSecretPermission();
	getPermissions();
	getProviders();
	if (application.value) {
		aclStore.getAppAclStatus(application.value.name);
	}
});

onBeforeUnmount(() => {
	bus.off('entrance_state_event', updateApplication);
});
</script>
