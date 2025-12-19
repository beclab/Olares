<template>
	<div class="nft-background">
		<!--		<div v-if="userStore.cloud_login == CloudLoginState.YES">-->
		<!--			<div v-if="userStore.nfts?.length == 0">-->
		<!--				{{ t('profile.bind_nft_in_settings') }}-->
		<!--			</div>-->
		<!--			<div v-else>-->
		<!--				<div class="default-avatar-grid row">-->
		<!--					<template v-for="item in userStore.nfts" :key="item">-->
		<!--						<bio-avatar-selector-->
		<!--							:src="item.imageUrl"-->
		<!--							:selected="modelValue.avatar === item.vp"-->
		<!--							@click="onItemClick(item)"-->
		<!--						/>-->
		<!--					</template>-->
		<!--				</div>-->
		<!--			</div>-->
		<!--		</div>-->
		<div class="column justify-center items-center" style="height: 100%">
			<q-icon size="32px" name="sym_r_imagesmode" color="ink-2" />
			<div class="nft-image-title">
				{{ t('profile.go_to_the_settings_to_complete_the_binding_desc') }}
			</div>
			<div
				class="nft-bind-border row justify-center items-center cursor-pointer"
				@click="goBindCloud"
			>
				<q-icon color="ink-2" size="16px" name="sym_r_add_photo_alternate" />
				<div class="nft-bind-label">{{ t('profile.go_to_binding') }}</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { bus } from '@apps/profile/src/utils/bus';
import { useI18n } from 'vue-i18n';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
const { t } = useI18n();

defineProps({
	modelValue: {
		type: Object,
		required: true
	}
});

const onItemClick = (nft: any) => {
	console.log(nft);
	bus.emit('choice', { imageUrl: nft.imageUrl, avatar: nft.vp });
};

const goBindCloud = () => {
	BtNotify.show({
		type: NotifyDefinedType.WARNING,
		message: 'Not supported temporarily'
	});
};
</script>

<style scoped lang="scss">
.nft-background {
	width: 100%;
	height: 100%;

	.nft-image-title {
		width: 256px;
		font-family: Roboto;
		font-size: 12px;
		font-weight: 400;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: center;
		color: $ink-2;
		margin-top: 8px;
	}

	.nft-bind-border {
		margin-top: 12px;
		border: 1px solid $green-default;
		padding: 8px 21px;
		border-radius: 8px;

		.nft-bind-label {
			margin-left: 8px;
			font-family: Roboto;
			font-size: 12px;
			font-weight: 400;
			line-height: 16px;
			letter-spacing: 0em;
			text-align: left;
			color: $ink-2;
		}
	}
}
</style>
