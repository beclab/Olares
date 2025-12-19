<template>
	<MyCard square flat :title="t('DETAILS')">
		<div class="relative-position">
			<DetailPage :data="statusList"> </DetailPage>
			<q-inner-loading :showing="loading"> </q-inner-loading>
		</div>
	</MyCard>
</template>

<script setup lang="ts">
import DetailPage from '@apps/control-panel-common/src/containers/DetailPage.vue';
import { computed, ref } from 'vue';
import { getNodeRoles, getNodeStatus } from '@apps/control-hub/src/utils/node';
import { t } from '@apps/control-hub/src/boot/i18n';
import { getLocalTime } from '@apps/control-hub/src/utils';
import { get, isEmpty } from 'lodash';
import { useRoute } from 'vue-router';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';

interface Props {
	detail: Record<string, any>;
}
const props = withDefaults(defineProps<Props>(), {});
const statusList = computed(() => getAttrs(props.detail));
let loading = ref(false);

const getAttrs = (detail: any) => {
	if (isEmpty(detail)) {
		return;
	}

	const statusStr = getNodeStatus(detail);

	const address = get(detail, 'status.addresses[0].address', '-');
	const nodeInfo = detail.nodeInfo || {};

	return [
		{
			name: t('STATUS'),
			value: statusStr
		},
		{
			name: t('IP_ADDRESS'),
			value: address
		},
		{
			name: t('ROLE'),
			value:
				getNodeRoles(detail.labels).indexOf('master') === -1
					? t('WORKER')
					: t('CONTROL_PLANE')
		},
		{
			name: t('OS_VERSION'),
			value: nodeInfo.osImage
		},
		{
			name: t('OS_TYPE'),
			value: t(nodeInfo.operatingSystem.toUpperCase())
		},
		{
			name: t('KERNEL_VERSION'),
			value: nodeInfo.kernelVersion
		},
		{
			name: t('CONTAINER_RUNTIME'),
			value: nodeInfo.containerRuntimeVersion
		},
		{
			name: t('KUBELET_VERSION'),
			value: nodeInfo.kubeletVersion
		},
		{
			name: t('KUBE_PROXY_VERSION'),
			value: nodeInfo.kubeProxyVersion
		},
		{
			name: t('ARCHITECTURE'),
			value: nodeInfo.architecture.toUpperCase()
		},
		{
			name: t('CREATION_TIME_TCAP'),
			value: getLocalTime(detail.createTime).format('YYYY-MM-DD HH:mm:ss')
		}
	];
};
const route = useRoute();
</script>

<style></style>
