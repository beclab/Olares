import { defineStore } from 'pinia';
import { date } from 'quasar';
import { useTokenStore } from 'src/stores/settings/token';
import axios from 'axios';

export type DateFormat =
	| 'YYYY/MM/DD'
	| 'D/M/YY'
	| 'M/D/YY'
	| 'DD/MM/YYYY'
	| 'DD.MM.YYYY'
	| 'DD-MM-YYYY'
	| 'YYYY.MM.DD'
	| 'YYYY-MM-DD'
	| 'YY/MM/DD'
	| 'YY-M-D'
	| 'YY.M.D';

export interface WidgetPreferences {
	showWeight: boolean;
	is24HourFormat: boolean;
	dateFormat: DateFormat;
	showDashboard: boolean;
}

export const DESKTOP_WIDGET_DEFAULT: WidgetPreferences = {
	showWeight: true,
	is24HourFormat: true,
	dateFormat: 'YYYY/MM/DD',
	showDashboard: true
};

const WIDGET_PREFERENCES_CACHE_KEY = 'widget-preferences-cache-v1';

function loadWidgetPreferencesCache(): Partial<WidgetPreferences> {
	try {
		const raw = localStorage.getItem(WIDGET_PREFERENCES_CACHE_KEY);
		if (!raw) {
			return {};
		}
		return JSON.parse(raw) as Partial<WidgetPreferences>;
	} catch (error) {
		console.error('Load widget preferences cache failed:', error);
		return {};
	}
}

function saveWidgetPreferencesCache(value: WidgetPreferences) {
	try {
		localStorage.setItem(WIDGET_PREFERENCES_CACHE_KEY, JSON.stringify(value));
	} catch (error) {
		console.error('Save widget preferences cache failed:', error);
	}
}

export type WidgetPreferencesState = WidgetPreferences & {
	isLoaded: boolean;
};

export const useWidgetPreferencesStore = defineStore('widgetPreferences', {
	state: (): WidgetPreferencesState => {
		return {
			...DESKTOP_WIDGET_DEFAULT,
			...loadWidgetPreferencesCache(),
			isLoaded: false
		};
	},
	actions: {
		async getWidget() {
			try {
				const tokenStore = useTokenStore();
				const data: any = await axios.get(`${tokenStore.url}/api/widget`);
				this.save(data);
				return true;
			} catch (error) {
				console.error('Get widget preferences failed:', error);
				return false;
			}
		},
		formatNow() {
			const now = new Date();
			const result = {
				date: date.formatDate(now, this.dateFormat),
				time: date.formatDate(now, this.is24HourFormat ? 'HH:mm' : 'hh:mm'),
				week: date.formatDate(now, 'dddd'),
				year: date.formatDate(now, 'YYYY'),
				month: date.formatDate(now, 'MM'),
				day: date.formatDate(now, 'DD'),
				isAM: now.getHours() < 12
			};
			return result;
		},
		save(saved?: Partial<WidgetPreferences>) {
			try {
				this.showWeight =
					saved?.showWeight ?? DESKTOP_WIDGET_DEFAULT.showWeight;
				this.is24HourFormat =
					saved?.is24HourFormat ?? DESKTOP_WIDGET_DEFAULT.is24HourFormat;
				this.dateFormat =
					(saved?.dateFormat as DateFormat | undefined) ??
					DESKTOP_WIDGET_DEFAULT.dateFormat;
				this.showDashboard =
					saved?.showDashboard ?? DESKTOP_WIDGET_DEFAULT.showDashboard;
				this.isLoaded = true;
				saveWidgetPreferencesCache({
					showWeight: this.showWeight,
					is24HourFormat: this.is24HourFormat,
					dateFormat: this.dateFormat,
					showDashboard: this.showDashboard
				});
			} catch (e) {
				console.error('Error saving widget', saved);
			}
		},
		async update() {
			const tokenStore = useTokenStore();
			await axios.post(`${tokenStore.url}/api/widget`, {
				widget: {
					showWeight: this.showWeight,
					is24HourFormat: this.is24HourFormat,
					dateFormat: this.dateFormat,
					showDashboard: this.showDashboard
				}
			});
		}
	}
});
