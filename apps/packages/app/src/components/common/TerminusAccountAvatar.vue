<template>
	<div class="terminus-account-avatar">
		<TerminusAvatar
			class="avatar-circle q-mr-sm"
			:info="userStore.terminusInfo()"
			:size="40"
			@click="onOKChoose"
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
	</div>
</template>

<script setup lang="ts">
import { useUserStore } from '../../stores/user';
import { useTermipassStore } from '../../stores/termipass';
import { UserStatusActive } from '../../utils/checkTerminusState';
import { useRouter } from 'vue-router';

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

const $router = useRouter();

function onOKChoose() {
	$router.push('/accounts');
}
</script>

<style scoped lang="scss">
.terminus-account-avatar {
	position: relative;
	.user_status {
		width: 12px;
		height: 12px;
		border-radius: 6px;
		overflow: hidden;
		position: absolute;
		left: 28px;
		bottom: 0px;
		border: 1px white solid;
	}
}
</style>
