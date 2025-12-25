<template>
	<q-card flat class="row items-center">
		<div class="q-mr-lg">
			<q-circular-progress
				show-value
				:value="status.current / count"
				:min="0"
				:max="1"
				:thickness="0.32"
				size="100px"
				track-color="background-3"
				color="blue-default"
			>
				<div class="replication-status-value">
					{{ status.current || 0 }}/{{ count }}
				</div>
			</q-circular-progress>
		</div>
		<div class="column justify-center" v-if="onScale">
			<QButtonStyle>
				<q-btn dense unelevated color="blue-default" @click="add">
					<q-icon name="add" color="ink-on-brand" />
				</q-btn>
			</QButtonStyle>

			<QButtonStyle class="q-mt-lg">
				<q-btn
					dense
					unelevated
					color="blue-default"
					@click="remove"
					:disable="removeDisabled"
				>
					<q-icon name="remove" color="ink-on-brand" />
				</q-btn>
			</QButtonStyle>
		</div>
		<CountdownDialog
			:count="5"
			:title="t('ADJUST_REPLICAS')"
			:content="
				t('REPLICAS_SCALE_NOTIFY_CONTENT', {
					num: count
				})
			"
			v-model="visible"
			@cancel="cancel"
			@confirm="change"
		/>
	</q-card>
</template>

<script setup lang="ts">
import { t } from '@apps/control-hub/src/boot/i18n';
import { computed, ref, watch } from 'vue';
import CountdownDialog from './CountdownDialog.vue';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';

interface Props {
	status: {
		current: number;
		desire: number;
	};
	onScale: boolean;
}
const props = withDefaults(defineProps<Props>(), {
	onScale: true
});
const emits = defineEmits(['change']);

const visible = ref(false);
const count = ref(0);

const removeDisabled = computed(() => count.value <= 0);
watch(
	() => props.status.desire,
	(newValue, oldValue) => {
		if (newValue !== oldValue) {
			count.value = props.status.desire;
		}
	},
	{
		immediate: true
	}
);

function change() {
	emits('change', count.value);
}

function add() {
	visible.value = true;
	count.value = props.status.desire + 1;
}

function remove() {
	if (count.value <= 0) {
		return;
	}
	visible.value = true;
	count.value = props.status.desire - 1;
}

function cancel() {
	visible.value = false;
	count.value = props.status.current;
}
</script>

<style lang="scss" scoped>
.replication-status-value {
	color: $ink-1;
	font-size: 20px;
	font-weight: 600;
	line-height: 28px;
}
</style>
