<template>
	<div class="index-root row justify-center bg-background-1">
		<div v-if="!userStore.user" class="index-loading">
			<BtLoading
				:show="true"
				textColor="#ffffff"
				color="#ffffff"
				text="Loading"
				backgroundColor="rgba(0, 0, 0, 0.4)"
			/>
		</div>
		<div v-else class="index-layout row justify-center">
			<div class="left-editor column items-center">
				<div class="editor-top column">
					<div
						v-if="userStore.isMobile"
						class="mobile-brand row justify-center items-center"
					>
						<q-img class="brand-image" src="icons/profile-128*128.png" />
						<div class="brand-title text-ink-1 q-ml-md">
							{{ t('base.profile') }}
						</div>
					</div>
					<div class="editor-title row justify-center">
						<q-tabs
							v-if="tab && tabs"
							:model-value="tab"
							@update:modelValue="onTabUpdate"
							dense
							inline-label
							narrow-indicator
							active-color="light-green-default"
							indicator-color="light-green-default"
							active-class="tab-item-title-selected"
						>
							<template v-for="item in tabs" :key="item.value">
								<!--							:class="userStore.isMobile ? 'text-subtitle2' : 'text-body2'"-->
								<q-tab
									:name="item.value"
									:label="item.label"
									:icon="userStore.isMobile ? undefined : item.icon"
								/>
							</template>
						</q-tabs>
					</div>
				</div>

				<div class="editor-workspace row justify-center scroll">
					<q-tab-panels v-model="tab" class="full-width" animated keep-alive>
						<q-tab-panel :name="tabs[0].value">
							<profile-page />
						</q-tab-panel>

						<q-tab-panel :name="tabs[1].value">
							<social-page />
						</q-tab-panel>

						<q-tab-panel :name="tabs[2].value">
							<block-page />
						</q-tab-panel>

						<q-tab-panel :name="tabs[3].value">
							<appearance-page />
						</q-tab-panel>
					</q-tab-panels>
				</div>
			</div>

			<div
				v-if="userStore.isMobile"
				class="preview-btn row justify-center items-center bg-tooltip-bg text-white"
				@click="goPreview"
			>
				<q-icon size="24px" name="sym_r_visibility" />
				<div class="text-h6 q-ml-sm">{{ t('base.preview') }}</div>
			</div>

			<div v-else class="right-preview">
				<bio-share-header />
				<div class="preview-background row justify-center">
					<q-responsive :ratio="9 / 18" class="preview">
						<preview-layout :user="userStore.user" :font-size="10.24" />
					</q-responsive>
				</div>
			</div>
		</div>
	</div>
</template>
<script setup lang="ts">
import PreviewLayout from '@apps/profile/src/components/profile/layout/PreviewLayout.vue';
import BioShareHeader from '@apps/profile/src/components/profile/share/BioShareHeader.vue';
import AppearancePage from '@apps/profile/src/pages/profile/AppearancePage.vue';
import ProfilePage from '@apps/profile/src/pages/profile/ProfilePage.vue';
import SocialPage from '@apps/profile/src/pages/profile/SocialPage.vue';
import BlockPage from '@apps/profile/src/pages/profile/BlockPage.vue';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { User } from '@apps/profile/src/types/User';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { debounce } from 'quasar';
import { ref, watch } from 'vue';
import axios from 'axios';

const userStore = useUserStore();
const { t } = useI18n();
const router = useRouter();

const tabs = [
	{
		label: t('base.header'),
		value: 'Header',
		icon: 'sym_r_responsive_layout'
	},
	{
		label: t('base.social'),
		value: 'Social',
		icon: 'sym_r_diversity_2'
	},
	{
		label: t('base.blocks'),
		value: 'Blocks',
		icon: 'sym_r_note_stack'
	},
	{
		label: t('base.appearance'),
		value: 'Appearance',
		icon: 'sym_r_draw'
	}
];

const tab = ref(tabs[0].value);

const onTabUpdate = (value: string) => {
	tab.value = value;
};

const updateUserInfo = debounce(async function (user: User) {
	axios.post('/api/profile', user);
}, 500);

const goPreview = () => {
	router.push({
		path: '/preview'
	});
};

watch(
	() => userStore.user,
	(value: User | null) => {
		if (value) {
			console.log('update user');
			console.log(value);

			updateUserInfo(value);
		}
	},
	{
		deep: true
	}
);
</script>

<style lang="scss">
.index-root {
	width: 100vw;
	height: 100vh;

	.index-loading {
		width: 100vw;
		height: 100vh;
	}

	@media (max-width: 768px) {
		.index-layout {
			width: 100%;
			height: 100vh;
			position: relative;

			.left-editor {
				height: 100vh;
				width: 100%;

				.editor-top {
					width: 100%;
					box-shadow: var(--page-editor-tabs-shadow);

					.mobile-brand {
						height: 56px;
						width: 100%;

						.brand-image {
							width: 24px;
							height: 24px;
						}

						.brand-title {
							font-family: Inter;
							font-size: 16px;
							font-style: normal;
							font-weight: 600;
							line-height: 16px;
						}
					}

					.editor-title {
						height: 56px;
						width: 100%;
					}
				}

				.editor-workspace {
					margin-top: 4px;
					height: calc(100% - 116px);
					width: 100%;
				}
			}
		}

		.preview-btn {
			padding: 10px 32px;
			position: absolute;
			bottom: 32px;
			border-radius: 999px;
			box-shadow: 0 4px 20px 0 rgba(0, 0, 0, 0.1);
		}
	}

	@media (min-width: 768px) {
		.index-layout {
			width: 100%;
			min-width: 810px;
			max-width: 1280px;
			height: 100vh;

			.left-editor {
				height: 100vh;
				width: 60%;
				max-width: 758px;
				min-width: 478px;
				border-right: 1px solid $separator;

				.editor-title {
					height: 56px;
					width: 440px;
				}

				.editor-workspace {
					height: calc(100% - 56px);
					width: 100%;
				}
			}

			.right-preview {
				width: 40%;
				height: 100vh;
				overflow: scroll;
				min-width: 322px;
				max-width: 522px;

				.preview-background {
					width: 100%;
					background-image: url('/public/profile/preview_phone_background.png');
					background-size: cover;
					background-position: center;
					padding-top: 32px;
					padding-left: 19%;
					padding-right: 19%;

					.preview {
						position: relative;
						width: 100%;
						border: 8px solid #000;
						border-radius: 32px;
						background: #fff;
						overflow: hidden;
						box-shadow: inset 0 0.1rem 0.4rem rgba(0, 0, 0, 0.25);
						filter: drop-shadow(0 0.8rem 2.4rem rgba(0, 0, 0, 0.25));
					}
				}
			}
		}
	}
}
</style>
