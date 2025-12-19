<template>
	<div class="column no-wrap flex-gap-y-xl">
		<TerminusSetLanguage titleClass="text-h5"></TerminusSetLanguage>
		<TerminusSetTheme class="q-pa-none" titleClass="text-h5"></TerminusSetTheme>
		<BexBadge></BexBadge>
	</div>
</template>

<script setup lang="ts">
import TerminusSetTheme from 'src/components/common/TerminusSetTheme.vue';
import TerminusSetLanguage from 'src/components/common/TerminusSetLanguage.vue';
import BexBadge from 'src/components/common/BexBadge.vue';
</script>
