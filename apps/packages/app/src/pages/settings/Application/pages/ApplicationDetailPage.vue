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

		<bt-list
			first
			:label="t('entrances')"
			v-if="application?.entrances && application.entrances.length > 0"
		>
			<template v-for="(entrance, index) in application.entrances" :key="index">
				<application-item
					:icon="
						entrance.icon
							? entrance.icon
							: application?.icon
							? application?.icon
							: ''
					"
					:app-name="application.name"
					:raw-app-name="application.rawAppName"
					:cs-app="application.isClusterScoped"
					:title="entrance.title"
					:status="entrance.state"
					:width-separator="index !== application.entrances.length - 1"
					:margin-top="index !== 0"
					@click="gotoEntrance(entrance)"
					:hide-status="false"
				>
					<div class="text-body1 text-ink-1">
						{{ getAuthLevel(entrance.authLevel) }}
					</div>
				</application-item>
			</template>
		</bt-list>

		<bt-list
			first
			:label="t('Shared entrances')"
			v-if="
				application?.sharedEntrances && application.sharedEntrances.length > 0
			"
		>
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
					:app-name="application.name"
					:raw-app-name="application.rawAppName"
					:cs-app="application.isClusterScoped"
					:title="entrance.title || entrance.name"
					:hideStatus="true"
					:width-separator="index !== application.sharedEntrances.length - 1"
					:margin-top="index !== 0"
					@click="gotoEntrance(entrance, true)"
				/>
			</template>
		</bt-list>

		<bt-list
			first
			:label="t('Environment variables')"
			v-if="application?.owner && application.owner == adminStore.user.name"
		>
			<bt-form-item
				:title="t('Manage environment variables')"
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

		<bt-list
			first
			:label="t('permissions')"
			v-if="
				(application && application.ports && application.ports.length > 0) ||
				(aclStore.appAclList && aclStore.appAclList.length > 0)
			"
		>
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
				:width-separator="
					appRegisterProviders && appRegisterProviders.length > 0
				"
				:chevron-right="true"
			/>

			<!--			<bt-form-item-->
			<!--				v-if="appRegisterProviders && appRegisterProviders.length > 0"-->
			<!--				:title="t('Providers')"-->
			<!--				:margin-top="false"-->
			<!--				:width-separator="false"-->
			<!--				:chevron-right="true"-->
			<!--				@click="gotoProvider"-->
			<!--			/>-->
		</bt-list>

		<div class="full-width q-mb-lg" />
	</bt-scroll-area>
</template>

<script setup lang="ts">
import ApplicationOperateItem from 'src/components/settings/application/ApplicationOperateItem.vue';
import ApplicationItem from 'src/components/settings/application/ApplicationItem.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { AppPermission, PermissionProviderRegister } from 'src/constant/global';
import { useApplicationStore } from 'src/stores/settings/application';
import { useDeviceStore } from 'src/stores/settings/device';
import { useSecretStore } from 'src/stores/settings/secret';
import { useAdminStore } from 'src/stores/settings/admin';
import { useAclStore } from 'src/stores/settings/acl';
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { TerminusEntrance } from '@bytetrade/core';
import { useRoute, useRouter } from 'vue-router';
import { authLevelOptions } from 'src/constant';
import { useI18n } from 'vue-i18n';
import { bus } from 'src/utils/bus';

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

const getAuthLevel = (authLevel: string) => {
	const data = authLevelOptions().find((item) => {
		return item.value === authLevel;
	});
	if (data) {
		return data.label;
	}
	return '';
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

const gotoProvider = () => {
	router.push({
		path: `/application/${application.value?.name}/provider`
	});
};

const gotoManagerEnvironment = (query: any) => {
	router.push({
		path: '/application/environment/manager',
		query
	});
};

async function checkSecretPermission() {
	// const res = await secretStore.checkSecretPermission(application.value?.name);
	// if (res && res.permission === true) {
	// 	secretPermission.value = true;
	// } else {
	// 	secretPermission.value = false;
	// }
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
		if (application.value?.owner === adminStore.user.name) {
			appRegisterProviders.value =
				await applicationStore.getProviderRegistryList(application.value?.name);
		}
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
