import { defineStore } from 'pinia';
import {
	APP_STATUS,
	AppEntry,
	CLUSTER_TYPE,
	ClusterApp,
	DEPENDENCIES_TYPE,
	Dependency,
	TerminusResource,
	User,
	UserResource
} from 'src/constant/constants';
import { intersection } from 'src/utils/utils';
import { CFG_TYPE } from 'src/constant/config';
import { Range, Version } from 'src/utils/version';
import {
	ErrorGroup,
	ErrorGroupHandler,
	errorsToStructuredText,
	findLevel1Error,
	findLevel2Error,
	sortErrorGroups
} from 'src/constant/errorGroupHandler';
import { getSystemStatus } from 'src/api/market/centerApi';
import { getComputeResources } from 'src/api/settings/accelerator';
import type { ComputeResourceNode } from 'src/constant/compute';
import { useAppStore } from 'src/stores/market/appStore';
import { OLARES_ROLE } from 'src/constant';
import { OPERATION_TYPE } from 'src/services/market/appOperation/operationTypes';
import {
	APP_OP_REJECT_CODE,
	buildAppOpContext,
	evaluateAppOp
} from 'src/services/market/appOperation/appOperationPolicy';
import { bus, BUS_EVENT } from 'src/utils/bus';
import { CacheRequest } from 'src/stores/market/CacheRequest';
import { unwrapSystemStatusPayload } from 'src/utils/market/appResource';

export type PreCheckState = {
	userInfo: User | null;
	initialized: boolean;
	osVersion: string | null;
	userResource: UserResource | null;
	systemResource: TerminusResource | null;
	computeNodes: ComputeResourceNode[];
};

function applySystemStatusPayload(
	store: {
		userResource: UserResource | null;
		systemResource: TerminusResource | null;
		osVersion: string | null;
		userInfo: User | null;
		initialized: boolean;
	},
	payload: any
) {
	const data = unwrapSystemStatusPayload(payload);
	if (!data) return;

	store.userResource = data.cur_user_resource ?? null;
	store.systemResource = data.cluster_resource ?? null;
	store.osVersion = data.terminus_version?.version ?? null;
	store.userInfo = data.user_info ?? null;
	store.initialized = true;
}

