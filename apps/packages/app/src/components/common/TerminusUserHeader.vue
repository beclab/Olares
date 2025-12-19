<template>
	<div class="user-header_bg">
		<div class="user-header row items-center justify-between">
			<div class="row items-center">
				<TerminusAccountAvatar />
				<div class="text-ink-1 text-h5 user-header__title">
					{{ title }}
				</div>
			</div>

			<div class="row items-center justify-end">
				<div v-if="$slots.add">
					<slot name="add"></slot>
				</div>
			</div>
		</div>

		<TerminusUserHeaderReminder />
	</div>
</template>

<script setup lang="ts">
import { useUserStore } from '../../stores/user';
import { ref } from 'vue';
import TerminusUserHeaderReminder from './TerminusUserHeaderReminder.vue';
import TerminusAccountAvatar from './TerminusAccountAvatar.vue';

defineProps({
	title: {
		type: String,
		required: true,
		default: ''
	}
});

const userStore = useUserStore();
const userNameRef = ref('');

getUser();

userStore.$subscribe(() => {
	getUser();
});

function getUser() {
	const user = userStore.users!.items.get(userStore.current_id!)!;
	if (user) {
		userNameRef.value = user.name ? user.name : '';
	}
}
</script>

<style scoped lang="scss">
.user-header_bg {
	width: 100%;

	.user-header {
		width: 100%;
		height: 56px;
		padding-left: 20px;
		padding-right: 20px;
		position: relative;

		&__title {
			display: flex;
			align-items: center;
		}
	}
}
</style>
