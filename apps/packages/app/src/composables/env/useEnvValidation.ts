import { useI18n } from 'vue-i18n';
import { BaseEnv } from 'src/constant';

export interface EnvValidationResult {
	right: boolean;
	error: string;
}

const emailReg = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function useEnvValidation(): {
	validateScalarConstraints: (
		item: BaseEnv,
		valueToCheck: string,
		allowedValues?: string[]
	) => EnvValidationResult;
	validateEnvValue: (
		item: BaseEnv,
		valueToCheck: string,
		allowedValues?: string[]
	) => EnvValidationResult;
} {
	const { t } = useI18n();

	/**
	 * The allowed-value whitelist comes from either the resolved option list
	 * (passed in by the caller — this covers `remoteOptions`, whose entries are
	 * fetched at runtime and never written back to `item.options`) or, as a
	 * fallback, the statically declared `item.options`.
	 */
	function resolveAllowedValues(
		item: BaseEnv,
		allowedValues?: string[]
	): string[] {
		if (allowedValues && allowedValues.length) {
			return allowedValues.map((v) => String(v));
		}
		return item.options?.map((o) => String(o.value)) ?? [];
	}

	function validateScalarConstraints(
		item: BaseEnv,
		valueToCheck: string,
		allowedValues?: string[]
	): EnvValidationResult {
		const trimmed = String(valueToCheck ?? '').trim();
		const allowed = resolveAllowedValues(item, allowedValues);
		if (allowed.length && trimmed !== '') {
			const ok = allowed.some((v) => v.trim() === trimmed);
			if (!ok) {
				return {
					right: false,
					error: t('Application variable validation failed', {
						type: item.type ?? 'value'
					})
				};
			}
		}

		if (item.regex) {
			try {
				const reg = new RegExp(item.regex);
				if (!reg.test(valueToCheck)) {
					return {
						right: false,
						error: t(
							'does not meet format requirements (regex validation failed)',
							{ regex: item.regex }
						)
					};
				}
			} catch {
				return {
					right: false,
					error: t(
						'does not meet format requirements (regex validation failed)',
						{ regex: String(item.regex) }
					)
				};
			}
		}

		switch (item.type) {
			case 'number':
			case 'int':
				if (isNaN(Number(valueToCheck))) {
					return { right: false, error: t('must be a number') };
				}
				break;
			case 'email':
				if (!emailReg.test(valueToCheck)) {
					return { right: false, error: t('invalid email format') };
				}
				break;
			case 'bool':
				{
					const v = String(valueToCheck).trim().toLowerCase();
					if (v !== 'true' && v !== 'false') {
						return {
							right: false,
							error: t('Application variable validation failed', {
								type: item.type
							})
						};
					}
				}
				break;
			default:
				break;
		}

		return { right: true, error: '' };
	}

	function validateEnvValue(
		item: BaseEnv,
		valueToCheck: string,
		allowedValues?: string[]
	): EnvValidationResult {
		if (item.multiSelect) {
			const sp = item.splitter || ',';
			const parts = String(valueToCheck ?? '')
				.split(sp)
				.map((part) => part.trim())
				.filter((part) => part !== '');
			if (parts.length === 0) {
				if (item.required) {
					return { right: false, error: t('cannot be empty') };
				}
				return { right: true, error: '' };
			}
			for (const part of parts) {
				const res = validateScalarConstraints(item, part, allowedValues);
				if (!res.right) {
					return res;
				}
			}
			return { right: true, error: '' };
		}

		if (!valueToCheck && !!item.required) {
			return { right: false, error: t('cannot be empty') };
		}

		return validateScalarConstraints(item, valueToCheck, allowedValues);
	}

	return { validateScalarConstraints, validateEnvValue };
}
