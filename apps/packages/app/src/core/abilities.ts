export interface AbilityItem {
	running: boolean;
	id: string;
	url: string;
	name: string;
}
export interface AbilityData {
	vault: boolean;
	wise: AbilityItem & { title: string };
	translate: AbilityItem & { title: string };
	ytdlp: AbilityItem & { title: string };
}
