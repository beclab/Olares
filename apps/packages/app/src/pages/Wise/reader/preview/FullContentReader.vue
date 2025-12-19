<template>
	<div
		id="document-text-content"
		:class="marginTop ? 'q-mt-lg' : ''"
		v-if="readerStore.readingEntry && readerStore.realContent"
		v-html="readerStore.realContent"
	/>
</template>

<script setup lang="ts">
import { useReaderStore } from '../../../../stores/rss-reader';
import { nextTick, onUnmounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import hljs from 'highlight.js';
import 'highlight.js/styles/github.css';
import { getApplication } from '../../../../application/base';

defineProps({
	marginTop: {
		type: Boolean,
		default: false
	}
});

const readerStore = useReaderStore();
const { t } = useI18n();

watch(
	() => [readerStore.readingEntry, readerStore.realContent],
	() => {
		if (readerStore.readingEntry && readerStore.realContent) {
			nextTick(() => {
				// quote
				document
					.querySelectorAll('#document-text-content blockquote')
					.forEach(function (el: any) {
						el.style.borderLeft = '';
					});
				// table
				document
					.querySelectorAll(
						'#document-text-content table th, #document-text-content table td'
					)
					.forEach(function (el: any) {
						el.style.border = '';
					});

				// pre code
				document
					.querySelectorAll('#document-text-content pre')
					.forEach((pre) => {
						let codeElement = pre.querySelector('code');
						let codeText = codeElement
							? codeElement.textContent
							: pre.textContent;

						let lang = codeElement ? codeElement.getAttribute('lang') : null;

						if (!lang && codeElement) {
							let className = codeElement.className;
							let langMatch = className.match(/language-(\w+)/);
							lang = langMatch ? langMatch[1] : null;
						}

						let highlightedCode: any;
						if (lang && hljs.getLanguage(lang)) {
							try {
								highlightedCode = hljs.highlight(codeText, {
									language: lang,
									ignoreIllegals: true
								}).value;
							} catch (error) {
								console.error('Highlighting error:', error);
								highlightedCode = hljs.highlightAuto(codeText).value;
							}
						} else {
							highlightedCode = hljs.highlightAuto(codeText).value;
						}

						const container = document.createElement('section');
						container.className = 'code-container';

						const lineNumbers = document.createElement('div');
						lineNumbers.className = 'line-numbers';

						const copyButton = document.createElement('button');
						copyButton.className = 'copy-button';
						copyButton.textContent = t('copy');
						const copyHandler = () => {
							getApplication()
								.copyToClipboard(pre.textContent)
								.then(() => {
									BtNotify.show({
										type: NotifyDefinedType.SUCCESS,
										message: t('copy_success')
									});

									copyButton.textContent = t('copy_success');

									setTimeout(() => {
										copyButton.textContent = t('copy');
									}, 3000);
								})
								.catch((e) => {
									BtNotify.show({
										type: NotifyDefinedType.FAILED,
										message: t('copy_failure_message', e.message)
									});
								});
						};
						copyButton.addEventListener('click', copyHandler);
						(copyButton as any)._copyHandler = copyHandler;

						const lines = pre.textContent.split('\n').length;
						let lineNumberHtml = '';
						for (let i = 1; i <= lines; i++) {
							lineNumberHtml += `<div class="line-number">${i}</div>`;
						}
						lineNumbers.innerHTML = lineNumberHtml;

						const codeContent = document.createElement('div');
						codeContent.className = 'code-content';
						codeContent.innerHTML = `<pre><code class="hljs ${lang}">${highlightedCode}</code></pre>`;

						function generateLineNumbers() {
							const preElement = codeContent.querySelector('pre');
							const codeHeight = preElement.offsetHeight;
							const lineHeight = parseFloat(
								getComputedStyle(preElement).lineHeight
							);

							const lines = Math.ceil(codeHeight / lineHeight);

							let lineNumberHtml = '';
							for (let i = 1; i <= lines; i++) {
								lineNumberHtml += `<div class="line-number">${i}</div>`;
							}
							lineNumbers.innerHTML = lineNumberHtml;
						}

						pre.parentNode.replaceChild(container, pre);
						container.appendChild(lineNumbers);
						container.appendChild(codeContent);
						container.appendChild(copyButton);

						generateLineNumbers();

						const scrollListener = () => {
							lineNumbers.scrollTop = codeContent.scrollTop;
						};
						const resizeListener = generateLineNumbers;

						codeContent.addEventListener('scroll', scrollListener);
						window.addEventListener('resize', resizeListener);

						(container as any)._scrollListener = scrollListener;
						(container as any)._resizeListener = resizeListener;
					});
			});
		}
	},
	{
		immediate: true
	}
);

onUnmounted(() => {
	document.querySelectorAll('.code-container').forEach((container) => {
		container.querySelectorAll('.copy-button').forEach((button) => {
			const handler = (button as any)._copyHandler;
			if (handler) {
				button.removeEventListener('click', handler);
				delete (button as any)._copyHandler;
			}
		});

		const scrollListener = (container as any)._scrollListener;
		const resizeListener = (container as any)._resizeListener;

		if (scrollListener) {
			const codeContent = container.querySelector('.code-content');
			codeContent.removeEventListener('scroll', scrollListener);
		}

		if (resizeListener) {
			window.removeEventListener('resize', resizeListener);
		}

		delete (container as any)._scrollListener;
		delete (container as any)._resizeListener;
	});
});
</script>

<style scoped lang="scss"></style>
