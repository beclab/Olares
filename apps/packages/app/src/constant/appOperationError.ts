import {
	ComputeAvailabilityResponse,
	ComputeBindingValidation,
	ComputeModeOption,
	normalizeComputeAvailabilityResponse
} from './compute';

export const APP_OPERATION_INPUT_TYPE = {
	APP_ENV: 'appenv',
	APP_ENTRANCE: 'appEntrance',
	COMPUTE_MODE_SELECT: 'computeModeSelect',
	COMPUTE_BINDING_REQUIRED: 'computeBindingRequired',
	COMPUTE_BINDING_UNAVAILABLE: 'computeBindingUnavailable'
} as const;

export type AppOperationInputType =
	(typeof APP_OPERATION_INPUT_TYPE)[keyof typeof APP_OPERATION_INPUT_TYPE];

export interface AppUserInputError<T = unknown> {
	type: string;
	data: T;
	validation?: unknown;
}

export interface ComputeModeSelectError {
	type: typeof APP_OPERATION_INPUT_TYPE.COMPUTE_MODE_SELECT;
	data: ComputeModeOption[];
}

export interface ComputeBindingRequiredError {
	type: typeof APP_OPERATION_INPUT_TYPE.COMPUTE_BINDING_REQUIRED;
	data: ComputeAvailabilityResponse;
	validation?: ComputeBindingValidation;
}

export interface ComputeBindingUnavailableError {
	type: typeof APP_OPERATION_INPUT_TYPE.COMPUTE_BINDING_UNAVAILABLE;
	data: ComputeAvailabilityResponse;
	validation?: ComputeBindingValidation;
}

export type ComputeBindingPromptError =
	| ComputeBindingRequiredError
	| ComputeBindingUnavailableError;

export type ComputeUserInputError =
	| ComputeModeSelectError
	| ComputeBindingPromptError;

export interface ComputeBindingPromptPayload {
	availability?: ComputeAvailabilityResponse;
	validation?: ComputeBindingValidation;
	[key: string]: unknown;
}

// Backend may nest availability under `Data.availability`.
export function normalizeComputeBindingPromptData(raw: unknown): {
	availability: ComputeAvailabilityResponse;
	validation?: ComputeBindingValidation;
} {
	if (!raw || typeof raw !== 'object') {
		return {
			availability: normalizeComputeAvailabilityResponse({
				schedulable: false,
				requirement: { mode: '', requiredGpu: 0 },
				scope: 'card',
				nodes: []
			})
		};
	}

	const payload = raw as ComputeBindingPromptPayload;
	const availability =
		payload.availability && typeof payload.availability === 'object'
			? payload.availability
			: (raw as ComputeAvailabilityResponse);

	const validation =
		payload.validation ??
		availability.validation ??
		(raw as { validation?: ComputeBindingValidation }).validation;

	return {
		availability: normalizeComputeAvailabilityResponse(availability),
		validation
	};
}

const hasType = (value: unknown, type: string): boolean =>
	!!value &&
	typeof value === 'object' &&
	(value as { type?: unknown }).type === type;

export const isAppEnvError = (value: unknown): value is AppUserInputError =>
	hasType(value, APP_OPERATION_INPUT_TYPE.APP_ENV);

export const isAppEntranceError = (
	value: unknown
): value is AppUserInputError =>
	hasType(value, APP_OPERATION_INPUT_TYPE.APP_ENTRANCE);

export function isComputeModeSelectError(
	value: unknown
): value is ComputeModeSelectError {
	return hasType(value, APP_OPERATION_INPUT_TYPE.COMPUTE_MODE_SELECT);
}

export function isComputeBindingRequiredError(
	value: unknown
): value is ComputeBindingRequiredError {
	return hasType(value, APP_OPERATION_INPUT_TYPE.COMPUTE_BINDING_REQUIRED);
}

export function isComputeBindingUnavailableError(
	value: unknown
): value is ComputeBindingUnavailableError {
	return hasType(value, APP_OPERATION_INPUT_TYPE.COMPUTE_BINDING_UNAVAILABLE);
}

export function isComputeBindingPromptError(
	value: unknown
): value is ComputeBindingPromptError {
	return (
		isComputeBindingRequiredError(value) ||
		isComputeBindingUnavailableError(value)
	);
}
