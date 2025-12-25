<template>
	<app-intro-card
		v-if="showDialog"
		class="q-mt-lg"
		:title="t('detail.required_permissions')"
	>
		<template v-slot:content>
			<div class="permission-request-grid column justify-start">
				<permission-request
					v-if="filePermissionData.length > 0"
					name="sym_r_drive_file_move"
					:title="t('Access to Files')"
					:nodes="filePermissionData"
				/>
				<permission-request
					v-if="publicEntrancePermission.length > 0"
					name="sym_r_dynamic_feed"
					:title="t('Provide Public Entrance')"
					:nodes="publicEntrancePermission"
				/>
				<permission-request
					v-if="noVisibleEntrancePermission.length > 0"
					name="sym_r_remove_selection"
					:title="t('No Visible Entrance')"
					:nodes="noVisibleEntrancePermission"
				/>

				<permission-request
					v-if="clusterScopedApp"
					name="sym_r_inbox_text_share"
					:title="t('Shared App')"
					:nodes="[
						{
							label: t(
								'This app is shared by all users in the same Olares cluster.'
							)
						}
					]"
				/>
				<permission-request
					v-if="adminPermission"
					name="sym_r_admin_panel_settings"
					:title="t('Administrator Only')"
					:nodes="[
						{
							label: t(
								'This app requires Olares administrator privileges to install.'
							)
						}
					]"
				/>

				<permission-request
					v-if="clonePermission"
					name="sym_r_select_window_2"
					:title="t('Multiple Instances')"
					:nodes="[
						{
							label: t(
								'This app can be cloned into multiple independent instances after installation.'
							)
						}
					]"
				/>

				<permission-request
					v-if="systemEnvsPermission.length > 0"
					name="sym_r_stacks"
					:title="t('Using System Environment Variables')"
					:nodes="systemEnvsPermission"
				/>

				<permission-request
					v-if="providerPermission.length > 0"
					name="sym_r_modeling"
					:title="t('Connect to Other Apps')"
					:nodes="providerPermission"
				/>

				<permission-request
					v-if="middlewarePermission.length > 0"
					name="sym_r_dataset"
					:title="t('Using Middleware on Olares')"
					:nodes="middlewarePermission"
				/>

				<permission-request
					v-if="portPermission.length > 0"
					name="sym_r_hub"
					:title="t('Expose Ports for Remote Access')"
					:nodes="portPermission"
				/>
			</div>
		</template>
	</app-intro-card>
</template>

<script setup lang="ts">
import PermissionRequest from '../../../../components/appintro/PermissionRequest.vue';
import AppIntroCard from '../../../../components/appintro/AppIntroCard.vue';
import { PermissionNode } from '../../../../constant/constants';
import { computed, PropType } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	appEntry: {
		type: Object as PropType<any>,
		require: true
	},
	appName: {
		type: String,
		require: true
	},
	sourceId: {
		type: String,
		require: true
	}
});

const { t } = useI18n();

const showDialog = computed(() => {
	return (
		filePermissionData.value.length > 0 ||
		publicEntrancePermission.value.length > 0 ||
		noVisibleEntrancePermission.value.length > 0 ||
		clusterScopedApp.value ||
		adminPermission.value ||
		clonePermission.value ||
		systemEnvsPermission.value.length > 0 ||
		providerPermission.value.length > 0 ||
		middlewarePermission.value.length > 0 ||
		portPermission.value.length > 0
	);
});

const clusterScopedApp = computed(
	() => props.appEntry.options?.appScope?.clusterScoped ?? false
);

const adminPermission = computed(() => props.appEntry.onlyAdmin);

const clonePermission = computed(
	() => props.appEntry?.options?.allowMultipleInstall
);

type PermissionText =
	| 'Data, Cache and User directories'
	| 'Data and Cache directories'
	| 'Data and User directories'
	| 'Cache and User directories'
	| 'User directory'
	| 'Data directory'
	| 'Cache directory';

const filePermissionData = computed<PermissionNode[]>(() => {
	const permission = props.appEntry.permission;
	if (!permission) return [];

	const hasAppData = permission.appData;
	const hasAppCache = permission.appCache;
	const hasUserDir = permission.userData && permission.userData.length > 0;

	let displayText: PermissionText;

	switch (true) {
		case hasAppData && hasAppCache && hasUserDir:
			displayText = 'Data, Cache and User directories';
			break;
		case hasAppData && hasAppCache && !hasUserDir:
			displayText = 'Data and Cache directories';
			break;
		case hasAppData && !hasAppCache && hasUserDir:
			displayText = 'Data and User directories';
			break;
		case !hasAppData && hasAppCache && hasUserDir:
			displayText = 'Cache and User directories';
			break;
		case !hasAppData && !hasAppCache && hasUserDir:
			displayText = 'User directory';
			break;
		case hasAppData && !hasAppCache && !hasUserDir:
			displayText = 'Data directory';
			break;
		case !hasAppData && hasAppCache && !hasUserDir:
			displayText = 'Cache directory';
			break;
		default:
			return [];
	}

	return [{ label: t(displayText), children: [] }];
});

