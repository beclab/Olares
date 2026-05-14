import { ref } from 'vue';
import { PubSub } from '../../utils/event';

export const bus = new PubSub();

export const verticalObj = {
	top: '0px',
	center: '50%',
	bottom: '100%'
};

export const horizontalObj = {
	left: '0px',
	middle: '50%',
	right: '100%',
	start: '0px',
	end: '100%'
};
