<template>
	<page-title-component :show-back="true" :title="t('Manage sub policies')" />

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<app-menu-feature
			image="settings/imgs/root/env.svg"
			:label="t('Fine grained access control')"
			:description="t('Set access policies for specific URLs')"
			:button="t('Add sub policy')"
			@on-button-click="addPolicies"
		/>

		<bt-list v-if="policyStore.sub_policies.length > 0">
			<bt-form-item
				v-for="(policy, index) in policyStore.sub_policies"
				:key="index"
				:width-separator="index !== policyStore.sub_policies.length - 1"
			>
				<template v-slot:title>
					<div class="column justify-start">
						<div class="text-ink-1 text-body1">{{ policy.uri }}</div>
						<div class="row justify-start">
							<div class="text-info text-caption policy-label bg-blue-alpha">
								{{ policy.policy }}
							</div>

							<div
								v-if="policy.valid_duration > 0"
								class="text-info text-caption policy-label bg-blue-alpha q-ml-sm"
							>
								{{ policy.valid_duration }}s
							</div>
						</div>
					</div>
				</template>
				<div class="row justify-end">
					<q-icon
						class="cursor-pointer"
						name="sym_r_edit_square"
						color="ink-2"
						size="24px"
						@click="editPolicy(policy, index)"
					/>
					<q-icon
						class="q-ml-md cursor-pointer"
						name="sym_r_delete"
						color="ink-2"
						size="26px"
						@click="deletePolicy(policy)"
					/>
				</div>
			</bt-form-item>
		</bt-list>

		<div v-if="policyStore.resultCode != 3" class="row justify-end">
			<q-btn
				dense
				class="confirm-btn submit-btn-margin q-px-md"
				:disable="policyStore.isLoading || !policyStore.hasPoliciesChanges"
				:label="t('submit')"
				@click="onSubmit"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import BtList from 'src/components/settings/base/BtList.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import AppMenuFeature from 'src/components/settings/AppMenuFeature.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import PolicyDialog from 'src/components/settings/application/dialog/PolicyDialog.vue';
import { useEntrancePolicyStore } from 'src/stores/settings/entrancePolicy';
import { EntrancePolicy } from 'src/constant';
import { useRoute } from 'vue-router';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { onMounted } from 'vue';

const { t } = useI18n();
const Route = useRoute();
const $q = useQuasar();
const policyStore = useEntrancePolicyStore();

const application_name = Route.params.name as string;
const entrance_name = Route.params.entrance as string;

onMounted(async () => {
	if (
		policyStore.applicationName !== application_name ||
		policyStore.entranceName !== entrance_name
	) {
		await policyStore.init(application_name, entrance_name);
	}
});

async function onSubmit() {
	await policyStore.submitPolicies(t);
}

const defaultPolicy: EntrancePolicy = {
	one_time: true,
	policy: 'one_factor',
	uri: '',
	valid_duration: 0
};

const addPolicies = () => {
	const tempArray = policyStore.sub_policies;
	$q.dialog({
		component: PolicyDialog,
		componentProps: {
			policy: defaultPolicy
		}
	}).onOk((data: EntrancePolicy) => {
		tempArray?.push(data);
		policyStore.sub_policies = tempArray;
	});
};

const editPolicy = (policy: EntrancePolicy, index: number) => {
	const tempArray = policyStore.sub_policies;
	$q.dialog({
		component: PolicyDialog,
		componentProps: {
			policy,
			editMode: true
		}
	}).onOk((data: EntrancePolicy) => {
		tempArray.splice(index, 1, data);
		policyStore.sub_policies = tempArray;
	});
};

const deletePolicy = (policy: EntrancePolicy) => {
	const tempArray = policyStore.sub_policies?.filter(
		(item) => item.uri !== policy.uri
	);
	policyStore.sub_policies = tempArray;
};
</script>

<style scoped lang="scss">
.submit-btn-margin {
	margin-top: 20px;
}

.policy-label {
	border-radius: 4px;
	padding: 2px 4px;
}
</style>
