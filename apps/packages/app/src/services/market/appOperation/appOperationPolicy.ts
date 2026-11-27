// App-operation access-control policy (cross-cutting layer).
//
// Why a dedicated module?
//   The market is shifting from CSV2 (v2) to Shared (v3) apps, and each
//   version has version- and role-specific gates (e.g. blocked operations,
//   greyed-out install, restricted Open). Embedding these rules inside the
//   per-component UI logic and `appService.ts` makes them hard to find,
//   evolve, and audit.
//
// Architecture:
//   1. `evaluateAppOp` — pure rule evaluation. Given an `AppOpContext`, it
//      returns an `AppOpDecision` describing whether the operation is
//      allowed and why. No UI / store side-effects.
//   2. `buildAppOpContext` — gathers role / app-version inputs from the
//      stores so callers don't have to repeat the wiring.
//   3. `gateAppOp` — high-level entry used by `appService.ts` & UI handlers:
//      builds the context, runs the rules, surfaces a blocking dialog when
//      the rule is `hard`, and returns whether the caller may proceed.
//   4. `showAppOpRejectedDialog` — single-button BtDialog used by `gateAppOp`
//      and reused by hover-tooltips that need an explainer popup.
//
// New rules are added by appending to `APP_OP_RULES`. Each rule is local,
// declarative, and individually testable.

import { BtDialog } from '@bytetrade/ui';
import { i18n } from 'src/boot/i18n';
import { bus, BUS_EVENT } from 'src/utils/bus';
import { isCloneApp, isTemplateApp, uninstalledApp } from 'src/constant/config';
import { useMeStore } from 'src/stores/settings/me';
import { useAppStore } from 'src/stores/market/appStore';
import { readMarketDebugFlags } from '../marketDebugFlags';
import { OPERATION_TYPE } from './operationTypes';
import {
	AppSimpleInfoLatest,
	isCSV2,
	isSharedV3
} from 'src/constant/constants';
import {
	ErrorGroup,
	ErrorGroupHandler,
	errorsToStructuredText,
	findLevel1Error,
	level1TitleKey
} from 'src/constant/errorGroupHandler';

// ---------- public types ----------

/**
 * Distinguishes the "shape" of a rejection so callers can react differently:
 *   - `hard`   : action is forbidden; clicking must show a blocking dialog.
 *   - `disable`: the control should be visually disabled (greyed-out) with
 *                a hover tooltip; clicking should not trigger the action.
 */
export type AppOpRejectMode = 'hard' | 'disable';

/** Stable identifier for the rejection (used for i18n + tests). */
export type AppOpRejectKey =
	| 'csv2.install_blocked'
	| 'csv2.upgrade_blocked'
	| 'sharedv3.subuser_install_blocked'
	| 'sharedv3.subuser_operation_blocked'
	| 'template.install_blocked';

/**
 * How a policy rejection should be surfaced to the user:
 *   - 'dialog'     : BtDialog single-button modal (default; eye-catching).
 *   - 'errorGroup' : Same pipeline as `preCheckStore.frontendPreflight` —
 *                    converts the decision into an `ErrorGroup[]` and emits
 *                    `BUS_EVENT.APP_BACKEND_ERROR` so the existing error
 *                    banner UI renders it.
 *   - 'none'       : No UI side-effect; caller handles the returned
 *                    decision / errorGroups itself.
 */
export type AppOpPresenter = 'dialog' | 'errorGroup' | 'none';

export interface AppOpDecision {
	allowed: boolean;
	mode?: AppOpRejectMode;
	key?: AppOpRejectKey;
	/** Pre-translated message (uses i18n.global.t under the hood). */
	message?: string;
}

export const APP_OP_REJECT_CODE: Record<AppOpRejectKey, ErrorGroupHandler> = {
	'csv2.install_blocked': ErrorGroupHandler.G022,
	'csv2.upgrade_blocked': ErrorGroupHandler.G023,
	'sharedv3.subuser_install_blocked': ErrorGroupHandler.G024,
	'sharedv3.subuser_operation_blocked': ErrorGroupHandler.G025,
	'template.install_blocked': ErrorGroupHandler.G026
};

/**
 * Single source of truth for the i18n path of a reject key. Both the
 * dialog (decision.message) and the ErrorGroup (group.title) presentation
 * paths resolve through the shared `error.*` key registered in level1Map.
 */
function appOpI18nKey(key: AppOpRejectKey): string {
	return level1TitleKey(APP_OP_REJECT_CODE[key]) ?? 'error.unknown_error';
}

export interface AppOpContext {
	op: OPERATION_TYPE;
	appName: string;
	sourceId: string;
	isV2: boolean;
	isV3: boolean;
	isAdmin: boolean;
	isInstalled: boolean;
	isTemplate: boolean;
}

// ---------- rule table ----------

interface AppOpRule {
	key: AppOpRejectKey;
	mode: AppOpRejectMode;
	when: (ctx: AppOpContext) => boolean;
}

const APP_OP_RULES: AppOpRule[] = [
	// CSV2 (v2) apps: install / upgrade are globally retired. Every user
	// (including admin/owner) sees the same blocking dialog. Other lifecycle
	// operations on already-installed v2 apps remain allowed.
	{
		key: 'csv2.install_blocked',
		mode: 'hard',
		when: (c) =>
			c.isV2 &&
			c.op === OPERATION_TYPE.INSTALL &&
			!readMarketDebugFlags().bypassV2Install
	},
	{
		key: 'csv2.upgrade_blocked',
		mode: 'hard',
		when: (c) =>
			c.isV2 &&
			c.op === OPERATION_TYPE.UPGRADE &&
			!readMarketDebugFlags().bypassV2Upgrade
	},
	{
		key: 'sharedv3.subuser_operation_blocked',
		mode: 'hard',
		when: (c) => c.isV3 && !c.isAdmin && c.op !== OPERATION_TYPE.REMOVE
	},
	{
		key: 'template.install_blocked',
		mode: 'hard',
		when: (c) => c.isTemplate && c.op === OPERATION_TYPE.INSTALL
	}
];

