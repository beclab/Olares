import userSpace from '@apps/control-hub/src/assets/user-space.svg';
import osSystem from '@apps/control-hub/src/assets/os-system.svg';
import userSystem from '@apps/control-hub/src/assets/user-system.svg';
import defaultIcon from '@apps/control-hub/src/assets/default.svg';
import gpuIcon from '@apps/control-hub/src/assets/gpu.svg';
import { useAppDetailStore } from '@apps/control-hub/src/stores/AppDetail';
import kubeIcon from '@apps/control-hub/src/assets/kube.png';
import kubesphereIcon from '@apps/control-hub/src/assets/kubesphere.png';
import { useAppList } from '@apps/control-hub/src/stores/AppList';
import { get } from 'lodash';

const appList = useAppList();
const USERSPACE = 'user-space';
const USERSYSTEM = 'user-system';
const appDetail = useAppDetailStore();
let username: undefined | string = '';
let icons: Record<string, string> = {};

export const placeholderIcon = defaultIcon;

export const customNamesapceIcon: any = (username: string) => ({
	[`user-system-${username}`]: userSystem,
	[`user-space-${username}`]: userSpace
});
export const namespaceIcon: any = (username: string) => ({
	[`user-system-${username}`]: userSystem,
	[`user-space-${username}`]: userSpace,
	'os-framework': osSystem,
	'os-platform': osSystem,
	'os-protected': osSystem,
	'os-network': osSystem,
	'os-gpu': osSystem,
	default: userSpace,
	'kubekey-system': kubesphereIcon,
	'kubesphere-monitoring-federated': kubesphereIcon,
	'kubesphere-controls-system': kubesphereIcon,
	'kubesphere-system': kubesphereIcon,
	'kubesphere-monitoring-system': kubesphereIcon,
	'kube-system': kubeIcon,
	'kube-public': kubeIcon,
	'kube-node-lease': kubeIcon,
	'gpu-system': gpuIcon
});

export const getNamespaceIcon = (namespace: string, user?: string) => {
	username = appDetail.data?.user.username;
	const app = namespace.substring(0, namespace.lastIndexOf('-'));
	icons = { ...customNamesapceIcon(username), ...namespaceIcon(username) };
	const apps: any = get(appList, `data.${user}`, []);
	const appTarget = apps.find((item) => item.spec.namespace === namespace);
	return namespace.includes(USERSPACE)
		? userSpace
		: namespace.includes(USERSYSTEM)
		? userSystem
		: appTarget
		? appTarget.spec.icon
		: icons[namespace]
		? icons[namespace]
		: icons[app]
		? icons[app]
		: defaultIcon;
};
