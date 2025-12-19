<template>
	<div
		class="text-subtitle3 text-light-blue-default row items-center justify-center export-root"
		:style="{ '--height': height + 'px' }"
		:class="{
			'border-class': border
		}"
		@click="exportMnemonics"
	>
		{{
			userStore.currentUserBackup
				? $t('export_mnemonic_phrase')
				: $t('backup_mnemonic_phrase')
		}}
	</div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useUserStore } from '../../stores/user';
import { busEmit } from '../../utils/bus';

defineProps({
	height: {
		type: Number,
		default: 30,
		required: false
	},
	border: {
		type: Boolean,
		default: false,
		required: false
	}
});

const router = useRouter();
const userStore = useUserStore();

const exportMnemonics = async () => {
	if (!(await userStore.unlockFirst(undefined, { hide: true }))) {
		return;
	}
	if (!userStore.passwordReseted) {
		busEmit('configPassword');
		return;
	}
	router.push({
		path: '/backup_mnemonics',
		query: {
			backup: userStore.currentUserBackup ? 0 : 1
		}
	});
};
</script>

<style scoped lang="scss">
.export-root {
	width: 100%;
	height: var(--height, 30px);
	// height: 48px;
}

.border-class {
	border: 1px solid $separator;
	border-radius: 8px;
}
</style>
