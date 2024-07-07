package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
)

// CreateUserRequest holds the information required to create a user.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}

// User represents the full user details to present to the client.
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int    `json:"age"`
}

// UserResponse object to send to the client.
type UserResponse struct {
	Result *User `json:"result"`
}

// Render implements the render.Render interface.
func (*UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusCreated)
	return nil
}

// CreateUser creates a user in the application.
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	req := new(CreateUserRequest)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		logger.Error(r.Context(), "decoding create user request", err)
		err = render.Render(w, r, apperror.BadRequestHTTP(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	user, aErr := h.userService.Create(r.Context(), &muzz.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Gender:   req.Gender,
		Age:      req.Age,
	})
	if aErr != nil {
		logger.MaybeError(
			r.Context(),
			renderingErrorMessage,
			render.Render(w, r, aErr.ToHTTP()),
		)
		return
	}

	err = render.Render(w, r, &UserResponse{
		Result: &User{
			ID:       user.ID,
			Email:    user.Email,
			Password: user.Password,
			Name:     user.Name,
			Gender:   user.Gender.String(),
			Age:      user.Age,
		},
	})
	if err != nil {
		logger.Error(r.Context(), "rendering create user response", err)
	}
}

// UserDetails represents the public user details to present to the client.
type UserDetails struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Age    int    `json:"age"`
}

// DiscoverResponse object to send to the client.
type DiscoverResponse struct {
	Results []*UserDetails `json:"results"`
}

// Render implements the render.Render interface.
func (*DiscoverResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	return nil
}

// Discover all un-swiped users from the application.
func (h *Handlers) Discover(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authorizer.UserFromContext(r.Context())
	if err != nil {
		logger.Error(r.Context(), "getting authenticated user from context", err)
		err = render.Render(w, r, apperror.BadRequestHTTP(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	filters, err := muzz.UserFiltersFromParams(
		r.URL.Query().Get("maxAge"),
		r.URL.Query().Get("minAge"),
		r.URL.Query().Get("genders"),
	)
	if err != nil {
		logger.Error(r.Context(), "getting filters from query params", err)
		err = render.Render(w, r, apperror.BadRequestHTTP(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	appUsers, aErr := h.userService.Discover(r.Context(), &muzz.GetUsersInput{
		UserID:  userID,
		Filters: filters,
	})
	if aErr != nil {
		logger.MaybeError(
			r.Context(),
			renderingErrorMessage,
			render.Render(w, r, aErr.ToHTTP()),
		)
		return
	}

	users := make([]*UserDetails, 0, len(appUsers))
	for _, user := range appUsers {
		users = append(users, &UserDetails{
			ID:     user.ID,
			Name:   user.Name,
			Gender: user.Gender.String(),
			Age:    user.Age,
		})
	}

	err = render.Render(w, r, &DiscoverResponse{Results: users})
	if err != nil {
		logger.Error(r.Context(), "rendering discover response", err)
	}
}
