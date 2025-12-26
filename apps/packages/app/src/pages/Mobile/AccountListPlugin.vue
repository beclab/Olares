<template>
	<div class="account-list-container">
		<terminus-select-all
			ref="terminusSelect"
			:items="totalUsersIds"
			@show-select-mode="showSelectMode"
			@item-on-unable-select="itemOnUnableSelect"
		>
			<template v-slot="{ file }">
				<terminus-account-item
					class="q-mt-xs list-item-wrapper"
					:class="{
						'bg-light-blue-soft':
							file.id === userStore.current_id || file.id === approvalUserIdRef
					}"
					:selected="
						file.id === userStore.current_id || file.id === approvalUserIdRef
					"
					:user="userStore.users?.items.get(file.id)"
					@click="choose(file.id)"
					:bordered="false"
				>
					<template v-slot:side>
						<q-icon
							class="list-item-hover"
							name="sym_r_sync_alt"
							color="light-blue-default"
							size="20px"
						/>
					</template>
				</terminus-account-item>
			</template>
		</terminus-select-all>
	</div>
</template>

<script lang="ts" setup>
import { useUserStore } from 'src/stores/user';
import TerminusAccountItem from 'src/components/common/TerminusAccountItem.vue';
import TerminusSelectAll from 'src/components/common/TerminusSelectAll.vue';
import { useAccountList } from 'src/composables/mobile/useAccountList';

const userStore = useUserStore();
const {
	approvalUserIdRef,
	totalUsersIds,
	terminusSelect,
	choose,
	showSelectMode,
	itemOnUnableSelect,
	t
} = useAccountList();
</script>

<style lang="scss" scoped>
.account-list-container {
	width: 260px;
	.list-item-wrapper {
		.list-item-hover {
			display: none;
		}
		&:hover {
			.list-item-hover {
				display: block;
			}
		}
	}
}
</style>
