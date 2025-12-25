import { i18n } from 'src/boot/i18n';

interface SubErrorGroup {
	parentCode: string;
	code: string;
	title: string;
	variables?: Record<string, string | number>;
}

export interface ErrorGroup {
	code: ErrorGroupHandler;
	title: string;
	variables?: Record<string, string | number>;
	subGroups: SubErrorGroup[];
}

export enum ErrorGroupHandler {
	G001 = 'G001',
	G002 = 'G002',
	G003 = 'G003',
	G004 = 'G004',
	G005 = 'G005',
	G006 = 'G006',
	G007 = 'G007',
	G008 = 'G008',
	G009 = 'G009',
	G010 = 'G010',
	G011 = 'G011',
	G012 = 'G012',
	G013 = 'G013',
	G014 = 'G014',
	G015 = 'G015',
	G016 = 'G016',
	G017 = 'G017',
	G018 = 'G018',
	G019 = 'G019',
	G020 = 'G020',
	G021 = 'G021',
	G012_SG001 = 'G012-SG001',
	G012_SG002 = 'G012-SG002',
	G018_SG001 = 'G018-SG001',
	G018_SG002 = 'G018-SG002',
	G019_SG001 = 'G019-SG001',
	G021_SG001 = 'G021-SG001'
}

const level1Map = new Map<ErrorGroupHandler, string>([
	[ErrorGroupHandler.G001, 'error.unknown_error'],
	[ErrorGroupHandler.G002, 'error.app_info_get_failure'],
	[ErrorGroupHandler.G003, 'error.failed_get_user_role'],
	[ErrorGroupHandler.G004, 'error.only_be_installed_by_the_admin'],
	[ErrorGroupHandler.G005, 'error.not_admin_role_install_middleware'],
	[ErrorGroupHandler.G006, 'error.not_admin_role_install_cluster_app'],
	[ErrorGroupHandler.G007, 'error.failed_to_get_os_version'],
	[ErrorGroupHandler.G008, 'error.app_is_not_compatible_terminus_os'],
	[ErrorGroupHandler.G009, 'error.failed_to_get_user_resource'],
	[ErrorGroupHandler.G010, 'error.user_not_enough_cpu'],
	[ErrorGroupHandler.G011, 'error.user_not_enough_memory'],
	[ErrorGroupHandler.G012, 'error.need_to_install_dependent_app_first'],
	[ErrorGroupHandler.G013, 'error.failed_to_get_system_resource'],
	[ErrorGroupHandler.G014, 'error.terminus_not_enough_cpu'],
	[ErrorGroupHandler.G015, 'error.terminus_not_enough_memory'],
	[ErrorGroupHandler.G016, 'error.terminus_not_enough_disk'],
	[ErrorGroupHandler.G017, 'error.terminus_not_enough_gpu'],
	[ErrorGroupHandler.G018, 'error.cluster_not_support_platform'],
	[ErrorGroupHandler.G019, 'error.app_install_conflict'],
	[ErrorGroupHandler.G020, 'error.source_conflict'],
	[ErrorGroupHandler.G021, 'error.the_dependent_app_must_be_installed_first']
]);

const level2Map = new Map<ErrorGroupHandler, string>([
	[ErrorGroupHandler.G012_SG001, 'error.middleware_not_install_details'],
	[ErrorGroupHandler.G012_SG002, 'error.app_not_install_details'],
	[ErrorGroupHandler.G018_SG001, 'error.no_supportarch_specified_in_app_chart'],
	[
		ErrorGroupHandler.G018_SG002,
		'error.failed_to_get_your_olares_platform_architecture'
	],
	[ErrorGroupHandler.G019_SG001, 'error.app_install_conflict_details'],
	[ErrorGroupHandler.G021_SG001, 'error.app_not_install_details']
]);

export function findLevel1Error(
	code: string,
	variables?: Record<string, string | number>
): ErrorGroup | null {
	const level1Result = level1Map.get(code as ErrorGroupHandler);
	if (level1Result) {
		return {
			code: code as ErrorGroupHandler,
			title: level1Result,
			variables,
			subGroups: []
		};
	}
	return null;
}

export function findLevel2Error(
	code: ErrorGroupHandler,
	variables?: Record<string, string | number>
): SubErrorGroup | null {
	const level2Result = level2Map.get(code);
	if (level2Result) {
		const [parentCode, subCode] = code.split('-');
		return { parentCode, code: subCode, title: level2Result, variables };
	}
	return null;
}

export function sortErrorGroups(errorGroups: ErrorGroup[]): ErrorGroup[] {
	errorGroups.sort((a, b) => a.code.localeCompare(b.code));

	errorGroups.forEach((group) => {
		group.subGroups.sort((a, b) => a.code.localeCompare(b.code));
	});

	return errorGroups;
}

export function errorsToStructuredText(errorGroups: ErrorGroup[]): string {
	if (!errorGroups.length) {
		return '';
	}

	const lines: string[] = [];

	errorGroups.forEach((group) => {
		lines.push(`• ${i18n.global.t(group.title, group.variables || {})}`);

		group.subGroups.forEach((sub) => {
			lines.push(`• ${i18n.global.t(sub.title, sub.variables || {})}`);
		});
	});

	return lines.join('\n');
}
