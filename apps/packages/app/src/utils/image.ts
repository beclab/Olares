export function handleSiteIcon(iconContent: string, iconType: string): string {
	return iconContent ? `data:${iconContent}` : '';
}
