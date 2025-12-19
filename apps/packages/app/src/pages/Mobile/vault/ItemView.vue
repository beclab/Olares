<template>
	<div class="item-view-root">
		<ItemView :itemID="itemID" :isNew="menuStore.isEdit" />
	</div>
	<router-view></router-view>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useRoute } from 'vue-router';
import ItemView from '../../Items/ItemView.vue';

import { useMenuStore } from '../../../stores/menu';

const Route = useRoute();
let itemID = ref((Route.params.itemid as string) || '');
const menuStore = useMenuStore();
</script>

<style lang="scss" scoped>
.item-view-root {
	height: 100%;
	width: 100%;
}
</style>