// ---------- pure evaluator ----------

export function evaluateAppOp(ctx: AppOpContext): AppOpDecision {
	const matched = APP_OP_RULES.find((r) => r.when(ctx));
	if (!matched) return { allowed: true };

	return {
		allowed: false,
		mode: matched.mode,
		key: matched.key,
		message: i18n.global.t(appOpI18nKey(matched.key))
	};
}

// ---------- store-aware helpers ----------

interface BuildContextInput {
	op: OPERATION_TYPE;
	appName: string;
	sourceId: string;
	/** Optional pre-resolved simple info; falls back to appStore lookup.
	 *  v2/v3 classification now lives on simple info, so we no longer
	 *  need the (much heavier) full info blob to gate operations.
	 */
	simple?: AppSimpleInfoLatest | null;
}

export function buildAppOpContext(input: BuildContextInput): AppOpContext {
	const appStore = useAppStore();
	const meStore = useMeStore();

	const aggregation = appStore.getAppAggregationInfo(
		input.appName,
		input.sourceId
	);
	const simple = input.simple ?? aggregation?.app_simple_latest ?? null;
	const statusInfo = aggregation?.app_status_latest?.status;

	// "Installed" here means: backend has a status row AND it is not in any of
	// the uninstalled meta-states tracked by `config.uninstalledApp`.
	const installed = !!statusInfo && !uninstalledApp(statusInfo);

	return {
		op: input.op,
		appName: input.appName,
		sourceId: input.sourceId,
		isV2: isCSV2(simple ?? undefined),
		isV3: isSharedV3(simple ?? undefined),
		isAdmin: !!meStore.isAdmin,
		isInstalled: installed,
		isTemplate:
			isTemplateApp(simple?.app_simple_info) && !isCloneApp(statusInfo)
	};
}

/**
 * Convert a (rejecting) decision into an `ErrorGroup[]` matching the shape
 * produced by `preCheckStore.frontendPreflight`. Built from the reject key's
 * unified G-code via `findLevel1Error`, so dialog and errorGroup modes share
 * a single `error.*` i18n entry per reject reason. Returns an empty array for
 * `allowed` decisions.
 */
export function decisionToErrorGroups(decision: AppOpDecision): ErrorGroup[] {
	if (decision.allowed || !decision.key) return [];
	const group = findLevel1Error(APP_OP_REJECT_CODE[decision.key]);
	return group ? [group] : [];
}

/**
 * Show the standard single-button reject dialog. Returns a promise that
 * resolves when the user dismisses the dialog, so callers can `await` it.
 */
export async function showAppOpRejectedDialog(
	decision: AppOpDecision
): Promise<void> {
	if (decision.allowed) return;
	const t = i18n.global.t;
	try {
		await BtDialog.show({
			title: t('app_policy.dialog_title'),
			message: decision.message ?? t('app_policy.dialog_default_message'),
			okText: t('base.ok'),
			cancel: false
		});
	} catch (e) {
		// BtDialog rejects on cancel; for a single-button dialog this is benign.
		console.debug('app-policy reject dialog dismissed', e);
	}
}

/**
 * Pre-check-style presentation: emit the rejection through the global
 * `APP_BACKEND_ERROR` bus event so the existing error banner picks it up.
 * Returns the `ErrorGroup[]` for callers that want to bubble it via the
 * `@on-error` chain (matches `InstallButton.vue`'s `emit('onError', …)`
 * pipeline today).
 */
export function emitAppOpRejectedAsErrorGroup(
	decision: AppOpDecision
): ErrorGroup[] {
	const groups = decisionToErrorGroups(decision);
	if (groups.length === 0) return groups;
	bus.emit(BUS_EVENT.APP_BACKEND_ERROR, errorsToStructuredText(groups));
	return groups;
}

/**
 * Dispatcher used by `gateAppOp` and the UI composable. Pick a presenter
 * once at the call site; the implementation lives in one place.
 */
export async function presentAppOpRejection(
	decision: AppOpDecision,
	present: AppOpPresenter = 'errorGroup'
): Promise<ErrorGroup[]> {
	if (decision.allowed) return [];
	switch (present) {
		case 'dialog':
			await showAppOpRejectedDialog(decision);
			return decisionToErrorGroups(decision);
		case 'errorGroup':
			return emitAppOpRejectedAsErrorGroup(decision);
		case 'none':
		default:
			return decisionToErrorGroups(decision);
	}
}

/**
 * High-level gate: builds context, evaluates rules, and (if rejected)
 * surfaces the result with the requested presenter. Returns `true` when the
 * caller may proceed. Default presenter is `'dialog'` to preserve existing
 * `appService.ts` behaviour.
 *
 * Used by `appService.ts` entry points so every code path that performs an
 * app operation funnels through the policy layer (defence in depth). UI
 * components that prefer the pre-check banner style can pass
 * `present: 'errorGroup'`.
 */
export async function gateAppOp(
	input: BuildContextInput & { present?: AppOpPresenter }
): Promise<boolean> {
	const decision = evaluateAppOp(buildAppOpContext(input));
	if (decision.allowed) return true;
	// `disable`-mode rejections shouldn't normally be triggered from a click
	// path (UI should have disabled the control); but if we get here we still
	// honour the rule via the chosen presenter and block the call.
	await presentAppOpRejection(decision, input.present);
	return false;
}
