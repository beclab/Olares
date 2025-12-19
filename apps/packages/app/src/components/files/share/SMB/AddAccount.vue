<template>
	<div class="text-ink-2 text-subtitle1">{{ t('name') }}</div>
	<terminus-edit
		:inputHeight="40"
		class="q-mr-md"
		style="flex: 1"
		v-model="accountValue"
		:show-password-img="false"
	>
	</terminus-edit>

	<div class="text-ink-2 text-subtitle1 q-mt-md">
		{{ t('files.Add password') }}
	</div>

	<GeneratePassword v-model="passwordValue" class="q-mt-xs" />
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import GeneratePassword from '../GeneratePassword.vue';
import TerminusEdit from 'src/components/common/TerminusEdit.vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	account: {
		type: String,
		required: true
	},
	password: {
		type: String,
		required: true
	}
});

const accountValue = computed({
	get: () => props.account,
	set: (value) => {
		emit('update:account', value);
	}
});

const passwordValue = computed({
	get: () => props.password,
	set: (value) => {
		emit('update:password', value);
	}
});

const { t } = useI18n();

const emit = defineEmits(['update:account', 'update:password']);
</script>

<style scoped lang="scss"></style>
