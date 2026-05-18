<template>
	<div class="my-date-picker-container">
		<el-config-provider :locale="lang">
			<slot />
		</el-config-provider>
	</div>
</template>

<script lang="ts" setup>
import { ElConfigProvider } from 'element-plus';
import { computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import zhCn from 'element-plus/dist/locale/zh-cn.mjs';
import en from 'element-plus/dist/locale/en.mjs';
import { useQuasar } from 'quasar';
import 'element-plus/dist/index.css';
import 'element-plus/theme-chalk/dark/css-vars.css';

const $q = useQuasar();
const { locale } = useI18n();

const lang = computed(() =>
	locale.value.substring(0, 2) === 'zh' ? zhCn : en
);

watch(
	() => $q.dark.isActive,
	(isDark) => {
		if (isDark) {
			document.documentElement.classList.add('dark');
		} else {
			document.documentElement.classList.remove('dark');
		}
	},
	{ immediate: true }
);
</script>

<style scoped lang="scss">
.my-date-picker-container {
	::v-deep(.el-date-picker),
	::v-deep(.el-date-editor) {
		border-radius: 8px;
		width: 340px;
	}

	::v-deep(.el-date-picker.el-input__wrapper),
	::v-deep(.el-date-picker .el-input__wrapper),
	::v-deep(.el-date-editor.el-input__wrapper),
	::v-deep(.el-date-editor .el-input__wrapper) {
		box-shadow: 0 0 0 1px $input-stroke inset;
		background-color: $background-1 !important;
	}

	::v-deep(.el-date-picker .el-range-input),
	::v-deep(.el-date-editor .el-range-input) {
		color: $ink-2;
		font-size: 12px;
	}

	::v-deep(.el-date-picker .el-range-separator),
	::v-deep(.el-date-editor .el-range-separator) {
		color: $ink-2;
		font-size: 12px;
	}

	::v-deep(.el-date-picker .el-input__icon.el-range__icon),
	::v-deep(.el-date-editor .el-input__icon.el-range__icon) {
		color: $ink-3;
		font-size: 12px;
	}

	::v-deep(.el-date-picker .el-input__icon .el-range__close-icon),
	::v-deep(.el-date-editor .el-input__icon .el-range__close-icon) {
		color: $ink-3;
		font-size: 12px;
	}
}
</style>
