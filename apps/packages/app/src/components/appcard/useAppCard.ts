import { getI18nValue, TRANSACTION_PAGE } from 'src/constant/constants';
import { useCenterStore } from 'src/stores/market/center';
import { useColor } from '@bytetrade/ui';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import { isCloneApp, uninstalledApp } from 'src/constant/config';

export default function useAppCard(props) {
	const centerStore = useCenterStore();
	const router = useRouter();
	const { locale } = useI18n();
	const appAggregation = computed(() => {
		return centerStore.getAppAggregationInfo(props.appName, props.sourceId);
	});
	const { color: separatorColor } = useColor('separator');

	const clusterScopedApp = computed(
		() =>
			appAggregation.value?.app_full_info?.app_info?.app_entry?.options
				?.appScope?.clusterScoped ?? false
	);

	const appIcon = computed(
		() =>
			appAggregation.value?.app_simple_latest?.app_simple_info?.app_icon ?? ''
	);

	const appTitle = computed(() => {
		if (isCloneApp(appAggregation.value?.app_status_latest?.status)) {
			return appAggregation.value?.app_status_latest?.status.title;
		} else {
			return (
				getI18nValue(
					appAggregation.value?.app_simple_latest?.app_simple_info?.app_title,
					locale.value
				) ?? ''
			);
		}
	});

	const appDesc = computed(
		() =>
			getI18nValue(
				appAggregation.value?.app_simple_latest?.app_simple_info
					?.app_description,
				locale.value
			) ?? ''
	);

	const appVersion = computed(
		() =>
			appAggregation.value?.app_simple_latest?.app_simple_info?.app_version ??
			''
	);

	const myAppVersion = computed(() => {
		if (uninstalledApp(appAggregation.value?.app_status_latest?.status)) {
			return '';
		}
		return appAggregation.value?.app_status_latest?.version ?? '';
	});

	const appFeaturedImage = computed(
		() =>
			appAggregation.value?.app_full_info?.app_info?.app_entry?.featuredImage ??
			appAggregation.value?.app_simple_latest?.app_simple_info?.app_icon ??
			''
	);

	function goAppDetails() {
		console.log('click app');
		if (props.disabled) {
			return;
		}
		router.push({
			name: TRANSACTION_PAGE.App,
			params: {
				appName: props.appName,
				sourceId: props.sourceId
			},
			query: {
				...router.currentRoute.value.query
			}
		});
	}

	const sourceName = computed(() => {
		const source = centerStore.sources.find(
			(item) => item.id === props.sourceId
		);
		if (source) {
			return source.name;
		} else {
			return props.sourceId;
		}
	});

	return {
		separatorColor,
		appAggregation,
		clusterScopedApp,
		appIcon,
		appTitle,
		appDesc,
		appVersion,
		myAppVersion,
		appFeaturedImage,
		sourceName,
		goAppDetails
	};
}
