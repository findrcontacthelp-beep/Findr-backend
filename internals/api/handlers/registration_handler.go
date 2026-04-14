package handlers

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/model"
	authmodel "github.com/findr-app/findr-backend/internals/model"
	"github.com/findr-app/findr-backend/internals/repository"
)

// Allowed values for registration validation.
var (
	allowedColleges = map[string]bool{
		"Stanford University":                   true,
		"Massachusetts Institute of Technology": true,
		"UC Berkeley":                           true,
		"Georgia Tech":                          true,
		"Harvard University":                    true,
		"Carnegie Mellon University":            true,
	}
	allowedBranches = map[string]bool{
		"Computer Science":        true,
		"Mechanical Engineering":  true,
		"Electrical Engineering":  true,
		"Fine Arts":               true,
		"Business Administration": true,
		"Data Science":            true,
	}
	allowedGraduationYears = map[string]bool{
		"2024": true, "2025": true, "2026": true, "2027": true, "2028": true,
	}
	allowedInterests = map[string]bool{
		"Technology": true, "Design": true, "Business": true, "Marketing": true,
		"Sports": true, "Music": true, "Art": true, "Gaming": true,
		"Science": true, "Literature": true,
	}
)

// GetRegistrationOptions returns all allowed colleges, branches, graduation years,
func GetRegistrationOptions() gin.HandlerFunc {
	colleges := sortedKeysFromMap(allowedColleges)
	branches := sortedKeysFromMap(allowedBranches)
	graduationYears := sortedKeysFromMap(allowedGraduationYears)
	interests := sortedKeysFromMap(allowedInterests)

	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"colleges":         colleges,
			"branches":         branches,
			"graduation_years": graduationYears,
			"interests":        interests,
		})
	}
}

// them in sorted order. Sorting ensures deterministic API responses
func sortedKeysFromMap(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// generateOTP returns a cryptographically secure 6-digit OTP string.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generate OTP: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// Register handles user registration (Step 1 of 2).
//
// Flow:
//  1. Validate the request body against allowed values (college, branch, year, interests).
//  2. Generate a secure 6-digit OTP.
//  3. Store the registration data + OTP in Redis with a 2-minute TTL.
//  4. Return the OTP to the client (in production, this should be sent via email).
//
// The user must call VerifyRegistration with the correct OTP to complete registration.
func Register(pool *pgxpool.Pool, redisClient *redis.Client, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Parse and validate request body
		var req authmodel.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error: "invalid request body",
				Code:  "INVALID_REQUEST",
			})
			return
		}

		// Step 2: Validate each field against our whitelisted values.
		if !allowedColleges[req.CollegeName] {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error: "invalid college_name",
				Code:  "INVALID_COLLEGE",
			})
			return
		}
		if !allowedBranches[req.Branch] {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error: "invalid branch",
				Code:  "INVALID_BRANCH",
			})
			return
		}
		if !allowedGraduationYears[req.GraduationYear] {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error: "invalid graduation_year",
				Code:  "INVALID_GRADUATION_YEAR",
			})
			return
		}
		for _, interest := range req.Interests {
			if !allowedInterests[interest] {
				c.JSON(http.StatusBadRequest, model.ErrorResponse{
					Error: "invalid interest: " + interest,
					Code:  "INVALID_INTEREST",
				})
				return
			}
		}

		// Step 3: Generate a cryptographically secure OTP
		otp, err := generateOTP()
		if err != nil {
			log.Error("failed to generate OTP", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error: "internal server error",
				Code:  "OTP_GENERATION_FAILED",
			})
			return
		}

		// Step 4: Bundle registration data with OTP and store in Redis.
		// If the user doesn't verify within 2 minutes, the data is auto-deleted.
		redisData := authmodel.RedisRegistrationData{
			OTP:     otp,
			Request: req,
		}

		dataBytes, err := json.Marshal(redisData)
		if err != nil {
			log.Error("marshal redis data failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error: "internal server error",
				Code:  "MARSHAL_FAILED",
			})
			return
		}

		ctx := c.Request.Context()
		err = redisClient.Set(ctx, "registration:"+req.Email, dataBytes, 2*time.Minute).Err()
		if err != nil {
			log.Error("set redis data failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error: "internal server error",
				Code:  "REDIS_SET_FAILED",
			})
			return
		}

		// Step 5: Return OTP to client.
		// TODO: In production, send OTP via email instead of returning in response.
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Registration data temporarily saved. Please verify with OTP.",
			"email":   req.Email,
			"otp":     otp,
		})
	}
}

// VerifyRegistration completes user registration (Step 2 of 2).
//
// Flow:
//  1. Fetch the pending registration data from Redis using the email.
//  2. Compare the provided OTP with the stored OTP.
//  3. If OTP matches, persist the user to PostgreSQL via the user repository.
//  4. Clean up the temporary Redis key after successful registration.
//
// Error cases:
//   - Registration expired (Redis TTL exceeded) -> 404
//   - Invalid OTP -> 401
//   - Duplicate user/email -> 409
func VerifyRegistration(pool *pgxpool.Pool, redisClient *redis.Client, log *zap.Logger) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(pool)

	return func(c *gin.Context) {
		// Step 1: Parse the verification request (email + OTP)
		var vr authmodel.VerifyRequest
		if err := c.ShouldBindJSON(&vr); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error: "invalid request body",
				Code:  "INVALID_REQUEST",
			})
			return
		}

		// Step 2: Retrieve pending registration from Redis.
		ctx := c.Request.Context()
		val, err := redisClient.Get(ctx, "registration:"+vr.Email).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				c.JSON(http.StatusNotFound, model.ErrorResponse{
					Error: "registration expired or not found",
					Code:  "REGISTRATION_NOT_FOUND",
				})
				return
			}
			log.Error("failed to get from redis", zap.String("email", vr.Email), zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error: "internal server error",
				Code:  "REDIS_GET_FAILED",
			})
			return
		}

		// Step 3: Deserialize the stored registration data
		var redisData authmodel.RedisRegistrationData
		if err := json.Unmarshal([]byte(val), &redisData); err != nil {
			log.Error("failed to unmarshal redis registration data", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error: "internal server error",
				Code:  "UNMARSHAL_FAILED",
			})
			return
		}

		// Step 4: Verify OTP matches
		if vr.OTP != redisData.OTP {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: "invalid OTP",
				Code:  "INVALID_OTP",
			})
			return
		}

		// Step 5: OTP is valid — persist user to the database.
		// Handle known conflict errors (duplicate user/email) separately.
		_, err = userRepo.CreateUser(ctx, redisData.Request)
		if err != nil {
			if errors.Is(err, repository.ErrUserAlreadyExists) {
				c.JSON(http.StatusConflict, model.ErrorResponse{
					Error: "user already exists",
					Code:  "USER_EXISTS",
				})
				return
			}
			if errors.Is(err, repository.ErrEmailAlreadyExists) {
				c.JSON(http.StatusConflict, model.ErrorResponse{
					Error: "email already exists",
					Code:  "EMAIL_EXISTS",
				})
				return
			}
			log.Error("register user failed", zap.String("email", vr.Email), zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{
				Error: "failed to register user",
				Code:  "REGISTRATION_FAILED",
			})
			return
		}

		// Step 6: Clean up the temporary registration key from Redis.
		if err := redisClient.Del(ctx, "registration:"+vr.Email).Err(); err != nil {
			log.Warn("failed to delete registration key from redis", zap.String("email", vr.Email), zap.Error(err))
		}

		c.JSON(http.StatusCreated, model.SuccessResponse{Message: "Registration is successful"})
	}
}
