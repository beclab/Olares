import { WebSocketStatusEnum } from '@bytetrade/core';

declare const self: SharedWorkerGlobalScope;

const connections = new Set<MessagePort>();
let leaderPort: MessagePort | null = null;
let lastStatus = WebSocketStatusEnum.close;
// Last durable-state message (e.g. copy/cut clipboard) flagged with `retain`,
// replayed to tabs that register later. Transient events are never stored here.
let retainedMessage: unknown = null;
// True once a tab has used worker-mediated election (the fallback when the
// Web Locks API is unavailable). In this mode the worker grants leadership to
// a single tab and re-grants it when the leader disconnects.
let claimMode = false;
// A tab is `leaderPort` from the moment it is elected, but it cannot handle
// forwarded work until it has built its connect params and created the socket
// bean. `leaderReady` flips on the `leader-ready` signal; until then outbound
// send/apply work is buffered so it is not dropped during leader setup or
// hand-off. Capped to avoid unbounded growth if no leader ever becomes ready.
let leaderReady = false;
const MAX_PENDING = 500;
let pendingForLeader: unknown[] = [];
// Liveness tracking for worker-mediated election (the fallback when the Web
// Locks API is unavailable, e.g. an http:// origin that is not a secure
// context). Every connected tab — not just the leader — sends a periodic
// 'client-heartbeat'; the worker records the last time it heard from each port
// in `lastSeen`. This is what lets the worker tell live tabs from dead ones:
// a SharedWorker gets no event when a tab closes/crashes, and tabs that unload
// without sending 'disconnect' would otherwise linger forever in `connections`
// and poison election (leadership could be granted to a dead port, leaving the
// app with no working leader). The watchdog prunes ports that have gone silent
// longer than CLIENT_DEAD_MS and elects only among ports still considered live.
const WATCHDOG_INTERVAL_MS = 500;
// A tab heartbeats every ~500ms (see the store). Allow three missed beats
// before presuming it dead so brief main-thread jank is tolerated, while a
// crashed/killed leader is detected in ~1.5s (plus one watchdog tick) to keep
// the inbound-message loss window — the gap where the WebSocket is down and
// server pushes have no leader to receive them — as small as a timeout-based
// scheme can without over-aggressively replacing a briefly-janky leader.
const CLIENT_DEAD_MS = 1500;
const lastSeen = new Map<MessagePort, number>();
let watchdogTimer: ReturnType<typeof setInterval> | null = null;

function touch(port: MessagePort) {
	lastSeen.set(port, Date.now());
}

function isLive(port: MessagePort): boolean {
	return Date.now() - (lastSeen.get(port) ?? 0) <= CLIENT_DEAD_MS;
}

function broadcast(message: unknown) {
	connections.forEach((p) => p.postMessage(message));
}

// Pick the most-recently-seen live port (optionally excluding one), so election
// prefers the tab most likely to still be alive and never targets a dead port.
function pickLiveLeader(exclude?: MessagePort | null): MessagePort | null {
	let best: MessagePort | null = null;
	let bestAt = -1;
	const now = Date.now();
	for (const candidate of connections) {
		if (candidate === exclude) {
			continue;
		}
		const seen = lastSeen.get(candidate) ?? 0;
		if (now - seen <= CLIENT_DEAD_MS && seen > bestAt) {
			best = candidate;
			bestAt = seen;
		}
	}
	return best;
}

function grantLeadership(port: MessagePort) {
	leaderPort = port;
	leaderReady = false;
	touch(port);
	port.postMessage({ type: 'grant-leader' });
}

// Remove a port that is gone: drop it from tracking and, if it was the leader,
// revoke it (a no-op if truly dead, but if it was merely throttled this makes it
// tear down its socket so we never end up with two live connections) and clear
// leadership so a live tab can be elected.
function dropPort(port: MessagePort) {
	connections.delete(port);
	lastSeen.delete(port);
	if (port === leaderPort) {
		port.postMessage({ type: 'leader-revoked' });
		leaderPort = null;
		leaderReady = false;
		lastStatus = WebSocketStatusEnum.close;
		broadcast({ type: 'status', data: { status: lastStatus } });
	}
}

function startWatchdog() {
	if (watchdogTimer != null) {
		return;
	}
	watchdogTimer = setInterval(() => {
		if (!claimMode) {
			return;
		}
		// Prune dead ports. Never prune the sole remaining connection: with no
		// failover target, dropping it would strand the app; let that lone tab
		// manage its own socket until another tab appears.
		for (const port of [...connections]) {
			if (connections.size <= 1) {
				break;
			}
			if (!isLive(port)) {
				dropPort(port);
			}
		}
		// Ensure a live leader exists. Covers a leader that was just pruned, one
		// that resigned, or a first grant once a second tab has appeared.
		if (!leaderPort) {
			const next = pickLiveLeader();
			if (next) {
				grantLeadership(next);
			}
		}
	}, WATCHDOG_INTERVAL_MS);
}

