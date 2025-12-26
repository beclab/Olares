<template>
	<div class="column" v-if="blockRef">
		<block-nickname
			class="q-mt-sm"
			:type="blockRef.type"
			v-model:text="blockRef.nickName"
		/>
		<div class="edit-label full-width q-mt-lg">
			{{ t('blocks.url_or_email') + '(' + t('base.required') + ')' }}
		</div>
		<edit-view
			v-if="userStore.user"
			class="label-width q-mt-xs"
			:placeholder="t('blocks.link_url')"
			v-model="blockRef.url"
		/>
		<div class="link-root row q-pa-lg q-mt-lg">
			<bio-image v-model:img-url="blockRef.icon" />
			<div class="column justify-between link-text">
				<edit-view
					:placeholder="t('base.title') + '(' + t('base.required') + ')'"
					v-model="blockRef.title"
				/>
				<edit-view
					:placeholder="t('base.subtitle') + '(' + t('base.optional') + ')'"
					v-model="blockRef.subTitle"
				/>
			</div>
		</div>
		<div class="edit-label full-width q-mt-lg">
			{{ t('blocks.link_block_size') }}
		</div>
		<grid-picker-group class="q-mt-xs" :grid="false" v-model="blockRef.size">
			<picker-component text="S" :value="SIZE_TYPE.SMALL" />
			<picker-component class="q-ml-md" text="L" :value="SIZE_TYPE.LARGER" />
		</grid-picker-group>
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
			:label="t('blocks.link_outline')"
			v-model="blockRef.outline"
		/>
		<switch-component
			:label="t('blocks.link_shadow')"
			v-model="blockRef.shadow"
		/>
	</div>
</template>

<script lang="ts" setup>
import SwitchComponent from '@apps/profile/src/components/profile/base/SwitchComponent.vue';
import GridPickerGroup from '@apps/profile/src/components/profile/base/GridPickerGroup.vue';
import PickerComponent from '@apps/profile/src/components/profile/base/PickerComponent.vue';
import BlockNickname from '@apps/profile/src/components/profile/block/BlockNickname.vue';
import BioImage from '@apps/profile/src/components/profile/base/BioImage.vue';
import EditView from '@apps/profile/src/components/profile/base/EditView.vue';
import { SIZE_TYPE, ALIGNMENT_TYPE } from '@apps/profile/src/types/User';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
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

<style scoped lang="scss">
.link-root {
	width: 100%;
	height: 128px;
	border: 1px solid $input-stroke;
	border-radius: 8px;

	.link-text {
		margin-left: 12px;
		width: calc(100% - 100px);
		height: 88px;
	}
}
</style>
