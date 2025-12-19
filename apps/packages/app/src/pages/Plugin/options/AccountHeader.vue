<template>
	<div class="row no-wrap items-center flex-gap-sm account-header-container">
		<div class="avatar-wrapper">
			<TerminusAvatar :info="userStore.terminusInfo()" :size="40" />
		</div>

		<div>
			<span class="text-subtitle2 text-ink-1" :class="[sizeOPtion[size]]">
				{{ userStore.current_user?.local_name }}
			</span>
			<div class="text-body3 text-ink-3">
				@{{ userStore.current_user?.domain_name }}
			</div>
		</div>
		<!-- <div style="display: inline-block">
				<terminus-user-status />
			</div> -->
	</div>
</template>

<script setup lang="ts">
import { useUserStore } from 'src/stores/user';
import TerminusUserStatus from 'src/components/common/TerminusUserStatus.vue';

const userStore = useUserStore();

interface Props {
	size?: 'sm' | 'md';
}

const props = withDefaults(defineProps<Props>(), {
	size: 'md'
});

const sizeOPtion = {
	sm: 'text-body1',
	md: 'text-h5'
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
