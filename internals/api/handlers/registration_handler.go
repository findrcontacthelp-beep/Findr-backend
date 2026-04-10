package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
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

func Register(pool *pgxpool.Pool, redisClient *redis.Client, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req authmodel.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		// Validate allowed values
		if !allowedColleges[req.CollegeName] {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid college_name"})
			return
		}
		if !allowedBranches[req.Branch] {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid branch"})
			return
		}
		if !allowedGraduationYears[req.GraduationYear] {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid graduation_year"})
			return
		}
		for _, interest := range req.Interests {
			if !allowedInterests[interest] {
				c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid interest: " + interest})
				return
			}
		}

		otp := fmt.Sprintf("%06d", rand.Intn(1000000))
		redisData := authmodel.RedisRegistrationData{
			OTP:     otp,
			Request: req,
		}

		dataBytes, err := json.Marshal(redisData)
		if err != nil {
			log.Error("marshal redis data failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
			return
		}

		ctx := context.Background()
		err = redisClient.Set(ctx, "registration:"+req.Email, dataBytes, 2*time.Minute).Err()
		if err != nil {
			log.Error("set redis data failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
			return
		}

		// Typically, we would send this OTP via email here.
		// For now, we will return it in the response.
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Registration data temporarily saved. Please verify with OTP.",
			"email":   req.Email,
			"otp":     otp,
		})
	}
}

func VerifyRegistration(pool *pgxpool.Pool, redisClient *redis.Client, log *zap.Logger) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(pool)

	return func(c *gin.Context) {
		var vr authmodel.VerifyRequest
		if err := c.ShouldBindJSON(&vr); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		ctx := context.Background()
		val, err := redisClient.Get(ctx, "registration:"+vr.Email).Result()
		if err != nil {
			if err == redis.Nil {
				c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "registration expired or invalid"})
				return
			}
			log.Error("failed to get from redis", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
			return
		}

		var redisData authmodel.RedisRegistrationData
		if err := json.Unmarshal([]byte(val), &redisData); err != nil {
			log.Error("failed to unmarshal redis registration data", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal server error"})
			return
		}

		if vr.OTP != redisData.OTP {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "invalid OTP"})
			return
		}

		_, err = userRepo.CreateUser(c.Request.Context(), redisData.Request)
		if err != nil {
			if err == repository.ErrUserAlreadyExists {
				c.JSON(http.StatusConflict, model.ErrorResponse{Error: "user already exists"})
				return
			}
			if err == repository.ErrEmailAlreadyExists {
				c.JSON(http.StatusConflict, model.ErrorResponse{Error: "email already exists"})
				return
			}
			log.Error("register user failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to register user"})
			return
		}

		// Delete key from redis on success
		redisClient.Del(ctx, "registration:"+vr.Email)

		c.JSON(http.StatusCreated, model.SuccessResponse{Message: "Registration is successful"})
	}
}
