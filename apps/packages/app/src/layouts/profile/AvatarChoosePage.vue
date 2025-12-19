<template>
	<div class="column full-width bg-ink-1 full-height">
		<div class="avatar-dialog q-pa-lg bg-background-1">
			<div class="avatar-title text-ink-1 text-subtitle2">
				{{ t('profile.profile_picture') }}
			</div>

			<bt-scroll-area class="full-width" style="height: calc(100vh - 157px)">
				<div class="column full-width full-height">
					<div class="row justify-center items-center">
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

					<div class="avatar-crop-view row justify-center items-center q-mt-lg">
						<profile-avatar-board size="184px" :show-border="tab === 'nft'">
							<q-img
								v-if="selected"
								class="avatar-preview"
								:src="selected.imageUrl"
							/>
						</profile-avatar-board>
					</div>
					<div class="avatar-choose-view">
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
					</div>
				</div>
			</bt-scroll-area>

			<div class="bottom-button row justify-between items-center">
				<bio-button
					width="161px"
					height="48px"
					:default-selected="false"
					class="text-subtitle2"
					:label="t('base.cancel')"
					@click="router.back"
				/>
				<bio-button
					width="161px"
					height="48px"
					class="text-subtitle2"
					:label="t('base.confirm')"
					@click="onOKClick"
				/>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import ProfileAvatarBoard from '@apps/profile/src/components/profile/avatar/ProfileAvatarBoard.vue';
import DefaultAvatar from '@apps/profile/src/components/profile/avatar/DefaultAvatar.vue';
import UploadAvatar from '@apps/profile/src/components/profile/avatar/UploadAvatar.vue';
import NFTAvatar from '@apps/profile/src/components/profile/avatar/NFTAvatar.vue';
import BtTabItem from '@apps/profile/src/components/profile/base/BtTabItem.vue';
import BioButton from '@apps/profile/src/components/profile/base/BioButton.vue';
import { onMounted, onUnmounted, ref } from 'vue';
import { useUserStore } from '@apps/profile/src/stores/user';
import { useRouter } from 'vue-router';
import { bus } from '@apps/profile/src/utils/bus';
import { useI18n } from 'vue-i18n';

const userStore = useUserStore();
const selected = ref();
const tab = ref<string>('default');
const { t } = useI18n();
const router = useRouter();

async function onOKClick() {
	if (selected.value.avatar != userStore.info.avatar) {
		userStore.info.avatar = selected.value.avatar;
	}
	router.back();
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

<style scoped lang="scss">
.avatar-dialog {
	margin-top: 11px;
	height: calc(100% - 11px);
	width: 100%;
	border-radius: 30px 30px 0 0;
	overflow: hidden;

	.avatar-title {
		padding-top: 10px;
		padding-bottom: 10px;
		height: 44px;
		text-align: center;
	}

	.avatar-choose-view {
		margin-top: 20px;
		border-radius: 8px;
		border: 1px solid $separator-2;
		height: auto;
		width: 100%;
		margin-right: 6px;
	}

	.avatar-crop-view {
		background: $background-3;
		width: 100%;
		border-radius: 8px;
		height: 335px;

		.avatar-preview {
			width: 100%;
			height: 100%;
			border-radius: 50%;
		}
	}

	.bottom-button {
		padding-top: 20px;
		width: 100%;
		background: $background-1;
		bottom: 20px;
	}
}
</style>
