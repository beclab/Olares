export const chartData = {
	title: 'test',
	dimensions: ['product', '2015', '2016', '2017', '2018'],
	source: new Array(20).fill('').map((item: any, index: number) => ({
		product: `name${index}`,
		'2015': Number(Math.random() * 400).toFixed(1),
		'2016': Number(Math.random() * 100).toFixed(1),
		'2017': Number(Math.random() * 20).toFixed(1),
		'2018': Number(Math.random() * 100).toFixed(1)
	}))
};
