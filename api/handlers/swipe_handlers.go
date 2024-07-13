package handlers

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
)

// SwipeRequest holds the required information to perform a swipe on a user.
type SwipeRequest struct {
	// UserID is the id of the user record that the swipe action is being performed against.
	UserID int `json:"userID"`
	// Preference is whether the authenticated user would prefer to match with the swiped user.
	Preference bool `json:"preference"`
}

// Match represents the match details to present to the client.
type Match struct {
	// Matched, true if the swipe resulted in a match.
	Matched bool `json:"matched"`
	// MatchID is the unique identifier of the match record. Empty if Matched is false.
	MatchID int `json:"matchID,omitempty"`
}

// SwipeResponse object to return to the client.
type SwipeResponse struct {
	// Result is the match record.
	Result *Match `json:"result"`
}

// Render implements the render.Render interface.
func (*SwipeResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusCreated)
	return nil
}

// Swipe sets a users swipe preference against another user in the application.
//
// If both users have swiped with preference as true to each other, match details are returned.
func (h *Handlers) Swipe(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authorizer.UserFromContext(r.Context())
	if err != nil {
		logger.Error(r.Context(), "getting authenticated user from context", err)
		logger.MaybeError(r.Context(), renderingErrorMessage, render.Render(w, r, apperror.BadRequestHTTP(err)))
		return
	}

	req, err := decodeRequest[SwipeRequest](w, r)
	if err != nil {
		logger.Error(r.Context(), "decoding swipe request", err)
		return
	}

	res, aErr := h.matchService.Swipe(r.Context(), &muzz.CreateSwipeInput{
		UserID:       userID,
		SwipedUserID: req.UserID,
		Preference:   req.Preference,
	})
	if aErr != nil {
		logger.MaybeError(r.Context(), renderingErrorMessage, render.Render(w, r, aErr.ToHTTP()))
		return
	}

	err = encodeResponse(
		w,
		r,
		h.config.DebugEnabled(),
		&SwipeResponse{Result: &Match{Matched: res.Matched, MatchID: res.ID}},
	)
	if err != nil {
		logger.Error(r.Context(), "rendering swipe response", err)
	}
}
