<template>
	<BtSelect
		dense
		v-model="selected"
		:options="options"
		outlined
		@update:model-value="change"
	/>
</template>

<script setup lang="ts">
import { getNamespacesList } from '@apps/control-hub/src/network';
import { ref, onMounted, withDefaults } from 'vue';
import { useI18n } from 'vue-i18n';
import BtSelect from '@apps/control-panel-common/src/components/Select.vue';
const { t } = useI18n();

interface Props {
	modelValue: string;
}

const defaultOption = [
	{
		label: t('ALL_PROJECTS'),
		value: undefined
	}
];
const props = withDefaults(defineProps<Props>(), {});

const emits = defineEmits(['update:modelValue']);

const options = ref();
const selected = ref(defaultOption[0]);

const fetchData = async () => {
	try {
		const params = {
			labelSelector: '!kubesphere.io/devopsproject',
			sortBy: 'createTime'
		};
		const {
			data: { items, totalItems }
		} = await getNamespacesList(params);
		const namespaces = items.map((item) => ({
			label: item.metadata.name,
			value: item.metadata.name
		}));
		options.value = defaultOption.concat(namespaces);
	} catch (error) {
		//
	}
};

const change = () => {
	emits('update:modelValue', selected.value.value);
};

onMounted(() => {
	fetchData();
});
</script>

<style></style>
