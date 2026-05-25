/**
 * Reader Styles Composable
 *
 * Manages reader typography and theme settings
 * Supports customization for text size, font, line height, etc.
 */

import { computed, reactive } from 'vue';
import { useColor } from '@bytetrade/ui';
import { useQuasar } from 'quasar';

export interface ReaderStyleProps {
	/** Font size in pixels (default: 20) */
	fontSize?: number;
	/** Margin in pixels (default: 290) */
	margin?: number;
	/** Line height percentage (default: 150) */
	lineHeight?: number;
	/** Max width as percentage (0 = auto) */
	maxWidthPercentage?: number;
	/** Font family (default: 'Robot') */
	fontFamily?: string;
	/** Use high contrast text (default: false) */
	highContrastText?: boolean;
	/** Justify text alignment (default: false) */
	justifyText?: boolean;
}

export interface ReaderStyleOverrides {
	maxWidthPercentage?: number | null;
	lineHeight?: number | null;
	fontFamily?: string | null;
	highContrastText?: boolean;
	justifyText?: boolean;
}

const DEFAULT_PROPS: Required<ReaderStyleProps> = {
	fontSize: 20,
	margin: 290,
	lineHeight: 150,
	maxWidthPercentage: 0,
	fontFamily: 'Robot',
	highContrastText: false,
	justifyText: false
};

export function useReaderStyles(
	props: ReaderStyleProps = {},
	overrides: ReaderStyleOverrides = {}
) {
	const $q = useQuasar();
	const { color: background1 } = useColor('background-1');
	const { color: ink2 } = useColor('ink-2');

	// Merge props with defaults
	const mergedProps = computed(() => ({
		...DEFAULT_PROPS,
		...props
	}));

	// Theme colors
	const theme = reactive({
		readerTableHeader: '#FFFFFF',
		readerFontHighContrast: '#0A0806',
		get readerFont() {
			return ink2.value;
		},
		get readerBg() {
			return background1.value;
		}
	});

	// Text color based on contrast setting
	const textColorValue = (isHighContrast: boolean): string => {
		return isHighContrast ? theme.readerFontHighContrast : theme.readerFont;
	};

	// Justify text value
	const justifyTextValue = (isJustified: boolean): string => {
		return isJustified ? 'justify' : 'start';
	};

	// Computed styles
	const styles = computed(() => {
		const p = mergedProps.value;
		return {
			fontSize: p.fontSize,
			margin: p.margin,
			maxWidthPercentage: overrides.maxWidthPercentage ?? p.maxWidthPercentage,
			lineHeight: overrides.lineHeight ?? p.lineHeight,
			fontFamily: overrides.fontFamily ?? p.fontFamily,
			readerFontColor: textColorValue(
				overrides.highContrastText ?? p.highContrastText
			),
			readerTableHeaderColor: theme.readerTableHeader,
			readerHeadersColor: theme.readerFont
		};
	});

	// Max width styles for different screen sizes
	const maxWidthStyles = computed(() => {
		const s = styles.value;
		return {
			default: s.maxWidthPercentage
				? `${s.maxWidthPercentage}%`
				: `${1024 - s.margin}px`,
			small: s.maxWidthPercentage
				? `${s.maxWidthPercentage}%`
				: `calc(${120 - Math.round((s.margin * 10) / 100) / 100}% * 100vw)`
		};
	});

	// CSS variables for article content
	const cssVariables = computed(() => {
		const s = styles.value;
		const isSmallScreen = $q.screen.sm || $q.screen.xs;

		return {
			'--text-align': justifyTextValue(
				overrides.justifyText ?? mergedProps.value.justifyText
			),
			'--text-font-size': `${s.fontSize}px`,
			'--line-height': `${s.lineHeight}%`,
			'--blockquote-padding': isSmallScreen ? '1em 2em' : '0.5em 1em',
			'--blockquote-icon-font-size': isSmallScreen ? '1.7rem' : '1.3rem',
			'--figure-margin': isSmallScreen ? '2.6875rem auto' : '1.6rem auto',
			'--hr-margin': isSmallScreen ? '2em' : '1em',
			'--font-color': s.readerFontColor,
			'--table-header-color': s.readerTableHeaderColor
		};
	});

	// Content max width based on screen size
	const contentMaxWidth = computed(() => {
		const isSmallScreen = $q.screen.sm || $q.screen.xs;
		return isSmallScreen
			? maxWidthStyles.value.small
			: maxWidthStyles.value.default;
	});

	// Content padding based on screen size
	const contentPadding = computed(() => {
		const isSmallScreen = $q.screen.sm || $q.screen.xs;
		return isSmallScreen ? 15 : 30;
	});

	return {
		/** Theme colors */
		theme,
		/** Computed styles */
		styles,
		/** Max width styles */
		maxWidthStyles,
		/** CSS variables for article content */
		cssVariables,
		/** Content max width based on screen */
		contentMaxWidth,
		/** Content padding based on screen */
		contentPadding,
		/** Text color helper */
		textColorValue,
		/** Justify text helper */
		justifyTextValue,
		/** Screen size helpers */
		isSmallScreen: computed(() => $q.screen.sm || $q.screen.xs)
	};
}
