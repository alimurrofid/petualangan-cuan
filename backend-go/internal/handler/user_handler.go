package handler

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"cuan-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserHandler interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	UpdateProfile(c *fiber.Ctx) error
	ChangePassword(c *fiber.Ctx) error
	GoogleLogin(c *fiber.Ctx) error
	GoogleCallback(c *fiber.Ctx) error
	GetProfile(c *fiber.Ctx) error
	RefreshToken(c *fiber.Ctx) error
}

type userHandler struct {
	userService service.UserService
	frontendURL string
}

func NewUserHandler(userService service.UserService, frontendURL string) UserHandler {
	return &userHandler{
		userService: userService,
		frontendURL: frontendURL,
	}
}

func (h *userHandler) setRefreshCookie(c *fiber.Ctx, refreshToken string) {
	cookie := new(fiber.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = refreshToken
	cookie.Expires = time.Now().Add(72 * time.Hour) 
	cookie.HTTPOnly = true
	cookie.Secure = false 
	cookie.SameSite = "Lax" 
	
	c.Cookie(cookie)
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterInput true "Register Request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/auth/register [post]
func (h *userHandler) Register(c *fiber.Ctx) error {
	var input service.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Warn().Str("request_id", reqID).Err(err).Msg("Invalid request body payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user, accessToken, refreshToken, err := h.userService.Register(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	h.setRefreshCookie(c, refreshToken)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token": accessToken,
		"user":  user,
	})
}

// Login godoc
// @Summary Login user
// @Description Login with email and password to get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginInput true "Login Request"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/auth/login [post]
func (h *userHandler) Login(c *fiber.Ctx) error {
	var input service.LoginInput
	if err := c.BodyParser(&input); err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Warn().Str("request_id", reqID).Err(err).Msg("Invalid request body payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user, accessToken, refreshToken, err := h.userService.Login(input)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	h.setRefreshCookie(c, refreshToken)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": accessToken,
		"user":  user,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Logout current user (Invalidate token client-side)
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/auth/logout [post]
func (h *userHandler) Logout(c *fiber.Ctx) error {
	c.ClearCookie("refresh_token")

	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		_ = h.userService.Logout(refreshToken)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

// RefreshToken godoc
// @Summary Refresh Access Token
// @Description Get a new access and refresh token pair using valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Refresh Token Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/auth/refresh [post]
func (h *userHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No refresh token provided"})
	}

	accessToken, newRefreshToken, err := h.userService.RefreshToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	h.setRefreshCookie(c, newRefreshToken)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": accessToken,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update user name and email
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body service.UpdateProfileInput true "Update Profile Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/user/profile [put]
func (h *userHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var input service.UpdateProfileInput
	if err := c.BodyParser(&input); err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Warn().Str("request_id", reqID).Err(err).Msg("Invalid request body payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user, err := h.userService.UpdateProfile(uint(userID), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":    user,
		"message": "Profile updated successfully",
	})
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change user password
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body service.ChangePasswordInput true "Change Password Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/user/password [put]
func (h *userHandler) ChangePassword(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var input service.ChangePasswordInput
	if err := c.BodyParser(&input); err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Warn().Str("request_id", reqID).Err(err).Msg("Invalid request body payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := h.userService.ChangePassword(uint(userID), input)
	if err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password updated successfully",
	})
}



func (h *userHandler) GoogleLogin(c *fiber.Ctx) error {
	conf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	url := conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return c.Redirect(url)
}

func (h *userHandler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Code not found"})
	}

	conf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to exchange token"})
	}

	client := conf.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Error().Str("request_id", reqID).Err(err).Msg("Failed to get user info from Google")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user info"})
	}
	defer resp.Body.Close()

	// Make sure the body is fully read as we've fixed the syntax issue here
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Error().Str("request_id", reqID).Err(err).Msg("Failed to read response body from Google")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read user info"})
	}

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		ID    string `json:"id"`
	}

	if err := json.Unmarshal(body, &userInfo); err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Error().Str("request_id", reqID).Err(err).Msg("Failed to parse user info from Google")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse user info"})
	}

	user, jwtToken, refreshToken, err := h.userService.LoginOrRegisterGoogle(userInfo.Email, userInfo.Name, userInfo.ID)
	if err != nil {
		reqID, _ := c.Locals("requestid").(string)
		log.Error().Str("request_id", reqID).Err(err).Msg("Internal server error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	
	_ = user

	return c.Redirect(h.frontendURL + "/auth/google/callback?token=" + jwtToken + "&refresh_token=" + refreshToken)
}

func (h *userHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": user,
	})
}
