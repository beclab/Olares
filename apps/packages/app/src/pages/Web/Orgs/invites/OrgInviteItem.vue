<template>
	<div class="row items-center justify-start">
		<q-card-section
			class="row items-center justify-between q-pa-none text-ink-1"
		>
			<q-icon name="sym_r_mail" size="24px" />
		</q-card-section>
		<q-card-section class="column items-start justify-start q-pa-none q-ml-sm">
			<div class="text-weight-bold text-body1 text-ink-1">
				{{ invite.did }}
			</div>
			<div class="row items-center justify-start">
				<div
					class="tag text-subtitle3 text-ink-2 row items-center justify-center q-mr-sm"
					:class="status.class"
				>
					<q-icon :name="status.icon" size="20px" class="q-mr-xs" />
					{{ status.text }}
				</div>
			</div>
		</q-card-section>
	</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { Invite } from '@didvault/sdk/src/core';
import { formatDateFromNow } from 'src/utils/format';
import { useI18n } from 'vue-i18n';

export default defineComponent({
	name: 'InviteItem',
	props: {
		// ...your custom props
		invite: {
			type: Invite,
			required: true
		}
	},
	components: {},

	emits: [],
	setup(props: any) {
		const { t } = useI18n();
		const status = props.invite.expired
			? { icon: 'sym_r_schedule', class: 'warning', text: t('expired') }
			: props.invite.accepted
			? {
					icon: 'sym_r_check_small',
					class: 'highlight',
					text: t('accepted')
			  }
			: {
					icon: 'sym_r_schedule',
					class: '',
					text: (() => formatDateFromNow(props.invite.expires, false))()
			  };
		return {
			status
		};
	}
});
</script>

<style lang="scss" scoped>
.tags {
	display: flex;
	overflow-x: auto;
	align-items: center;
	-webkit-overflow-scrolling: touch;
	margin-top: 4px;
}

.tag {
	font-weight: bold;
	border-radius: 4px;
	padding: 0 4px;
	line-height: normal;
	text-align: center;
	border: 1px solid $separator;
}
.tag:not(:last-child) {
	margin-right: 6px;
}
.tag.highlight {
	border: 1px solid rgb(89, 198, 255);
	color: rgb(89, 198, 255);
}
.tag.warning {
	border: 1px solid $red;
	color: $red-4;
}
</style>
