<template>
	<div class="secret-parent">
		<!-- <terminus-user-header :title="t('secret')" v-if="!hasVCCardRef" />
		<v-c-empty-view class="bottom-content" v-if="!hasVCCardRef" />
		<q-scroll-area
			ref="scrollAreaRef"
			class="secret-root justify-start items-start"
			v-else
		>
			<terminus-user-header :title="t('secret')" />
			<div
				v-if="vcCardHeightRef"
				class="secret-vc"
				:style="'height : ' + vcCardHeightRef + 'px'"
			>
				<v-c-card-container
					class="secret-vc-item"
					:style="`top: calc(${index} * 72px)`"
					v-for="(item, index) in cardItemsRef"
					:item="item"
					:min-size="true"
					@click="goVCCardList"
					:key="item.id"
				/>
				<img
					v-show="cardItemsRef.length > 0"
					class="secret-vc-mask"
					src="../../../assets/secret/vc_card_mask.png"
				/>
			</div>
			<div
				v-if="showListRef"
				class="secret-list-parent column justify-start items-start"
				:style="{ '--reactHeight': topHeightRef + 'px' }"
			>
				<item-list-title-bar
					@onSearch="goManager"
					@onCheckBox="goManager"
					:hide-menu="true"
					:title="t('secret')"
					@onAddItemClick="onItemClicked"
				/>
				<item-list-v2
					:remove-type-array="[VaultType.VC, VaultType.TerminusTotp]"
					@onItemClick="onItemClicked"
					class="secret-list"
				/>
			</div>
			<div class="float-title">
				<item-list-title-bar
					@onSearch="goManager"
					:title="t('secret')"
					@onCheckBox="goManager"
					@onAddItemClick="onItemClicked"
					v-show="showFloatTitleRef"
				/>
			</div>
			<q-scroll-observer @scroll="onScroll" />
		</q-scroll-area> -->
	</div>
</template>

<script lang="ts" setup>
// import { useRouter } from 'vue-router';
import { useMenuStore } from '../../../stores/menu';
import { onMounted, onUnmounted, ref } from 'vue';
import { app } from '../../../globals';
import { debounce, VaultItem, VaultType } from '@didvault/sdk/src/core';
import { useUserStore } from '../../../stores/user';
import { VCCardItem, convertVault2CardItem } from '../../../utils/vc';
import { busOn, busOff } from '../../../utils/bus';

// import { useI18n } from 'vue-i18n';

// const Router = useRouter();
const menuStore = useMenuStore();
const cardItemsRef = ref([] as VCCardItem[]);
const vcCardHeightRef = ref(0);
const hasAuthRef = ref();
const topHeightRef = ref();
const userStore = useUserStore();
// const showFloatTitleRef = ref(false);
// const router = useRouter();
const hasVCCardRef = ref<boolean>(false);
const showListRef = ref(false);
// const scrollAreaRef = ref();
// let isScrolling = false;
// const { t } = useI18n();

userStore.$subscribe(() => {
	updateList();
});

function updateList() {
	const user = userStore.users!.items.get(userStore.current_id!)!;
	if (user.name && !hasVCCardRef.value) {
		hasVCCardRef.value = true;
	}
	showListRef.value = !!user.name && hasVCCardRef.value;
}

// function onItemClicked(itemId: string) {
// 	Router.push({
// 		path: '/items/' + (itemId ? itemId : '')
// 	});
// }

// const goVCCardList = () => {
// 	router.push({
// 		path: '/VC_card_list'
// 	});
// };

// const onScroll = (info: any) => {
// 	if (info.direction === 'down') {
// 		if (info.position.top >= topHeightRef.value) {
// 			showFloatTitleRef.value = true;
// 		} else {
// 			if (isScrolling) {
// 				return;
// 			}
// 			isScrolling = true;
// 			scrollAreaRef.value.setScrollPosition(
// 				'vertical',
// 				topHeightRef.value,
// 				500
// 			);
// 			window.setTimeout(() => {
// 				showFloatTitleRef.value = true;
// 				isScrolling = false;
// 			}, 500);
// 		}
// 	}

// 	if (info.position.top < topHeightRef.value && info.direction === 'up') {
// 		showFloatTitleRef.value = false;
// 	}
// };

onMounted(() => {
	stateUpdate();
	updateList();
	busOn('appSubscribe', stateUpdate);

	menuStore.$subscribe(() => {
		updateItems();
	});
});

let updateItems = debounce(() => {
	stateUpdate();
}, 50);

onUnmounted(() => {
	busOff('appSubscribe', stateUpdate);
});

function stateUpdate() {
	const vaultList: VaultItem[] = [];
	for (const vault of app.state.vaults) {
		for (const item of vault.items) {
			if (item.type === VaultType.TerminusTotp) {
				const auth = item.fields.find((value) => {
					return value.type === 'totp';
				});
				hasAuthRef.value = !!auth;
			}

			if (item.type === VaultType.VC) {
				hasVCCardRef.value = true;
				vaultList.push(item);
			}
		}
	}

	vaultList.sort((a, b) => {
		return b.updated.getTime() - a.updated.getTime();
	});

	cardItemsRef.value = (
		vaultList.length > 2 ? vaultList.slice(0, 3) : vaultList
	).flatMap((item) => {
		return convertVault2CardItem(item)!;
	});

	vcCardHeightRef.value =
		cardItemsRef.value.length > 0
			? (cardItemsRef.value.length - 1) * 72 + 106
			: 0;

	topHeightRef.value =
		56 +
		(hasAuthRef.value ? 80 + 12 : 180 + 12) +
		(vcCardHeightRef.value > 0 ? vcCardHeightRef.value + 20 : 0);
}
</script>

<style lang="scss" scoped>
.secret-parent {
	height: 100%;
	width: 100%;

	.bottom-content {
		width: 100%;
		height: calc(100% - 56px);
	}

	.secret-root {
		width: 100%;
		height: 100%;

		.bottom-content2 {
			width: 100%;
			height: calc(100% - 56px);
		}

		.float-title {
			position: fixed;
			top: 0;
			width: 100%;
		}

		.secret-auth {
			width: calc(100% - 48px);
			margin-left: 24px;
			margin-right: 24px;
			margin-top: 12px;
		}

		.secret-vc {
			width: calc(100% - 48px);
			margin-left: 24px;
			margin-right: 24px;
			margin-top: 20px;
			position: relative;

			.secret-vc-item {
				position: absolute;
				width: 100%;
				left: 0;
			}

			.secret-vc-mask {
				position: absolute;
				height: 34px;
				width: 100%;
				left: 0;
				right: 0;
				bottom: 0;
				background-size: contain;
			}
		}

		.secret-list-parent {
			width: 100%;
			height: calc(100% - var(--reactHeight, 200px));

			.secret-list {
				width: 100%;
				height: 100vh;
			}
		}
	}
}
</style>
