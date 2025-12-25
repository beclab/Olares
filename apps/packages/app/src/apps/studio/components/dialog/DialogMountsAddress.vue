<template>
	<q-dialog class="card-dialog" v-model="show" ref="dialogRef">
		<q-card class="card-container" flat>
			<terminus-dialog-bar
				:label="t('docker.add_storage_volume')"
				icon=""
				titAlign="text-left"
				@close="onDialogCancel"
			/>

			<card-form-item
				:name="t('docker.host_path')"
				:required="true"
				:tip="t('docker.host_path')"
				:keyWidth="130"
			>
				<q-input
					ref="envHostPathRef"
					dense
					borderless
					no-error-icon
					v-model="hostSuffix"
					class="form-item-input"
					input-class="text-ink-2"
					:placeholder="ruleConfig.hostPath.placeholder"
					:rules="ruleConfig.hostPath.rules"
					style="padding-left: 0"
				>
					<template v-slot:prepend>
						<q-select
							dense
							options-dense
							map-options
							borderless
							v-model="hostPrefix"
							:options="hostOptions"
							dropdown-icon="sym_r_keyboard_arrow_down"
							class="host-path"
							popup-content-class="options_selected_content"
						>
							<template v-slot:option="{ itemProps, opt }">
								<q-item
									dense
									v-bind="itemProps"
									style="border-radius: 4px; overflow: hidden"
								>
									<q-item-section class="text-ink-2">
										<q-item-label>{{ opt }}</q-item-label>
									</q-item-section>
								</q-item>
							</template>
						</q-select>
					</template>
				</q-input>
			</card-form-item>

			<card-form-item
				:name="t('docker.container_path')"
				:required="true"
				:tip="t('docker.container_path')"
				:keyWidth="130"
			>
				<q-input
					ref="envContainerPathRef"
					dense
					borderless
					no-error-icon
					v-model.trim="config.value"
					class="form-item-input"
					input-class="text-ink-2"
					:placeholder="ruleConfig.containerPath.placeholder"
					:rules="ruleConfig.containerPath.rules"
				>
				</q-input>
			</card-form-item>

			<TerminusFormFooter
				:ok-text="t('submit')"
				:cancel-text="t('cancel')"
				@submit="submit"
				@cancel="onDialogCancel"
			/>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { ref, reactive, PropType } from 'vue';
import { useDialogPluginComponent } from 'quasar';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { ruleConfig } from './../../types/config';
import { VariableType } from '@apps/studio/src/types/core';
import { useI18n } from 'vue-i18n';
import CardFormItem from './../common/CardFormItem.vue';
import TerminusDialogBar from '../common/TerminusDialogBar.vue';
import TerminusFormFooter from '../common/TerminusFormFooter.vue';

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const props = defineProps({
	defaultMounts: {
		type: Object as PropType<VariableType>,
		required: false
	},
	mountsConfig: {
		type: Object as PropType<VariableType[]>,
		required: false
	}
});

const { t } = useI18n();

const config = reactive({
	key: props.defaultMounts ? props.defaultMounts.key : '',
	value: props.defaultMounts ? props.defaultMounts.value : '',
	id: props.defaultMounts ? props.defaultMounts.id : undefined
});

const show = ref(true);
const hostOptions = ['/app/data', '/app/cache', '/Home'];
const hostPrefix = ref('/app/data');
const hostSuffix = ref();

const envHostPathRef = ref();
const envContainerPathRef = ref();

if (props.defaultMounts) {
	hostPrefix.value =
		hostOptions.find((option) => props.defaultMounts?.key.startsWith(option)) ||
		'';
	hostSuffix.value = props.defaultMounts?.key.slice(hostPrefix.value.length);
}

const verifyDuplicate = (): boolean => {
	if (
		props.mountsConfig?.find(
			(item) => item.key === config.key && item.id !== config.id
		)
	) {
		BtNotify.show({
			type: NotifyDefinedType.WARNING,
			message: t('host_path_repeat')
		});

		return false;
	}

	if (
		props.mountsConfig?.find(
			(item) => item.value === config.value && item.id !== config.id
		)
	) {
		BtNotify.show({
			type: NotifyDefinedType.WARNING,
			message: t('container_path_repeat')
		});
		return false;
	}

	return true;
};

const submit = () => {
	envHostPathRef.value.validate();
	if (envHostPathRef.value.hasError) return;
	envContainerPathRef.value.validate();
	if (envContainerPathRef.value.hasError) return;

	config.key = hostPrefix.value + (hostSuffix.value || '');

	if (!verifyDuplicate()) return;

	onDialogOK(config);
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-container {
		width: 560px;
		border-radius: 12px;
		padding: 20px;
		.dialog-content {
			flex: 1;
			padding: 20px 0 32px;
		}
	}
}

.host-path {
	width: 124px;
	padding: 0 10px 0 10px;
	color: $ink-2;
	background-color: $background-6;
	border-radius: 8px;
}
</style>
../../types/config../common/CardFormItem.vue
