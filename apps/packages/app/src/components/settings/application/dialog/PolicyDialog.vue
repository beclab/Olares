<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="
			t('action_sub_policies', {
				action: editMode ? t('edit') : t('add')
			})
		"
		:skip="false"
		:ok="t('submit')"
		:cancel="t('cancel')"
		size="medium"
		:platform="deviceStore.platform"
		@onSubmit="onOKClick"
		:okDisabled="!canSubmit"
	>
		<div class="policy-content column justify-start">
			<bt-form v-model:can-submit="canSubmit">
				<error-message-tip
					style="width: 100%; max-width: 100%"
					:is-error="!uriRef"
					:width-separator="false"
				>
					<terminus-edit
						v-model="uriRef"
						:label="t('effected_urls')"
						:show-password-img="false"
						style="width: 100%"
						:is-error="!uriRef"
					/>
				</error-message-tip>

				<bt-form-border-select
					class="q-mt-md"
					v-model="factorMode"
					:title="t('second_factor_model')"
					:options="factorModelOptions()"
					:border="true"
				/>

				<bt-form-border-item
					v-if="factorMode === FACTOR_MODEL.Two"
					class="q-mt-md"
					:title="t('one_time')"
					:item-title="t('one_time')"
					sideClasses="bt-switch-margin-right"
				>
					<bt-switch truthy-track-color="blue-default" v-model="oneTimeMode" />
				</bt-form-border-item>

				<error-message-tip
					style="width: 100%; max-width: 100%"
					:is-error="validDuration < 0 || validDuration > 100"
					:width-separator="false"
					:error-message="t('errors.please_enter_a_valid_number')"
				>
					<bt-form-border-item
						v-if="factorMode === FACTOR_MODEL.Two"
						class="q-mt-md"
						:title="t('valid_duration')"
						:item-title="validDuration + ' s'"
					>
						<bt-time-picker
							v-model="validDuration"
							unit=" s"
							:input-disabled="true"
							:number-show="false"
						/>
					</bt-form-border-item>
				</error-message-tip>
				<!-- </q-list> -->
			</bt-form>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import BtForm from '../../base/BtForm.vue';
import { ref, onMounted, PropType } from 'vue';
import {
	EntrancePolicy,
	FACTOR_MODEL,
	factorModelOptions
} from 'src/constant/index';
import BtTimePicker from '../../base/BtTimePicker.vue';
import ErrorMessageTip from '../../base/ErrorMessageTip.vue';
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import BtFormBorderItem from '../../base/BTFormBorderItem.vue';
import BtFormBorderSelect from '../../base/BTFormBorderSelect.vue';

const props = defineProps({
	editMode: {
		type: Boolean,
		default: false
	},
	policy: {
		type: Object as PropType<EntrancePolicy>,
		require: true
	}
});

const { t } = useI18n();

const uriRef = ref();
const factorMode = ref();
const oneTimeMode = ref();
const validDuration = ref();

const canSubmit = ref(props.editMode);

onMounted(() => {
	if (props.policy) {
		uriRef.value = props.policy.uri;
		factorMode.value = props.policy.policy;
		oneTimeMode.value = props.policy.one_time;
		validDuration.value = props.policy.valid_duration;
	}
});
const CustomRef = ref();
const deviceStore = useDeviceStore();

async function onOKClick() {
	if (props.policy) {
		const data: EntrancePolicy = {
			uri: uriRef.value,
			policy: factorMode.value,
			one_time: oneTimeMode.value,
			valid_duration: validDuration.value
		};
		CustomRef.value.onDialogOK(data);
	}
}
</script>
<style lang="scss" scoped>
.policy-dialog {
	max-width: 400px;
	width: 400px;
	height: auto;
	padding: 0;
	position: relative;

	.policy-dialog-header {
		width: 100%;
		height: 32px;
		background: $background-3;
		padding-left: 12px;
		padding-right: 12px;
	}

	.policy-content {
		width: 100%;
		// height: calc(100% - 32px);
		padding: 20px;
	}
}
</style>
