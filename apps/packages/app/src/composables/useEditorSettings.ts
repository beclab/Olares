import { ref, watch } from 'vue';

export interface EditorSettings {
	fontSize: number;
	tabSize: number;
	wordWrap: 'on' | 'off' | 'wordWrapColumn' | 'bounded';
	minimap: boolean;
	lineNumbers: boolean;
	folding: boolean;
	renderWhitespace: 'none' | 'boundary' | 'selection' | 'all';
	formatOnSave: boolean;
}

const STORAGE_KEY = 'studio-editor-settings';

const defaultSettings: EditorSettings = {
	fontSize: 14,
	tabSize: 2,
	wordWrap: 'on',
	minimap: true,
	lineNumbers: true,
	folding: true,
	renderWhitespace: 'selection',
	formatOnSave: true
};

const loadSettings = (): EditorSettings => {
	try {
		const stored = localStorage.getItem(STORAGE_KEY);
		if (stored) {
			const parsed = JSON.parse(stored);
			return { ...defaultSettings, ...parsed };
		}
	} catch (error) {
		console.error('Failed to load editor settings:', error);
	}
	return { ...defaultSettings };
};

const saveSettings = (settings: EditorSettings) => {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
	} catch (error) {
		console.error('Failed to save editor settings:', error);
	}
};

const editorSettings = ref<EditorSettings>(loadSettings());

watch(
	editorSettings,
	(newSettings) => {
		saveSettings(newSettings);
	},
	{ deep: true }
);

export const useEditorSettings = () => {
	const updateSetting = <K extends keyof EditorSettings>(
		key: K,
		value: EditorSettings[K]
	) => {
		editorSettings.value[key] = value;
	};

	const resetSettings = () => {
		editorSettings.value = { ...defaultSettings };
	};

	return {
		settings: editorSettings,
		updateSetting,
		resetSettings
	};
};
