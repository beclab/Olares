import { app } from 'src/globals';
import { useUserStore } from 'src/stores/user';
import { sendLock } from 'src/utils/bexFront';

let _lockTimeout = 0;
let _syncTimeout: any = null;

export const homeMounted = () => {
	resume();
	addEventListener('click', startLockTimer);
	addEventListener('keydown', startLockTimer);
};

export const homeUnMounted = () => {
	_pause();
	removeEventListener('click', startLockTimer);
	removeEventListener('keydown', startLockTimer);
};

function startLockTimer() {
	_cancelAutoLock();
	if (app.settings.autoLock && !app.state.locked) {
		_lockTimeout = window.setTimeout(
			() => _doLock(),
			app.settings.autoLockDelay * 60 * 1000
		);
	}
}

function _cancelAutoLock() {
	if (_lockTimeout) {
		clearTimeout(_lockTimeout);
	}
}

async function _doLock() {
	if (!app.settings.autoLock) {
		return false;
	}
	if (app.state.syncing) {
		startLockTimer();
		return;
	}

	await app.lock();
	const user = useUserStore();
	await user.users?.lock();
	sendLock();
}

function _pause() {
	if (_syncTimeout) {
		clearTimeout(_syncTimeout);
	}
}

function resume() {
	startLockTimer();
	_doSync();
}

async function _doSync() {
	_syncTimeout = setTimeout(async () => {
		if (app.state.loggedIn && !app.state.locked) {
			try {
				await app.synchronize();
			} catch (e) {
				console.error(e);
			}
		}
		_doSync();
	}, app.settings.syncInterval * 60 * 1000);
}
