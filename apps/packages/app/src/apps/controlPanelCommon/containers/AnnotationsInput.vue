<template>
	<div class="annotations-input-container">
		<div v-for="(item, index) in values" :key="index" class="input-wrapper">
			<div class="input-item">
				<InputSelection
					v-model="item.key"
					:options="optionsFilter"
					dense
					outlined
					bg-color="white"
					:placeholder="t('KEY')"
				/>
			</div>
			<div class="input-item input-item-second">
				<q-input
					v-model="item.value"
					type="text"
					dense
					outlined
					bg-color="white"
					:placeholder="t('VALUE')"
				/>
			</div>
			<div class="input-tool">
				<q-btn
					color="red"
					flat
					dense
					no-caps
					icon="delete"
					@click="onClick(index)"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { isEmpty } from 'lodash-es';
import { INGRESS_ANNOTATIONS } from '@apps/control-panel-common/src/utils/constants';
import { computed, ref, watch } from 'vue';
import InputSelection from './InputSelection.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface Props {
	modelValue: any;
	options: string[];
	hiddenKeys: any;
	readOnlyKeys: any;
}

const props = withDefaults(defineProps<Props>(), {
	hiddenKeys: [],
	readOnlyKeys: [],
	options: INGRESS_ANNOTATIONS as any
});

const optionsFilter = computed(() => {
	const valuesKeys = values.value.map((item) => item.key);
	return props.options.filter((item) => !valuesKeys.includes(item));
});
const data = ref(props.modelValue);
const getValues = () => {
	const propsValue: any = data.value;
	const hiddenValues: any[] = [];
	const readOnlyValues: any[] = [];
	const arrayValues: any = [];

	Object.keys(propsValue).forEach((key) => {
		if (
			props.hiddenKeys.some((hiddenKey: string) =>
				new RegExp(hiddenKey).test(key)
			)
		) {
			hiddenValues.push({
				key,
				value: propsValue[key]
			});
		} else if (
			props.readOnlyKeys.some((readOnlyKey: string) =>
				new RegExp(readOnlyKey).test(key)
			)
		) {
			readOnlyValues.push({
				key,
				value: propsValue[key]
			});
		} else {
			arrayValues.push({
				key,
				value: propsValue[key]
			});
		}
	});

	if (isEmpty(arrayValues) && isEmpty(readOnlyValues)) {
		arrayValues.push({ key: '' });
	}

	return [...hiddenValues, ...readOnlyValues, ...arrayValues];
};

const values = ref(getValues());

const model = ref();
const selectedOption = ref();
console.log('values', values.value);

const onClick = (index: number) => {
	values.value = values.value.filter((item, itemIndex) => itemIndex !== index);
};

const addItem = () => {
	values.value.push({ key: '', value: '' });
};

const updateValue = (value: any) => {
	model.value = value;
};

const emits = defineEmits(['update:modelValue']);

watch(
	() => values.value,
	() => {
		let obj: Record<string, string> = {};
		values.value.forEach((item, index) => {
			if (item.key) {
				obj[item.key] = item.value;
			}
		});
		console.log('obj: ' + obj);
		emits('update:modelValue', obj);
	},
	{
		deep: true
	}
);

defineExpose({
	addItem
});
</script>

<style lang="scss" scoped>
.annotations-input-container {
	.input-wrapper {
		display: flex;
		align-items: center;
		padding: 10px 12px;
		border-radius: 8px;
		background: $grey-3;
		margin-bottom: 12px;
		.input-item {
			flex: 1;
			&.input-item-second {
				margin-left: 12px;
			}
		}
		.input-tool {
			margin-left: 12px;
		}
	}
}
</style>
