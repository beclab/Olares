<template>
	<div class="member_wrap row items-center justify-center">
		<div class="stretch row items-center justify-start">
			<div class="q-mr-md avator">
				<TerminusAvatar
					:info="userStore.getUserTerminusInfo(member.id || '')"
					:size="28"
				/>
			</div>
			<div class="column justify-between">
				<div class="did text-body text-ink-1">{{ member.name }}</div>
				<div>
					<!-- TODO: snowning.com error-->
					<span class="text-ink-2">@{{ userStore.getCurrentDomain() }}</span>
					<span
						class="tag text-subtitle3 owner"
						v-if="member.role === OrgRole.Owner"
						>{{ t('owner') }}</span
					>
					<span
						class="tag text-subtitle3 admin"
						v-if="member.role === OrgRole.Admin"
						>{{ t('admin') }}</span
					>
					<span
						class="tag text-3ubtitle3 suspended"
						v-if="org?.isSuspended(member)"
						>{{ t('suspend') }}</span
					>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { OrgMember, OrgRole } from '@didvault/sdk/src/core';
import { app } from '../../../../globals';
import { useMenuStore } from '../../../../stores/menu';
import { useUserStore } from '../../../../stores/user';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	member: {
		type: OrgMember,
		required: true
	},
	isSelected: {
		type: Boolean,
		default: false,
		required: false
	}
});

const meunStore = useMenuStore();
const userStore = useUserStore();
const org = computed(function () {
	return app.orgs.find((org) => org.id == meunStore.org_id);
});

const { t } = useI18n();
</script>

<style lang="scss" scoped>
.member_wrap {
	width: 100%;

	.stretch {
		width: 100%;
		height: 40px;
		border-radius: 8px;
		.avator {
			width: 28px;
			height: 28px;
			border-radius: 14px;
			overflow: hidden;
		}
		.did {
			width: 90%;
		}
		.tag {
			padding: 0 4px;
			border-radius: 4px;
			margin-left: 10px;
			&.owner {
				border: 1px solid $deep-orange-5;
				color: $deep-orange-5;
			}
			&.admin {
				border: 1px solid $blue-4;
				color: $blue-4;
			}
			&.provisioned {
				border: 1px solid $grey-8;
				color: $grey-8;
			}
			&.suspended {
				border: 1px solid $grey-8;
				color: $grey-8;
			}
		}
	}
}
</style>
