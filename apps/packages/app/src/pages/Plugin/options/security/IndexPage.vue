<template>
	<div class="column no-wrap flex-gap-y-xl">
		<TerminusSetSecurityPassword></TerminusSetSecurityPassword>
		<MnemonicsCard></MnemonicsCard>
		<TerminusSetSecurityAutoLock></TerminusSetSecurityAutoLock>
		<CookieManage></CookieManage>
	</div>
</template>

<script setup lang="ts">
import TerminusSetSecurityPassword from 'src/components/common/TerminusSetSecurityPasswordBex.vue';
import TerminusSetSecurityAutoLock from 'src/components/common/TerminusSetSecurityAutoLockBex.vue';
import MnemonicsCard from './MnemonicsCard.vue';
import CookieManage from './CookieManage.vue';
</script>
