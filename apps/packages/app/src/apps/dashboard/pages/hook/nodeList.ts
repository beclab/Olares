import { PodItem } from '@apps/control-panel-common/network/network';
import { getNodesList } from '@apps/dashboard/src/network';
import { ref } from 'vue';

const params = {
	sortBy: 'createTime',
	limit: -1
};

export const useNodeList = () => {
	const nodeList = ref<PodItem[]>([]);

	const requestNodeList = async () => {
		if (nodeList.value.length > 0) {
			return;
		}
		const res = await getNodesList(params);
		nodeList.value = res.data.items;
	};

	return { nodeList: nodeList, requestNodeList };
};
