package appstate

import (
	appsv1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"

	"k8s.io/klog/v2"
)

// GetRefsForImageManager resolves the image references that need to be
// downloaded for the given application config by performing a Helm dry-run.
func GetRefsForImageManager(appConfig *appcfg.ApplicationConfig, values map[string]interface{}) (refs []appsv1.Ref, err error) {
	switch {
	case appConfig.APIVersion == appcfg.V2 && appConfig.IsMultiCharts():
		var chartRefs []appsv1.Ref
		for _, chart := range appConfig.SubCharts {
			chartRefs, err = utils.GetRefFromResourceList(chart.ChartPath(appConfig.RawAppName), values, appConfig.Images)
			if err != nil {
				klog.Errorf("get refs from chart %s failed %v", chart.Name, err)
				return
			}

			refs = append(refs, chartRefs...)
		}
	default:
		refs, err = utils.GetRefFromResourceList(appConfig.ChartsName, values, appConfig.Images)
	}
	return
}
