export interface CustomData {
	phase: string;
	selected: boolean;
}

export interface CustomEvents {
	onCustomEvent: (event: MouseEvent) => void;
}
