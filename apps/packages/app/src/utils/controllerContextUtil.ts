import { Object } from 'ts-toolbelt';
import { Controller as Controller } from 'src/platform/interface/bex/background/controller';

type IExtractFromPromise<T> = T extends Promise<infer U> ? U : T;

export type ControllerType = Object.Merge<
	{
		[key in keyof Controller]: Controller[key] extends (
			...args: infer ARGS
		) => infer RET
			? <T extends IExtractFromPromise<RET> = IExtractFromPromise<RET>>(
					...args: ARGS
			  ) => Promise<IExtractFromPromise<T>>
			: Controller[key];
	},
	Record<string, <T = any>(...params: any) => Promise<T>>
>;
