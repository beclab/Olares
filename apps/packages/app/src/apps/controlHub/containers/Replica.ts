import { ref } from 'vue';

interface ReplicaStatus {
	current?: number;
	desire?: number;
}
export const replicaStatus = ref<ReplicaStatus>({});
