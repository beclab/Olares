/**
 * Lightweight registry that lets transient UI (e.g. a Google Play
 * "prominent disclosure" dialog that must not be silently dismissed by
 * pressing the Android back button) opt into blocking the platform-level
 * back handler while it is on screen.
 *
 * Each guard is a predicate that returns `true` to swallow the back press.
 * Guards are evaluated in LIFO order; the first one that returns `true`
 * wins. This keeps nested overlays composable without coordination.
 */
type BackButtonGuardFn = () => boolean;

const guards: BackButtonGuardFn[] = [];

export function pushBackButtonGuard(guard: BackButtonGuardFn): () => void {
	guards.push(guard);
	return () => {
		const idx = guards.indexOf(guard);
		if (idx >= 0) {
			guards.splice(idx, 1);
		}
	};
}

export function shouldBlockBackButton(): boolean {
	for (let i = guards.length - 1; i >= 0; i--) {
		try {
			if (guards[i]()) {
				return true;
			}
		} catch {
			// A throwing guard should never break the back handler; skip it.
		}
	}
	return false;
}