export const usePreCheckStore = defineStore('preCheck', {
	state: () => {
		return {
			userInfo: null,
			initialized: false,
			osVersion: null,
			userResource: null,
			systemResource: null,
			computeNodes: []
		} as PreCheckState;
	},
	actions: {
		initSystemRequest() {
			return new CacheRequest('cache_system_status', getSystemStatus, {
				onData: (data) => {
					applySystemStatusPayload(this, data);
				}
			});
		},
		async refreshSystemStatus(force = false): Promise<void> {
			const request = this.initSystemRequest();
			if (force) {
				await request.forceRefresh();
				return;
			}
			await request.get();
		},
		/**
		 * Pull per-node compute capabilities (mode/gpuType) so the detail card
		 * can tell which accelerator modes the cluster actually exposes. A
		 * failure leaves `computeNodes` empty, which makes the evaluator fall
		 * back to the arch + aggregate-GPU heuristic.
		 */
		async refreshComputeResources(): Promise<void> {
			try {
				const nodes = await getComputeResources();
				this.computeNodes = Array.isArray(nodes) ? nodes : [];
			} catch (e) {
				console.error('refreshComputeResources failed', e);
				this.computeNodes = [];
			}
		},
		hasAdminPermissions(): boolean {
			return (
				!!this.userInfo &&
				(this.userInfo.role === OLARES_ROLE.OWNER ||
					this.userInfo.role === OLARES_ROLE.ADMIN)
			);
		},
		frontendPreflight(
			appEntry: AppEntry,
			policyCtx?: {
				op: OPERATION_TYPE;
				appName: string;
				sourceId: string;
			}
		): ErrorGroup[] {
			console.log('frontendPreflight');
			const errorGroups: ErrorGroup[] = [];

			if (policyCtx) {
				const decision = evaluateAppOp(buildAppOpContext(policyCtx));
				if (!decision.allowed && decision.key) {
					this.appPushError(errorGroups, APP_OP_REJECT_CODE[decision.key]);
					bus.emit(
						BUS_EVENT.APP_BACKEND_ERROR,
						errorsToStructuredText(errorGroups)
					);
					return errorGroups;
				}
			}

			const role = this._userRolePreflight(errorGroups, appEntry);
			const terminus = this._terminusOSVersionPreflight(errorGroups, appEntry);
			const systemResource = this._systemResourcePreflight(
				errorGroups,
				appEntry
			);
			const appDependencies = this._appDependenciesPreflight(
				errorGroups,
				appEntry
			);
			const appConflicts = this._appConflictsPreflight(errorGroups, appEntry);
			if (
				role &&
				terminus &&
				systemResource &&
				appDependencies &&
				appConflicts
			) {
				console.log('ok');
				return errorGroups;
			} else {
				if (!errorGroups || errorGroups.length === 0) {
					this.appPushError(errorGroups, ErrorGroupHandler.G001);
				}
				const group = sortErrorGroups(errorGroups);

				bus.emit(BUS_EVENT.APP_BACKEND_ERROR, errorsToStructuredText(group));

				return group;
			}
		},
		_userRolePreflight(errorGroups: ErrorGroup[], app: AppEntry): boolean {
			if (!this.userInfo || !this.userInfo.role) {
				this.appPushError(errorGroups, ErrorGroupHandler.G003);
				return false;
			}

			if (
				app.onlyAdmin &&
				this.userInfo.role !== OLARES_ROLE.OWNER &&
				this.userInfo.role !== OLARES_ROLE.ADMIN
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G004);
				return false;
			}
			if (
				app.cfgType === CFG_TYPE.MIDDLEWARE &&
				this.userInfo.role !== OLARES_ROLE.OWNER &&
				this.userInfo.role !== OLARES_ROLE.ADMIN
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G005);
				return false;
			}
			return true;
		},

		_terminusOSVersionPreflight(
			errorGroups: ErrorGroup[],
			app: AppEntry
		): boolean {
			if (!this.osVersion) {
				this.appPushError(errorGroups, ErrorGroupHandler.G007);
				return false;
			}
			if (
				app.options &&
				app.options.dependencies &&
				app.options.dependencies.length > 0
			) {
				for (let i = 0; i < app.options.dependencies.length; i++) {
					const appInfo = app.options.dependencies[i];
					if (appInfo.type === DEPENDENCIES_TYPE.system) {
						if (appInfo.version == '>=0.5.0-0') {
							this.appPushError(errorGroups, ErrorGroupHandler.G008);
							return false;
						}

						try {
							const range = new Range(appInfo.version);
							if (range) {
								const result = range.satisfies(new Version(this.osVersion));
								if (result) {
									return true;
								}
							}
						} catch (e) {
							console.log(e);
							this.appPushError(errorGroups, ErrorGroupHandler.G008);
							return false;
						}
					}
				}
			}
			this.appPushError(errorGroups, ErrorGroupHandler.G008);
			return false;
		},

		_appDependenciesPreflight(
			errorGroups: ErrorGroup[],
			app: AppEntry
		): boolean {
			let isOK = true;
			if (
				app.options &&
				app.options.dependencies &&
				app.options.dependencies.length > 0
			) {
				const appStore = useAppStore();
				const allApps = appStore.installStatusList.map((item) => {
					return {
						name: item.status.name,
						state: item.status.state,
						type: CLUSTER_TYPE.app,
						version: ''
					};
				});
				const middlewares: ClusterApp[] = [];
				if (
					this.systemResource &&
					this.systemResource.apps &&
					this.systemResource.apps.length > 0
				) {
					this.systemResource.apps.forEach((app: ClusterApp) => {
						if (app.type === CLUSTER_TYPE.app) {
							allApps.push(app);
						} else if (app.type === CLUSTER_TYPE.middleware) {
							middlewares.push(app);
						}
					});
				}

				for (let i = 0; i < app.options.dependencies.length; i++) {
					const dependency = app.options.dependencies[i];
					if (dependency.type === DEPENDENCIES_TYPE.middleware) {
						if (
							!middlewares.find(
								(item) =>
									item.name === dependency.name &&
									item.state === APP_STATUS.RUNNING
							)
						) {
							this.appPushError(errorGroups, ErrorGroupHandler.G012_SG001, {
								name: dependency.name,
								version: dependency.version
							});
							isOK = false;
						}
					}

					if (dependency.type === DEPENDENCIES_TYPE.application) {
						if (dependency.mandatory) {
							if (
								!allApps.find(
									(item) =>
										item.name === dependency.name &&
										item.state === APP_STATUS.RUNNING
								)
							) {
								if (dependency.name === app.name) {
									this.appPushError(errorGroups, ErrorGroupHandler.G021_SG001, {
										name: dependency.name,
										version: dependency.version
									});
									isOK = false;
								} else {
									this.appPushError(errorGroups, ErrorGroupHandler.G012_SG002, {
										name: dependency.name,
										version: dependency.version
									});
									isOK = false;
								}
							}
						}
					}
				}
			}
			return isOK;
		},

		_systemResourcePreflight(
			errorGroups: ErrorGroup[],
			app: AppEntry
		): boolean {
			// System resource sufficiency (CPU/memory/disk/GPU) is no longer enforced
			// on the frontend for any app (including accelerator apps) — the backend is
			// authoritative. Only architecture compatibility still blocks here.
			if (!app.supportArch || app.supportArch.length === 0) {
				this.appPushError(errorGroups, ErrorGroupHandler.G018_SG001);
				return false;
			}

			const nodes = this.systemResource?.nodes;
			if (!nodes || nodes.length === 0) {
				// No cluster node data → cannot determine architecture; do not block.
				return true;
			}

			if (intersection(nodes, app.supportArch).length === 0) {
				this.appPushError(errorGroups, ErrorGroupHandler.G018);
				return false;
			}
			return true;
		},

		_appConflictsPreflight(errorGroups: ErrorGroup[], app: AppEntry): boolean {
			let isOK = true;
			if (
				app.options &&
				app.options.conflicts &&
				app.options.conflicts.length > 0
			) {
				const appStore = useAppStore();
				const appApps = appStore.installStatusList.map(
					(item) => item.status.name
				);

				if (
					this.systemResource &&
					this.systemResource.apps &&
					this.systemResource.apps.length > 0
				) {
					this.systemResource.apps.forEach((app: ClusterApp) => {
						if (app.type === CLUSTER_TYPE.app) {
							appApps.push(app.name);
						}
					});
				}

				const conflictingApps = app.options.conflicts.filter(
					(conflict) =>
						conflict.type === DEPENDENCIES_TYPE.application &&
						appApps.includes(conflict.name)
				);

				if (conflictingApps.length > 0) {
					this.appPushError(errorGroups, ErrorGroupHandler.G019, {
						count: conflictingApps.length
					});
					conflictingApps.forEach((conflict) => {
						this.appPushError(errorGroups, ErrorGroupHandler.G019_SG001, {
							name: conflict.name
						});
					});
					isOK = false;
				}
			}

			const appStore2 = useAppStore();
			const appApps2 = appStore2.installStatusList.map(
				(item) => item.status.name
			);
			if (appApps2.includes(app.name)) {
				this.appPushError(errorGroups, ErrorGroupHandler.G020);
				isOK = false;
			}
			return isOK;
		},

		appPushError(
			errorGroups: ErrorGroup[],
			code: ErrorGroupHandler,
			variables?: Record<string, string | number>
		) {
			if (code.indexOf('-') === -1) {
				const level1Error = findLevel1Error(code, variables);
				if (level1Error && !errorGroups.find((e) => e.code === code)) {
					errorGroups.push(level1Error);
				}
				return;
			}

			const level2Error = findLevel2Error(code, variables);
			if (level2Error) {
				const parentCode = level2Error.parentCode;
				let parentError: ErrorGroup | undefined | null = errorGroups.find(
					(e) => e.code === parentCode
				);

				if (!parentError) {
					parentError = findLevel1Error(parentCode, variables);
					if (parentError) {
						errorGroups.push(parentError);
					}
				}

				if (parentError) {
					parentError.subGroups.push(level2Error);
				}
			}
		},

		getUnInstallDependencies(app: AppEntry): Dependency[] {
			const dependencies: Dependency[] = [];
			if (
				app.options &&
				app.options.dependencies &&
				app.options.dependencies.length > 0
			) {
				const appStore = useAppStore();
				const allApps = appStore.installStatusList.map((item) => {
					return {
						name: item.status.name,
						state: item.status.state,
						type: CLUSTER_TYPE.app,
						version: ''
					};
				});
				const middlewares: ClusterApp[] = [];
				if (
					this.systemResource &&
					this.systemResource.apps &&
					this.systemResource.apps.length > 0
				) {
					this.systemResource.apps.forEach((app: ClusterApp) => {
						if (app.type === CLUSTER_TYPE.app) {
							allApps.push(app);
						} else if (app.type === CLUSTER_TYPE.middleware) {
							middlewares.push(app);
						}
					});
				}

				for (let i = 0; i < app.options.dependencies.length; i++) {
					const dependency = app.options.dependencies[i];
					if (dependency.type === DEPENDENCIES_TYPE.middleware) {
						if (
							!middlewares.find(
								(item) =>
									item.name === dependency.name &&
									item.state === APP_STATUS.RUNNING
							)
						) {
							dependencies.push(dependency);
						}
					}

					if (dependency.type === DEPENDENCIES_TYPE.application) {
						if (dependency.mandatory) {
							if (
								!allApps.find(
									(item) =>
										item.name === dependency.name &&
										item.state === APP_STATUS.RUNNING
								)
							) {
								dependencies.push(dependency);
							}
						}
					}
				}
			}
			return dependencies;
		},
		hasUninstallDependencies(app: AppEntry): boolean {
			const dependencies = this.getUnInstallDependencies(app);
			return dependencies.length > 0;
		}
	}
});
