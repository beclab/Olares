<template>
	<q-item
		clickable
		dense
		class="terminus-account-root row items-center q-pa-sm q-pr-md flex-gap-x-sm"
		:class="{
			'bg-light-blue-soft': selected,
			bordered: bordered
		}"
	>
		<terminus-avatar
			v-if="user.name"
			:info="userStore.getUserTerminusInfo(user.id)"
			:size="32"
			class="avatar-circle"
		/>
		<div
			class="terminus-account-root__img row items-center justify-center"
			v-else
		>
			<q-icon name="sym_r_person" size="20px" color="text-ink-1" />
		</div>

		<div
			class="row no-wrap items-center justify-between"
			style="flex: 1; overflow: hidden"
		>
			<div class="column justify-between full-width">
				<div class="row items-center full-width">
					<div
						class="text-subtitle3 ellipsis"
						style="text-align: left"
						:class="{
							'text-ink-1': user.name,
							'text-ink-2': !user.name
						}"
					>
						{{ user.name ? user.local_name : t('olares_id_not_created') }}
					</div>
					<!-- <div
						style="max-width: 60%"
						v-if="user.name && user.id == userStore.current_user?.id"
					>
						<terminus-user-status class="q-ml-sm" />
					</div> -->
				</div>

				<div class="text-overline text-ink-3 q-mt-xs" style="text-align: left">
					{{ subInfo }}
				</div>
			</div>
		</div>
		<div class="column justify-center" v-if="$slots.side">
			<slot name="side" />
		</div>
	</q-item>
</template>

<script setup lang="ts">
import { onMounted, PropType, ref } from 'vue';
import { UserItem } from '@didvault/sdk/src/core';
import { generateStringEllipsis } from '../../utils/utils';
import { useI18n } from 'vue-i18n';
import { useUserStore } from '../../stores/user';

const { t } = useI18n();
const userStore = useUserStore();

const props = defineProps({
	user: {
		type: Object as PropType<UserItem>,
		required: true
	},
	itemHeight: {
		type: Number,
		default: 64,
		required: false
	},
	selected: {
		type: Boolean,
		default: false
	},
	bordered: {
		type: Boolean,
		default: true
	}
});

const subInfo = ref();

onMounted(() => {
	if (props.user.name) {
		subInfo.value = '@' + props.user.domain_name;
	} else {
		subInfo.value = props.user.id
			? generateStringEllipsis(props.user.id as string, 23)
			: '';
	}
});
</script>

<style scoped lang="scss">
.terminus-account-root {
	width: 100%;
	// height: 64px;
	border-radius: 8px;
	&.bordered {
		border: 1px solid $separator;
	}
	&__img {
		width: 40px;
		height: 40px;
		border-radius: 20px;
		background: $background-3;
	}
}
</style>
