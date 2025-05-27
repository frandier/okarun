package anime

var AddionalAnimeInfoJsCode = `
(() => {
	const listItems = document.querySelectorAll('.card-bod ul li');
	const data = {};

	listItems.forEach(li => {
		const labelSpan = li.querySelector('span');
		if (!labelSpan) return;

		const key = labelSpan.textContent.replace(':', '').trim().toLowerCase();

		const contentNodes = Array.from(li.childNodes).slice(1);

		const values = contentNodes
		.map(node => {
			if (node.nodeType === Node.TEXT_NODE) {
			const text = node.textContent.trim();
			// Filtra comas sueltas u otros separadores vacíos
			return text === ',' || text === '' ? null : text;
			} else if (node.nodeType === Node.ELEMENT_NODE && node.tagName === 'A') {
			return node.textContent.trim();
			} else if (node.nodeType === Node.ELEMENT_NODE && node.classList.contains('currently')) {
			return node.textContent.trim();
			}
			return null;
		})
		.filter(text => text); // filtra null y ''

		data[key] = values.length === 1 ? values[0] : values;
	});

  return data;
})()
`
