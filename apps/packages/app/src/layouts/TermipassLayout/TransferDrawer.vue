<template>
	<q-drawer
		show-if-above
		behavior="desktop"
		:width="240"
		:bordered="false"
		class="myDrawer"
	>
		<q-scroll-area
			style="height: 100%; width: 100%"
			:thumb-style="scrollBarStyle.thumbStyle"
		>
			<bt-menu
				:items="menus"
				:default-active="`${transferStore.activeItem}`"
				@select="updateCurrentMenu"
				style="width: 100%"
				class="title-norla"
				active-class="text-subtitle2 bg-yellow-soft text-ink-1"
			>
			</bt-menu>
		</q-scroll-area>
	</q-drawer>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useTransfer2Store } from '../../stores/transfer2';
import { scrollBarStyle } from '../../utils/contact';
import { filesIsV2 } from 'src/api';

const transferStore = useTransfer2Store();

const items = transferStore.menus();

const menus = ref(items);

const updateCurrentMenu = (item: { key: any }) => {
	console.log('updateCurrentMenu', item);
	transferStore.activeItem = Number(item.key);
	transferStore.filesInFolder = [];
	transferStore.filesInFolderMap = {};
};

watch(
	() => transferStore.downloading,
	(newVal) => {
		let count: string | number = newVal.length;
		if (count > 99) {
			count = '99+';
		}

		if (!count) {
			menus.value[0].children[1].count = '';
		} else {
			menus.value[0].children[1].count = count;
		}
	},
	{
		deep: true,
		immediate: true
	}
);

watch(
	() => transferStore.uploading,
	(newVal) => {
		let count: string | number = newVal.length;
		if (count > 99) {
			count = '99+';
		}

		if (!count) {
			delete menus.value[0].children[0].count;
		} else {
			menus.value[0].children[0].count = count;
		}
	},
	{
		deep: true,
		immediate: true
	}
);

watch(
	() => transferStore.clouding,
	(newVal) => {
		let count: string | number = newVal.length;
		if (count > 99) {
			count = '99+';
		}

		if (!count) {
			delete menus.value[0].children[2].count;
		} else {
			menus.value[0].children[2].count = count;
		}
	},
	{
		deep: true,
		immediate: true
	}
);

if (filesIsV2()) {
	watch(
		() => transferStore.copying,
		(newVal) => {
			let count: string | number = newVal.length;
			if (count > 99) {
				count = '99+';
			}

			if (!count) {
				delete menus.value[0].children[3].count;
			} else {
				menus.value[0].children[3].count = count;
			}
		},
		{
			deep: true,
			immediate: true
		}
	);
}
</script>

<style lang="scss">
.files-active-link {
	background: rgba(255, 235, 59, 0.1);
}

.myDrawer {
	overflow: hidden;
	padding-top: 6px;
	// border-right: 1px solid $separator;
}
</style>
