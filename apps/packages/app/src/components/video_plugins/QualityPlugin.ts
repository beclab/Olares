import videojs from 'video.js';

interface QualitySources {
	[key: string]: string;
}

const QualityPlugin = function (this: any) {
	const player = this;
	const qualitySources: QualitySources = {
		low: 'YOUR_LOW_QUALITY_VIDEO_SOURCE.mp4',
		medium: 'YOUR_MEDIUM_QUALITY_VIDEO_SOURCE.mp4',
		high: 'YOUR_HIGH_QUALITY_VIDEO_SOURCE.mp4'
	};

	const qualityButton = videojs.dom.createEl('button', {
		className: 'vjs-quality-button vjs-control',
		innerHTML: 'Quality'
	});

	const qualityPanel = videojs.dom.createEl('div', {
		className: 'vjs-quality-panel',
		innerHTML: `
      <ul>
        <li data-quality="low">Low</li>
        <li data-quality="medium">Medium</li>
        <li data-quality="high">High</li>
      </ul>
    `
	});

	const parent = document.getElementById('video-previewer');
	const popup = document.getElementsByClassName('vjs-quality-panel');

	parent &&
		parent.addEventListener('click', function (event: any) {
			let isDescendant = false;
			for (let i = 0; i < popup.length; i++) {
				if (popup[i].contains(event.target)) {
					isDescendant = true;
					break;
				}
			}

			if (!isDescendant) {
				qualityPanel.classList.remove('active');
			}
			event.stopPropagation();
		});

	qualityButton.addEventListener('click', (event) => {
		qualityPanel.classList.toggle('active');

		const settingsPanel = document.getElementsByClassName('vjs-settings-panel');
		settingsPanel[0].classList.remove('active');
		event.stopPropagation();
	});

	qualityPanel.querySelectorAll('li').forEach((item) => {
		item.addEventListener('click', (event) => {
			const target = event.target as HTMLElement;
			const quality = target.getAttribute('data-quality');
			if (quality && qualitySources[quality]) {
				player.src({ type: 'video/mp4', src: qualitySources[quality] });
				player.play();
			}
			qualityPanel.classList.remove('active');
			event.stopPropagation();
		});
	});

	player.controlBar.el().appendChild(qualityButton);
	player.el().appendChild(qualityPanel);
};

export default QualityPlugin;
