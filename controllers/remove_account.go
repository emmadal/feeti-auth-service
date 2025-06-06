package controllers

import (
	"encoding/json"
	"github.com/emmadal/feeti-auth/helpers"
	"github.com/emmadal/feeti-auth/models"
	jwt "github.com/emmadal/feeti-module/auth"
	status "github.com/emmadal/feeti-module/status"
	"github.com/emmadal/feeti-module/subject"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// RemoveAccount remove user account
func RemoveAccount(c *gin.Context) {
	// Increment counter for HTTP requests total to prometheus
	helpers.HttpRequestsTotal.WithLabelValues(c.Request.URL.Path, c.Request.Method).Inc()

	body := models.RemoveUserAccount{}
	var response helpers.ResponsePayload

	// Validate request body
	if err := c.ShouldBindJSON(&body); err != nil {
		status.HandleError(c, http.StatusBadRequest, "Bad request", err)
		return
	}

	// search if a user exists in DB
	user, err := models.GetUserByPhoneNumber(body.PhoneNumber)
	if err != nil {
		status.HandleError(c, http.StatusNotFound, "Invalid phone number or user PIN", err)
		return
	}

	// verify user identity with context data
	if user.ID != jwt.GetUserIDFromGin(c) {
		status.HandleError(c, http.StatusForbidden, "Unauthorized user", err)
		return
	}

	// verify user password
	if !helpers.VerifyPassword(body.Pin, user.Pin) {
		status.HandleError(c, http.StatusUnauthorized, "invalid password or phone number", err)
		return
	}

	// publish a request to get wallet data
	request := helpers.RequestPayload{
		Subject: subject.SubjectWalletLock,
		Data:    user.ID.String(),
	}
	resp, err := request.PublishEvent()
	if err != nil {
		status.HandleError(c, http.StatusInternalServerError, "Unable to process wallet", err)
		return
	}

	// Unmarshal the wallet data
	_ = json.Unmarshal(resp.Data, &response)
	if !response.Success {
		status.HandleError(c, http.StatusUnprocessableEntity, response.Error, nil)
		return
	}

	// remove a user account
	if err := user.DeactivateUserAccount(); err != nil {
		status.HandleError(c, http.StatusInternalServerError, "Failed to remove account", err)
		return
	}

	// Send success response and delete cookie
	jwt.ClearAuthCookie(c, "")

	// record auth log
	go func() {
		authLog := models.AuthLog{
			UserID:      user.ID,
			PhoneNumber: user.PhoneNumber,
			DeviceToken: user.DeviceToken,
			Activity:    "remove_account",
			Metadata:    `{"source": "remove_account"}`,
		}
		if err := authLog.CreateAuthLog(); err != nil {
			log.Printf("Error creating auth log: %v\n", err)
		}
	}()

	// Send success response
	status.HandleSuccess(c, "Account removed successfully")
}
