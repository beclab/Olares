import { getCurrentOrigin } from 'src/utils/interface/device';

export const BroadcastChannelBex = 'BroadcastChannelBex';

export const useChannelBexPost = <T>(key: string, value: T) => {
	const channelBexPost = new BroadcastChannel(BroadcastChannelBex);
	channelBexPost.postMessage({
		key: key,
		newValue: value,
		from: getCurrentOrigin()
	});
	channelBexPost.close();
};
