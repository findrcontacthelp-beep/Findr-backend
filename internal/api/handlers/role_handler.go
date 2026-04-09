package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/model"
)

func ListAvailableRoles(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, name FROM available_roles ORDER BY name`)
		if err != nil {
			log.Error("list roles failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list roles"})
			return
		}
		defer rows.Close()

		var roles []model.AvailableRole
		for rows.Next() {
			var r model.AvailableRole
			if err := rows.Scan(&r.ID, &r.Name); err != nil {
				continue
			}
			roles = append(roles, r)
		}

		c.JSON(http.StatusOK, roles)
	}
}

func CreateRole(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CreateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		var role model.AvailableRole
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO available_roles (name) VALUES ($1) RETURNING id, name`,
			req.Name,
		).Scan(&role.ID, &role.Name)
		if err != nil {
			log.Error("create role failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create role"})
			return
		}

		c.JSON(http.StatusCreated, role)
	}
}

func SubmitRoleRequest(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CreateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}

		var roleReq model.RoleRequest
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO role_requests (name) VALUES ($1) RETURNING id, name`,
			req.Name,
		).Scan(&roleReq.ID, &roleReq.Name)
		if err != nil {
			log.Error("submit role request failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to submit role request"})
			return
		}

		c.JSON(http.StatusCreated, roleReq)
	}
}

func ListRoleRequests(pool *pgxpool.Pool, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, name FROM role_requests ORDER BY name`)
		if err != nil {
			log.Error("list role requests failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list role requests"})
			return
		}
		defer rows.Close()

		var requests []model.RoleRequest
		for rows.Next() {
			var r model.RoleRequest
			if err := rows.Scan(&r.ID, &r.Name); err != nil {
				continue
			}
			requests = append(requests, r)
		}

		c.JSON(http.StatusOK, requests)
	}
}
