<template>
	<q-input filled v-model="storageName" hint="Name and surname" />
	<q-input filled v-model="alias" hint="Name and surname" />
	<q-input filled v-model="desc" hint="Name and surname" />
</template>

<script setup lang="ts">
import { useQuasar } from 'quasar';
import { ref, withDefaults, onMounted } from 'vue';

interface Props {
	data: any;
}

const props = withDefaults(defineProps<Props>(), {});

onMounted(() => {
	storageName.value = props.data.name;
	alias.value = props.data.aliasName;
	desc.value = props.data.description;
});

const $q = useQuasar();

const storageName = ref();
const alias = ref();
const desc = ref();

const onReset = () => {
	storageName.value = undefined;
	alias.value = undefined;
	desc.value = undefined;
};
</script>
