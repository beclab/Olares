package handler

import (
	"encoding/json"
	"errors"
	"integration/pkg/hertz/biz/model/api/infisical"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func FormatErrorMessage(m string) error {
	if strings.Contains(m, "{") && strings.Contains(m, "}") {
		pos := strings.Index(m, "{")
		errMsg := m[pos:]
		var errObj *infisical.ErrorMessage
		if err := json.Unmarshal([]byte(errMsg), &errObj); err != nil {
			return err
		}

		return errors.New(errObj.Message)
	}

	return errors.New(m)
}

func ErrorResponse(c *app.RequestContext, msg error) {
	c.AbortWithStatusJSON(consts.StatusInternalServerError,
		utils.H{
			"code":    consts.StatusInternalServerError,
			"message": msg.Error(),
		})
}

func SuccessResponse(c *app.RequestContext, result interface{}) {
	c.JSON(consts.StatusOK, result)
}
