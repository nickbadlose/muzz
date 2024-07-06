package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
)

// SwipeRequest holds the required information to perform a swipe on a user.
type SwipeRequest struct {
	UserID     int  `json:"userID"`
	Preference bool `json:"preference"`
}

type Match struct {
	Matched bool `json:"matched"`
	MatchID int  `json:"matchID,omitempty"`
}

// SwipeResponse object to return to the client.
type SwipeResponse struct {
	Result *Match `json:"result"`
}

// Render implements the render.Render interface.
func (*SwipeResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	return nil
}

// Swipe sets a users swipe preference against another user in the application.
func (h *Handlers) Swipe(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authorizer.UserFromContext(r.Context())
	if err != nil {
		logger.Error(r.Context(), "getting authenticated user from context", err)
		err = render.Render(w, r, apperror.BadRequestHTTP(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	req := new(SwipeRequest)
	err = json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		logger.Error(r.Context(), "decoding swipe request", err)
		err = render.Render(w, r, apperror.BadRequestHTTP(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	res, aErr := h.matchService.Swipe(r.Context(), &muzz.CreateSwipeInput{
		UserID:       userID,
		SwipedUserID: req.UserID,
		Preference:   req.Preference,
	})
	if aErr != nil {
		logger.MaybeError(
			r.Context(),
			renderingErrorMessage,
			render.Render(w, r, aErr.ToHTTP()),
		)
		return
	}

	err = render.Render(w, r, &SwipeResponse{Result: &Match{Matched: res.Matched, MatchID: res.ID}})
	if err != nil {
		logger.Error(r.Context(), "rendering swipe response", err)
	}
}
