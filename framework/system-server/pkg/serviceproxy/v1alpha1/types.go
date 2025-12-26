package serviceproxy

import (
	"strconv"

	sysv1alpha1 "bytetrade.io/web3os/system-server/pkg/apis/sys/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
)

const (
	LEAGCY_PATCH    = "/legacy/v1alpha1"
	ParamSubPath    = "subpath"
	LEAGCY_PATCH_V2 = "/system-server/v2"
)

type ProxyRequest struct {
	Op       string      `json:"op"`
	DataType string      `json:"datatype"`
	Version  string      `json:"version"`
	Group    string      `json:"group"`
	AppKey   string      `json:"appkey"`
	Param    interface{} `json:"param,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Token    string
}

type GetOpParam struct {
	DataID string `json:"dataid"`
}

type ListOpParam struct {
	Filters map[string][]string `json:"filters,omitempty"`
	Page    Pagination          `json:"page,omitempty"`
}

type Pagination struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit"`
}

type UpdateOpParam struct {
	DataID string `json:"dataid"`
}

type DispatchRequest struct {
	ProxyRequest
	Result interface{} `json:"result"`
}

// NewProxyRequestFromOpRequest constructs a new ProxyRequest.
func NewProxyRequestFromOpRequest(appkey, op string, req *restful.Request) (*ProxyRequest, error) {
	datatype := req.PathParameter(api.ParamDataType)
	version := req.PathParameter(api.ParamVersion)
	group := req.PathParameter(api.ParamGroup)
	token := req.Request.Header.Get(api.AccessTokenHeader)

	var param interface{}
	var data interface{}
	switch op {
	case sysv1alpha1.Get:
		id := req.PathParameter(api.ParamDataID)
		param = GetOpParam{
			DataID: id,
		}

	case sysv1alpha1.List:
		q := req.Request.URL.Query()
		filters := make(map[string][]string)

		for k, v := range q {
			if k != "offset" && k != "limit" {
				filters[k] = v
			}
		}

		listParam := ListOpParam{
			Filters: filters,
		}

		l, ok := q["limit"]
		if ok && len(l) > 0 {
			limit, err := strconv.Atoi(l[0])
			if err != nil {
				klog.Error("query param limit is invalid numeric string, ", l[0])
			} else {
				listParam.Page.Limit = limit
				if o, ok := q["offset"]; ok {
					if offset, err := strconv.Atoi(o[0]); err != nil {
						listParam.Page.Offset = offset
					}
				}
			}

		} // end if page param

		param = listParam

	default:
		var body map[string]interface{}
		err := req.ReadEntity(&body)
		if err != nil {
			return nil, err
		}

		data = body
		if op == sysv1alpha1.Update || op == sysv1alpha1.Delete {
			id := req.PathParameter(api.ParamDataID)
			param = UpdateOpParam{
				DataID: id,
			}
		}
	} // end switch

	return &ProxyRequest{
		Op:       op,
		DataType: datatype,
		Version:  version,
		Group:    group,
		AppKey:   appkey,
		Param:    param,
		Data:     data,
		Token:    token,
	}, nil
}

// NewDispatchRequest constructs a new DispatchRequest object.
func NewDispatchRequest(pr *ProxyRequest, result any) *DispatchRequest {
	return &DispatchRequest{
		*pr,
		result,
	}
}
