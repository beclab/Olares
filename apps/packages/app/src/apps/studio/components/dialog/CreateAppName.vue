<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('home_create')"
		:cancel="t('btn_cancel')"
		:ok="t('btn_confirm')"
		:okLoading="loading ? $t('loading') : false"
		size="medium"
		@onSubmit="submit"
	>
		<div class="form-item row">
			<div class="form-item-key text-subtitle2 text-ink-1">
				{{ t('home_appname') }} <span class="text-red-default">*</span>
			</div>
			<div class="form-item-value q-mb-lg">
				<q-input
					ref="appNameRef"
					dense
					borderless
					no-error-icon
					v-model="appTitle"
					class="form-item-input"
					input-class="text-ink-2"
					:placeholder="ruleConfig.appNameRules.placeholder"
					:rules="ruleConfig.appNameRules.rules"
				>
				</q-input>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { ruleConfig } from './../../types/config';
import { useDockerStore } from './../../stores/docker';
import { useDevelopingApps } from '../../stores/app';
import { useMenuStore } from '../../stores/menu';

const { t } = useI18n();
const router = useRouter();

const dockerStore = useDockerStore();
const appStores = useDevelopingApps();
const menuStore = useMenuStore();

const CustomRef = ref();
const appNameRef = ref();
const loading = ref(false);
const appTitle = ref();

const submit = async () => {
	console.log('click submit');
	appNameRef.value.validate();
	if (appNameRef.value.hasError) return;

	try {
		loading.value = true;
		await dockerStore.create_name(appTitle.value);
		await appStores.getApps();
		await menuStore.updateApplications();
		dockerStore.appStatus = undefined;
		const appName = await getAppName(appTitle.value);
		router.push({ path: '/app/' + appName });
		menuStore.currentItem = '/app/' + appName;
		loading.value = false;
		CustomRef.value.onDialogOK();
	} catch (error) {
		loading.value = false;
	}
};

async function getAppName(title: string) {
	const current_app = appStores.apps.find((item) => item.title == title);

	return current_app.appName;
}
</script>

<style lang="scss" scoped>
.form-item {
	.form-item-key {
		width: 100px;
		height: 40px;
		line-height: 40px;
		text-align: center;
	}
	.form-item-value {
		flex: 1;
	}
}
</style>
