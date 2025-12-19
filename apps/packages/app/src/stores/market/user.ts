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
import { useCenterStore } from 'src/stores/market/center';
import { OLARES_ROLE } from 'src/constant';
import { bus, BUS_EVENT } from 'src/utils/bus';
import { CacheRequest } from 'src/stores/market/CacheRequest';

export type UserState = {
	user: User | null;
	initialized: boolean;
	osVersion: string | null;
	userResource: UserResource | null;
	systemResource: TerminusResource | null;
};

export const useUserStore = defineStore('userStore', {
	state: () => {
		return {
			user: null,
			initialized: false,
			osVersion: null,
			userResource: null,
			systemResource: null
		} as UserState;
	},

	actions: {
		init() {
			return new CacheRequest('cache_system_status', getSystemStatus, {
				onData: (data) => {
					this.userResource = data.cur_user_resource;
					this.systemResource = data.cluster_resource;
					this.osVersion = data.terminus_version.version;
					this.user = data.user_info;
					this.initialized = true;
				}
			});
		},
		hasAdminPermissions() {
			return (
				this.user &&
				(this.user.role === OLARES_ROLE.OWNER ||
					this.user.role === OLARES_ROLE.ADMIN)
			);
		},
		frontendPreflight(appEntry: AppEntry): ErrorGroup[] {
			console.log('frontendPreflight');
			const errorGroups: ErrorGroup[] = [];
			const role = this._userRolePreflight(errorGroups, appEntry);
			const terminus = this._terminusOSVersionPreflight(errorGroups, appEntry);
			const userResource = this._userResourcePreflight(errorGroups, appEntry);
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
				userResource &&
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
			if (!this.user || !this.user.role) {
				this.appPushError(errorGroups, ErrorGroupHandler.G003);
				return false;
			}

			if (
				app.onlyAdmin &&
				this.user.role !== OLARES_ROLE.OWNER &&
				this.user.role !== OLARES_ROLE.ADMIN
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G004);
				return false;
			}
			if (
				app.cfgType === CFG_TYPE.MIDDLEWARE &&
				this.user.role !== OLARES_ROLE.OWNER &&
				this.user.role !== OLARES_ROLE.ADMIN
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G005);
				return false;
			}
			if (
				app.options &&
				app.options.appScope &&
				app.options.appScope.clusterScoped &&
				this.user.role !== OLARES_ROLE.OWNER &&
				this.user.role !== OLARES_ROLE.ADMIN
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G006);
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
						//temp
						if (appInfo.version == '>=0.5.0-0') {
							// console.log('intercept by temporary version : >=0.5.0-0');
							this.appPushError(errorGroups, ErrorGroupHandler.G008);
							return false;
						}

						try {
							const range = new Range(appInfo.version);
							if (range) {
								const result = range.satisfies(new Version(this.osVersion));
								// console.log(
								// 	'version satisfies : ' +
								// 		result +
								// 		' version : ' +
								// 		this.osVersion +
								// 		' range : ' +
								// 		appInfo.version
								// );
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

		_userResourcePreflight(errorGroups: ErrorGroup[], app: AppEntry): boolean {
			if (
				!this.userResource ||
				!this.userResource.cpu ||
				!this.userResource.memory
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G009);
				return false;
			}

			let isOK = true;

			const availableCpu =
				this.userResource.cpu.total - this.userResource.cpu.usage;
			if (
				app.requiredCPU &&
				this.userResource.cpu.total &&
				Number(app.requiredCPU) > availableCpu
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G010);
				isOK = false;
			}

			const availableMemory =
				this.userResource.memory.total - this.userResource.memory.usage;
			if (
				app.requiredMemory &&
				this.userResource.memory.total &&
				Number(app.requiredMemory) > availableMemory
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G011);
				isOK = false;
			}
			return isOK;
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
				const centerStore = useCenterStore();
				const allApps = centerStore.installStatusList.map((item) => {
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
							//cluster app
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
								// temp dify dependency dify(for cluster in this.systemResource.apps)
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
			if (
				!this.systemResource ||
				!this.systemResource.metrics ||
				!this.systemResource.nodes
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G013);
				return false;
			}

			let isOK = true;

			const availableCpu =
				this.systemResource.metrics.cpu.total -
				this.systemResource.metrics.cpu.usage;
			if (
				app.requiredCPU &&
				this.systemResource.metrics.cpu.total &&
				Number(app.requiredCPU) > availableCpu
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G014);
				isOK = false;
			}

			const availableMemory =
				this.systemResource.metrics.memory.total -
				this.systemResource.metrics.memory.usage;
			if (
				app.requiredMemory &&
				this.systemResource.metrics.memory.total &&
				Number(app.requiredMemory) > availableMemory
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G015);
				isOK = false;
			}

			const availableDisk =
				this.systemResource.metrics.disk.total -
				this.systemResource.metrics.disk.usage;
			if (
				app.requiredDisk &&
				this.systemResource.metrics.disk.total &&
				Number(app.requiredDisk) > availableDisk
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G016);
				isOK = false;
			}

			const availableGpu = this.systemResource.metrics.gpu.total > 0;
			if (app.requiredGPU && Number(app.requiredGPU) && !availableGpu) {
				this.appPushError(errorGroups, ErrorGroupHandler.G017);
				isOK = false;
			}

			if (!app.supportArch || app.supportArch.length === 0) {
				this.appPushError(errorGroups, ErrorGroupHandler.G018_SG001);
				isOK = false;
			} else if (
				!this.systemResource.nodes ||
				this.systemResource.nodes.length === 0
			) {
				this.appPushError(errorGroups, ErrorGroupHandler.G018_SG002);
				isOK = false;
			} else {
				const intersectedArray = intersection(
					this.systemResource.nodes,
					app.supportArch
				);
				if (intersectedArray.length === 0) {
					this.appPushError(errorGroups, ErrorGroupHandler.G018);
					isOK = false;
				}
			}
			return isOK;
		},

		_appConflictsPreflight(errorGroups: ErrorGroup[], app: AppEntry): boolean {
			let isOK = true;
			if (
				app.options &&
				app.options.conflicts &&
				app.options.conflicts.length > 0
			) {
				const centerStore = useCenterStore();
				const appApps = centerStore.installStatusList.map(
					(item) => item.status.name
				);

				if (
					this.systemResource &&
					this.systemResource.apps &&
					this.systemResource.apps.length > 0
				) {
					this.systemResource.apps.forEach((app: ClusterApp) => {
						if (app.type === CLUSTER_TYPE.app) {
							//cluster app
							appApps.push(app.name);
						}
					});
				}

				for (let i = 0; i < app.options.conflicts.length; i++) {
					const conflict = app.options.conflicts[i];

					if (conflict.type === DEPENDENCIES_TYPE.application) {
						if (appApps.includes(conflict.name)) {
							this.appPushError(errorGroups, ErrorGroupHandler.G019_SG001, {
								name: conflict.name
							});
							isOK = false;
						}
					}
				}
			}

			const centerStore = useCenterStore();
			const appApps2 = centerStore.installStatusList.map(
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
				const centerStore = useCenterStore();
				const allApps = centerStore.installStatusList.map((item) => {
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
							//cluster app
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
