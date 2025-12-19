<template>
	<div v-if="user" :style="parentStyle" class="container-wrapper">
		<q-scroll-area class="full-width full-height">
			<div
				class="preview-container column items-center"
				:class="
					user.header.style === HEADER_STYLE_TYPE.PORTRAIT
						? ''
						: 'preview-container-padding'
				"
			>
				<header-banner :header="user.header" />
				<div
					v-if="user.appearance.theme.style === THEME_TYPE.IMAGE"
					class="overlay"
					:style="backgroundOverlay"
				/>
				<div
					class="column items-center full-width"
					:style="{ maxWidth: pc ? '37.5em' : '', zIndex: '1' }"
				>
					<header-preview
						:header="user.header"
						:social="user.social"
						:text-color="user.appearance.theme.header.textColor"
					/>

					<div v-if="isDefault" class="default-view">
						<img src="/profile/profile-default.svg" alt="logo" />
						<div class="content">
							{{ t('base.let_people_own_their_data_again') }}
						</div>
					</div>
					<div
						class="blocks-margin full-width"
						:style="{
							paddingLeft:
								user.header.style === HEADER_STYLE_TYPE.PORTRAIT
									? '1.25em'
									: '',
							paddingRight:
								user.header.style === HEADER_STYLE_TYPE.PORTRAIT ? '1.25em' : ''
						}"
					>
						<template v-for="item in enabledLinks" :key="item.id">
							<div
								class="blocks-div row items-center"
								v-if="item.type === BLOCK_TYPE.LINK"
								@click="onOpenWindow(item.url)"
								:style="{
									boxShadow: item.shadow
										? '0px 8px 0px 0px rgba(0, 0, 0, 0.20)'
										: '',
									border: item.outline ? `2px solid ${ink1.color.value}` : '',
									borderRadius: blockRadius,
									background: blockBackground,
									color: user.appearance.theme.link.textColor
								}"
							>
								<q-img
									:class="
										item.size === SIZE_TYPE.SMALL
											? 'link-img-small'
											: 'link-img-larger'
									"
									v-if="item.icon && item.icon !== ''"
									:src="item.icon"
								/>
								<div
									class="link-text column"
									:style="{
										width: item.icon
											? `calc(100% - ${
													item.size === SIZE_TYPE.SMALL ? '4em' : '4.75em'
											  })`
											: '100%',
										textAlign:
											item.textAlignment === ALIGNMENT_TYPE.LEFT
												? 'left'
												: item.textAlignment === ALIGNMENT_TYPE.RIGHT
												? 'right'
												: 'center'
									}"
								>
									<div
										:class="
											item.size === SIZE_TYPE.SMALL
												? 'link-title-small'
												: 'link-title-larger'
										"
									>
										{{ item.title }}
									</div>
									<div
										:class="
											item.size === SIZE_TYPE.SMALL
												? 'link-sub-title-small'
												: 'link-sub-title-larger'
										"
									>
										{{ item.subTitle }}
									</div>
								</div>
							</div>
							<div
								v-if="item.type === BLOCK_TYPE.TEXT"
								class="blocks-div column"
								:style="{
									boxShadow: user.appearance.block.shadow
										? '0px 8px 0px 0px rgba(0, 0, 0, 0.20)'
										: '',
									border: user.appearance.block.outline
										? `2px solid ${ink1.color.value}`
										: '',
									borderRadius: blockRadius,
									background: item.transparent
										? blockBackground
										: user.appearance.theme.block.background,
									color: user.appearance.theme.block.textColor,
									textAlign:
										item.textAlignment === ALIGNMENT_TYPE.LEFT
											? 'left'
											: item.textAlignment === ALIGNMENT_TYPE.RIGHT
											? 'right'
											: 'center'
								}"
							>
								<div class="text-title">
									{{ item.title ? item.title : t('blocks.text_block_title') }}
								</div>
								<div class="text-description">
									{{
										item.description
											? item.description
											: t('blocks.text_block_description')
									}}
								</div>
							</div>
							<div
								class="image-parent column"
								v-if="item.type === BLOCK_TYPE.IMAGE"
							>
								<div
									class="image-title"
									:style="{
										color: user.appearance.theme.block.textColor
									}"
								>
									{{ item.title }}
								</div>
								<div
									class="image-description"
									:style="{
										color: user.appearance.theme.block.textColor
									}"
								>
									{{ item.description }}
								</div>
								<q-img
									:ratio="item.ratio"
									fit="cover"
									:class="item.link ? 'cursor-pointer' : ''"
									:style="{
										marginTop: '0.5em',
										borderRadius: blockRadius,
										boxShadow: user.appearance.block.shadow
											? '0px 8px 0px 0px rgba(0, 0, 0, 0.20)'
											: '',
										border: user.appearance.block.outline
											? `2px solid ${ink1.color.value}`
											: ''
									}"
									@click="onOpenWindow(item.link)"
									:src="item.img ? item.img : 'profile/block_image_default.svg'"
								/>
							</div>
						</template>
					</div>
				</div>

				<slot />
			</div>
		</q-scroll-area>
	</div>
