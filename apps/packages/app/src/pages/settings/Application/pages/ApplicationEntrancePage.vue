<template>
	<page-title-component :show-back="true" :title="entrance_name" />

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list
			first
			:label="t('Access policies')"
			v-if="adminStore.isAdmin || !isDemo"
		>
			<bt-form-item :title="t('auth_level')">
				<bt-select
					:model-value="policyStore.authorizationLevel"
					:options="authLevelOptions()"
					@update:modelValue="
						(value) => {
							policyStore.authorizationLevel = value;
							policyStore.factorMode = factorModeOptions[0].value;
						}
					"
				/>
			</bt-form-item>
			<bt-form-item :title="t('second_factor_model')" :margin-top="false">
				<bt-select
					v-model="policyStore.factorMode"
					:options="factorModeOptions"
				/>
			</bt-form-item>

			<bt-form-item
				v-if="policyStore.factorMode === FACTOR_MODEL.Two"
				:title="t('one_time')"
			>
				<bt-switch
					truthy-track-color="blue-default"
					v-model="policyStore.oneTimeMode"
				/>
			</bt-form-item>

			<error-message-tip
				:is-error="
					policyStore.validDuration < 0 || policyStore.validDuration > 100
				"
				:error-message="t('errors.please_enter_a_valid_number')"
				:width-separator="false"
			>
				<bt-form-item
					v-if="policyStore.factorMode === FACTOR_MODEL.Two"
					:title="t('valid_duration')"
					:width-separator="false"
				>
					<bt-time-picker
						v-model="policyStore.validDuration"
						unit=" s"
						:input-disabled="true"
					/>
				</bt-form-item>
			</error-message-tip>

			<bt-form-item
				:title="t('Manage sub policies')"
				:margin-top="false"
				:width-separator="false"
				@click="goToPoliciesPage"
			>
				<div class="row justify-end items-center">
					<div class="text-body1 text-ink-1">
						{{ policyStore.policiesCount }}
					</div>
					<q-icon
						class="q-ml-xs"
						name="sym_r_chevron_right"
						color="ink-2"
						size="20px"
					/>
				</div>
			</bt-form-item>
		</bt-list>

		<div v-if="policyStore.resultCode != 3" class="row justify-end">
			<q-btn
				dense
				class="confirm-btn submit-btn-margin q-px-md"
				:disable="policyStore.isLoading || policyStore.resultCode === 3"
				:label="t('submit')"
				@click="onSubmit"
			/>
		</div>

		<application-domain :name="application.name" :entrance="entrance_name" />
	</bt-scroll-area>
</template>

<script setup lang="ts">
import BtList from 'src/components/settings/base/BtList.vue';
import BtSelect from 'src/components/settings/base/BtSelect.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BtTimePicker from 'src/components/settings/base/BtTimePicker.vue';
import ErrorMessageTip from 'src/components/settings/base/ErrorMessageTip.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ApplicationDomain from 'src/pages/settings/Application/pages/ApplicationDomain.vue';
import { useEntrancePolicyStore } from 'src/stores/settings/entrancePolicy';
import { useApplicationStore } from 'src/stores/settings/application';
import { useAdminStore } from 'src/stores/settings/admin';
import { useRoute, useRouter } from 'vue-router';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
	authLevelOptions,
	FACTOR_MODEL,
	factorModelOptions
} from 'src/constant';

const adminStore = useAdminStore();
const policyStore = useEntrancePolicyStore();

const factorModeOptions = computed(() => {
	return factorModelOptions(policyStore.authorizationLevel);
});

const isDemo = computed(() => {
	return !!process.env.DEMO;
});

const { t } = useI18n();

const applicationStore = useApplicationStore();
const Route = useRoute();
const router = useRouter();

const application = ref(
	applicationStore.getApplicationById(Route.params.name as string)
);

const application_name = ref(Route.params.name as string);
const entrance_name = Route.params.entrance as string;

onMounted(async () => {
	await policyStore.init(application_name.value, entrance_name);
});

function goToPoliciesPage() {
	router.push(
		'/application/domain/' +
			application.value?.name +
			'/' +
			entrance_name +
			'/policies'
	);
}

async function onSubmit() {
	await policyStore.submitAll(t);
	const updatedEntrance =
		applicationStore.entrances[application_name.value]?.[entrance_name];
	if (updatedEntrance && application.value) {
		const entrances = application.value.entrances.map((item) =>
			item.name === entrance_name
				? { ...item, authLevel: updatedEntrance.authLevel }
				: item
		);
		applicationStore.updateOneApplication({
			...application.value,
			entrances
		});
	}
}
</script>

<style scoped lang="scss">
.submit-btn-margin {
	margin-top: 20px;
}

.policies-count {
	font-size: 14px;
	color: var(--ink-2);
	font-weight: 500;
}
</style>
