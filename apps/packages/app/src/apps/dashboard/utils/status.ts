import { NamedColor } from 'quasar';

export function resourceStatusColor(value: number): NamedColor | undefined {
	return isNaN(value)
		? 'white'
		: Number(value) > 80
		? 'negative'
		: Number(value) > 50
		? 'warning'
		: 'positive';
}
