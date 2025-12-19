<template>
	<div
		class="policies-card-root column justify-center q-py-lg"
		:style="deviceStore.isMobile ? '	min-height: 56px;' : '	min-height: 64px;'"
		:class="deviceStore.isMobile ? '' : ''"
	>
		<div
			class="row items-center justify-between item-margin-left item-margin-right text-body1 text-ink-1"
		>
			<span>{{ t('policies') }}</span>
			<bt-icon name="add" @click="addPolicies" />
		</div>
		<template v-for="(policy, index) in policies" :key="index">
			<div
				class="policies-content row items-center justify-between bg-background-3"
				v-if="policies.length > 0"
			>
				<div
					class="row items-center justify-start"
					style="width: calc(100% - 60px)"
				>
					<div class="policy-icon q-mr-md text-ink-2">
						<q-icon class="cursor-pointer" name="sym_r_article" size="16px" />
					</div>
					<div style="width: calc(100% - 48px)">
						<div class="row items-center" style="width: 100%">
							<div
								class="text-ink-1 text-subtitle3 q-mr-md"
								style="
									text-overflow: ellipsis;
									overflow: hidden;
									text-align: left;
									flex: 1;
								"
							>
								{{ policy.uri }}
							</div>
							<div class="text-ink-2 text-caption q-px-md q-py-xs policy-label">
								{{ policy.policy }}
							</div>
						</div>
						<!-- <div class="text-overline text-ink-2 q-mt-xs">
							{{ policy.valid_duration }}
						</div> -->
					</div>
				</div>

				<div class="row items-center justify-center text-ink-2">
					<q-icon
						class="cursor-pointer"
						name="sym_r_edit_square"
						size="16px"
						@click="editPolicy(policy, index)"
					/>
					<q-icon
						class="q-ml-md cursor-pointer"
						name="sym_r_delete"
						size="16px"
						@click="deletePolicy(policy)"
					/>
				</div>
			</div>
		</template>
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { useQuasar } from 'quasar';

import DialogPolicies from './dialog/PolicyDialog.vue';
import { EntrancePolicy } from 'src/constant/index';
import { useDeviceStore } from 'src/stores/settings/device';
import BtIcon from '../base/BtIcon.vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	policies: {
		type: Object as PropType<EntrancePolicy[]>,
		required: true
	},
	disabledRemove: {
		type: Boolean,
		required: false,
		default: false
	}
});

const $q = useQuasar();

const { t } = useI18n();

const deviceStore = useDeviceStore();

const emit = defineEmits(['update:policies']);

const defaultPolicy: EntrancePolicy = {
	one_time: true,
	policy: 'one_factor',
	uri: '',
	valid_duration: 0
};

const addPolicies = () => {
	const tempArray = props.policies;
	$q.dialog({
		component: DialogPolicies,
		componentProps: {
			policy: defaultPolicy
		}
	}).onOk((data: EntrancePolicy) => {
		tempArray?.push(data);
		emit('update:policies', tempArray);
	});
};

const editPolicy = (policy: EntrancePolicy, index: number) => {
	const tempArray = props.policies;
	$q.dialog({
		component: DialogPolicies,
		componentProps: {
			policy,
			editMode: true
		}
	}).onOk((data: EntrancePolicy) => {
		tempArray.splice(index, 1, data);
		emit('update:policies', tempArray);
	});
};

const deletePolicy = (policy: EntrancePolicy) => {
	const tempArray = props.policies?.filter((item) => item.uri !== policy.uri);
	emit('update:policies', tempArray);
};
</script>

<style lang="scss" scoped>
.policies-card-root {
	width: 100%;

	.policies-content {
		border-radius: 12px;
		margin-top: 12px;
		padding: 12px;
		margin: 12px 12px 0;

		.policy-icon {
			width: 32px;
			height: 32px;
			display: flex;
			align-items: center;
			justify-content: center;
			border-radius: 16px;
		}

		.policy-label {
			border: 1px solid $separator;
			border-radius: 20px;
		}
	}
}
</style>
