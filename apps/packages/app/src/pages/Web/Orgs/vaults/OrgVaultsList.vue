<template>
	<div class="vault-list bg-background-1" :class="{ borderRight: isWeb }">
		<div class="row items-center justify-between header">
			<div class="row items-center" style="flex: 1">
				<div class="row items-center q-pl-md" @click="goBack">
					<q-icon v-if="isMobile" name="sym_r_chevron_left" size="24px" />
					<q-icon :name="heading.icon" size="20px" class="q-pa-xs q-mr-xs" />
				</div>
				<div class="column q-ml-sm mobile-content" v-if="!isMobile">
					<div class="text-ink-3 text-overline ellipsis full-width">
						{{ org?.name }}
					</div>
					<div
						class="text-subtitle2 text-ink-1 text-weight-bold ellipsis full-width"
					>
						{{ heading.title }}
					</div>
				</div>
				<div class="column q-pl-md mobile-content" v-else>
					<div class="text-ink-3 text-overline ellipsis full-width">
						{{ org?.name }}
					</div>
					<div
						class="text-subtitle2 text-ink-1 text-weight-bold ellipsis full-width"
					>
						{{ heading.title }}
					</div>
				</div>
			</div>

			<div class="row items-center q-py-xs q-my-md" style="flex: 0 0 40px">
				<q-icon
					class="q-mr-md cursor-pointer"
					name="sym_r_add"
					size="24px"
					clickable
					color="ink-1"
					@click="onCreate"
				>
					<q-tooltip>{{ t('add_vault') }}</q-tooltip>
				</q-icon>
			</div>
		</div>
		<q-list style="width: 100%; height: calc(100% - 60px); overflow: hidden">
			<template v-if="itemList.length > 0">
				<terminus-scroll-area style="height: 100%">
					<template v-for="(item, index) in itemList" :key="index">
						<div class="card-wrap full-width">
							<q-card
								clickable
								v-ripple
								@click="selectItem(item)"
								:active="isSelected(item)"
								active-class="text-blue"
								flat
								class="vaultsCard row items-center justify-start q-my-md q-pa-sm"
								:class="isSelected(item) ? 'vaultActive' : ''"
							>
								<q-card-section
									class="row items-center justify-between q-pa-none"
								>
									<q-icon name="sym_r_deployed_code" size="24px" />
								</q-card-section>
								<q-card-section
									class="column items-start justify-start q-pa-none q-ml-sm name-section"
								>
									<div class="name">
										{{ item.name }}
									</div>
									<div class="row items-center justify-start">
										<div
											class="members text-body3 row items-center justify-center"
										>
											<q-icon name="sym_r_person" size="14px" class="q-mr-xs" />
											<span>{{ org?.getMembersForVault(item)?.length }}</span>
										</div>
									</div>
								</q-card-section>
							</q-card>
						</div>
					</template>
				</terminus-scroll-area>
			</template>
			<div
				class="text-ink-2 column items-center justify-center full-height"
				v-else
			>
				<img src="../../../../assets/layout/nodata.svg" />
				<span class="q-mt-md">
					{{ t('not_have_any_vaults_yet') }}
				</span>
			</div>
		</q-list>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { app } from '../../../../globals';
import { debounce, Vault } from '@didvault/sdk/src/core';
import { useMenuStore } from '../../../../stores/menu';
import { scrollBarStyle } from '../../../../utils/contact';
import { busOn, busOff } from '../../../../utils/bus';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { watch } from 'vue';
import TerminusScrollArea from 'src/components/common/TerminusScrollArea2.vue';

const router = useRouter();
const route = useRoute();
const meunStore = useMenuStore();
const $q = useQuasar();
const isMobile = ref(
	process.env.PLATFORM == 'MOBILE' ||
		process.env.PLATFORM == 'BEX' ||
		$q.platform.is.mobile
);
const isWeb = ref(process.env.APPLICATION == 'VAULT');

const org = ref();

const initOrg = () => {
	console.log('meunStore.org_id ===>', meunStore.org_id);

	org.value = app.orgs.find((org) => org.id == meunStore.org_id);
};

const heading = computed(function () {
	return {
		icon: 'sym_r_apps',
		title: 'Vaults'
	};
});

async function onCreate() {
	router.push({
		path: '/org/Vaults/new'
	});
}
function _getItems() {
	// console.log('app.vaults ===>', app.account?.orgs);
	if (!org.value) {
		return [];
	}

	console.log('org.value ===>', org.value);

	// const orgs = app.getOrg(org.value!);
	// console.log('org =>', orgs?.vaults);

	const vault = org.value.vaults; //app.vaults.filter(({ id }) => app.mainVault?.id != id);
	return vault;
}
async function selectItem(item: Vault) {
	router.push({
		path: '/org/Vaults/' + (item.id ? item.id : '')
	});
	meunStore.org_mode_id = item.id;
}

function isSelected(item: Vault): boolean {
	return meunStore.org_mode_id == item.id;
}
let itemList = ref<Vault[]>(_getItems());

function stateUpdate() {
	initOrg();
	itemList.value = _getItems();
}

const goBack = () => {
	if (isMobile.value) {
		router.go(-1);
	}
};

onMounted(() => {
	stateUpdate();
	busOn('orgSubscribe', stateUpdate);
	meunStore.$subscribe(() => {
		updateItems();
	});
});

onUnmounted(() => {
	busOff('orgSubscribe', stateUpdate);
});

let updateItems = debounce(() => {
	itemList.value = _getItems();
}, 50);

const { t } = useI18n();

watch(
	() => route.params.org_type,
	() => {
		if (!route.params.org_type || route.params.org_type == 'new') {
			meunStore.org_mode_id = '';
		} else {
			meunStore.org_mode_id = route.params.org_type as string;
		}
	}
);
</script>

<style lang="scss" scoped>
.vault-list {
	height: 100vh;
	&.borderRight {
		border-right: 1px solid $separator;
	}
}
.card-wrap {
	display: flex;
	align-items: center;
	justify-content: center;
	border-bottom: 1px solid $separator;
	.vaultsCard {
		width: calc(100% - 24px);
		border: 0;
		border-radius: 0;
		box-sizing: border-box;
		position: relative;
		border-radius: 8px;
		cursor: pointer;

		&:hover {
			background: $background-hover;
		}

		.groups,
		.members {
			height: 20px;
			border: 1px solid $separator;
			border-radius: 4px;
			padding: 0px 6px;
			box-sizing: border-box;
		}

		&.vaultActive {
			background: $background-hover;
		}

		.name-section {
			width: calc(100% - 40px);
			overflow: hidden;

			.name {
				width: 100%;
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;
			}
		}
	}
}

.header {
	flex: 0 0 auto;
	height: 60px;
	width: 100%;

	.mobile-content {
		overflow: hidden;
		text-overflow: ellipsis;
		width: calc(100% - 70px);
	}
}
</style>
