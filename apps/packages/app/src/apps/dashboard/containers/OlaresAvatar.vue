<template>
	<TerminusAvatar :info="userInfo" :size="32" class="terminu-avatar-wrapper" />
</template>

<script setup lang="ts">
import { getUserInto } from '@apps/dashboard/src/network/user';
import { ref } from 'vue';
const userInfo = ref({});

getUserInto().then((res) => {
	try {
		userInfo.value = res.data.data.info;
	} catch (error) {
		userInfo.value = {};
	}
});
</script>

<style></style>
