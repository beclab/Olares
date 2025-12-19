<template>
	<q-page class="column orgs" :style-fn="myTweak">
		<OrgDashboardIndex v-if="orgMode === OrgMenu.DASHBOARD" />
		<OrgVaultsIndex v-if="orgMode === OrgMenu.VAULTES" />
		<OrgInvitesIndex v-if="orgMode === OrgMenu.INVITES" />
		<OrgMembersIndex v-if="orgMode === OrgMenu.MEMBERS" />
		<OrgSettingsIndex v-if="orgMode === OrgMenu.SETTINGS" />
	</q-page>
</template>

<script lang="ts">
import { defineComponent, ref, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import { OrgMenu } from '../../../globals';
import OrgDashboardIndex from './dashboard/OrgDashboardIndex.vue';
import OrgVaultsIndex from './vaults/OrgVaultsIndex.vue';
import OrgInvitesIndex from './invites/OrgInvitesIndex.vue';
import OrgMembersIndex from './members/OrgMembersIndex.vue';
import OrgSettingsIndex from './settings/OrgSettingsIndex.vue';
import { useQuasar } from 'quasar';

export default defineComponent({
	name: 'OrgIndexPage',
	components: {
		OrgDashboardIndex,
		OrgVaultsIndex,
		OrgInvitesIndex,
		OrgMembersIndex,
		OrgSettingsIndex
	},
	setup() {
		const Route = useRoute();
		const orgMode = ref(Route.params.org_mode);

		const $q = useQuasar();

		const orgHeight = computed(() => {
			if (process.env.PLATFORM === 'DESKTOP') {
				if ($q.platform.is.win) {
					return 'calc(100vh - 40px) !important';
				}
				return 'calc(100vh - 16px) !important';
			} else {
				return '100vh !important';
			}
		});

		watch(
			() => Route.params.org_mode,
			(newValue, oldValue) => {
				if (newValue == oldValue) {
					return;
				}
				orgMode.value = newValue;
			}
		);

		const title = computed(() => {
			return 'Org/' + orgMode.value;
		});

		const myTweak = () => {
			if (process.env.PLATFORM === 'DESKTOP') {
				if ($q.platform.is.win) {
					return { height: 'calc(100vh - 40px) !important' };
				}
				return { height: 'calc(100vh - 16px) !important' };
			} else {
				return { height: '100vh !important' };
			}
		};

		return {
			orgMode,
			myTweak,
			title,
			OrgMenu,
			orgHeight
		};
	}
});
</script>

<style scoped lang="scss">
.orgs {
	border-radius: 12px;
	.settingItem {
		height: 58px;
		line-height: 58px;
		border-bottom: 0.5px solid #ececec;
		box-sizing: border-box;

		&.itemActive {
			border-left: 2px solid $blue;
		}
	}
}
</style>