</template>
<script lang="ts" setup>
import { computed, PropType } from 'vue';
import {
	BLOCK_TYPE,
	THEME_TYPE,
	User,
	HEADER_STYLE_TYPE,
	SIZE_TYPE,
	ALIGNMENT_TYPE,
	getGradientColor,
	BLOCK_STYLE_TYPE,
	IMAGE_FILTER,
	LinkBlock
} from '@apps/profile/src/types/User';
import { useI18n } from 'vue-i18n';
import { useColor } from '@bytetrade/ui';
import { colorsRgba } from 'quasar/dist/types/utils/colors';
import { colors } from 'quasar';
import hexToRgb = colors.hexToRgb;
import rgbToHex = colors.rgbToHex;
import HeaderPreview from '@apps/profile/src/components/profile/layout/HeaderPreview.vue';
import HeaderBanner from '@apps/profile/src/components/profile/layout/HeaderBanner.vue';
import { useRoute } from 'vue-router';

const props = defineProps({
	user: {
		type: Object as PropType<User>,
		required: true
	},
	fontSize: {
		type: Number,
		default: 16
	},
	pc: {
		type: Boolean,
		default: false
	}
});

const { t } = useI18n();
const ink1 = useColor('ink-1');
const route = useRoute();

console.log(route.path);

const parentStyle = computed(() => {
	const baseStyle = {
		fontSize: `${props.fontSize}px`,
		fontFamily: props.user.appearance.font
	};
	let backgroundStyle = {};
	switch (props.user.appearance.theme.style) {
		case THEME_TYPE.SOLID:
			backgroundStyle = {
				backgroundColor: props.user.appearance.theme.background
			};
			break;
		case THEME_TYPE.GRADIENT:
			backgroundStyle = {
				background: getGradientColor(
					props.user.appearance.theme.gradientTopColor,
					props.user.appearance.theme.gradientBottomColor
				)
			};
			break;
		case THEME_TYPE.IMAGE:
			if (props.user.appearance.theme.useUpload) {
				backgroundStyle = {
					background: `url("${props.user.appearance.theme.uploadImg}")`,
					backgroundSize: 'cover'
				};
			} else {
				backgroundStyle = {
					backgroundImage: `url("/profile/${props.pc ? 'pc' : 'mobile'}/${
						props.user.appearance.theme.localImg
					}")`,
					backgroundSize: 'cover'
				};
			}
			break;
	}
	return { ...backgroundStyle, ...baseStyle };
});

const backgroundOverlay = computed(() => {
	switch (props.user.appearance.theme.filter) {
		case IMAGE_FILTER.DARK:
			return {
				background: 'rgba(0, 0, 0, 0.3)'
			};
		case IMAGE_FILTER.Light:
			return {
				background: 'rgba(255, 255, 255, 0.30)'
			};
		default:
			return '';
	}
});

const blockRadius = computed(() => {
	switch (props.user.appearance.block.style) {
		case BLOCK_STYLE_TYPE.ROUND:
			return '2em';
		case BLOCK_STYLE_TYPE.SQUARE:
			return '0';
		case BLOCK_STYLE_TYPE.CUSTOM:
			return props.user.appearance.block.cornerRadius / 16 + 'em';
		default:
			return '2em';
	}
});

