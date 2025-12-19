<template>
	<div class="row no-wrap items-center flex-gap-md account-header-container">
		<div class="avatar-wrapper">
			<TerminusAvatar :info="userStore.terminusInfo()" :size="44" />
		</div>

		<div>
			<div class="row no-wrap flex-gap-xs">
				<span class="text-ink-1" :class="[sizeOPtion[size]]">
					{{ userStore.current_user?.local_name }}
				</span>
				<span class="text-body1 text-ink-3"
					>@{{ userStore.current_user?.domain_name }}</span
				>
			</div>
			<div style="display: inline-block">
				<TerminusUserStatus2 />
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useUserStore } from 'src/stores/user';
import TerminusUserStatus2 from 'src/components/common/TerminusUserStatus2.vue';

const userStore = useUserStore();

interface Props {
	size?: 'sm' | 'md';
}

const props = withDefaults(defineProps<Props>(), {
	size: 'md'
});

const sizeOPtion = {
	sm: 'text-body1',
	md: 'text-h6'
};
</script>

<style lang="scss" scoped>
.account-header-container {
	.avatar-wrapper {
		border-radius: 50%;
		overflow: hidden;
	}
}
</style>
