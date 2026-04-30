package controller

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

var errRegistrationIPLimited = errors.New("registration ip limit exceeded")

type registrationIPCheckError struct {
	err error
}

func (e *registrationIPCheckError) Error() string {
	return fmt.Sprintf("check registration ip limit: %v", e.err)
}

func (e *registrationIPCheckError) Unwrap() error {
	return e.err
}

func setRegistrationIP(c *gin.Context, user *model.User) error {
	ip := strings.TrimSpace(c.ClientIP())
	if ip == "" {
		return nil
	}

	since := time.Now().Add(-model.RegisterIPLimitWindow).Unix()
	limited, err := model.IsRegistrationIPLimited(ip, since)
	if err != nil {
		return &registrationIPCheckError{err: err}
	}
	if limited {
		return errRegistrationIPLimited
	}

	user.RegistrationIP = ip
	return nil
}

func respondRegistrationIPError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, errRegistrationIPLimited) {
		common.ApiErrorI18n(c, i18n.MsgUserRegisterIPLimited)
		return true
	}

	var checkErr *registrationIPCheckError
	if errors.As(err, &checkErr) {
		common.SysLog(checkErr.Error())
		common.ApiErrorI18n(c, i18n.MsgDatabaseError)
		return true
	}

	return false
}
