
var jsrp = require('jsrp');
var client = new jsrp.client();

i = process.env.OWNER
p = process.env.PASSWORD

client.init({ username: i, password: p }, function () {
	client.createVerifier(function(err, result) {
		console.log("%s:%s", result.salt, result.verifier);
	});
});