<template>
	<div class="account-list-root">
		<terminus-title-bar
			v-if="!approvalUserIdRef && selectIds === null"
			:title="t('switch_accounts')"
		>
			<template v-slot:right>
				<q-btn
					class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
					icon="sym_r_checklist"
					text-color="ink-2"
					@click="intoCheckedMode"
				>
				</q-btn>
			</template>
		</terminus-title-bar>
		<terminus-select-header
			v-else
			:selectIds="selectIds"
			@handle-close="handleClose"
			@handle-select-all="handleSelectAll"
			@handle-remove="handleRemove"
		/>
		<terminus-scroll-area class="account-list-scroll">
			<template v-slot:content>
				<terminus-select-all
					ref="terminusSelect"
					:items="totalUsersIds"
					@show-select-mode="showSelectMode"
					@item-on-unable-select="itemOnUnableSelect"
				>
					<template v-slot="{ file }">
						<terminus-account-item
							:user="userStore.users?.items.get(file.id)"
							@click="choose(file.id)"
							style="margin-top: 12px"
						>
							<template
								v-slot:side
								v-if="
									file.id === userStore.current_id ||
									file.id === approvalUserIdRef
								"
							>
								<q-icon
									name="sym_r_check_circle"
									size="24px"
									color="light-blue-default"
								/>
							</template>
						</terminus-account-item>
					</template>
				</terminus-select-all>

				<terminus-item
					img-bg-classes="bg-background-3"
					style="margin-top: 12px"
					icon-name="sym_r_person_add"
					:img-b-g-size="40"
					:border-radius="12"
					@click="addAccount"
					v-if="!approvalUserIdRef && selectIds === null"
					title-classes="add-new-terminus-name"
				>
					<template v-slot:title>
						{{ t('add_new_olares_id') }}
					</template>
				</terminus-item>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script lang="ts" setup>
import { useUserStore } from '../../stores/user';
import TerminusItem from '../../components/common/TerminusItem.vue';
import TerminusTitleBar from '../../components/common/TerminusTitleBar.vue';
import TerminusAccountItem from '../../components/common/TerminusAccountItem.vue';
import TerminusScrollArea from '../../components/common/TerminusScrollArea.vue';
import TerminusSelectAll from './../../components/common/TerminusSelectAll.vue';
import TerminusSelectHeader from '../../components/common/TerminusSelectHeader.vue';
import { useAccountList } from 'src/composables/mobile/useAccountList';

const userStore = useUserStore();
const {
	approvalUserIdRef,
	selectIds,
	totalUsersIds,
	terminusSelect,
	choose,
	addAccount,
	intoCheckedMode,
	showSelectMode,
	handleSelectAll,
	handleClose,
	handleRemove,
	itemOnUnableSelect,
	t
} = useAccountList();
</script>

<style lang="scss" scoped>
.account-list-root {
	width: 100%;
	height: 100%;

	.account-list-scroll {
		height: calc(100% - 56px);
		width: 100%;
		padding-left: 20px;
		padding-right: 20px;
	}

	.current-user {
		padding: 4px 8px;
		border-radius: 4px;
		text-align: center;
		border: 1px solid $blue-4;
		color: $blue-4;
	}

	.request-user {
		@extend .current-user;
		border: 1px solid $green;
		color: $green;
	}
}
</style>
