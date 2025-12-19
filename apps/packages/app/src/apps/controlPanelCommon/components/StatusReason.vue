<template>
	<div>
		<q-icon name="sym_r_info" class="status-color" :class="status" />
		<span style="margin-left: 4px" class="status-color" :class="status">
			{{ t(reason) }}
		</span>
		<q-tooltip anchor="top middle" self="bottom middle">
			<div class="status-tooltip-wrapper">
				<div className="tooltip-title">{{ t('STATUS_INFORMATION') }}</div>
				<div v-for="cd in data.status.conditions" :key="cd">
					<div class="conditions row items-center">
						<q-icon
							v-if="isSuccess(type, cd)"
							color="positive"
							name="check_circle"
							class="status-icon"
						/>
						<q-icon v-else name="cancel" color="negative" class="status-icon" />
						<span>
							{{
								t(`${type.toUpperCase()}_CONDITION_${cd.type.toUpperCase()}`, {
									defaultValue: cd.type
								})
							}}
						</span>
					</div>
					<div v-if="cd.status">
						{{
							t('STATUS_VALUE', {
								value: cd.status === 'True' ? 'True' : 'False'
							})
						}}
					</div>
					<div v-if="cd.reason">
						{{
							t('REASON_VALUE', {
								value: cd.reason
							})
						}}
					</div>
					<div v-if="cd.message">
						{{ t('MESSAGE_VALUE', { value: cd.message }) }}
					</div>
				</div>
			</div>
		</q-tooltip>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface Props {
	data: any;
	status: string;
	reason: string;
	type?: string;
}

withDefaults(defineProps<Props>(), {});

const isSuccess = (type: string, condition: any) => {
	const conditionType = condition.type;
	const conditionStatus = condition.status;

	if (type === 'volume') {
		return conditionStatus === 'True';
	}

	return conditionType === 'ReplicaFailure'
		? conditionStatus === 'False'
		: conditionStatus === 'True';
};
</script>

<style lang="scss" scoped>
.status-tooltip-wrapper {
	color: #fff;
	.conditions {
		font-weight: 600;
		margin-top: 12px;
		.status-icon {
			background: #fff;
			border-radius: 50%;
			margin-right: 4px;
		}
	}
}

.status-color {
	&.draft,
	&.pending-review,
	&.to-be-reviewed,
	&.creating {
		color: $primary;
	}

	&.success,
	&.succeeded,
	&.successful,
	&.ready,
	&.running,
	&.jobrunning,
	&.active,
	&.normal,
	&.bound,
	&.available,
	&.healthy,
	&.ok,
	&.working,
	&.passed,
	&.published {
		color: $status-success;
	}

	&.creating,
	&.jobrunning {
		&.flicker {
			animation: flicker 1.5s linear infinite;
		}
	}

	&.updating,
	&.upgrading,
	&.syncing,
	&.building,
	&.notrunning,
	&.warning,
	&.waiting,
	&.terminating,
	&.released,
	&.unfinished,
	&.firing,
	&.major,
	&.deleting,
	&.in-review {
		color: $status-warning;
	}

	&.deleting,
	&.deleted,
	&.error,
	&.lost,
	&.unschedulable,
	&.failed,
	&.critical,
	&.rejected {
		color: $status-failed;
	}

	&.stopped,
	&.disabled,
	&.pending,
	&.unknown,
	&.draining,
	&.terminated,
	&.completed,
	&.minor,
	&.paused,
	&.inactive,
	&.suspended {
		color: $status-pending;
	}
}
</style>
