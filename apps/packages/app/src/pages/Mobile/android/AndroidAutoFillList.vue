<template>
	<div class="auto-fill-root">
		<div
			class="auto-fill-title row justify-between items-center q-px-md q-py-sm"
		>
			<div class="auto-fill-label text-ink-1">{{ t('Vault') }}</div>
			<div
				class="auto-fill-add-background row justify-center items-center"
				@click="onAddClicked"
			>
				<q-icon size="24px" name="sym_r_add" />
			</div>
		</div>
		<q-scroll-area class="auto-fill-scroll" v-if="itemListRef.length > 0">
			<div
				style="width: 100%"
				v-for="(item, index) in itemListRef"
				:key="item.vault.id"
			>
				<q-card
					clickable
					v-ripple
					@click="onItemClicked(item)"
					:active="true"
					active-class="text-blue"
					flat
					bordered
					class="vaultCard bg-grey-1 col-6 q-mx-lg q-mb-md"
				>
					<q-card-section
						class="row items-center justify-between q-pa-none q-mb-xs"
					>
						<div
							class="field-name"
							:style="{
								width: `calc(100% - ${showTags(item.item).tagWidth}px)`
							}"
						>
							<q-icon
								:class="item.class"
								:name="showItemIcon(item.item.icon)"
								size="24px"
							/>
							<div class="item-name q-ml-sm">
								<div class="label text-body3 text-ink-1">
									{{ item.vault.name }}
								</div>
								<div class="name text-subtitle2">
									{{ item.item.name ? item.item.name : t('new_item') }}
								</div>
							</div>
						</div>
						<div
							class="tag-wrap"
							:style="{
								width: `${showTags(item.item).tagWidth}px`
							}"
						>
							<div
								class="tag q-mr-sm"
								v-for="(tag, index) in showTags(item.item).tags"
								:key="index"
							>
								<q-icon :name="tag.icon" />
								<span class="q-ml-xs tag-name" v-if="tag.name">{{
									tag.name
								}}</span>
							</div>
						</div>
					</q-card-section>

					<q-icon
						class="west"
						name="sym_r_arrow_circle_left"
						size="24px"
						v-if="arrowItemObj[item.item.id]"
						@click.stop="moveItem($event, 'west', index)"
					/>

					<q-icon
						class="east"
						name="sym_r_arrow_circle_right"
						size="24px"
						v-if="arrowItemObj[item.item.id]"
						@click.stop="moveItem($event, 'east', index)"
					/>

					<q-scroll-area
						ref="vaultItemRef"
						:visible="true"
						:thumb-style="contentStyle as any"
						style="height: 54px; width: 100%"
						@scroll="scrollItem($event, item.item.id)"
					>
						<q-card-section horizontal>
							<div
								v-for="(filed, index2) in item.item.fields"
								class="item-unit cursor-pointer q-px-sm q-py-xs"
								:key="`f` + index2"
							>
								<div class="text-blue-4 item-header">
									<q-icon :name="filed.icon" size="20px" />
									<span class="text-caption text-body1 q-ml-xs">
										{{ filed.name }}
									</span>
								</div>
								<div
									v-if="filed.value"
									class="text-grey-9 text-left item-unit-content q-ml-xs"
								>
									<span v-if="filed.type === 'totp'">
										<Totp :secret="filed.value" ref="myTotps" />
									</span>
									<span v-else>
										{{ filed.format(true) }}
									</span>
								</div>
								<div v-else class="text-color-title">[{{ t('empty') }}]</div>
								<div class="hideCopied text-body3" v-if="filed.value" />
							</div>

							<div
								v-for="(filed, index2) in item.item.attachments"
								class="item-unit cursor-pointer q-px-sm q-py-xs"
								:key="`f` + index2"
							>
								<div class="text-blue-4 item-header">
									<q-icon name="sym_r_attach_file" size="20px" />
									<span class="text-caption text-body1 q-ml-xs">
										{{ filed.name }}
									</span>
								</div>
								<div
									v-if="filed.size"
									class="text-grey-9 text-left item-unit-content"
								>
									<span v-if="filed.type === 'totp'">
										<Totp :secret="filed.value" ref="myTotps" />
									</span>
									<span v-else>
										{{ format.humanStorageSize(filed.size) }}
									</span>
								</div>
								<div v-else class="text-color-title">[{{ t('empty') }}]</div>
							</div>

							<div
								v-if="
									item.item &&
									item.item.fields.length <= 0 &&
									item.item.attachments.length <= 0
								"
							>
								<div class="text-body2">
									{{ t('vault_t.no_fields') }}
								</div>
								<div class="text-caption">
									{{ t('vault_t.this_item_has_no_fields') }}
								</div>
							</div>
						</q-card-section>
					</q-scroll-area>
				</q-card>
			</div>
		</q-scroll-area>

		<div class="auto-fill-empty column justify-center items-center" v-else>
			<img src="../../../assets/layout/nodata.svg" />
			<span class="q-mb-md text-ink-1" style="margin-top: 32px">
				{{ t('vault_t.no_matching_vault') }}
			</span>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ListItem } from '@didvault/sdk/src/types';
