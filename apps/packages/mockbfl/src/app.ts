// server.ts
import express, { Express, Request, Response, NextFunction } from 'express';
import { ApiRouter } from './routes/api-routes';

const port = 5010;

const app: Express = express();
const apiRouter = new ApiRouter();

function getOrigin(request: Request) {
	const origin = request.headers.origin;
	if (origin) {
		return origin;
	}
	return '*';
}
app.use((request: Request, response: Response, next: NextFunction) => {
	const origin = getOrigin(request);
	console.log('origin', origin);
	if (origin.startsWith('http://localhost')) {
		response.setHeader('Access-Control-Allow-Origin', origin);
	}

	response.setHeader('Access-Control-Allow-Credentials', 'true');
	response.setHeader(
		'Access-Control-Allow-Headers',
		'Origin, X-Requested-With, Content-Type, Accept, Authorization'
	);
	response.setHeader(
		'Access-Control-Allow-Methods',
		'GET, POST, PATCH, DELETE'
	);

	next();
});

app.use(express.json());
app.use('', apiRouter.router);

app.use((request: Request, response: Response) => {
	response.type('text/plain');
	response.status(404);
	response.send('Page is not found.');
});

async function start() {
	app.listen(port, async () => {
		console.log(`server is running on http://localhost:` + port);
	});
}

start();
