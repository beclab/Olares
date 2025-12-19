<template>
	<div
		class="card_wrap"
		:class="{
			'select-web': !isMobile,
			'select-not': !selectIds
		}"
	>
		<q-card
			clickable
			v-ripple
			@click="selectItem(vaultItem as ListItem)"
			:active="isSelected(vaultItem as ListItem)"
			active-class="text-blue"
			flat
			bordered
			class="vaultCard col-6 q-mx-md"
			:class="isSelected(vaultItem as ListItem) ? 'vaultCardActive' : ''"
			@mouseenter="handleMouseEnter(vaultItem.item.id)"
			@mouseleave="handleMouseLeave"
		>
			<q-card-section
				class="row items-center justify-between q-pa-none q-mb-xs"
			>
				<div
					class="field-name"
					:style="{
						width: `calc(100% - ${showTags(vaultItem.item).tagWidth}px)`
					}"
				>
					<q-icon
						:class="vaultItem.item.class"
						:name="showItemIcon(vaultItem.item.icon)"
						size="24px"
						color="ink-1"
					/>
					<div class="item-name q-ml-sm">
						<div class="label text-body3 text-ink-3">
							{{ vaultItem.vault.name }}
						</div>
						<div class="name text-subtitle2 text-ink-1">
							{{ vaultItem.item.name ? vaultItem.item.name : t('new_item') }}
						</div>
					</div>
				</div>
				<div
					class="tag-wrap text-ink-2"
					:style="{
						width: `${showTags(vaultItem.item).tagWidth}px`
					}"
				>
					<div
						class="tag q-mr-sm"
						v-for="(tag, index) in showTags(vaultItem.item).tags"
						:key="index"
					>
						<q-icon :name="tag.icon" />
						<span class="q-ml-xs tag-name text-overline" v-if="tag.name">{{
							tag.name
						}}</span>
					</div>
				</div>
			</q-card-section>
			<q-icon
				class="west"
				name="sym_r_arrow_circle_left"
				size="24px"
				v-if="
					isHoveredId === vaultItem.item.id &&
					arrowItemObj[vaultItem.item.id]?.left
				"
				@click.stop="moveItem($event, 'west')"
			/>

			<q-icon
				class="east"
				name="sym_r_arrow_circle_right"
				size="24px"
				v-if="
					isHoveredId === vaultItem.item.id &&
					arrowItemObj[vaultItem.item.id]?.right
				"
				@click.stop="moveItem($event, 'east')"
			/>

			<q-scroll-area
				ref="vaultItemRef"
				:thumb-style="{ height: '0px' }"
				:visible="true"
				style="height: 54px; width: 100%; padding: 0 20px"
				@scroll="scrollItem($event, vaultItem.item.id)"
			>
				<q-card-section horizontal>
					<div
						v-for="(filed, index2) in vaultItem.item.fields"
						class="item-unit cursor-pointer q-px-sm q-py-xs"
						:key="`f` + index2"
					>
						<div class="text-light-blue-default item-header">
							<q-icon :name="filed.icon" size="20px" />
							<span class="text-body3 q-ml-xs">
								{{ translate(filed.name) }}
							</span>
						</div>
						<div
							v-if="filed.value"
							class="text-ink-2 text-left item-unit-content q-ml-xs"
						>
							<span v-if="filed.type === 'totp'">
								<Totp :secret="filed.value" ref="myTotps" />
							</span>
							<span class="text-body3" v-else>
								{{ filed.format(true) }}
							</span>
						</div>
						<div v-else class="text-body3">[{{ t('empty') }}]</div>
						<div
							class="hideCopied text-body3 text-ink-1"
							v-if="filed.value"
							@click="copyItem(filed, $event)"
						>
							<q-icon name="sym_r_check_circle" size="16px" class="q-mr-xs" />
							{{ t('copied') }}
						</div>
					</div>

					<div
						v-for="(filed, index2) in vaultItem.item.attachments"
						class="item-unit cursor-pointer q-px-sm q-py-xs"
						:key="`f` + index2"
					>
						<div class="text-light-blue-default item-header">
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
						<div v-else class="text-body3">[{{ t('empty') }}]</div>
					</div>

					<div
						v-if="
							vaultItem.item &&
							vaultItem.item.fields.length <= 0 &&
							vaultItem.item.attachments.length <= 0
						"
						style="height: 54px; line-height: 44px"
					>
						{{ t('vault_t.no_fields') }}
					</div>
				</q-card-section>
			</q-scroll-area>
		</q-card>
		<!-- <q-separator v-if="!$q.platform.is.mobile" /> -->
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, PropType } from 'vue';
import { useQuasar } from 'quasar';
import { useRoute } from 'vue-router';

