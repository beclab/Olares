<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('detail.dependency_not_installed')"
		@onSubmit="onDialogOK"
		:ok="t('base.ok')"
	>
		<div class="text-ink-2 text-body2">
			{{ t('detail.require_dependencies_for_full') }}
		</div>
		<recommend-app-card
			:key="dependency.name"
			v-for="(dependency, index) in dependencies"
			:source-id="sourceId"
			:app-name="dependency.name"
			:is-last-line="index === dependencies.length - 1"
		/>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import RecommendAppCard from 'src/components/appcard/RecommendAppCard.vue';
import { useCenterStore } from 'src/stores/market/center';
import { useUserStore } from 'src/stores/market/user';
import { Dependency } from 'src/constant/constants';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	appName: {
		type: String,
		required: true
	},
	sourceId: {
		type: String,
		required: true
	}
});

const userStore = useUserStore();
const centerStore = useCenterStore();
const dependencies = ref<Dependency[]>([]);
const onDialogOK = () => {
	customRef.value.onDialogOK();
};
const { t } = useI18n();
const customRef = ref();

onMounted(() => {
	const fullInfo = centerStore.getAppFullInfo(props.appName, props.sourceId);
	if (fullInfo) {
		dependencies.value = userStore.getUnInstallDependencies(
			fullInfo.app_info?.app_entry
		);
	}
});
</script>

<style scoped lang="scss"></style>
