<template>
	<div class="add-header row items-center">
		<CustomButton
			@click="addFileToUpload(item)"
			class="q-mr-sm bg-yellow-default"
		>
			<template #label>
				<div class="row items-center text-body3text-ink-2">
					<q-icon class="q-mr-xs" name="sym_r_add" size="20px" />
					{{ t('files.upload_files') }}
				</div>
			</template>
		</CustomButton>

		<template v-if="filesIsV2()">
			<WiseAbilityTooltipContainer>
				<CustomButton
					@click="addCloudTask"
					outline
					:disable="!appAbilitiesStore.wise.running"
					class="q-mr-sm"
				>
					<template #label>
						<div class="text-body3 text-ink-2">
							<q-icon class="q-mr-xs" name="sym_r_link" size="20px" />
							{{ t('files.Link Download') }}
						</div>
					</template>
				</CustomButton>
			</WiseAbilityTooltipContainer>
		</template>
		<template v-else>
			<CustomButton @click="addCloudTask" outline class="q-mr-sm">
				<template #label>
					<div class="text-body3 text-ink-2">
						<q-icon class="q-mr-xs" name="sym_r_link" size="20px" />
						{{ t('files.Link Download') }}
					</div>
				</template>
			</CustomButton>
		</template>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import WiseAbilityTooltipContainer from '../../../components/WiseAbilityTooltipContainer.vue';
import { useAppAbilitiesStore } from '../../../stores/appAbilities';
import CustomButton from '../../Plugin/components/CustomButton.vue';
import { filesIsV2 } from '../../../api';
const appAbilitiesStore = useAppAbilitiesStore();

const { t } = useI18n();

const addCloudTask = () => {
	emits('addCloudTask');
};

const addFileToUpload = async () => {
	emits('addUploadTask');
};

const emits = defineEmits(['addUploadTask', 'addCloudTask']);
</script>

<style scoped lang="scss">
.add-header {
	height: 56px;
	width: 100%;

	.link-btn {
		padding: 0px 8px;
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		height: 32px;
	}

	.upload1-btn {
		padding: 0px 8px;
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		height: 32px;
		color: $grey-10;
		background: $yellow-default;
	}
}
</style>
