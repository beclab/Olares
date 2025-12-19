<template>
	<q-item>
		<q-item-section>
			<div
				class="row items-center justify-between text-ink-1"
				:class="[titleClass]"
			>
				{{ t('settings.themes.title') }}
			</div>
			<terminus-check-box
				class="q-mt-md"
				:model-value="isThemeAuto"
				:label="t('settings.themes.follow_system_theme')"
				@update:modelValue="updateTheme(ThemeDefinedMode.AUTO)"
			/>
			<div class="q-mt-md text-ink-3 text-body3">
				{{
					t(
						"After being selected, LarePass will follow the device's system settings to switch theme modes"
					)
				}}
			</div>
			<div class="q-mt-md row items-center justify-between theme-select">
				<div
					class="theme-item-common"
					:class="isThemeLight ? 'theme-item-select' : ''"
					@click="updateTheme(ThemeDefinedMode.LIGHT)"
				>
					<q-img src="../../assets/setting/theme-light.svg" class="image" />
					<div class="content">
						<q-radio
							v-model="deviceStore.theme"
							:val="ThemeDefinedMode.LIGHT"
							:label="t('settings.themes.light')"
							color="yellow-default"
							@update:model-value="updateTheme"
						/>
					</div>
				</div>
				<div
					class="theme-item-common"
					:class="isThemeDark ? 'theme-item-select' : ''"
					@click="updateTheme(ThemeDefinedMode.DARK)"
				>
					<q-img src="../../assets/setting/theme-dark.svg" class="image" />
					<div class="content">
						<q-radio
							v-model="deviceStore.theme"
							:val="ThemeDefinedMode.DARK"
							:label="t('settings.themes.dark')"
							color="yellow-default"
							@update:model-value="updateTheme"
						/>
					</div>
				</div>
			</div>
		</q-item-section>
	</q-item>
</template>

<script setup lang="ts">
import { ThemeDefinedMode } from '@bytetrade/ui';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDeviceStore } from '../../stores/device';
import TerminusCheckBox from './TerminusCheckBox.vue';

interface Props {
	titleClass?: string;
}

withDefaults(defineProps<Props>(), {
	titleClass: 'text-h6'
});

const { t } = useI18n();

const deviceStore = useDeviceStore();

const isThemeAuto = computed(function () {
	return deviceStore.theme == ThemeDefinedMode.AUTO;
});

const isThemeDark = computed(function () {
	return deviceStore.theme == ThemeDefinedMode.DARK;
});

const isThemeLight = computed(function () {
	return deviceStore.theme == ThemeDefinedMode.LIGHT;
});

const updateTheme = (theme: ThemeDefinedMode) => {
	deviceStore.setTheme(theme);
};
</script>

<style scoped lang="scss">
.checkbox-content {
	width: 100%;
	height: 30px;
	.checkbox-common {
		width: 16px;
		height: 16px;
		margin-right: 10px;
		border-radius: 4px;
	}

	.checkbox-unselect {
		border: 1px solid $separator-2;
	}

	.checkbox-selected-green {
		background: $positive;
	}
}
.theme-select {
	width: 440px;
	height: 144px;

	.theme-item-common {
		height: 144px;
		width: 210px;
		border: 1px solid $separator;
		border-radius: 12px;
		overflow: hidden;
		.image {
			width: 100%;
			height: 100px;
		}
		.content {
			width: 100%;
			height: 44px;
		}
	}

	.theme-item-select {
		border: 1px solid $yellow-default;
	}
}
</style>
