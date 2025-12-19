<template>
	<page-title-component :show-back="true" :title="t('secrets')" />

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div class="row justify-between items-center">
			<div class="text-subtitle1 secret-title">
				{{ t('app_secrets') }}
			</div>
			<list-bottom-func-btn @funcClick="newSecret" :title="t('add_secret')" />
		</div>

		<q-list
			v-if="secretList && secretList.length > 0"
			class="q-list-class q-pa-lg"
		>
			<path-location>
				<template v-slot:start>
					<div>{{ t('key') }}</div>
				</template>
				<template v-slot:midden>
					<div>{{ t('last_updated') }}</div>
				</template>
				<template v-slot:end>
					<div>{{ t('action') }}</div>
				</template>
			</path-location>
			<bt-separator class="q-mt-md" />
			<template v-for="item in secretList" :key="item.Key">
				<secret-item
					:secret="item"
					@on-update="updateSecret"
					@on-delete="deleteSecret"
				/>
			</template>
		</q-list>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useQuasar } from 'quasar';
import { useApplicationStore } from 'src/stores/settings/application';
import { useSecretStore } from 'src/stores/settings/secret';
import NewSecret from 'src/components/settings/application/dialog/secret/NewSecret.vue';
import UpdateSecret from 'src/components/settings/application/dialog/secret/UpdateSecret.vue';
import DeleteSecret from 'src/components/settings/application/dialog/secret/DeleteSecret.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import PathLocation from 'src/components/settings/PathLocation.vue';
import SecretItem from 'src/components/settings/application/SecretItem.vue';
import ListBottomFuncBtn from 'src/components/settings/ListBottomFuncBtn.vue';
import BtSeparator from 'src/components/settings/base/BtSeparator.vue';
import { Secret } from 'src/constant/index';

import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const $q = useQuasar();
const applicationStore = useApplicationStore();
const secretStore = useSecretStore();
const Route = useRoute();

const application = ref(
	applicationStore.getApplicationById(Route.params.name as string)
);

const secretList = ref<Secret[]>([]);

async function updateSecretList() {
	secretList.value = await secretStore.getSecrets(application.value?.name);
}

onMounted(async () => {
	application.value = applicationStore.getApplicationById(
		Route.params.name as string
	);
	await updateSecretList();
});

async function newSecret() {
	$q.dialog({
		component: NewSecret,
		persistent: true,
		componentProps: {}
	}).onOk(async (data) => {
		await secretStore.newSecret(application.value?.name, data.key, data.value);
		await updateSecretList();
	});
}

async function updateSecret(key: string) {
	$q.dialog({
		component: UpdateSecret,
		persistent: true,
		componentProps: {
			key
		}
	}).onOk(async (data) => {
		await secretStore.updateSecret(application.value?.name, key, data.value);
		await updateSecretList();
	});
}

async function deleteSecret(key: string) {
	$q.dialog({
		component: DeleteSecret,
		persistent: true,
		componentProps: {
			id: key
		}
	}).onOk(async () => {
		await secretStore.deleteSecret(application.value?.name, key);
		await updateSecretList();
	});
}
</script>

<style scoped lang="scss">
.secret-title {
	margin-top: 12px;
	margin-bottom: 8px;
	color: $ink-1;
}
</style>
