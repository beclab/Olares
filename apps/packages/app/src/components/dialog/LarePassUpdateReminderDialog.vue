<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Software update reminder')"
		:skip="false"
		:ok="t('Installing the Update')"
		:cancel="false"
		size="medium"
		:noRouteDismiss="true"
		@onSubmit="onUpdate"
	>
		<div class="dialog-content q-py-sm">
			<q-item class="q-pa-none" bordered>
				<q-item-section avatar>
					<div
						class="bg-light-blue-default row items-center justify-center"
						style="width: 40px; height: 40px; border-radius: 20px"
					>
						<q-icon
							name="sym_r_download_2"
							class="text-white"
							style="font-size: 20px"
						/>
					</div>
				</q-item-section>
				<q-item-section>
					<q-item-label class="text-body2 full-width text-ink-1">
						{{ t('Download the update now?') }}
					</q-item-label>
					<q-item-label class="text-ink-2 full-width text-body2">
						{{
							t(
								'The latest version of {productName} is {lastVersion}. Your current version is {currentVersion}',
								{
									productName: 'LarePass',
									lastVersion: lastVersion,
									currentVersion: currentVersion
								}
							)
						}}
					</q-item-label>
				</q-item-section>
			</q-item>

			<terminus-check-box
				class="q-mt-lg"
				:model-value="autoUpdate"
				:label="t('Automatically download and install updates in the future')"
				@update:modelValue="autoUpdate = !autoUpdate"
			/>
		</div>
		<template v-slot:footerMore>
			<q-item
				clickable
				dense
				class="but-skip-update row justify-center items-center"
				@click="onSkip"
			>
				{{ t('Skip this version') }}
			</q-item>
		</template>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import TerminusCheckBox from '../common/TerminusCheckBox.vue';

import { ref } from 'vue';

defineProps({
	currentVersion: {
		type: String,
		required: true
	},
	lastVersion: {
		type: String,
		required: true
	}
});

const { t } = useI18n();

const CustomRef = ref();

const autoUpdate = ref(false);

const onUpdate = () => {
	console.log('submit');
	CustomRef.value.onDialogOK({
		action: 'update',
		autoUpdate: autoUpdate.value
	});
};

const onSkip = () => {
	console.log('onSkip ===>');

	CustomRef.value.onDialogOK({
		action: 'skip',
		autoUpdate: autoUpdate.value
	});
};
</script>

<style lang="scss" scoped>
.but-skip-update {
	min-width: 100px;
	border-radius: 8px;
	font-weight: 500;
	border: 1px solid $btn-stroke;
	color: $ink-2;
	font-size: 16px;
	padding: 7px 8px;
	line-height: 24px;
}
</style>