import { useRoute, useRouter } from 'vue-router';
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { FieldType, VaultItem } from '@didvault/sdk/src/core';
import { app } from '../../../globals';
import { updateUIToAddWeb } from '../../../platform/addItem';
import { busOn, busOff } from '../../../utils/bus';
import { useI18n } from 'vue-i18n';
import { format } from '../../../utils/format';
import { showItemIcon } from '../../../utils/utils';
import { AUTO_FILL_TYPE } from '../../../platform/capacitor/android/platform';
import AndroidPlugins from 'src/platform/interface/capacitor/android/androidPlugins';

const router = useRouter();
const route = useRoute();
const { t } = useI18n();
const uriRef = ref<string>('');
const isFrameworkRef = ref<boolean>(
	!!(route.query.type as string) &&
		(route.query.type as string) == AUTO_FILL_TYPE.FRAMEWORK
);
const contentStyle = ref({
	height: 0
});
const searchDataRef = ref('');
const isSearchModeRef = ref(false);
const itemListRef = ref<ListItem[]>([]);
const vaultItemRef = ref();
const arrowItemObj = ref({});

watch(
	() => route.query.uri,
	(newValue) => {
		if (newValue) {
			uriRef.value = newValue as string;
			itemListRef.value = _getItems();
		}
	},
	{
		immediate: true
	}
);

const scrollItem = (e: any, id: string) => {
	const scrollLeft = e.horizontalPosition;
	if (scrollLeft < 20) {
		arrowItemObj.value[id] = false;
	} else {
		arrowItemObj.value[id] = true;
	}
};
const moveItem = (_e: any, direction: string, index: number) => {
	if (direction === 'west') {
		vaultItemRef.value[index].setScrollPosition('horizontal', 0);
	} else {
		vaultItemRef.value[index].setScrollPosition('horizontal', 10000000);
	}
};

const onBackClick = () => {
	if (isFrameworkRef.value) {
		AndroidPlugins.AndroidUniversal.finish();
	} else {
		router.back();
	}
};

onMounted(() => {
	busOn('backButton', onBackClick);
});

onUnmounted(() => {
	busOff('backButton', onBackClick);
});

interface TagInter {
	name: string;
	icon: string;
	class: string;
}

const showTags = (item: any) => {
	const tags: TagInter[] = [];
	let tagWidth = 0;
	if (item.tags.length) {
		tags.push({
			icon: 'sym_r_style',
			name: item.tags[0],
			class: ''
		});
		tagWidth += 80;
		if (item.tags.length > 1) {
			tags.push({
				icon: 'sym_r_style',
				name: `+${item.tags.length - 1}`,
				class: ''
			});
			tagWidth += 54;
		}
	}
	const attCount = (item.attachments && item.attachments.length) || 0;
	if (attCount) {
		tags.push({
			name: attCount.toString(),
			icon: 'sym_r_attach_file',
			class: ''
		});
		tagWidth += 42;
	}
	if (app.account!.favorites.has(item.id)) {
		tags.push({
			name: '',
			icon: 'sym_r_grade',
			class: 'text-red'
		});
		tagWidth += 32;
	}

	return {
		tags,
		tagWidth
	};
};

// function getPassword(item: ListItem) {
// 	const value = getFieldValue(item, FieldType.Password);
// 	let password = '';
// 	// eslint-disable-next-line @typescript-eslint/no-unused-vars
// 	for (const _i in value as any) {
// 		password = password + '*';
// 	}
// 	return password;
// }

function _getItems(): ListItem[] {
	let items: ListItem[];

	items = app.state.vaults.flatMap((vault) =>
		[...vault.items]
			.filter((item) =>
				filterByString(
					isSearchModeRef.value ? searchDataRef.value : uriRef.value,
					item
				)
			)
			.map((item) => ({ vault, item }))
	);

	items.sort((a, b) => {
		const x = a.item.name.toLowerCase();
		const y = b.item.name.toLowerCase();
		return x > y ? 1 : x < y ? -1 : 0;
	});

	return items;
}

