package handlers

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
	"github.com/nickbadlose/muzz/internal/logger"
	"github.com/paulmach/orb"
)

// CreateUserRequest holds the information required to create a user record.
type CreateUserRequest struct {
	// Email of the user.
	Email string `json:"email"`
	// Password of the user.
	Password string `json:"password"`
	// Name of the user.
	Name string `json:"name"`
	// Gender of the user.
	Gender string `json:"gender"`
	// Age of the user.
	Age int `json:"age"`
	// Location of the user at current.
	Location Location `json:"location"`
}

// Location represents the geographic coordinates as longitude and latitude.
type Location struct {
	// Lat is the latitude of the location.
	Lat float64 `json:"latitude"`
	// Lon is the longitude of the location.
	Lon float64 `json:"longitude"`
}

// User represents the full user details to present to the client.
type User struct {
	// ID is the unique identifier of the user record.
	ID int `json:"id"`
	// Email of the user record.
	Email string `json:"email"`
	// Password of the user record.
	Password string `json:"password"`
	// Name of the user record.
	Name string `json:"name"`
	// Gender of the user record.
	Gender string `json:"gender"`
	// Age of the user record.
	Age int `json:"age"`
	// Location of the user record on last login.
	Location Location `json:"location"`
}

// UserResponse object to send to the client.
type UserResponse struct {
	// Result is the user record.
	Result *User `json:"result"`
}

// Render implements the render.Render interface.
func (*UserResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusCreated)
	return nil
}

// CreateUser creates a user record in the application.
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	req, err := decodeRequest[CreateUserRequest](w, r)
	if err != nil {
		logger.Error(r.Context(), "decoding create user request", err)
		return
	}

	user, aErr := h.userService.Create(r.Context(), &muzz.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Gender:   req.Gender,
		Age:      req.Age,
		Location: orb.Point{req.Location.Lon, req.Location.Lat},
	})
	if aErr != nil {
		logger.MaybeError(
			r.Context(),
			renderingErrorMessage,
			render.Render(w, r, aErr.ToHTTP()),
		)
		return
	}

	err = encodeResponse(w, r, h.config.DebugEnabled(), &UserResponse{
		Result: &User{
			ID:       user.ID,
			Email:    user.Email,
			Password: user.Password,
			Name:     user.Name,
			Gender:   user.Gender.String(),
			Age:      user.Age,
			Location: Location{
				Lat: user.Location.Lat(),
				Lon: user.Location.Lon(),
			},
		},
	})
	if err != nil {
		logger.Error(r.Context(), "rendering create user response", err)
	}
}

// UserDetails represents the public user details to present to the client.
type UserDetails struct {
	// ID is the unique identifier of the user record.
	ID int `json:"id"`
	// Email of the user record.
	// Name of the user record.
	Name string `json:"name"`
	// Gender of the user record.
	Gender string `json:"gender"`
	// Age of the user record.
	Age int `json:"age"`
	// DistanceFromMe is the user records distance from the authenticated user.
	DistanceFromMe float64 `json:"distanceFromMe"`
}

// DiscoverResponse object to send to the client.
type DiscoverResponse struct {
	// Results are the user records.
	Results []*UserDetails `json:"results"`
}

// Render implements the render.Render interface.
func (*DiscoverResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	return nil
}

// Discover attempts to retrieve potential matches from the application data. Previously swiped users are excluded from
// results and the results are filterable and sortable, based on the provided query parameters:
//   - sort: muzz.SortType
//   - maxAge: int
//   - minAge: int
//   - genders: []muzz.Gender
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

	sortType := muzz.SortValues[r.URL.Query().Get("sort")]

	location, err := h.location.ByIP(r.Context(), r.RemoteAddr)
	if err != nil {
		err = render.Render(w, r, apperror.InternalServerHTTP(err))
		logger.MaybeError(r.Context(), renderingErrorMessage, err)
		return
	}

	appUsers, aErr := h.userService.Discover(r.Context(), &muzz.GetUsersInput{
		UserID:   userID,
		Location: location,
		SortType: sortType,
		Filters:  filters,
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
			ID:             user.ID,
			Name:           user.Name,
			Gender:         user.Gender.String(),
			Age:            user.Age,
			DistanceFromMe: user.DistanceFromMe,
		})
	}

	err = encodeResponse(w, r, h.config.DebugEnabled(), &DiscoverResponse{Results: users})
	if err != nil {
		logger.Error(r.Context(), "rendering discover response", err)
	}
}
