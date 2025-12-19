<template>
	<page-title-component :show-back="false" :title="t('notification')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<q-list class="q-list-class">
			<q-item class="list_section" clickable @click="gotoPage('/job')">
				<q-item-section> {{ t('job') }} </q-item-section>
				<q-item-section side>
					<q-icon name="bi-chevron-right" size="16px" />
				</q-item-section>
			</q-item>
			<q-separator class="q-list-sep-line" />
			<q-item class="list_section" clickable @click="gotoPage('/sender')">
				<q-item-section> {{ t('sender') }}</q-item-section>
				<q-item-section side>
					<q-icon name="bi-chevron-right" size="16px" />
				</q-item-section>
			</q-item>

			<q-separator class="q-list-sep-line" />
			<q-item class="list_section" clickable @click="gotoPage('/recipients')">
				<q-item-section> {{ t('recipients') }} </q-item-section>
				<q-item-section side>
					<q-icon name="bi-chevron-right" size="16px" />
				</q-item-section>
			</q-item>

			<q-separator class="q-list-sep-line" />
			<q-item class="list_section" clickable @click="gotoPage('/template')">
				<q-item-section> {{ t('template') }} </q-item-section>
				<q-item-section side>
					<q-icon name="bi-chevron-right" size="16px" />
				</q-item-section>
			</q-item>

			<q-separator class="q-list-sep-line" />
			<q-item class="list_section" clickable @click="gotoPage('/notify')">
				<q-item-section> {{ t('notify_policy') }} </q-item-section>
				<q-item-section side>
					<q-icon name="bi-chevron-right" size="16px" />
				</q-item-section>
			</q-item>
		</q-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { useNotificationStore } from 'src/stores/settings/notification';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();

const router = useRouter();

const notificationStore = useNotificationStore();
onMounted(async () => {
	await notificationStore.init();
});

function gotoPage(path: string) {
	router.push({ path });
}
</script>
