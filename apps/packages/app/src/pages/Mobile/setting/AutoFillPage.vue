<template>
	<div class="auto-fill-root">
		<terminus-title-bar
			:title="$q.platform.is.android ? t('autofill') : t('password_autofill')"
		/>
		<div v-if="$q.platform.is.android">
			<q-list class="q-mt-lg">
				<q-item v-if="frameworkSupport">
					<terminus-settings-module-item :title="t('autofill_service')">
						<template v-slot:side>
							<bt-switch
								truthy-track-color="blue-default"
								v-model="autofillFramework"
								@update:model-value="changeAutofillFramework"
							/>
						</template>
						<template v-slot:content>
							<div class="q-ma-md text-body3 text-ink-1">
								{{ t('autofill_service_desc') }}
							</div>
						</template>
					</terminus-settings-module-item>
				</q-item>

				<q-item>
					<terminus-settings-module-item :title="t('use_accessibility')">
						<template v-slot:side>
							<bt-switch
								truthy-track-color="blue-default"
								v-model="accessibilityAutofill"
								@update:model-value="changeAccessibilityAutofill"
							/>
						</template>
						<template v-slot:content>
							<div class="q-ma-md text-body3 text-ink-1">
								{{ t('use_accessibility_desc') }}
							</div>
						</template>
					</terminus-settings-module-item>
				</q-item>

				<q-item>
					<terminus-settings-module-item :title="t('use_draw_over')">
						<template v-slot:side>
							<bt-switch
								truthy-track-color="blue-default"
								v-model="drawOver"
								@update:model-value="changeOverlaySupport"
							/>
						</template>
						<template v-slot:content>
							<div class="q-ma-md text-body3 text-ink-1">
								{{ t('use_draw_over_desc') }}
							</div>
						</template>
					</terminus-settings-module-item>
				</q-item>
			</q-list>
		</div>
		<div v-else-if="$q.platform.is.ios">
			<q-list class="q-mt-lg">
				<q-item>
					<terminus-settings-module-item
						:title="t('get_instant_access_to_your_passwords')"
					>
						<template v-slot:content>
							<div class="q-pb-md">
								<div class="q-ma-md text-body3 text-ink-1">
									{{ t('autofill_instructions') }}<br /><br />
									1. {{ t('go_to_the_ios_settings') }}<br />
									2. {{ t('tap_passwords') }} <br />
									3. {{ t('tap_autofill_passwords') }} <br />
									4. {{ t('turn_on_autoFill') }}<br />
									5. {{ t('select_larepass') }}
								</div>
								<div class="row items-center justify-center">
									<img
										src="../../../assets/setting/autofill_ios_show.png"
										height="210"
										style="margin-top: 24px"
									/>
								</div>
							</div>
						</template>
					</terminus-settings-module-item>
				</q-item>
			</q-list>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import AndroidPlugins from 'src/platform/interface/capacitor/android/androidPlugins';
import { busOn, busOff } from '../../../utils/bus';
import { useQuasar } from 'quasar';
import TerminusDialog from '../../Items/dialog/TerminusDialog.vue';
import { useI18n } from 'vue-i18n';
import TerminusSettingsModuleItem from '../../../components/common/TerminusSettingsModuleItem.vue';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';

const $q = useQuasar();
const frameworkSupport = ref(false);
const autofillFramework = ref(false);
const { t } = useI18n();

const changeAutofillFramework = () => {
	if (autofillFramework.value) {
		AndroidPlugins.AutofillFramework.openSettingPage();
	} else {
		AndroidPlugins.AutofillFramework.closeService();
	}
};

const accessibilityAutofill = ref(false);

const changeAccessibilityAutofill = () => {
	if (accessibilityAutofill.value) {
		$q.dialog({
			component: TerminusDialog,
			componentProps: {
				title: t('accessibility_service_disclosure'),
				content: t('accessibility_service_disclosure_desc'),
				leftText: t('decline'),
				rightText: t('accept')
			}
		})
			.onOk(() => {
				AndroidPlugins.Accessibility.openAccessibilitySettingPage();
			})
			.onCancel(async () => {
				const data2 =
					await AndroidPlugins.Accessibility.isAccessibilityEnable();
				accessibilityAutofill.value = data2.isEnable;
			});
		return;
	}
	AndroidPlugins.Accessibility.openAccessibilitySettingPage();
};

const drawOver = ref(false);

const changeOverlaySupport = () => {
	if (drawOver.value && !accessibilityAutofill.value) {
		notifyFailed(t('open_accessibility_first'));
		drawOver.value = false;
		return;
	}
	AndroidPlugins.Accessibility.openOverlaySettingPage();
};

const callback = async () => {
	if (!$q.platform.is.android) {
		return;
	}
	const data = await AndroidPlugins.AutofillFramework.isEnable();
	autofillFramework.value = data.isEnable;
	const data2 = await AndroidPlugins.Accessibility.isAccessibilityEnable();
	accessibilityAutofill.value = data2.isEnable;
	const data3 = await AndroidPlugins.Accessibility.isOverlayEnable();
	drawOver.value = data3.isEnable;
};

onMounted(async () => {
	if (!$q.platform.is.android) {
		return;
	}
	const support = await AndroidPlugins.AutofillFramework.isSupport();
	frameworkSupport.value = support.isSupport;
	await callback();
	busOn('appStateChange', callback);
});

onUnmounted(() => {
	if (!$q.platform.is.android) {
		return;
	}
	busOff('appStateChange', callback);
});
</script>

<style scoped lang="scss">
.auto-fill-root {
	width: 100%;
	height: 100%;
}
</style>
