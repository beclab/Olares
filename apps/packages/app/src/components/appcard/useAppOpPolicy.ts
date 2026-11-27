// Reactive composable that exposes the current app-operation policy
// decision for a single (appName, sourceId, op) tuple to UI components.
//
// Components (InstallButton, useAppAction menu items, etc.) consume:
//   - decision      : the full AppOpDecision (kept for future branching).
//   - errorGroups   : `ErrorGroup[]` derived from the decision; same shape
//                     used by preCheckStore.frontendPreflight, so it can be
//                     handed to `emit('onError', …)` straight away.
//   - isDisabled    : convenience boolean for `disable`-mode rejections.
//   - isHardBlocked : convenience boolean for `hard`-mode rejections.
//   - tooltip       : pre-translated string for q-tooltip / title attr.
//   - present(mode) : surfaces the rejection via the requested presenter
//                     ('dialog' | 'errorGroup' | 'none'). Returns the
//                     ErrorGroup[] so callers can re-emit if needed.
//
// Reactivity: depends on meStore.isAdmin + appStore aggregation (both
// Pinia state), so wrapping in `computed` is enough to stay live.

import { computed, Ref, unref } from 'vue';
import {
	AppOpPresenter,
	buildAppOpContext,
	decisionToErrorGroups,
	evaluateAppOp,
	presentAppOpRejection
} from 'src/services/market/appOperation/appOperationPolicy';
import { OPERATION_TYPE } from 'src/services/market/appOperation/operationTypes';

type MaybeRef<T> = T | Ref<T>;

interface UseAppOpPolicyInput {
	op: MaybeRef<OPERATION_TYPE | null>;
	appName: MaybeRef<string>;
	sourceId: MaybeRef<string>;
}

export function useAppOpPolicy(input: UseAppOpPolicyInput) {
	const decision = computed(() => {
		const op = unref(input.op);
		if (op === null) return { allowed: true } as const;
		return evaluateAppOp(
			buildAppOpContext({
				op,
				appName: unref(input.appName),
				sourceId: unref(input.sourceId)
			})
		);
	});

	const errorGroups = computed(() => decisionToErrorGroups(decision.value));

	const isDisabled = computed(
		() => !decision.value.allowed && decision.value.mode === 'disable'
	);

	const isHardBlocked = computed(
		() => !decision.value.allowed && decision.value.mode === 'hard'
	);

	const tooltip = computed(() =>
		decision.value.allowed ? '' : decision.value.message ?? ''
	);

	/**
	 * Surface the current rejection via the chosen presenter and return the
	 * resulting `ErrorGroup[]` (always returned, regardless of presenter, so
	 * callers can also bubble it via `emit('onError', …)`).
	 */
	async function present(
		mode: AppOpPresenter = 'dialog'
	): Promise<ReturnType<typeof decisionToErrorGroups>> {
		return presentAppOpRejection(decision.value, mode);
	}

	return {
		decision,
		errorGroups,
		isDisabled,
		isHardBlocked,
		tooltip,
		present
	};
}