import { getPlatform } from '@didvault/sdk/src/core';
import { app } from '../../globals';
import { ListItem } from '@didvault/sdk/src/types';
import Totp from './totp.vue';
import { showItemIcon } from './../../utils/utils';

import { notifyFailed } from '../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import { translate } from '@didvault/sdk/src/util';
import { format } from '../../utils/format';

defineProps({
	vaultItem: {
		type: Object as PropType<ListItem>,
		required: true
	},
	selectIds: {
		type: Array as PropType<string[]>,
		required: false
	}
});

const emits = defineEmits(['selectItem']);

const Route = useRoute();
const vaultItemRef = ref();
const arrowItemObj = ref({});
const isHoveredId = ref();
const myTotps = ref();

const { t } = useI18n();
const $q = useQuasar();
const isMobile = ref(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile);

async function selectItem(item: ListItem) {
	emits('selectItem', item);
}

function isSelected(item: ListItem): boolean {
	if (item && item.item.id === Route.params.itemid) {
		return true;
	}
	return false;
}

onMounted(async () => {});
onUnmounted(() => {});

const copyItem = (value: any, e: any) => {
	e.stopPropagation();
	let copyTxt = value.format(true);
	if (
		[
			'password',
			'pin',
			'apiSecret',
			'mnemonic',
			'cryptoaddress',
			'credit'
		].includes(value.type)
	) {
		copyTxt = value.value;
	}
	if (value.type === 'totp') {
		copyTxt = myTotps.value[0].token;
	}
	const fieldEl = e.target as HTMLElement;
	fieldEl.classList.add('copied');
	setTimeout(() => fieldEl.classList.remove('copied'), 1000);
	getPlatform()
		.setClipboard(copyTxt)
		.catch((e) => {
			notifyFailed(
				t('copy_failure_message', {
					message: e.message
				})
			);
		});
};
const scrollItem = (e: any, id: string) => {
	const scrollLeft = e.horizontalPosition;
	const scrollRight =
		e.horizontalSize - e.horizontalContainerSize - e.horizontalPosition + 40;

	arrowItemObj.value[id] = {
		left: false,
		right: true
	};
	if (scrollLeft < 20) {
		arrowItemObj.value[id].left = false;
	} else {
		arrowItemObj.value[id].left = true;
	}
	if (scrollRight < 10) {
		arrowItemObj.value[id].right = false;
	} else {
		arrowItemObj.value[id].right = true;
	}
};
const moveItem = (_e: any, direction: string) => {
	if (direction === 'west') {
		vaultItemRef.value.setScrollPosition('horizontal', 0);
	} else {
		vaultItemRef.value.setScrollPosition('horizontal', 10000000);
	}
};

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

const handleMouseEnter = (id: string) => {
	isHoveredId.value = id;
};

const handleMouseLeave = () => {
	isHoveredId.value = null;
};
</script>

<style lang="scss" scoped>
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
		margin-bottom: 4px;
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
		border-radius: 8px;

		display: flex;
		align-items: center;
		justify-content: center;
		background: $grey-2;
		border: 1px solid $separator;
	}

	.copied {
		opacity: 1;
		border: 1px solid $yellow-default;
		background: $background-1;

		&:after {
			width: 100%;
			height: 100%;
			background: $yellow-alpha;
			content: '';
			position: absolute;
			top: 0;
			left: 0;
			z-index: 1;
		}
	}
}

.card_wrap {
	width: 100%;
	// border-bottom: 1px solid $separator;
	padding: 8px 0;
	position: relative;
	&::before {
		content: '';
		position: absolute;
		bottom: 0;
		right: 0px;
		width: 100%;
		height: 1px;
		background-color: $separator;
		transition: width 0.5s ease;
	}

	&.select-not::before {
		width: calc(100% - 36px);
	}

	&.select-web::before {
		width: calc(100% + 80px);
		right: -20px;
	}
}

.vaultCard {
	margin: 0px;
	border: 0;
	box-sizing: border-box;
	position: relative;
	padding: 8px;

	.tag-wrap {
		display: flex;
		align-items: center;
		justify-content: flex-end;

		.tag {
			border: 1px solid $ink-2;
			padding: 0 4px;
			border-radius: 4px;
			float: right;
			height: 20px;
			line-height: 20px;
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

	.west {
		position: absolute;
		left: 10px;
		bottom: 18px;
		margin: auto;
		z-index: 1;
		color: $ink-2;
		cursor: pointer;

		&:hover {
			color: $ink-1;
		}
	}

	.east {
		border-radius: 14px;
		position: absolute;
		right: 10px;
		bottom: 18px;
		margin: auto;
		z-index: 1;
		color: $ink-2;
		cursor: pointer;

		&:hover {
			color: $ink-1;
		}
	}
}
</style>
