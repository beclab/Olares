<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('app.stop')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		@onSubmit="onOK"
	>
		<div class="column">
			<div class="text-ink-2 text-body3">
				{{
					t('Are you sure you want to stop the shared app', {
						app: appName
					})
				}}
			</div>

			<bt-check-box
				:label="t('Also stop the shared server （affects all users)')"
				:link-label="t('Learn more')"
				link="https://docs.olares.com/manual/olares/market/market.html#free-up-resources-from-unused-apps"
				check-img="market/check_box.svg"
				uncheck-img="market/uncheck_box.svg"
				class="q-mt-md"
				:model-value="allRef"
				@update:model-value="onUpdate"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import BtCheckBox from '../rss/BtCheckBox.vue';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

const { t } = useI18n();
const customRef = ref();

const props = defineProps({
	appName: {
		type: String,
		required: true
	},
	all: {
		type: Boolean,
		required: true,
		default: true
	}
});

const allRef = ref(props.all);

const onUpdate = (status: boolean) => {
	allRef.value = status;
};

const onOK = () => {
	customRef.value.onDialogOK({
		all: allRef.value
	});
};
</script>

<style scoped lang="scss"></style>
