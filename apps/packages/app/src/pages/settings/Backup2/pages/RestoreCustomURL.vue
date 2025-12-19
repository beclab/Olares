<template>
	<page-title-component
		:show-back="true"
		:title="t('add_restore_from_custom_url')"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-form>
			<bt-list first :label="t('basic_backup_info')">
				<error-message-tip
					:is-error="!urlRule"
					:error-message="t('errors.naming_is_not_compliant')"
					:with-popup="true"
				>
					<bt-form-item :title="t('backup_url')" :width-separator="false">
						<bt-edit-view
							style="width: 200px"
							v-model="backupUrl"
							:right="true"
							:placeholder="t('backup_url')"
						/>
					</bt-form-item>
				</error-message-tip>

				<error-message-tip :is-error="password.length < 4">
					<bt-form-item :width-separator="false">
						<template v-slot:title>
							<div class="column">
								<div>{{ t('restore_password') }}</div>
								<div class="row" v-if="!deviceStore.isMobile">
									<div
										style="width: 16px; height: 16px"
										class="row items-center justify-center"
									>
										<q-icon
											:name="
												password.length >= 4 ? 'sym_r_check' : 'sym_r_clear'
											"
											class="text-ink-3"
											size="16px"
										/>
									</div>
									<div class="text-body3 text-ink-3">
										{{ t('must_have_at_least_4_characters') }}
									</div>
								</div>
							</div>
						</template>

						<bt-edit-view
							v-model="password"
							:is-password="true"
							style="width: 200px"
							:right="true"
							:placeholder="t('please_enter_a_password')"
						/>
					</bt-form-item>
					<template v-slot:reminder v-if="deviceStore.isMobile">
						<div
							class="row bg-background-3 items-center q-mb-sm q-px-xs"
							style="width: 100%; height: 24px; border-radius: 4px"
						>
							<div
								style="width: 16px; height: 16px"
								class="row items-center justify-center"
							>
								<q-icon
									:name="password.length >= 4 ? 'sym_r_check' : 'sym_r_clear'"
									class="text-ink-3"
									size="16px"
								/>
							</div>
							<div class="q-ml-sm text-overline-m text-ink-3">
								{{ t('must_have_at_least_4_characters') }}
							</div>
						</div>
					</template>
				</error-message-tip>

				<bt-form-item :title="t('Restore location')" :width-separator="false">
					<transfet-select-to
						class="q-mt-xs"
						@setSelectPath="setSelectPath"
						:origins="backupOriginsRef"
						:master-node="true"
					>
						<template v-slot:default>
							<div class="row justify-end items-center">
								<div
									class="text-body1 text-ink-1"
									style="width: calc(100% - 40px)"
									v-if="restorePathRef"
								>
									{{ restorePathRef }}
								</div>
								<q-btn
									class="text-ink-2 btn-size-sm btn-no-text btn-no-border"
									icon="sym_r_edit_square"
									outline
									no-caps
								/>
							</div>
						</template>
					</transfet-select-to>
				</bt-form-item>
			</bt-list>
		</bt-form>
		<div class="row justify-end">
			<q-btn
				dense
				flat
				:disable="!canSubmit"
				class="confirm-btn q-mt-lg q-mt-lg"
				:label="t('start_restore')"
				@click="onSubmit"
				:loading="isLoading"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import ErrorMessageTip from '../../../../components/settings/base/ErrorMessageTip.vue';
import PageTitleComponent from '../../../../components/settings/PageTitleComponent.vue';
import BtFormItem from '../../../../components/settings/base/BtFormItem.vue';
import TransfetSelectTo from '../../../Electron/Transfer/TransfetSelectTo.vue';
import BtEditView from '../../../../components/settings/base/BtEditView.vue';
import { backupOriginsRef, BackupResourcesType } from '../../../../constant';
import BtForm from '../../../../components/settings/base/BtForm.vue';
import { useDeviceStore } from '../../../../stores/settings/device';
import { useBackupStore } from '../../../../stores/settings/backup';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { FilePath } from '../../../../stores/files';
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import BtList from '../../../../components/settings/base/BtList.vue';

const { t } = useI18n();
const router = useRouter();
const restorePathRef = ref();
const password = ref('');
const backupUrl = ref('');
const isLoading = ref(false);
const backupStore = useBackupStore();
const deviceStore = useDeviceStore();

const setSelectPath = (fileSavePath: FilePath) => {
	restorePathRef.value = fileSavePath.decodePath;
};

const urlRule = computed(() => {
	return backupUrl.value.length > 0;
});

const onSubmit = () => {
	isLoading.value = true;
	backupStore
		.restoreCustomUrl(backupUrl.value, password.value, restorePathRef.value, '')
		.then(() => {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('success')
			});
			router.push({ path: '/backup' });
		})
		.catch((e) => {
			console.error(e);
		})
		.finally(() => {
			isLoading.value = false;
		});
};

const canSubmit = computed(() => {
	return (
		!!backupUrl.value &&
		!!password.value &&
		!!restorePathRef.value &&
		urlRule.value
	);
});
</script>

<style scoped lang="scss"></style>
