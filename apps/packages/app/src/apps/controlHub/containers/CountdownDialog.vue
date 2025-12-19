<template>
	<q-dialog v-model="showDialog" persistent>
		<q-card style="min-width: 200px">
			<q-card-section>
				<p class="text-h6">{{ title }}</p>
				<p v-html="content"></p>
			</q-card-section>
			<q-card-actions align="right">
				<q-btn
					no-caps
					dense
					:label="t('REPLICAS_SCALE_NOTIFY_CANCEL')"
					@click="cancelHandler"
				/>
				<q-btn
					color="primary"
					no-caps
					dense
					:label="t('REPLICAS_SCALE_NOTIFY_CONFIRM', { seconds: countdown })"
					@click="confirm"
				/>
			</q-card-actions>
		</q-card>
	</q-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, toRefs, watch } from 'vue';
import { t } from '@apps/control-hub/src/boot/i18n';

interface DialogData {
	title: string;
	content: string;
	count: number;
	modelValue: boolean;
}

const props = withDefaults(defineProps<DialogData>(), {});
const emits = defineEmits(['cancel', 'confirm', 'update:modelValue']);

const showDialog = ref(false);
const countdown = ref(0);
let countdownInterval: NodeJS.Timeout | undefined = undefined;
const { title, content } = toRefs(props);

watch(
	() => props.modelValue,
	() => {
		showDialog.value = props.modelValue;
		props.modelValue && init();
	},
	{
		immediate: true
	}
);

const cancelHandler = () => {
	emits('cancel');
	closeDialog();
};
const closeDialog = () => {
	clearLocker();
	emits('update:modelValue', false);
};

const init = () => {
	countdown.value = 0;
	if (countdown.value === 0) {
		countdown.value = props.count;
		countdownInterval = setInterval(() => {
			countdown.value--;
			if (countdown.value === 0) {
				confirm();
			}
		}, 1000);
	}
};

const clearLocker = () => {
	countdownInterval && clearInterval(countdownInterval);
	countdownInterval = undefined;
};

const confirm = () => {
	emits('confirm');
	closeDialog();
};
</script>
