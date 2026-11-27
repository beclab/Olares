import { useApplicationStore } from 'src/stores/desktop/app';
import { useBookmarkStore } from 'src/stores/desktop/bookmark';
import { DisplayItem, parseKey } from 'src/types/desktop-layout';

// Render-time resolver for a 'a:<id>' / 'b:<id>' layout key.
// M1 only implements the 'a:' branch; the 'b:' branch lands with M2's
// bookmark store. Anything we can't resolve is returned with kind='unknown'
// so the caller can decide whether to skip it or show a placeholder.

export function resolveDesktopItem(key: string): DisplayItem {
	const parsed = parseKey(key);
	if (!parsed) {
		return unknownItem(key);
	}

	if (parsed.kind === 'a') {
		const appStore = useApplicationStore();
		const app = appStore.myApps.find((a: any) => a.id === parsed.id);
		if (!app) {
			return unknownItem(key);
		}
		return {
			kind: 'app',
			key,
			id: parsed.id,
			icon: (app as any).icon,
			name: (app as any).name,
			title: (app as any).title,
			url: (app as any).url,
			openMethod: (app as any).openMethod,
			source: app
		};
	}

	if (parsed.kind === 'b') {
		const bookmarkStore = useBookmarkStore();
		const bm = bookmarkStore.byId[parsed.id];
		if (!bm) {
			return unknownItem(key);
		}
		return {
			kind: 'bookmark',
			key,
			id: parsed.id,
			icon: bm.icon || '',
			name: bm.name,
			title: bm.name,
			url: bm.url,
			openMethod: bm.openMode,
			source: bm
		};
	}

	return unknownItem(key);
}

function unknownItem(key: string): DisplayItem {
	return {
		kind: 'unknown',
		key,
		id: '',
		icon: '',
		name: '',
		title: ''
	};
}
