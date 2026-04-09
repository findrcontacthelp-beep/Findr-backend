package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/model"
	"github.com/findr-app/findr-backend/internals/api/middleware"
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

func Register(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	userRepo := repository.NewUserRepository(pool)

	return func(c *gin.Context) {
		firebaseUID := middleware.GetUserUID(c)
		if firebaseUID == "" {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "missing user identity"})
			return
		}

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

		u, err := userRepo.CreateUser(c.Request.Context(), firebaseUID, req)
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

		c.JSON(http.StatusCreated, u)
	}
}