function filterByString(uri: string, rec: VaultItem) {
	if (rec.icon !== 'web') {
		return false;
	}

	const userName = rec.fields.find((v) => v.type === FieldType.Username);
	const password = rec.fields.find((v) => v.type === FieldType.Password);

	if (password && userName) {
		const content = [...rec.fields.map((f) => f.value)].join(' ').toLowerCase();
		return content.search(uri.toLowerCase()) !== -1;
	}
	return false;
}

// function getFieldValue(item: ListItem, type: FieldType) {
// 	const field = item.item.fields.find((v) => v.type === type);
// 	return field ? field.value : '';
// }

function onItemClicked(item: ListItem) {
	const userName = item.item.fields.find((v) => v.type === FieldType.Username);
	const password = item.item.fields.find((v) => v.type === FieldType.Password);

	if (userName && password) {
		if (isFrameworkRef.value) {
			AndroidPlugins.AutofillFramework.setResult({
				uri: uriRef.value,
				userName: userName.value,
				password: password.value
			});
			AndroidPlugins.AndroidUniversal.finish();
		} else {
			AndroidPlugins.Accessibility.setResult({
				uri: uriRef.value,
				userName: userName.value,
				password: password.value
			});
			router.back();
			AndroidPlugins.AndroidUniversal.moveTaskToBack();
		}
	}
}

function onAddClicked() {
	updateUIToAddWeb(uriRef.value, router);
}
</script>

<style lang="scss" scoped>
.auto-fill-root {
	height: 100%;
	width: 100%;

	.auto-fill-title {
		height: 56px;
		width: 100%;

		.auto-fill-label {
			font-family: Roboto;
			font-size: 20px;
			font-weight: 800;
			line-height: 28px;
			text-align: left;
		}

		.auto-fill-add-background {
			width: 40px;
			height: 40px;
		}
	}

	.auto-fill-scroll {
		height: calc(100% - 56px);
		width: 100%;

		.vaultCard {
			border: 0;
			box-sizing: border-box;
			position: relative;
			padding: 12px 12px 0px 12px;
			border-radius: 12px;

			.tag-wrap {
				display: flex;
				align-items: center;
				justify-content: flex-end;

				.tag {
					border: 1px solid $grey-8;
					padding: 0 4px;
					border-radius: 4px;
					float: right;
					height: 26px;
					line-height: 26px;
					display: flex;
					align-items: center;
					justify-content: flex-start;

					.tag-name {
						max-width: 60px;
						overflow: hidden;
						text-overflow: ellipsis;
						white-space: nowrap;
					}
				}
			}

			.field-name {
				display: flex;
				align-items: center;
				justify-content: flex-start;

				.item-name {
					overflow: hidden;
					white-space: nowrap;
					text-overflow: ellipsis;

					.label {
						overflow: hidden;
						white-space: nowrap;
						text-overflow: ellipsis;
					}

					.name {
						overflow: hidden;
						white-space: nowrap;
						text-overflow: ellipsis;
					}
				}
			}

			.item-unit {
				border-radius: 5px;
				margin-right: 4px;
				white-space: nowrap;
				position: relative;
				max-width: 180px;
				min-width: 70px;
				overflow: hidden;
				white-space: nowrap;
				text-overflow: ellipsis;

				.item-header {
					width: 100%;
					overflow: hidden;
					white-space: nowrap;
					text-overflow: ellipsis;
				}

				.item-unit-content {
					line-height: 1 !important;
					white-space: nowrap;

					span {
						display: inline-block;
						width: 100%;
						overflow: hidden;
						text-overflow: ellipsis;
						white-space: nowrap;
						font-size: 14px;
					}
				}

				.hideCopied {
					position: absolute;
					width: 100%;
					height: 100%;
					left: 0;
					top: 0;
					opacity: 0;
					border-radius: 5px;
					color: $grey-10;

					display: flex;
					align-items: center;
					justify-content: center;
					background: $grey-2;
					border: 1px solid $separator;
				}
			}

			.west {
				position: absolute;
				left: 10px;
				bottom: 18px;
				margin: auto;
				z-index: 1;
				color: $grey-13;
				cursor: pointer;

				&:hover {
					color: $grey-10;
				}
			}

			.east {
				border-radius: 14px;
				position: absolute;
				right: 10px;
				bottom: 18px;
				margin: auto;
				z-index: 1;
				color: $grey-13;
				cursor: pointer;

				&:hover {
					color: $grey-10;
				}
			}
		}
	}

	.auto-fill-empty {
		width: 100%;
		height: calc(100% - 44px);

		.empty-img {
			width: 72px;
			height: 102px;
		}

		.empty_info {
			margin-top: 37px;
			text-align: center;
		}

		.empty_btn {
			margin-top: 8px;
			width: 126px;
			height: 32px;
		}
	}
}
</style>
../../../platform/addItem
