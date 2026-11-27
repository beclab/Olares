<template>
	<q-expansion-item
		dense
		style="padding: 0; border-radius: 10px"
		class="overflow-hidden"
		expandSeparator
		headerClass="header_extension text-white"
		expand-icon="sym_r_keyboard_arrow_down"
	>
		<template v-slot:header>
			<q-item-section style="width: 100%">
				<div class="row items-center" style="width: 100%">
					<div
						class="device-item-title row justify-between items-center"
						style="width: 100%"
					>
						<div class="row justify-start" style="width: 100%">
							<div
								class="bg-background-3 row items-center justify-center device-title-image"
							>
								<q-img class="icon" src="settings/person/headscale.png" />
								<div
									class="status_base"
									:style="`background:${timeStatus(
										new Date(device?.lastSeen || 0)
									)}`"
								></div>
							</div>
							<div
								class="column justify-center q-ml-sm"
								style="width: calc(100% - 50px)"
							>
								<div class="row items-center" style="width: 100%">
									<div
										class="text-subtitle1 text-ink-1"
										@click.stop="titleAction"
										style="
											max-width: calc(100% - 50px);
											overflow: hidden;
											text-overflow: ellipsis;
											white-space: nowrap;
										"
									>
										{{ device?.givenName }}
									</div>
									<q-icon
										@click.stop="titleAction"
										class="q-ml-sm text-ink-1"
										name="sym_r_edit_square"
										size="16px"
									/>
								</div>
								<div v-if="showTag" class="tags-bg row q-gutter-xs">
									<div
										class="bg-background-3 text-overline row items-center justify-center tags-add"
										@click.stop="addTag"
									>
										<q-icon name="sym_r_add" class="icon" size="12px" />
										<div>{{ t('add_tag') }}</div>
									</div>
									<div
										v-for="(i, index) in tags"
										@click.stop="deleteTag(index)"
										:key="index"
										class="bg-background-3 text-overline row items-center justiy-center tags-item"
									>
										<div>
											{{ i }}
										</div>
										<q-icon name="sym_r_close" class="icon" size="12px" />
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>
			</q-item-section>

			<q-item-section
				side
				v-if="!device.availableRoutes || device.availableRoutes.length == 0"
			>
				<q-icon name="sym_r_delete" size="20px" @click.stop="deleteDevice" />
			</q-item-section>
		</template>
		<q-card class="device-item-content bg-background-2">
			<bt-separator />
			<q-list :class="deviceStore.isMobile ? 'q-pb-md' : 'q-py-sm'">
				<AdaptiveLayout>
					<template v-slot:pc>
						<q-item dense style="padding: 0">
							<q-item-section class="text-body2 title">
								{{ t('device_last_seen') }}
							</q-item-section>
							<q-item-section class="text-body2 detail">
								{{
									getLocalTime(new Date(device.lastSeen).getTime()).format(
										'YYYY-MM-DD HH:mm:ss'
									)
								}}
							</q-item-section>
						</q-item>
					</template>
					<template v-slot:mobile>
						<q-item style="padding: 0" class="q-mt-md">
							<q-item-section>
								<q-item-label class="text-body3-m text-ink-3">
									{{ t('device_last_seen') }}
								</q-item-label>
								<q-item-label class="text-body3-m text-ink-1">
									{{
										getLocalTime(new Date(device.lastSeen).getTime()).format(
											'YYYY-MM-DD HH:mm:ss'
										)
									}}
								</q-item-label>
							</q-item-section>
						</q-item>
					</template>
				</AdaptiveLayout>
				<AdaptiveLayout>
					<template v-slot:pc>
						<q-item dense style="padding: 0">
							<q-item-section
								class="text-body2 title"
								:top="device.ipAddresses.length > 1"
							>
								{{ t('ip_addresses') }}
							</q-item-section>
							<q-item-section class="text-body2 detail column">
								<div
									v-for="(address, index) in device.ipAddresses"
									:key="address"
									class="row items-center"
									:style="index == 0 ? 'margin-bottom: 8px' : ''"
								>
									{{ address }}
								</div>
							</q-item-section>
						</q-item>
					</template>
					<template v-slot:mobile>
						<q-item style="padding: 0">
							<q-item-section>
								<q-item-label
									class="text-body3-m text-ink-3 q-mt-lg"
									:top="device.ipAddresses.length > 1"
								>
									{{ t('ip_addresses') }}
								</q-item-label>
								<q-item-label class="text-body3-m text-ink-1 column">
									<div
										v-for="(address, index) in device.ipAddresses"
										:key="address"
										class="row"
										:style="
											index + 1 < device.ipAddresses.length
												? 'margin-bottom: 8px'
												: ''
										"
									>
										{{ address }}
									</div>
								</q-item-label>
							</q-item-section>
						</q-item>
					</template>
				</AdaptiveLayout>

				<AdaptiveLayout>
					<template v-slot:pc>
						<q-item dense style="padding: 0">
							<q-item-section class="text-body2 title">
								{{ t('device_name') }}
							</q-item-section>
							<q-item-section class="text-body2 column detail">
								{{ device.name }}
							</q-item-section>
						</q-item>
					</template>
					<template v-slot:mobile>
						<q-item style="padding: 0">
							<q-item-section>
								<q-item-label class="text-body3-m text-ink-3 q-mt-lg">
									{{ t('device_name') }}
								</q-item-label>
								<q-item-label class="text-body3-m text-ink-1">
									{{ device.name }}
								</q-item-label>
							</q-item-section>
						</q-item>
					</template>
				</AdaptiveLayout>

				<AdaptiveLayout>
					<template v-slot:pc>
						<q-item dense style="padding: 0">
							<q-item-section
								class="text-body2 title"
								:top="device.availableRoutes.length > 1"
							>
								{{ t('device_routes') }}
							</q-item-section>
							<q-item-section class="text-body2 column detail">
								<div
									v-for="(route, index) in device.availableRoutes"
									:key="route"
									class="row items-center"
									:class="
										index + 1 < device.availableRoutes.length ? 'q-mb-xs' : ''
									"
								>
									<div>{{ route }}</div>
									<div
										class="row items-center justify-center text-caption route-toggle"
										:class="routeToggleClass(route)"
										@click="modifyDeviceRoutesAction(route)"
									>
										<q-spinner-dots
											v-if="isRouteApproving(route)"
											size="14px"
											:color="
												device.approvedRoutes.includes(route)
													? 'positive'
													: 'warning'
											"
										/>
										<template v-else>
											<div
												class="node q-mr-sm"
												:class="
													device.approvedRoutes.includes(route)
														? 'node-enable'
														: 'node-pendding'
												"
											/>
											{{
												device.approvedRoutes.includes(route)
													? t('enable')
													: t('pending')
											}}
										</template>
									</div>
								</div>
							</q-item-section>
						</q-item>
					</template>
					<template v-slot:mobile>
						<q-item
							style="padding: 0"
							v-if="!deviceStore.isMobile || device.availableRoutes.length > 0"
						>
							<q-item-section>
								<q-item-label
									class="text-body3-m text-ink-3 q-mt-lg"
									:top="device.availableRoutes.length > 1"
								>
									{{ t('device_routes') }}
								</q-item-label>
								<q-item-label class="text-body3-m text-ink-1">
									<div
										v-for="(route, index) in device.availableRoutes"
										:key="route"
										class="row items-center"
										:class="
											index + 1 < device.availableRoutes.length ? 'q-mb-xs' : ''
										"
									>
										<div>{{ route }}</div>
										<div
											class="row items-center justify-center text-caption route-toggle"
											:class="routeToggleClass(route)"
											@click="modifyDeviceRoutesAction(route)"
										>
											<q-spinner-dots
												v-if="isRouteApproving(route)"
												size="14px"
												:color="
													device.approvedRoutes.includes(route)
														? 'positive'
														: 'warning'
												"
											/>
											<template v-else>
												<div
													class="node q-mr-sm"
													:class="
														device.approvedRoutes.includes(route)
															? 'node-enable'
															: 'node-pendding'
													"
												></div>
												{{
													device.approvedRoutes.includes(route)
														? t('enable')
														: t('pending')
												}}
											</template>
										</div>
									</div>
								</q-item-label>
							</q-item-section>
						</q-item>
					</template>
				</AdaptiveLayout>
			</q-list>
		</q-card>
	</q-expansion-item>
</template>
<script lang="ts" setup>
import EditHeadscaleInfoDialog from 'src/components/settings/person/EditHeadscaleInfoDialog.vue';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import BtSeparator from 'src/components/settings/base/BtSeparator.vue';
import { useHeadScaleStore } from 'src/stores/settings/headscale';
import { useDeviceStore } from 'src/stores/settings/device';
import { getLocalTime } from 'src/utils/settings/utils';
import { HeadScaleDevice } from '@bytetrade/core';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { PropType, ref } from 'vue';

const props = defineProps({
	device: {
		type: Object as PropType<HeadScaleDevice>,
		required: true
	}
});

const headScaleStore = useHeadScaleStore();
const deviceStore = useDeviceStore();
const showTag = ref(false);

const $q = useQuasar();
const { t } = useI18n();

const approvingRoute = ref<string | null>(null);

function isRouteApproving(route: string) {
	return approvingRoute.value === route;
}

function routeToggleClass(route: string) {
	const base = props.device.approvedRoutes.includes(route)
		? 'route-bg-enable'
		: 'route-bg-pendding';
	if (approvingRoute.value === null) {
		return [base, 'route-toggle--interactive'];
	}
	if (approvingRoute.value === route) {
		return [base, 'route-toggle--loading'];
	}
	return [base, 'route-toggle--idle'];
}

async function modifyDeviceRoutesAction(route: string) {
	if (approvingRoute.value !== null) {
		return;
	}

	const currentList = props.device?.approvedRoutes || [];

	const hasRoute = currentList.includes(route);
	let newList: string[];

	if (hasRoute) {
		newList = currentList.filter((item) => item !== route);
	} else {
		newList = [...currentList, route];
	}

	approvingRoute.value = route;
	try {
		await headScaleStore.approveRoutes(props.device.id, newList);
	} finally {
		approvingRoute.value = null;
	}
}

function timeStatus(date: Date) {
	let currentTime = new Date();
	let timeDifference = Math.round(
		(currentTime.getTime() - date.getTime()) / 1000
	);
	if (timeDifference < 3600) {
		return '#29CC5F';
	} else if (timeDifference < 86400) {
		return '#FEBE01';
	} else {
		return '#ADADAD';
	}
}

const tags = ref(props.device.forcedTags);

async function deleteTag(index: number) {
	const newTags = tags.value;
	newTags.splice(index, 1);
	await headScaleStore.updateTags(props.device.id, newTags);
}

const addTag = () => {
	$q.dialog({
		component: EditHeadscaleInfoDialog,
		componentProps: {
			oldValue: '',
			title: t('add_tag'),
			editTitle: t('tag')
		}
	}).onOk(async (tag: string) => {
		if (tag.length > 0) {
			if (tags.value.includes(tag)) {
				return;
			}

			const newTags = tags.value;
			newTags.push(tag);
			await headScaleStore.updateTags(props.device.id, newTags);
		}
	});
};

const deleteDevice = () => {
	if (!props.device) {
		return;
	}

	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('delete_device'),
			message: t('delete_device_message', {
				device: props.device.givenName
			}),
			useCancel: true,
			confirmText: t('delete'),
			cancelText: t('cancel')
		}
	}).onOk(async () => {
		await headScaleStore.removeDevice(props.device.id);
		headScaleStore.getDevices();
	});
};

const titleAction = () => {
	if (!props.device) {
		return;
	}
	$q.dialog({
		component: EditHeadscaleInfoDialog,
		componentProps: {
			oldValue: props.device.givenName,
			title: t('edit'),
			editTitle: t('host_name')
		}
	}).onOk(async (name: string) => {
		if (name == props.device!.givenName) {
			return;
		}
		headScaleStore.renameDevice(props.device!.id, name);
	});
};
</script>

<style lang="scss" scoped>
.device-item-title {
	height: auto;

	.common-background {
		background-color: $background-3;
	}

	.device-title-image {
		@extend .common-background;
		height: 32px;
		width: 32px;
		border-radius: 16px;
		position: relative;

		.icon {
			width: 26px;
			height: 8px;
		}

		.status_base {
			position: absolute;
			right: 1px;
			bottom: 1px;
			width: 12px;
			height: 12px;
			border-radius: 6px;
		}
	}

	.tags-bg {
		width: 100%;
		// background-color: red;
		margin-top: 4px;
		height: auto;

		.tags-add {
			@extend .common-background;
			border-radius: 4px;
			padding: 0px 8px;
			height: 20px;
			// margin: 0;

			.icon {
				width: 12px;
				height: 12px;
				margin-right: 8px;
			}

			color: $ink-2;
		}

		.tags-item {
			@extend .common-background;
			padding: 0px 8px;
			border-radius: 4px;
			height: 20px;
			// margin: 0;
			// margin-top: 0px;

			color: $ink-1;

			.icon {
				width: 12px;
				height: 12px;
				margin-left: 8px;
				color: $ink-2;
			}

			/* 120% */
		}
	}
}

.device-item-content {
	padding: 0 20px;

	.title {
		color: $ink-2;
	}

	.detail {
		color: $ink-1;
	}

	.route-base {
		border-radius: 4px;
		height: 20px;
		margin-left: 12px;
		padding: 0 8px;

		.node {
			width: 8px;
			height: 8px;
			border-radius: 4px;
		}

		text-align: justify;
	}

	.route-bg-pendding {
		@extend .route-base;
		background: rgba($color: $warning, $alpha: 0.1);
		color: $warning;
	}

	.route-bg-enable {
		@extend .route-base;
		background: $green-1;
		color: $green;
	}

	.node-pendding {
		background: $warning;
	}

	.node-enable {
		background: $green;
	}

	.route-toggle {
		transition: opacity 0.2s ease, transform 0.2s ease;
		min-height: 20px;
	}

	.route-toggle--interactive {
		cursor: pointer;
	}

	.route-toggle--interactive:active {
		transform: scale(0.98);
	}

	.route-toggle--loading {
		cursor: wait;
		padding-left: 10px;
		padding-right: 10px;
		pointer-events: none;
	}

	.route-toggle--idle {
		cursor: not-allowed;
		opacity: 0.45;
		pointer-events: none;
	}
}
</style>
