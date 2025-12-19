import { ref } from 'vue';

export const assetsUrl = ref('/');

export const setAssetsUrl = (urlStr: string) => {
	assetsUrl.value = urlStr;
};
export function getPluginImageUrl(path: string): string {
	if (!path) {
		return '';
	}
	if (path.startsWith('http')) {
		return path;
	}
	return assetsUrl.value + 'www/plugin/' + path;
}
