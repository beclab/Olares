<template>
	<q-dialog ref="dialogRef">
		<q-card class="avatar-choose-dialog">
			<div class="avatar-choose-header row justify-between items-center">
				<div class="avatar-choose-title">{{ t('profile.avatar_setting') }}</div>
				<q-icon
					size="16px"
					name="sym_r_clear"
					color="ink-3"
					@click="onDialogCancel"
				/>
			</div>
			<div class="row justify-center">
				<div class="avatar-choose-view">
					<q-card-section style="padding: 0">
						<div
							class="row justify-center items-center"
							style="margin-top: 20px"
						>
							<bt-tab-item
								class="cursor-pointer"
								:selected="tab === 'default'"
								:label="t('profile.robot_avatar')"
								@click="tab = 'default'"
							/>
							<bt-tab-item
								class="cursor-pointer"
								:selected="tab === 'nft'"
								:label="t('profile.nft_avatar')"
								@click="tab = 'nft'"
							/>
							<bt-tab-item
								class="cursor-pointer"
								:selected="tab === 'upload'"
								:label="t('profile.custom_avatar')"
								@click="tab = 'upload'"
							/>
						</div>

						<q-tab-panels
							v-model="tab"
							animated
							style="width: 100%; height: 315px"
							keep-alive
						>
							<q-tab-panel name="default" style="width: 100%; height: 100%">
								<default-avatar v-model="selected" />
							</q-tab-panel>

							<q-tab-panel name="upload">
								<upload-avatar v-model="selected" />
							</q-tab-panel>

							<q-tab-panel name="nft">
								<n-f-t-avatar v-model="selected" />
							</q-tab-panel>
						</q-tab-panels>
					</q-card-section>
				</div>
				<div class="avatar-crop-view row justify-center items-center">
					<profile-avatar-board size="184px" :show-border="tab === 'nft'">
						<q-img class="avatar-preview" :src="selected.imageUrl" />
					</profile-avatar-board>
				</div>
			</div>

			<div class="row justify-end items-center q-mt-xl">
				<bio-button
					:default-selected="false"
					class="text-body3 q-mr-lg"
					:label="t('base.cancel')"
					@click="onDialogCancel"
				/>
				<bio-button
					width="75px"
					class="text-body3 q-mr-lg"
					:label="t('base.create')"
					@click="onOKClick"
				/>
			</div>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import ProfileAvatarBoard from './ProfileAvatarBoard.vue';
import BioButton from '@apps/profile/src/components/profile/base/BioButton.vue';
import { useDialogPluginComponent } from 'quasar';
import { ref, onMounted, onUnmounted } from 'vue';
import DefaultAvatar from './DefaultAvatar.vue';
import UploadAvatar from './UploadAvatar.vue';
import NFTAvatar from './NFTAvatar.vue';
import { bus } from '@apps/profile/src/utils/bus';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import BtTabItem from '../base/BtTabItem.vue';
import { useI18n } from 'vue-i18n';

const userStore = useUserStore();
const selected = ref();
const tab = ref<string>('default');
const { t } = useI18n();

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

async function onOKClick() {
	if (selected.value.avatar != userStore.info.avatar) {
		userStore.info.avatar = selected.value.avatar;
	}
	onDialogOK(selected.value.avatar);
}

onMounted(async () => {
	bus.on('choice', (data: any) => {
		console.log(data);
		selected.value = data;
	});

	selected.value = userStore.loadAvatar();
});

onUnmounted(() => {
	bus.off('choice');
});
</script>
<style lang="scss">
.avatar-choose-dialog {
	max-width: 800px;
	width: 800px;
	height: 500px;
	max-height: 500px;
	background: $background-1;
	padding: 0;
	position: relative;

	.avatar-choose-header {
		width: 100%;
		height: 32px;
		background: $background-3;
		padding-left: 12px;
		padding-right: 12px;

		.avatar-choose-title {
			font-family: Roboto;
			font-size: 12px;
			font-weight: 500;
			line-height: 16px;
			letter-spacing: 0;
			text-align: left;
			color: $ink-1;
		}
	}

	.avatar-choose-view {
		margin-top: 20px;
		border-radius: 8px;
		border: 1px solid $separator-2;
		height: 364px;
		width: 374px;
		margin-right: 6px;
	}

	.avatar-crop-view {
		margin-top: 20px;
		border-radius: 8px;
		border: 1px solid $separator-2;
		height: 364px;
		width: 374px;
		margin-left: 6px;
		background: $background-3;

		.avatar-preview {
			width: 100%;
			height: 100%;
			border-radius: 50%;
		}
	}

	.loading-view {
		width: 100%;
		height: 100%;
		position: absolute;
	}
}

.q-dialog__inner--minimized > div {
	max-width: 800px;
}
</style>
