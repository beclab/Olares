type EGoToStateFunc<State, Args extends State[] = State[]> = (
	...args: Args
) => State;
type EToState<State> = State | EGoToStateFunc<State>;

interface ITransitionDir<State> {
	from: State | State[] | '*';
	to: EToState<State>;
	onTransition?: (from: State, to: State) => void;
}

type ITransitions<T, State> = {
	[P in keyof T]: ITransitionDir<State> | ITransitionDir<State>[];
};

type TransitionCall<T, State, Args extends State[] = State[]> = {
	[P in keyof T]: (...args: Args) => void;
};

interface STOptions<T, State> {
	init: State;
	onBefore?: (from: State, to: State) => void;
	onAfter?: (from: State, to: State) => void;
	transitions: ITransitions<T, State>;
}

export function BuildTransition<State>(
	from: State | State[] | '*',
	to: EToState<State>,
	onTransition?: (from: State, to: State) => void
): ITransitionDir<State> {
	return { from, to, onTransition };
}

export class StateMachine<T, State> {
	private _transitions: TransitionCall<T, State>;
	private _curState: State;
	private _originTransitions: ITransitions<T, State>;

	private _isTransiting = false;

	onBefore?: (from: State, to: State) => void;
	onAfter?: (from: State, to: State) => void;
	onError?: (code: number, reason: string) => void;

	constructor(option: STOptions<T, State>) {
		const { init, transitions, onBefore, onAfter } = option;

		this._curState = init;
		this.onBefore = onBefore;
		this.onAfter = onAfter;

		this.setupTransitions(transitions);
	}

	private setupTransitions(transitions: ITransitions<T, State>) {
		this._originTransitions = transitions;
		this._transitions = {} as any;

		Object.keys(transitions).forEach((k) => {
			const key = k as keyof T;

			const value: ITransitionDir<State> | ITransitionDir<State>[] =
				transitions[key];

			this._transitions[key] = (...extraArgs: any[]) => {
				if (!this.can(key)) {
					this.postError(
						1000,
						`You can not '${String(key)}' now. Current state is ${
							this._curState
						}`
					);
					return;
				}

				if (this._isTransiting) {
					this.postError(
						1011,
						`This is transiting now. You cannot transition more times at one time.`
					);
					return;
				}

				const curState = this._curState;
				const dir = this._findDir(curState, value);

				if (dir == null) {
					this.postError(
						1000,
						`You can not '${String(key)}' now. Current state is ${
							this._curState
						}`
					);
					return;
				}

				const { to, onTransition } = dir;
				const toState =
					typeof to === 'function'
						? (to as EGoToStateFunc<State>)(...extraArgs)
						: to;

				if (toState != curState) {
					this._isTransiting = true;
					this.onBefore && this.onBefore(curState, toState);
					onTransition && onTransition(curState, toState);
					this._curState = toState;
					this.onAfter && this.onAfter(curState, toState);
					this._isTransiting = false;
				}
			};
		});
	}

	private postError(code: number, reason: string) {
		this.logError(reason);
		this.onError && this.onError(code, reason);
	}

	private logError(reason: any) {
		console.error(reason);
	}

	transition() {
		return this._transitions;
	}

	/** get cur state */
	state() {
		return this._curState;
	}

	/** current is `state` */
	is(state: State) {
		return this._curState == state;
	}

	/**
	 * Can transition `t` now.
	 * @param t
	 */
	can(t: keyof T) {
		const value: ITransitionDir<State> | ITransitionDir<State>[] =
			this._originTransitions[t];
		if (!value) {
			return false;
		}

		const dir = this._findDir(this._curState, value);
		return dir != null;
	}

	private _findDir(
		from: State,
		dirs: ITransitionDir<State> | ITransitionDir<State>[]
	) {
		if (Array.isArray(dirs)) {
			return this._findDirOfArray(from, dirs);
		}

		if (this._isIncludeState(dirs.from, from)) {
			return dirs;
		}

		return null;
	}

	private _findDirOfArray(from: State, dirs: ITransitionDir<State>[]) {
		for (const index in dirs) {
			const dir = dirs[index];
			if (this._isIncludeState(dir.from, from)) {
				return dir;
			}
		}

		return null;
	}

	private _isIncludeState(state: State | State[] | '*', targetState: State) {
		if (state === '*') {
			return true;
		}

		if (targetState === state) {
			return true;
		}

		if (Array.isArray(state) && state.indexOf(targetState) != -1) {
			return true;
		}

		return false;
	}
}
