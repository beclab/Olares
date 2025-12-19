import { useConfigStore } from 'src/stores/rss-config';
import axios from 'axios';

export async function queryKnowledgeRestore(): Promise<boolean> {
	const configStore = useConfigStore();
	return await axios.get(
		configStore.url + '/knowledge/backup/knowledge_restore_status'
	);
}
