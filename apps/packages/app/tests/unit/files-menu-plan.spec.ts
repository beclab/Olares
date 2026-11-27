jest.mock('src/api', () => ({
	isShareEnable: jest.fn(() => true)
}));

import { isShareEnable } from 'src/api';
import {
	MENU_REFRESH_RETRY_DELAYS_MS,
	MenuRefreshRetrySchedule,
	resolveMenuGroups
} from 'src/domain/files/menu-plan';
import { DriveType } from 'src/utils/interface/files';

const mockedIsShareEnable = isShareEnable as jest.MockedFunction<
	typeof isShareEnable
>;

describe('resolveMenuGroups', () => {
	beforeEach(() => {
		mockedIsShareEnable.mockReturnValue(true);
	});

	it('maps External, Cache, and Common to parent menu origins', () => {
		const groups = resolveMenuGroups([
			DriveType.Drive,
			DriveType.External,
			DriveType.Sync,
			DriveType.Data,
			DriveType.Cache,
			DriveType.Common,
			DriveType.GoogleDrive,
			DriveType.Share
		]);

		expect(groups.map((group) => group.menuOrigin)).toEqual([
			DriveType.Drive,
			DriveType.Sync,
			DriveType.Data,
			DriveType.GoogleDrive,
			DriveType.Share
		]);
	});

	it('skips child origins when parent is already included', () => {
		const groups = resolveMenuGroups([DriveType.Drive, DriveType.External]);

		expect(groups).toHaveLength(1);
		expect(groups[0].menuOrigin).toBe(DriveType.Drive);
	});

	it('omits Sync and Share when sharing is disabled', () => {
		mockedIsShareEnable.mockReturnValue(false);

		const groups = resolveMenuGroups([
			DriveType.Drive,
			DriveType.Sync,
			DriveType.Share
		]);

		expect(groups.map((group) => group.menuOrigin)).toEqual([DriveType.Drive]);
	});

	it('skips Data on mobile platform', () => {
		const originalPlatform = process.env.PLATFORM;
		process.env.PLATFORM = 'MOBILE';

		const groups = resolveMenuGroups([
			DriveType.Drive,
			DriveType.Data,
			DriveType.Cache
		]);

		expect(groups.map((group) => group.menuOrigin)).toEqual([DriveType.Drive]);

		process.env.PLATFORM = originalPlatform;
	});
});

describe('MenuRefreshRetrySchedule', () => {
	it('hands out every configured delay in order, starting with the first', () => {
		const schedule = new MenuRefreshRetrySchedule();
		expect(MENU_REFRESH_RETRY_DELAYS_MS).toEqual([2000, 5000, 15000, 30000]);

		const delays = [
			schedule.next(),
			schedule.next(),
			schedule.next(),
			schedule.next()
		];

		expect(delays).toEqual(MENU_REFRESH_RETRY_DELAYS_MS);
	});

	it('stops handing out delays once the last slot is used', () => {
		const schedule = new MenuRefreshRetrySchedule([1000]);

		expect(schedule.next()).toBe(1000);
		expect(schedule.next()).toBeUndefined();
	});

	it('resets attempt counters', () => {
		const schedule = new MenuRefreshRetrySchedule();
		schedule.next();
		schedule.reset();
		expect(schedule.getAttempt()).toBe(0);
		expect(schedule.next()).toBe(2000);
	});
});
