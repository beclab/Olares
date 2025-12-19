<template>
	<div class="terminus-account-avatar-bex">
		<div class="relative-position z-top">
			<TerminusAvatar
				class="avatar-icon"
				:info="userStore.terminusInfo()"
				:size="24"
				style="position: relative"
			/>
			<div
				class="user_status"
				:class="
					configIconClass(
						termipassStore.totalStatus?.isError || UserStatusActive.normal
					)
				"
			></div>
			<div
				class="user-list-popover q-px-md q-pb-md q-pt-sm q-mr-sm bg-background-2"
			>
				<AccountList></AccountList>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useUserStore } from '../../stores/user';
import { useTermipassStore } from '../../stores/termipass';
import { UserStatusActive } from '../../utils/checkTerminusState';
import AccountList from 'src/pages/Mobile/AccountListPlugin.vue';

const userStore = useUserStore();
const termipassStore = useTermipassStore();

const configIconClass = (active: UserStatusActive) => {
	if (active == UserStatusActive.error) {
		return 'bg-red';
	}
	if (active == UserStatusActive.normal) {
		return 'bg-grey';
	}
	return 'bg-green';
};
</script>

<style scoped lang="scss">
.terminus-account-avatar-bex {
	position: relative;
	width: 24px;
	height: 24px;
	.user-list-popover {
		position: absolute;
		bottom: 0;
		right: 100%;
		border-radius: 12px;
		box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2);
		display: none;
	}
	&::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: -8px;
		right: 0;
		height: 100%;
	}
	&:hover {
		.user-list-popover {
			display: block;
		}
	}
	.avatar-icon {
		border-radius: 50%;
		overflow: hidden;
	}
	.user_status {
		width: 8px;
		height: 8px;
		border-radius: 6px;
		overflow: hidden;
		position: absolute;
		left: 100%;
		transform: translateX(-50%);
		bottom: 0px;
		border: 1px white solid;
	}
}
</style>
