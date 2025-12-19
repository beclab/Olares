<template>
	<QSectionStyle>
		<q-select
			v-model="selectValue"
			:options="options"
			dense
			outlined
			@update:model-value="change"
		>
			<template v-slot:selected-item="scope">
				<div>{{ selectItemformat(scope.opt) }}</div>
			</template>
			<template v-slot:option="scope">
				<q-item v-bind="scope.itemProps">
					<q-item-section>
						<q-item-label> {{ selectItemformat(scope.opt) }}</q-item-label>
					</q-item-section>
				</q-item>
			</template>
		</q-select>
	</QSectionStyle>
</template>

<script lang="ts">
export type DateRangeItem = string;

export const options: DateRangeItem[] = [
	'10m',
	'20m',
	'30m',
	'1h',
	'2h',
	'3h',
	'5h',
	'8h',
	'12h',
	'1d',
	'2d',
	'3d',
	'7d'
];
</script>

<script setup lang="ts">
import { onMounted, ref, toRefs, watch } from 'vue';
import QSectionStyle from '../../components/QSectionStyle.vue';
import {
	getLastTimeStr,
	getStep,
	getTimeOptions,
	getTimes,
	timeRangeFormate
} from '@apps/control-panel-common/src/containers/Monitoring/utils';
import { useRoute } from 'vue-router';
interface Props {
	defaultValue?: DateRangeItem;
	times?: number;
	step?: string;
	modelValue: any;
}

const route = useRoute();
const props = withDefaults(defineProps<Props>(), {
	step: '10m',
	times: 30
});
const selectValue = ref<DateRangeItem>(getLastTimeStr(props.step, props.times));

const emit = defineEmits<{
	(e: 'change', data: any): void;
	(e: 'update:modelValue', data: any): void;
}>();

const params = ref({
	step: props.step || props.step,
	times: props.times || props.times
});

watch(
	() => props.modelValue,
	(newValue, oldValue) => {
		if (newValue != selectValue.value) {
			selectValue.value = props.modelValue;
		}
	},
	{
		immediate: true
	}
);

const selectItemformat = (opt: string) => {
	const tempOptions = getTimeOptions(options);
	const target = tempOptions.find((item: any) => item.value === opt);
	if (target) {
		return `last ${target.label}`;
	}
	return opt;
};
const initParams = () => {
	const paramsOption = { ...params.value };
	const createTime: any = route.params.createTime;
	if (createTime) {
		const create = new Date(createTime).valueOf() / 1000;
		const now = Date.now() / 1000;
		const interval = now - create;
		console.log('createTime', interval);

		switch (true) {
			// half an hour
			case interval <= 1800:
				paramsOption.times = 30;
				paramsOption.step = '1m';
				break;
			// an hour
			case interval <= 3600:
				paramsOption.times = 60;
				paramsOption.step = '1m';
				break;
			// five hours
			case interval <= 3600 * 5:
				paramsOption.times = 60;
				paramsOption.step = '5m';
				break;
			default:
				break;
		}
	}

	params.value = { ...paramsOption };
};

const change = (data) => {
	const value = props.modelValue;
	if (value) {
		params.value = timeRangeFormate(value, params.value.times);
	}
	emit('change', {
		step: params.value.step,
		times: params.value.times,
		start: '',
		end: '',
		lastTime: value
	});

	emit('update:modelValue', selectValue.value);
};

onMounted(() => {
	initParams();
});
</script>

<style></style>