function forwardToLeader(message: unknown) {
	if (leaderPort && leaderReady) {
		leaderPort.postMessage(message);
		return;
	}
	// No ready leader yet: buffer until one signals readiness.
	pendingForLeader.push(message);
	if (pendingForLeader.length > MAX_PENDING) {
		pendingForLeader.shift();
	}
}

function flushPendingToLeader() {
	if (!leaderPort || !leaderReady || pendingForLeader.length === 0) {
		return;
	}
	const queued = pendingForLeader;
	pendingForLeader = [];
	queued.forEach((message) => leaderPort?.postMessage(message));
}

self.onconnect = (event) => {
	const port = event.ports[0];

	port.onmessage = (messageEvent) => {
		const { type, data, retain, force } = messageEvent.data;

		// Any message from a port proves it is alive; refresh its liveness so the
		// watchdog does not prune a tab that is actively talking to the worker.
		// A port that reappears after being pruned (e.g. a throttled tab that
		// resumed) is re-added to `connections` here.
		touch(port);
		if (type !== 'disconnect' && type !== 'dispose') {
			connections.add(port);
		}

		switch (type) {
			case 'client-heartbeat':
				// Liveness only; timestamp already refreshed above.
				break;
			case 'register':
				connections.add(port);
				// Sync the connection status to a newly registered tab. Transient
				// app events are never replayed (that would re-run side-effecting
				// handlers and duplicate UI effects), but durable retained state
				// (e.g. the current clipboard) is replayed so the new tab is in sync.
				port.postMessage({
					type: 'status',
					data: { status: lastStatus }
				});
				if (retainedMessage != null) {
					port.postMessage({
						type: 'message',
						data: retainedMessage
					});
				}
				break;
			case 'set-leader':
				leaderPort = port;
				leaderReady = false;
				break;
			case 'leader-ready':
				// The elected leader has created its socket bean and can now
				// receive forwarded work. Flush anything buffered during setup.
				if (port === leaderPort) {
					leaderReady = true;
					flushPendingToLeader();
				}
				break;
			case 'claim-leader':
				// Web Locks unavailable: elect a single leader via the worker so
				// only one tab opens the main-thread WebSocket.
				claimMode = true;
				startWatchdog();
				if (!leaderPort) {
					grantLeadership(port);
				} else if (port !== leaderPort && !isLive(leaderPort)) {
					// The current leader has gone silent (crashed/killed without
					// sending 'disconnect'). Revoke it and hand leadership to this
					// live claimant so a dead leader cannot block re-election.
					leaderPort.postMessage({ type: 'leader-revoked' });
					grantLeadership(port);
				}
				break;
			case 'leader-status':
				lastStatus = data.status;
				broadcast({
					type: 'status',
					data: { status: lastStatus }
				});
				break;
			case 'leader-message':
				if (retain) {
					retainedMessage = data;
				}
				broadcast({
					type: 'message',
					data
				});
				break;
			case 'send':
				forwardToLeader({
					type: 'do-send',
					data
				});
				break;
			case 'recreate-leader':
				// Only meaningful once a leader is ready; while the leader is mid
				// setup a recreate is redundant and is safely ignored.
				if (leaderPort && leaderReady) {
					leaderPort.postMessage({
						type: 'do-recreate',
						data,
						force
					});
				}
				break;
			case 'resign-leader':
				// The granted tab could not open the connection (e.g. transient
				// empty URL). Drop it as leader and, in worker-mediated election
				// mode, hand leadership to a different live tab.
				if (port === leaderPort) {
					leaderPort = null;
					leaderReady = false;
					if (claimMode) {
						const next = pickLiveLeader(port);
						if (next) {
							grantLeadership(next);
						}
					}
				}
				break;
			case 'status':
				port.postMessage({
					type: 'status',
					data: { status: lastStatus }
				});
				break;
			case 'disconnect':
				connections.delete(port);
				lastSeen.delete(port);
				if (port === leaderPort) {
					leaderPort = null;
					leaderReady = false;
					lastStatus = WebSocketStatusEnum.close;
					broadcast({
						type: 'status',
						data: { status: lastStatus }
					});
					// In worker-mediated election mode, hand leadership to another
					// live tab so the connection is re-established. With Web Locks
					// the OS releases the lock and a waiting tab self-elects instead.
					if (claimMode) {
						const next = pickLiveLeader(port);
						if (next) {
							grantLeadership(next);
						}
					}
				}
				break;
			case 'dispose':
				connections.clear();
				lastSeen.clear();
				leaderPort = null;
				leaderReady = false;
				lastStatus = WebSocketStatusEnum.close;
				retainedMessage = null;
				claimMode = false;
				pendingForLeader = [];
				if (watchdogTimer != null) {
					clearInterval(watchdogTimer);
					watchdogTimer = null;
				}
				break;
			default:
				forwardToLeader({
					type: 'do-apply',
					data: messageEvent.data
				});
		}
	};
};
