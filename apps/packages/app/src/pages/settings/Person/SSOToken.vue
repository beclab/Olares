<template>
	<page-title-component
		:show-back="true"
		:title="t('list_of_sso_authorization_tokens')"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<AdaptiveLayout>
			<template v-slot:pc>
				<div class="q-pa-md column q-list-class">
					<q-table
						tableHeaderStyle="height: 32px;"
						table-header-class="text-body3 text-ink-2"
						flat
						:bordered="false"
						:rows="sso"
						:columns="columns"
						row-key="id"
						v-model:pagination="pagination"
						hide-pagination
						hide-selected-banner
						hide-bottom
					>
						<template v-slot:body-cell-actions="props">
							<q-td :props="props" class="text-ink-2">
								<q-icon
									v-if="
										props.row.termiPass &&
										!allCookies.includes(props.row.termiPass.sso)
									"
									name="sym_r_delete"
									size="16px"
									class="q-mr-sm"
									color="ink-2"
									@click.stop="revokeToken(props.row)"
								/>
							</q-td>
						</template>
					</q-table>

					<div class="row justify-center q-mt-md">
						<q-pagination
							v-model="pagination.page"
							:max="pagesNumber"
							input
							round
							icon-first="sym_r_keyboard_double_arrow_left"
							icon-last="sym_r_keyboard_double_arrow_right"
							icon-next="sym_r_keyboard_arrow_right"
							icon-prev="sym_r_keyboard_arrow_left"
							color="ink-3"
							input-class="text-ink-2 text-body-3"
						/>
					</div>
				</div>
			</template>
			<template v-slot:mobile>
				<bt-grid
					class="mobile-items-list"
					:repeat-count="2"
					v-for="(token, index) in sso"
					:key="index"
					:paddingY="12"
				>
					<template v-slot:title>
						<div
							class="text-subtitle3-m row justify-between items-center clickable-view q-mb-md"
						>
							<div>{{ t('SSO authorization token') }}</div>
							<q-icon
								name="sym_r_delete"
								color="ink-2"
								size="20px"
								v-if="
									token.termiPass && !allCookies.includes(token.termiPass.sso)
								"
								@click.stop="revokeToken(token)"
							/>
						</div>
					</template>
					<template v-slot:grid>
						<bt-grid-item
							:label="t('expire_time')"
							mobileTitleClasses="text-body3-m"
							:value="
								date.formatDate(token.sso.expireTime * 1000, 'YYYY-MM-DD HH:mm')
							"
						/>
						<bt-grid-item
							:label="t('auth_level')"
							mobileTitleClasses="text-body3-m"
							:value="token.sso.authLevel"
						/>
						<bt-grid-item
							:label="t('first_factor_time')"
							mobileTitleClasses="text-body3-m"
							:value="
								date.formatDate(
									token.sso.firstFactorTimestamp * 1000,
									'YYYY-MM-DD HH:mm'
								)
							"
						/>
						<bt-grid-item
							:label="t('second_factor_time')"
							mobileTitleClasses="text-body3-m"
							:value="
								token.sso.secondFactorTimestamp == 0
									? '--'
									: date.formatDate(
											token.sso.secondFactorTimestamp * 1000,
											'YYYY-MM-DD HH:mm'
									  )
							"
						/>
					</template>
				</bt-grid>
			</template>
		</AdaptiveLayout>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useAdminStore, SSOResult } from 'src/stores/settings/admin';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import BtGridItem from 'src/components/settings/base/BtGridItem.vue';
import BtGrid from 'src/components/settings/base/BtGrid.vue';
import { date, useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const adminStore = useAdminStore();
const $q = useQuasar();
const sso = ref<SSOResult[]>([]);
const allCookies = ref(document.cookie);

onMounted(async () => {
	// allCookies.value = document.cookie;
	sso.value = await adminStore.get_sso();
});

const revokeToken = async (token: SSOResult) => {
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('revoke_token'),
			message: t('revoke_token_message')
		}
	}).onOk(async () => {
		// deleteUserSureAction();
		$q.loading.show();
		try {
			await adminStore.revoke_token(token);
			sso.value = await adminStore.get_sso();
			$q.loading.hide();
		} catch (error) {
			console.log(error);
			$q.loading.hide();
		}
	});
};

const columns = [
	{
		name: 'expireTime',
		align: 'left',
		label: t('expire_time'),
		field: 'sso',
		format: (val: any) => {
			return date.formatDate(val.expireTime * 1000, 'YYYY-MM-DD HH:mm');
		},
		sortable: false
	},

	{
		name: 'authLevel',
		align: 'left',
		label: t('auth_level'),
		field: 'sso',
		format: (val: any) => {
			return val.authLevel;
		},
		sortable: false
	},
	{
		name: 'firstFactorTimestamp',
		align: 'left',
		label: t('first_factor_time'),
		field: 'sso',
		format: (val: any) => {
			return date.formatDate(
				val.firstFactorTimestamp * 1000,
				'YYYY-MM-DD HH:mm'
			);
		},
		sortable: false
	},
	{
		name: 'secondFactorTimestamp',
		align: 'left',
		label: t('second_factor_time'),
		field: 'sso',
		format: (val: any) => {
			if (val.secondFactorTimestamp == 0) {
				return '--';
			}
			return date.formatDate(
				val.secondFactorTimestamp * 1000,
				'YYYY-MM-DD HH:mm'
			);
		},
		sortable: false
	},
	{
		name: 'actions',
		label: t('actions'),
		field: 'termiPass',
		align: 'right'
	}
];

const pagesNumber = computed(() =>
	Math.ceil(sso.value.length / pagination.value.rowsPerPage)
);

const pagination = ref({
	sortBy: 'desc',
	descending: false,
	page: 1,
	rowsPerPage: 10
});
</script>
