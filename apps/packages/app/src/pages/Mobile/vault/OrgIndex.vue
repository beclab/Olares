<template>
	<q-page class="column orgs" :style-fn="myTweak">
		<OrgDashboardIndex v-if="orgMode === OrgMenu.DASHBOARD" />
		<OrgVaultsList v-if="orgMode === OrgMenu.VAULTES" />
		<OrgInvitesList v-if="orgMode === OrgMenu.INVITES" />
		<OrgMembersList v-if="orgMode === OrgMenu.MEMBERS" />
		<OrgSettingsIndex v-if="orgMode === OrgMenu.SETTINGS" />
	</q-page>
</template>

<script lang="ts">
import { defineComponent, ref, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import { OrgMenu } from '../../../globals';
import OrgDashboardIndex from './../../Web/Orgs/dashboard/OrgDashboardIndex.vue';
import OrgInvitesList from './../../Web/Orgs/invites/OrgInvitesList.vue';
import OrgMembersList from './../../Web/Orgs/members/OrgMembersList.vue';
import OrgVaultsList from './../../Web/Orgs/vaults/OrgVaultsList.vue';
import OrgSettingsIndex from './../../Web/Orgs/settings/OrgSettingsIndex.vue';

export default defineComponent({
	name: 'OrgIndexPage',
	components: {
		OrgDashboardIndex,
		OrgVaultsList,
		OrgInvitesList,
		OrgMembersList,
		OrgSettingsIndex
	},
	setup() {
		const Route = useRoute();
		const orgMode = ref(Route.params.org_mode);

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
			return { height: '100%' };
		};

		return {
			orgMode,
			myTweak,
			title,
			OrgMenu
		};
	}
});
</script>

<style scoped lang="scss">
.orgs {
	width: 100%;
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
