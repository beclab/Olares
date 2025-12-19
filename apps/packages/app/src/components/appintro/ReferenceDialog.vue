<template>
	<q-dialog class="card-dialog-pc" ref="dialogRef">
		<q-card class="card-container-pc column no-shadow">
			<base-dialog-bar
				:label="t('detail.reference_app_not_installed')"
				@close="onDialogCancel"
			/>
			<div class="q-pa-lg dialog-scroll">
				<div class="text-ink-2 text-body2">
					{{ t('detail.need_reference_app_to_use') }}
				</div>
				<template :key="app" v-for="(app, index) in references">
					<recommend-app-card
						:app-name="app"
						:source-id="settingStore.marketSourceId"
						:is-last-line="index === references.length - 1"
					/>
				</template>

				<div class="full-width row justify-end q-mt-lg">
					<q-btn
						class="bg-blue-default btn-ok text-subtitle1 text-white"
						:label="t('base.ok')"
						@click="onDialogOK"
					/>
				</div>
			</div>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import BaseDialogBar from '../../components/base/BaseDialogBar.vue';
import RecommendAppCard from '../appcard/RecommendAppCard.vue';
import { useSettingStore } from '../../stores/market/setting';
import { useCenterStore } from '../../stores/market/center';
import { useDialogPluginComponent } from 'quasar';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	app: {
		type: String,
		required: true
	}
});

const { onDialogOK, onDialogCancel, dialogRef } = useDialogPluginComponent();
const references = ref<string[]>([]);
const centerStore = useCenterStore();
const settingStore = useSettingStore();
const { t } = useI18n();

onMounted(() => {
	const fullInfo = centerStore.getAppFullInfo(
		props.app,
		settingStore.marketSourceId
	);
	if (fullInfo) {
		references.value =
			fullInfo.app_info?.app_entry?.options?.appScope?.appRef ?? [];
	}
});
</script>

<style scoped lang="scss">
.card-dialog-pc {
	.card-container-pc {
		border-radius: 12px;
		width: 400px;

		.dialog-scroll {
			width: 400px;
			max-height: 396px !important;

			.btn-ok {
				width: 100px;
				height: 40px;
			}
		}
	}
}
</style>
