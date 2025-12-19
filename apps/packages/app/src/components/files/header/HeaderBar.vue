<template>
	<header class="q-pa-xs">
		<slot />

		<div class="row items-center justify-center">
			<slot name="info" />
			<div
				id="dropdown"
				:class="{ active: store.show === 'more' }"
				class="q-pr-md"
			>
				<slot name="actions" />
			</div>
		</div>

		<!-- <action v-if="$slots.actions" id="more" icon="more_vert" :label="$t('buttons.more')"
      @action="store.showHover('more')" /> -->

		<!-- <div class="overlay" v-show="store.show == 'more'" @click="store.closeHovers()" /> -->
	</header>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { useDataStore } from '../../../stores/data';
import { logoURL } from '../../../utils/constants';

// import Action from './Action.vue';

export default defineComponent({
	name: 'header-bar',
	props: {
		showLogo: {
			type: Boolean,
			required: false
		},
		showMenu: {
			type: Boolean,
			required: false
		}
	},
	components: {
		// Action
	},
	setup() {
		const store = useDataStore();
		const openSidebar = () => {
			store.showHover('sidebar');
		};

		return {
			store,
			logoURL,
			openSidebar
		};
	}
});
</script>

<style scoped lang="scss">
header {
	z-index: 1000;
	/* border-bottom: 1px solid rgba(0, 0, 0, 0.075); */
	/* box-shadow: 0 0 5px rgba(0, 0, 0, 0.1); */
	position: fixed;
	top: 0;
	left: 0;
	/* height: 4em; */
	/* width: 100%; */
	padding: 0;
	display: flex;
	padding: 0.5em 0.5em 0.5em 1em;
	align-items: center;
	justify-content: space-between;
}

header > * {
	flex: 0 0 auto;
}

header title {
	display: block;
	flex: 1 1 auto;
	padding: 0 1em;
	overflow: hidden;
	text-overflow: ellipsis;
	font-size: 1.2em;
}

header .overlay {
	width: 0;
	height: 0;
}

header a,
header a:hover {
	color: inherit;
}

header > div:first-child > .action,
header img {
	margin-right: 1em;
}

header img {
	height: 2.5em;
}

header .action span {
	display: none;
}

header > div div {
	vertical-align: middle;
	position: relative;
}

header .search-button,
header .menu-button {
	display: none;
}
</style>
