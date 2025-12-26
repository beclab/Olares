import { useDockerStore } from '../stores/docker';

export const resetStatus = () => {
	const dockerStore = useDockerStore();

	dockerStore.appStatus = undefined;
	dockerStore.appInstallState = undefined;
};
