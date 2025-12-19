import type { User } from './classes';
import { HeadScaleDevice } from '@bytetrade/core';

export function sortUsers(users: User[]): User[] {
	const sortKey = localStorage.getItem('headscaleUserSort') || '';
	const sortDirection =
		localStorage.getItem('headscaleUserSortDirection') || '';
	let sortedUsers = users;

	const collator = new Intl.Collator([], { numeric: true });
	if (sortDirection == 'ascending') {
		switch (sortKey) {
			case 'id':
				sortedUsers = users.sort((a: User, b: User) =>
					collator.compare(a.id, b.id)
				);
				break;
			case 'createdAt':
				sortedUsers = users.sort(
					(a: User, b: User) => -collator.compare(a.createdAt, b.createdAt)
				);
				break;
			case 'name':
				sortedUsers = users.sort((a: User, b: User) =>
					collator.compare(a.name, b.name)
				);
				break;
		}
	}
	if (sortDirection == 'descending') {
		switch (sortKey) {
			case 'id':
				sortedUsers = users.sort(
					(a: User, b: User) => -collator.compare(a.id, b.id)
				);
				break;
			case 'createdAt':
				sortedUsers = users.sort((a: User, b: User) =>
					collator.compare(a.createdAt, b.createdAt)
				);
				break;
			case 'name':
				sortedUsers = users.sort(
					(a: User, b: User) => -collator.compare(a.name, b.name)
				);
				break;
		}
	}

	return sortedUsers;
}

export function sortDevices(devices: HeadScaleDevice[]): HeadScaleDevice[] {
	const sortKey = localStorage.getItem('headscaleDeviceSort') || '';
	const sortDirection =
		localStorage.getItem('headscaleDeviceSortDirection') || '';
	let sortedDevices = devices;

	const collator = new Intl.Collator([], { numeric: true });
	if (sortDirection == 'ascending') {
		switch (sortKey) {
			case 'id':
				sortedDevices = devices.sort((a: HeadScaleDevice, b: HeadScaleDevice) =>
					collator.compare(a.id, b.id)
				);
				break;
			case 'lastSeen':
				sortedDevices = devices.sort(
					(a: HeadScaleDevice, b: HeadScaleDevice) =>
						-collator.compare(a.lastSeen, b.lastSeen)
				);
				break;
			case 'givenName':
				sortedDevices = devices.sort((a: HeadScaleDevice, b: HeadScaleDevice) =>
					collator.compare(a.givenName, b.givenName)
				);
				break;
		}
	}
	if (sortDirection == 'descending') {
		switch (sortKey) {
			case 'id':
				sortedDevices = devices.sort(
					(a: HeadScaleDevice, b: HeadScaleDevice) =>
						-collator.compare(a.id, b.id)
				);
				break;
			case 'lastSeen':
				sortedDevices = devices.sort((a: HeadScaleDevice, b: HeadScaleDevice) =>
					collator.compare(a.lastSeen, b.lastSeen)
				);
				break;
			case 'givenName':
				sortedDevices = devices.sort(
					(a: HeadScaleDevice, b: HeadScaleDevice) =>
						-collator.compare(a.givenName, b.givenName)
				);
				break;
		}
	}

	return sortedDevices;
}
