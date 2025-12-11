package controllers

import (
	"context"
	"fmt"

	"bytetrade.io/web3os/bfl/internal/ingress/controllers/config"
	"bytetrade.io/web3os/bfl/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var patches = map[string][]func(ctx context.Context, r *NginxController, s *config.Server) (*config.Server, error){
	"files": {
		filesNodeApiPatch,
	},
	"settings": {
		filesNodeApiPatch,
	},
}

// var locationAdditionalsCommon = func(node string) []string {
// 	return []string{
// 		"auth_request /authelia-verify;",
// 		"auth_request_set $remote_token $upstream_http_remote_accesstoken;",
// 		"proxy_set_header Remote-Accesstoken $remote_token;",
// 		"proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;",
// 		"proxy_set_header X-Forwarded-Host $host;",
// 		"client_body_timeout 60s;",
// 		"keepalive_timeout 75s;",
// 		"proxy_read_timeout 60s;",
// 		"proxy_send_timeout 60s;",
// 		"proxy_set_header X-BFL-USER " + constants.Username + ";",
// 		"proxy_set_header X-Authorization $http_x_authorization;",
// 		"proxy_set_header X-Terminus-Node " + node + ";",
// 	}
// }

var locationAdditionalsForFilesOp = func(node string) []string {
	return []string{
		"auth_request /authelia-verify;",
		"auth_request_set $remote_token $upstream_http_remote_accesstoken;",
		"auth_request_set $authelia_nonce $upstream_http_authelia_nonce;",
		"proxy_set_header Remote-Accesstoken $remote_token;",
		"proxy_set_header X-Real-IP $remote_addr;",
		"proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;",
		"proxy_set_header X-Forwarded-Host $host;",
		"proxy_set_header X-Provider-Proxy $proxy_host;",
		"proxy_set_header Authelia-Nonce $authelia_nonce;",
		"client_body_timeout 60s;",
		"client_max_body_size 2000M;",
		"proxy_request_buffering off;",
		"keepalive_timeout 75s;",
		"proxy_read_timeout 60s;",
		"proxy_send_timeout 60s;",
		"proxy_set_header X-BFL-USER " + constants.Username + ";",
		"proxy_set_header X-Authorization $http_x_authorization;",
		"proxy_set_header X-Terminus-Node " + node + ";",
		"add_header Access-Control-Allow-Headers \"access-control-allow-headers,access-control-allow-methods,access-control-allow-origin,content-type,x-auth,x-unauth-error,x-authorization\";",
		"add_header Access-Control-Allow-Methods \"PUT, GET, DELETE, POST, OPTIONS\";",
	}
}

func filesNodeApiPatch(ctx context.Context, r *NginxController, s *config.Server) (*config.Server, error) {
	pods, err := r.getFilesPods(ctx)
	if err != nil {
		klog.Errorf("failed to list pods, %v", err)
		return nil, err
	}

	var nodes corev1.NodeList
	err = r.List(ctx, &nodes)
	if err != nil {
		klog.Errorf("failed to list nodes, %v", err)
		return nil, err
	}

	masterNode, podMap := r.getFileserverPodMap(nodes, pods)

	authRequest := config.Location{
		Prefix:      "= /authelia-verify",
		ProxyPass:   AUTHELIA_URL,
		DirectProxy: true,
		Additionals: []string{
			"internal;",
			"proxy_pass_request_body off;",
			"proxy_set_header Content-Length \"\";",
			"proxy_set_header X-Original-URL $scheme://$http_host$request_uri;",
			"proxy_set_header X-Original-Method $request_method;",
			"proxy_set_header X-Forwarded-For $remote_addr;",
			"proxy_set_header X-Forwarded-Proto $scheme;",
			"proxy_set_header X-BFL-USER " + constants.Username + ";",
			"proxy_set_header X-Authorization $http_x_authorization;",
			"proxy_set_header Cookie $http_cookie;",
		},
	}

	var apis []config.Location

	var nodeLocationPrefix = []string{
		"/api/resources/cache/",
		"/api/preview/cache/",
		"/api/raw/cache/",
		"/api/tree/cache/",
		"/api/resources/external/",
		"/api/preview/external/",
		"/api/raw/external/",
		"/api/tree/external/",
		"/api/mount/",
		"/api/unmount/external/",
		"/api/smb_history/",
		"/upload/upload-link/",
		"/upload/file-uploaded-bytes/",
		"/api/paste/",
		"/videos/",
		"/api/md5/cache/",
		"/api/md5/external/",
		"/api/permission/cache/",
		"/api/permission/external/",
		"/api/task/",
		"~ ^/api/resources/share/(.*)_",
		"~ ^/api/preview/share/(.*)_",
		"~ ^/api/tree/share/(.*)_",
		"~ ^/api/raw/share/(.*)_",
	}

	var masterLocation = []string{
		"/api/resources/cache/",
		"/api/preview/cache/",
		"/api/resources/external/",
		"/api/preview/external/",
		"/api/paste/",
		"/api/task/",
		"/api/repos/",
		"/seafhttp/",
		"/api/resources/share/",
		"/api/preview/share/",
		"/api/tree/share/",
		"/api/raw/share/",
	}

	for node := range podMap {
		proxyCfg := ProxyServiceConfig{node}
		for _, prefix := range nodeLocationPrefix {
			nodeApi := config.Location{
				Prefix:      fmt.Sprintf("%s%s/", prefix, node),
				Additionals: locationAdditionalsForFilesOp(node),

				ProxyPass:   proxyCfg.ServiceHost(),
				DirectProxy: true,
			}

			apis = append(apis, nodeApi)
		}

	} // end for each node

	s.Locations = append(s.Locations, authRequest)

	if _, ok := podMap[masterNode]; ok {
		var masterApis []config.Location
		proxyCfg := ProxyServiceConfig{masterNode}
		for _, l := range masterLocation {
			masterApi := config.Location{
				Prefix:      l,
				Additionals: locationAdditionalsForFilesOp(masterNode),

				ProxyPass:   proxyCfg.ServiceHost(),
				DirectProxy: true,
			}

			masterApis = append(masterApis, masterApi)
		}

		s.Locations = append(s.Locations, masterApis...)
	}

	s.Locations = append(s.Locations, apis...)

	return s, nil
}

func (r *NginxController) getFileserverPodMap(nodes corev1.NodeList, pods corev1.PodList) (masterNode string, podMap map[string]*corev1.Pod) {
	podMap = make(map[string]*corev1.Pod)
	for _, node := range nodes.Items {
		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			masterNode = node.Name
		}

		for _, pod := range pods.Items {
			if pod.Spec.NodeName == node.Name && isFileServerPod(&pod) {
				podMap[node.Name] = &pod
			}
		}
	}

	return
}

func (r *NginxController) getFilesPods(ctx context.Context) (corev1.PodList, error) {
	var pods corev1.PodList
	err := r.Client.List(ctx, &pods, client.MatchingLabels{"app": "files"})
	return pods, err
}
