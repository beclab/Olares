<template>
	<div
		:clickable="clickable"
		dense
		class="terminus-account-root row items-center q-pa-sm q-pr-md flex-gap-x-sm"
		:class="{
			bordered: bordered
		}"
		v-ripple="false"
	>
		<terminus-avatar
			v-if="user.name"
			:info="userStore.getUserTerminusInfo(user.id)"
			:size="avatarSize"
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
				<div class="row items-center full-width ellipsis">
					<div
						class="row ellipsis items-center no-wrap flex-gap-x-sm"
						style="text-align: left; flex: 1"
						:class="[user.name ? 'text-ink-1' : 'text-ink-2', userInfoClass]"
					>
						<div class="ellipsis">
							{{ user.name ? user.local_name : t('olares_id_not_created') }}
						</div>
						<TerminusUserStatus2
							v-if="user.id == userStore.current_user?.id"
						></TerminusUserStatus2>
					</div>

					<!-- <div
						style="max-width: 60%"
						v-if="user.name && user.id == userStore.current_user?.id"
					>
						<terminus-user-status class="q-ml-sm" />
					</div> -->
				</div>

				<div
					class="text-ink-3 q-mt-xs"
					:class="[userSubtitleClass]"
					style="text-align: left"
				>
					{{ subInfo }}
				</div>
			</div>
		</div>
		<div class="column justify-center" v-if="$slots.side">
			<slot name="side" />
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted, PropType, ref } from 'vue';
import { UserItem } from '@didvault/sdk/src/core';
import { generateStringEllipsis } from '../../utils/utils';
import { useI18n } from 'vue-i18n';
import { useUserStore } from '../../stores/user';
import TerminusUserStatus2 from 'components/common/TerminusUserStatus2.vue';

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
	},
	size: {
		type: String as PropType<'md' | 'lg'>,
		default: 'md'
	},
	clickable: {
		type: Boolean,
		default: true
	}
});

const subInfo = ref();

const avatarSize = computed(() => {
	if (props.size === 'lg') {
		return 40;
	} else if (props.size === 'md') {
		return 32;
	} else {
		return 32;
	}
});

const userInfoClass = computed(() => {
	if (props.size === 'lg') {
		return 'text-subtitle1';
	} else if (props.size === 'md') {
		return 'text-subtitle3';
	} else {
		return 'text-subtitle3';
	}
});

const userSubtitleClass = computed(() => {
	if (props.size === 'lg') {
		return 'text-body3';
	} else if (props.size === 'md') {
		return 'text-Overline';
	} else {
		return 'text-body3';
	}
});

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