const publicEntrancePermission = computed<PermissionNode[]>(() => {
	const entrances = props.appEntry?.entrances;
	if (!entrances || entrances.length === 0) return [];

	const hasPublicEntrance = entrances.some(
		(item) => item.authLevel === 'public'
	);

	if (!hasPublicEntrance) return [];

	return [
		{
			label: t(
				'This app provides an publicly accessible entrance that does not require authentication.'
			),
			children: []
		},
		{
			label: t(
				"All traffic to this entrance may consume your reverse proxy's bandwidth."
			),
			children: []
		}
	];
});

const noVisibleEntrancePermission = computed<PermissionNode[]>(() => {
	const entrances = props.appEntry?.entrances;
	if (!entrances || entrances.length === 0) return [];

	const desktopSize = entrances.filter((item) => !item.invisible).length;

	if (desktopSize !== 0) return [];

	return [
		{
			label: t(
				'This is a background service with no user interface. It provides an API for other apps to interact with.'
			),
			children: []
		}
	];
});

const systemEnvsPermission = computed<PermissionNode[]>(() => {
	const envs = props.appEntry?.envs;
	console.log(envs);
	if (!envs || !Array.isArray(envs) || envs.length === 0) {
		return [];
	}

	const envChildren: PermissionNode[] = envs
		.filter((env) => {
			return env.valueFrom && typeof env.valueFrom.envName === 'string';
		})
		.map((env) => {
			const label = `${env.envName}=[${env.valueFrom.envName}]`;
			return { label, children: [] };
		});

	if (envChildren.length === 0) {
		return [];
	}

	return [
		{
			label: t(
				'This app retrieves its env values from the following system environment variables.'
			),
			children: envChildren
		}
	];
});

const providerPermission = computed<PermissionNode[]>(() => {
	const providers = props.appEntry?.permission?.provider;
	if (!providers || !Array.isArray(providers) || providers.length === 0) {
		return [];
	}

	const providerChildren: PermissionNode[] = providers
		.filter((provider) => {
			return !!provider.providerName || !!provider.appName;
		})
		.map((provider) => {
			const label = provider.providerName || provider.appName;
			return { label, children: [] };
		});

	if (providerChildren.length === 0) {
		return [];
	}

	return [
		{
			label: t(
				'This app requires permission to call following providers to extend its functionality'
			),
			children: providerChildren
		}
	];
});

const middlewarePermission = computed<PermissionNode[]>(() => {
	const middleware = props.appEntry?.middleware;

	console.log(props.appEntry);
	console.log(middleware);
	if (
		!middleware ||
		typeof middleware !== 'object' ||
		Object.keys(middleware).length === 0
	) {
		return [];
	}

	const middlewareChildren: PermissionNode[] = Object.entries(middleware).map(
		([middlewareName, cfg]) => {
			let label = middlewareName;
			return { label, children: [] };
		}
	);

	console.log(middlewareChildren);

	return [
		{
			label: t('This app requires the following middleware.'),
			children: middlewareChildren
		}
	];
});

const portPermission = computed<PermissionNode[]>(() => {
	const ports = props.appEntry?.ports || [];
	const tailscaleAcls = props.appEntry?.tailscale?.acls || [];
	const isShow =
		tailscaleAcls.length > 0 || ports.some((port) => port.addToTailscaleAcl);

	if (!isShow) {
		return [];
	}

	const portGroupMap = new Map<string, string[]>();

	const targetPorts = ports.filter((port) => port.addToTailscaleAcl);
	targetPorts.forEach((port) => {
		const proto = port.protocol?.trim() || 'all';
		const portStr = port.exposePort ? `*:${port.exposePort}` : '*:random port';

		if (!portGroupMap.has(proto)) {
			portGroupMap.set(proto, []);
		}
		portGroupMap.get(proto)!.push(portStr);
	});

	tailscaleAcls.forEach((acl) => {
		acl.dst.forEach((dstStr) => {
			if (!dstStr) return;

			let proto = acl.proto || 'all';

			if (!portGroupMap.has(proto)) {
				portGroupMap.set(proto, []);
			}
			portGroupMap.get(proto)!.push(dstStr);
		});
	});

	if (portGroupMap.size === 0) {
		return [];
	}

	console.log(portGroupMap);

	const groupChildren: PermissionNode[] = Array.from(
		portGroupMap.entries()
	).map(([proto, portStrs]) => {
		const uniquePortStrs = [...new Set(portStrs)];
		const portChildren = uniquePortStrs.map((portStr) => ({
			label: portStr,
			children: []
		}));

		const protoLabel = proto === 'all' ? 'all' : proto.toLowerCase();
		return {
			label: protoLabel,
			children: portChildren
		};
	});

	return [
		{
			label: t(
				'This app opens the following ports for remote access when VPN is enabled'
			),
			children: groupChildren
		}
	];
});
</script>

<style scoped lang="scss">
.permission-request-grid {
	width: 100%;
	align-items: start;
	justify-items: center;
	justify-content: center;
	display: grid;
	grid-row-gap: 20px;
	grid-template-columns: repeat(1, minmax(0, 1fr));

	//@media (max-width: 1440px) {
	//	grid-template-columns: repeat(1, minmax(0, 1fr));
	//}
	//
	//@media (min-width: 1441px) {
	//	grid-template-columns: repeat(2, minmax(0, 1fr));
	//}
}
</style>
