<template>
	<q-page class="column orgs bg-white" :style-fn="myTweak">
		<OrgVaultsView v-if="orgMode === OrgMenu.VAULTES" />
		<OrgInvitesView v-if="orgMode === OrgMenu.INVITES" />
		<OrgMembersView v-if="orgMode === OrgMenu.MEMBERS" />
	</q-page>
</template>

<script lang="ts">
import { defineComponent, ref, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import { OrgMenu } from '../../../globals';
import OrgVaultsView from './../../Web/Orgs/vaults/OrgVaultsView.vue';
import OrgInvitesView from './../../Web/Orgs/invites/OrgInvitesView.vue';
import OrgMembersView from './../../Web/Orgs/members/OrgMembersView.vue';

export default defineComponent({
	name: 'OrgIndexPage',
	components: {
		OrgVaultsView,
		OrgInvitesView,
		OrgMembersView
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
