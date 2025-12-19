<template>
	<div class="column" v-if="blockRef">
		<block-nickname
			class="q-mt-sm"
			:type="blockRef.type"
			v-model:text="blockRef.nickName"
		/>
		<div class="block-desc q-mt-lg">{{ t('blocks.format_your_text') }}</div>
		<div class="edit-label full-width q-mt-lg">
			{{ t('blocks.add_title') }}
		</div>
		<edit-view
			v-if="userStore.user"
			class="label-width q-mt-xs"
			:placeholder="t('blocks.text_block')"
			v-model="blockRef.title"
		/>
		<div class="edit-label full-width q-mt-lg">
			{{ t('blocks.add_text') }}
		</div>
		<edit-view
			v-if="userStore.user"
			height="96px"
			class="label-width q-mt-xs"
			:placeholder="t('blocks.text_description')"
			v-model="blockRef.description"
		/>
		<div class="edit-label full-width q-mt-lg">
			{{ t('blocks.text_alignment') }}
		</div>
		<grid-picker-group
			class="q-mt-xs"
			:grid="false"
			v-model="blockRef.textAlignment"
		>
			<picker-component
				icon="sym_r_format_align_left"
				:value="ALIGNMENT_TYPE.LEFT"
			/>
			<picker-component
				class="q-ml-md"
				icon="sym_r_format_align_center"
				:value="ALIGNMENT_TYPE.CENTER"
			/>
			<picker-component
				class="q-ml-md"
				icon="sym_r_format_align_right"
				:value="ALIGNMENT_TYPE.RIGHT"
			/>
		</grid-picker-group>
		<switch-component
			:label="t('blocks.transparent_background')"
			v-model="blockRef.transparent"
		/>
	</div>
</template>

<script lang="ts" setup>
import SwitchComponent from '@apps/profile/src/components/profile/base/SwitchComponent.vue';
import GridPickerGroup from '@apps/profile/src/components/profile/base/GridPickerGroup.vue';
import PickerComponent from '@apps/profile/src/components/profile/base/PickerComponent.vue';
import BlockNickname from '@apps/profile/src/components/profile/block/BlockNickname.vue';
import EditView from '@apps/profile/src/components/profile/base/EditView.vue';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { ALIGNMENT_TYPE } from '@apps/profile/src/types/User';
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

const route = useRoute();
const userStore = useUserStore();
const blockRef = ref();
const { t } = useI18n();

onMounted(() => {
	if (route.params.id && userStore.user) {
		const block = userStore.user.block.data.find((item) => {
			return item.id === route.params.id;
		});
		console.log(block);
		if (block) {
			blockRef.value = block;
		}
	}
});
</script>

<style scoped lang="scss" />
