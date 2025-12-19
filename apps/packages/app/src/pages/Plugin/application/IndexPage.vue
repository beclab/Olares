<template>
	<PageCard :title="$t('application')">
		<template #extra>
			<div class="relative-position cursor-pointer" @click="actionHandler">
				<q-img
					:src="EditButtonIcon"
					:ratio="1"
					width="32px"
					spinner-size="0px"
				/>
				<q-tooltip>{{ $t('edit') }}</q-tooltip>
			</div>
		</template>
		<div class="column no-wrap flex-gap-lg">
			<div>
				<div class="text-h6 text-ink-1">{{ $t('bex.home') }}</div>
				<div class="q-mt-sm">
					<ForegroundApplications
						ref="ForegroundApplicationsRef"
						:show-action="edit"
						@updateOne="setEdit"
					></ForegroundApplications>
				</div>
			</div>
			<div>
				<div class="text-h6 text-ink-1">{{ $t('bex.all_apps') }}</div>
				<div class="q-mt-sm">
					<AllApplications :show-action="edit"></AllApplications>
				</div>
			</div>
		</div>
		<div
			class="fixed-bottom row justify-between items-center flex-gap-md q-mx-lg"
			style="bottom: 20px"
			v-show="edit"
		>
			<CustomButton
				:label="$t('cancel')"
				text-color="ink-2"
				outline
				style="width: 132px"
				@click="cancelHandler"
			></CustomButton>
			<CustomButton
				:label="$t('save')"
				style="width: 132px"
				color="yellow-default"
				@click="saveHandler"
			></CustomButton>
		</div>
	</PageCard>
</template>

<script setup lang="ts">
import ForegroundApplications from 'src/pages/Plugin/containers/ForegroundApplications.vue';
import AllApplications from 'src/pages/Plugin/containers/AllApplications.vue';
import PageCard from 'src/pages/Plugin/components/PageCard.vue';
import EditButtonIconLight from 'src/assets/plugin/edit-button.svg';
import EditButtonIconDark from 'src/assets/plugin/edit-button-dark.svg';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useAppsStore } from 'src/stores/bex/apps';
import CustomButton from 'src/pages/Plugin/components/CustomButton.vue';
import { useQuasar } from 'quasar';

const appsStore = useAppsStore();

const edit = ref(false);
const ForegroundApplicationsRef = ref();
const $q = useQuasar();

const EditButtonIcon = computed(() =>
	$q.dark.isActive ? EditButtonIconDark : EditButtonIconLight
);
const actionHandler = () => {
	edit.value = !edit.value;
};

const setEdit = () => {
	edit.value = true;
};
const cancelHandler = () => {
	edit.value = false;
	ForegroundApplicationsRef.value.reset();
	appsStore.appActionCancel();
};

const saveHandler = () => {
	edit.value = false;
	ForegroundApplicationsRef.value.reset();
	appsStore.appActionSave();
};

onMounted(() => {
	appsStore.init();
});

onBeforeUnmount(() => {
	if (edit.value) {
		appsStore.appActionCancel();
	}
});
</script>

<style></style>
