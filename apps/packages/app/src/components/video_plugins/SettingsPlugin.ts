import videojs from 'video.js';

const SettingsPlugin = function (this: any) {
	const player = this;

	const settingsButton = videojs.dom.createEl('button', {
		className: 'vjs-settings-button vjs-control',
		innerHTML: '<img src="/video/settings.svg" />'
	});

	const settingsPanel = videojs.dom.createEl('div', {
		className: 'vjs-settings-panel',
		innerHTML: `
			<div class="tabs">
          <div class="tab active" tab-id="tab1">Subtitles</div>
          <div class="tab" tab-id="tab2">Audio Track</div>
          <div class="tab" tab-id="tab3">Audio Channel</div>
      </div>
      <div class="tab-content" id="tab1" style="display: block;">
          <li data-event="tab1-1">这是选项卡 1 的内容。</li>
          <li>这是选项卡 1 的内容。</li>
          <li>这是选项卡 1 的内容。</li>
          <li>这是选项卡 1 的内容。</li>
          <li>这是选项卡 1 的内容。</li>

      </div>
      <div class="tab-content" id="tab2">
          <li>这是选项卡 2 的内容。</li>
          <li>这是选项卡 2 的内容。</li>
          <li>这是选项卡 2 的内容。</li>

      </div>
      <div class="tab-content" id="tab3">
          <li>这是选项卡 3 的内容。</li>
          <li>这是选项卡 3 的内容。</li>
          <li>这是选项卡 3 的内容。</li>
          <li>这是选项卡 3 的内容。</li>
          <li>这是选项卡 3 的内容。</li>
          <li>这是选项卡 3 的内容。</li>
          <li>这是选项卡 3 的内容。</li>
          <li>这是选项卡 3 的内容。</li>
          <li>这是选项卡 3 的内容。</li>
      </div>
    `
	});

	const popup = document.getElementsByClassName('vjs-settings-panel');
	const parent = document.getElementById('video-previewer');

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
				settingsPanel.classList.remove('active');
			}
		});

	settingsButton.addEventListener('click', (event) => {
		settingsPanel.classList.toggle('active');
		const qualityPanel = document.getElementsByClassName('vjs-quality-panel');
		qualityPanel[0].classList.remove('active');
		event.stopPropagation();
	});

	settingsPanel.querySelectorAll('.tab').forEach((item) => {
		item.addEventListener('click', (evt: any) => {
			const tabName = evt.target.getAttribute('tab-id');
			const tabcontent: any = document.getElementsByClassName('tab-content');
			for (let i = 0; i < tabcontent.length; i++) {
				tabcontent[i].style.display = 'none';
			}
			const tablinks = document.getElementsByClassName('tab');
			for (let i = 0; i < tablinks.length; i++) {
				tablinks[i].className = tablinks[i].className.replace(' active', '');
			}
			const tabName2: any = document.getElementById(tabName);
			tabName2.style.display = 'block';
			evt.currentTarget.className += ' active';

			evt.stopPropagation();
		});
	});

	settingsPanel.querySelectorAll('.tab-content').forEach((item) => {
		item.addEventListener('click', (evt: any) => {
			const target = evt.target.getAttribute('data-event');
			console.log('target', target);
			player.play();

			evt.stopPropagation();
		});
	});

	player.controlBar.el().appendChild(settingsButton);
	player.el().appendChild(settingsPanel);
};

export default SettingsPlugin;
