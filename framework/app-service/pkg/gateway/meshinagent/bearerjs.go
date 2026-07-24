package meshinagent

// BearerJS is the njs module loaded by RenderNginxConf for per-request JWT reads.
// The token file is refreshed by kubelet when the caller-jwt Secret is updated.
func BearerJS() string {
	path := JWTSecretMountPath + "/token"
	return "" +
		"var fs = require('fs');\n" +
		"\n" +
		"function readJWT(r) {\n" +
		"    try {\n" +
		"        var t = fs.readFileSync('" + path + "');\n" +
		"        if (t === undefined || t === null) {\n" +
		"            return '';\n" +
		"        }\n" +
		"        return t.toString().replace(/\\s+$/g, '');\n" +
		"    } catch (e) {\n" +
		"        return '';\n" +
		"    }\n" +
		"}\n" +
		"\n" +
		"export default {readJWT};\n"
}
