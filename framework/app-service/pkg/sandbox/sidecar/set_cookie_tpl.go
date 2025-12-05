package sidecar

import (
	corev1 "k8s.io/api/core/v1"
)

func getHTTProbePath(pod *corev1.Pod) (probesPath []string) {
	for _, c := range pod.Spec.Containers {
		if c.LivenessProbe != nil && c.LivenessProbe.HTTPGet != nil {
			probesPath = append(probesPath, c.LivenessProbe.HTTPGet.Path)
		}
		if c.ReadinessProbe != nil && c.ReadinessProbe.HTTPGet != nil {
			probesPath = append(probesPath, c.ReadinessProbe.HTTPGet.Path)
		}
		if c.StartupProbe != nil && c.StartupProbe.HTTPGet != nil {
			probesPath = append(probesPath, c.StartupProbe.HTTPGet.Path)
		}
	}
	return probesPath
}

const envoySetCookie = `
local pattern = "Domain=([^;]*)"
function split(str, sep)
    local ret = {}
    for s in string.gmatch(str, "([^"..sep.."]+)") do
        table.insert(ret, s)
    end
    return ret
end
function replace(str, pattern, replacement)
    return string.gsub(str, pattern, replacement)
end

function reset_cookie_domain(cookie)
    local reset_cookie = cookie
    -- get domain from set-cookie-string
    local set_cookie_domain = string.match(cookie, pattern)
    if set_cookie_domain == nil or set_cookie_domain == "" then
    else
        reset_cookie = replace(cookie, pattern, "Domain=")
    end
    return reset_cookie
end

function envoy_on_response(response_handle)
	local headers = response_handle:headers()
	local n = headers:getNumValues("Set-Cookie")
	local cookies = {}
	for i = 0, n - 1 do
  		local v = headers:getAtIndex("Set-Cookie", i)
  		if v and v ~= "" then
    		table.insert(cookies, v)
  		end
	end
	local first = true
	for _, cookie in pairs(cookies) do
		local updated = reset_cookie_domain(cookie)
		if first then
			response_handle:headers():replace("Set-Cookie", updated)
			first = false
		else
			response_handle:headers():add("Set-Cookie", updated)
		end
	end
end
`

func genEnvoySetCookieScript() []byte {
	return []byte(envoySetCookie)
}
