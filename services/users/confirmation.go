package users

import (
	"fmt"
	"github.com/donohutcheon/gowebserver/datalayer"
	"github.com/donohutcheon/gowebserver/lib/nonce"
	"github.com/donohutcheon/gowebserver/state"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func ConfirmUsersForever(state *state.ServerState) {
	defer state.ShutdownWG.Done()
	logger := state.Logger
	dl := state.DataLayer

	email := state.Providers.Email

	for u := range state.Channels.ConfirmUsers {
		logger.Printf("Received user to confirm from channel %s %s %s", u.Email.String, u.Password.String, u.State.String)
		confirmationNonce := nonce.GenerateNonce(32)

		err := dl.SetUserStateByID(u.ID, datalayer.UserStateProcessing)
		if err != nil {
			logger.Printf("failed to update user's state to %s %+v", u.Email.String, datalayer.UserStatePending)
			continue
		}

		nonceID, err := dl.CreateSignUpConfirmation(confirmationNonce, u.ID)
		if err != nil {
			logger.Printf("failed to create sign-up confirmation for user %s %+v", u.Email.String, confirmationNonce)
			continue
		}

		to := u.Email.String
		toList := []string{to}
		from := "noreply@someapp.com"
		message := fmt.Sprintf("Hello %s,\n Welcome to this app - whatever it is.  Please confirm your registration by clicking on this link " +
		"%s/api/users/confirm/%s", u.Email.String, state.URL, confirmationNonce)

		err = email.SendMail(toList, from, "Welcome to this app!", message)
		if err != nil {
			logger.Print("failed to send mail")
			continue
		}

		err = dl.SetUserStateByID(u.ID, datalayer.UserStatePending)
		if err != nil {
			logger.Printf("failed to update user's state to %s %+v", u.Email.String, datalayer.UserStatePending)
			continue
		}

		logger.Printf("Sent confirmation email for user %s with confirmationNonce %s nonceID %d", u.Email.String, confirmationNonce, nonceID)
	}
	logger.Print("ConfirmUsersForever done")
}
