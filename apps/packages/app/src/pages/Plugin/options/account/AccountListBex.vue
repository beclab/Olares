<template>
	<div class="account-list-root">
		<div class="action-wrapper z-top">
			<div v-if="!approvalUserIdRef && selectIds === null">
				<q-btn
					@click="intoCheckedMode"
					color="background-3"
					text-color="ink-2"
					padding="sm lg"
					no-caps
				>
					<div class="row inline items-center flex-gap-xs text-body1">
						<q-icon name="sym_r_supervisor_account" size="20px" />
						<span>{{ $t('Manage') }}</span>
					</div>
				</q-btn>
			</div>
			<div class="row items-center flex-gap-x-md" v-else>
				<q-btn
					color="background-3"
					text-color="ink-2"
					padding="sm lg"
					no-caps
					@click="handleClose"
				>
					<span class="text-body1">{{ $t('cancel') }}</span>
				</q-btn>

				<q-btn
					color="background-3"
					text-color="negative"
					padding="sm lg"
					no-caps
					@click="handleRemove"
				>
					<div class="row inline items-center flex-gap-xs text-body1">
						<q-icon name="sym_r_supervisor_account" size="20px" />
						<span>{{ $t('delete') }}</span>
					</div>
				</q-btn>
			</div>
		</div>
		<div>
			<terminus-select-all
				ref="terminusSelect"
				:items="totalUsersIds"
				@show-select-mode="showSelectMode"
				@item-on-unable-select="itemOnUnableSelect"
				item-class="list-item-wrapper q-mb-md q-pr-lg q-py-md"
			>
				<template v-slot="{ file }">
					<div>
						<TerminusAccountItem2
							class="q-mt-xs list-item-content"
							:selected="
								file.id === userStore.current_id ||
								file.id === approvalUserIdRef
							"
							:user="userStore.users?.items.get(file.id)"
							@click="choose(file.id)"
							:bordered="false"
							size="lg"
							style="padding: 20px 20px 20px 20px"
						>
							<template
								v-slot:side
								v-if="
									!(
										file.id === userStore.current_id ||
										file.id === approvalUserIdRef
									)
								"
							>
								<q-icon
									class="list-item-hover"
									name="sym_r_sync_alt"
									color="light-blue-default"
									size="20px"
								/>
							</template>
						</TerminusAccountItem2>
					</div>
				</template>
			</terminus-select-all>
			<terminus-item
				img-bg-classes="bg-background-3 terminus-item-img-wrapper"
				icon-name="sym_r_person_add"
				:img-b-g-size="40"
				:border-radius="12"
				@click="importAccount"
				v-if="!approvalUserIdRef && selectIds === null"
				title-classes="add-new-terminus-name"
				style="padding-left: 20px; padding-right: 20px"
				:item-height="82"
				:wholePictureSize="40"
				icon-color="ink-2"
			>
				<template v-slot:title>
					<span class="text-ink-3">{{ t('add_new_olares_id') }}</span>
				</template>
			</terminus-item>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useUserStore } from 'src/stores/user';
import TerminusItem from 'src/components/common/TerminusItem.vue';
import TerminusAccountItem2 from 'src/components/common/TerminusAccountItem2.vue';
import TerminusSelectAll from 'src/components/common/TerminusSelectAll2.vue';
import { useAccountList } from 'src/composables/mobile/useAccountList';
import { useRouter } from 'vue-router';
import { ROUTE_CONST } from 'src/router/route-const';

const router = useRouter();
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

const importAccount = () => {
	router.push({
		path: '/import_mnemonic'
	});
};
</script>

<style lang="scss" scoped>
.account-list-root {
	width: 100%;
	height: 100%;
	position: relative;
	.action-wrapper {
		position: absolute;
		top: -64px;
		right: 0;
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
	::v-deep(.list-item-wrapper) {
		border: 1px solid $separator;
		border-radius: 12px;
		width: 100%;
	}
}
.terminus-item-img-wrapper {
	border: 1px solid red;
	width: 40px !important;
	height: 40px !important;
}
.list-item-content {
	.list-item-hover {
		display: none;
	}
	&:hover {
		.list-item-hover {
			display: block;
		}
	}
}
</style>