const blockBackground = computed(() => {
	const rgba: colorsRgba = hexToRgb(
		props.user.appearance.theme.block.background
	);
	rgba.a = 100 - props.user.appearance.block.transparency;
	return rgbToHex(rgba);
});

const enabledLinks = computed(() => {
	return props.user.block.data.filter((item) => {
		if (item.type === BLOCK_TYPE.LINK) {
			if (!(item as LinkBlock).url || !(item as LinkBlock).title) {
				return false;
			}
		}
		return item.enabled;
	});
});

const isDefault = computed(() => {
	return !(props.user.block && props.user.block.data.length > 0);
});

const onOpenWindow = (url: string) => {
	if (url) {
		const hasProtocol = /^https?:\/\//i.test(url);
		const absoluteUrl = hasProtocol ? url : 'https://' + url;
		window.open(absoluteUrl);
	}
};
</script>
<style lang="scss">
::-webkit-scrollbar {
	display: none;
}

.container-wrapper {
	width: 100%;
	height: 100%;
	position: relative;

	.preview-container-padding {
		width: 100%;
		padding: 2.5em 1.25em 0;
	}

	.preview-container {
		width: 100%;
		height: 100%;

		.overlay {
			position: absolute;
			top: 0;
			left: 0;
			width: 100%;
			height: 100%;
			pointer-events: none;
		}

		.blocks-margin {
			width: 100%;
			margin-bottom: 6.25em;

			.image-parent {
				width: 100%;
				margin-top: 2em;

				.image-title {
					max-width: 100%;
					overflow-wrap: break-word;
					word-break: break-all;
					font-size: 1.125em;
					font-style: normal;
					font-weight: 500;
					text-align: center;
					line-height: 1.625em;
				}

				.image-description {
					max-width: 100%;
					overflow-wrap: break-word;
					word-break: break-all;
					white-space: pre-wrap;
					text-align: center;
					font-size: 0.875em;
					font-style: normal;
					font-weight: 400;
					line-height: 1.25em;
				}
			}

			.blocks-div {
				width: 100%;
				margin-top: 2em;
				padding: 1.25em;

				.text-title {
					max-width: 100%;
					overflow-wrap: break-word;
					word-break: break-all;
					font-size: 1.5em;
					font-style: normal;
					font-weight: 700;
					line-height: 2em;
				}

				.text-description {
					max-width: 100%;
					overflow-wrap: break-word;
					word-break: break-all;
					white-space: pre-wrap;
					font-size: 1em;
					font-style: normal;
					font-weight: 400;
					line-height: 1.5em;
				}

				.link-img-small {
					width: 3.25em;
					height: 3.25em;
					border-radius: 50%;
				}

				.link-img-larger {
					width: 4em;
					height: 4em;
					border-radius: 50%;
				}

				.link-text {
					margin-left: 0.75em;

					.link-title {
						max-width: 100%;
						white-space: pre-line;
						overflow-wrap: break-word;
						word-break: break-all;
						width: 100%;
						height: auto;
						cursor: pointer;
					}

					.link-title-small {
						@extend .link-title;
						font-size: 1.25em;
						font-style: normal;
						font-weight: 700;
						line-height: 1.75em;
					}

					.link-title-larger {
						@extend .link-title;
						font-size: 1.5em;
						font-style: normal;
						font-weight: 700;
						line-height: 2em;
					}

					.link-sub-title-small {
						@extend .link-title;
						font-size: 0.875em;
						font-style: normal;
						font-weight: 400;
						line-height: 1.25em;
					}

					.link-sub-title-larger {
						@extend .link-title;
						font-size: 1.125em;
						font-style: normal;
						font-weight: 400;
						line-height: 1.625em;
					}
				}
			}
		}

		.default-view {
			margin-top: 12.5em;
			text-align: center;
		}

		.default-view img {
			width: 3em;
			height: 2.57813em;
			margin-bottom: 1.25em;
		}

		.default-view .content {
			color: #adadad;
			text-align: center;
			font-family: Roboto;
			font-size: 1em;
			font-style: normal;
			font-weight: 400;
			line-height: 1.5em;
		}
	}
}
</style>
