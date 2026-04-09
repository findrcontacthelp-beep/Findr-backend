package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/findr-app/findr-backend/internal/api/handlers"
	"github.com/findr-app/findr-backend/internal/api/middleware"
	authrouter "github.com/findr-app/findr-backend/internals/api/router"
)

type Deps struct {
	Pool      *pgxpool.Pool
	Redis     *redis.Client
	Log       *zap.Logger
	JWTSecret string
	AdminUIDs []string
}

func Setup(r *gin.Engine, d *Deps) {
	auth := middleware.AuthRequired(d.Pool, d.JWTSecret, d.Log)
	admin := middleware.RequireAdmin(d.AdminUIDs)

	v1 := r.Group("/api/v1")

	// Health
	v1.GET("/health", handlers.HealthCheck(d.Pool))

	// Auth
	authrouter.Setup(r, &authrouter.Deps{
		Pool:      d.Pool,
		Redis:     d.Redis,
		Log:       d.Log,
		JWTSecret: d.JWTSecret,
	})

	// Users
	users := v1.Group("/users")
	{
		users.GET("/me", auth, handlers.GetCurrentUser(d.Pool, d.Log))
		users.PUT("/me", auth, handlers.UpdateUser(d.Pool, d.Log))
		users.DELETE("/me", auth, handlers.DeleteUser(d.Pool, d.Log))
		users.PATCH("/me/skills", auth, handlers.UpdateSkills(d.Pool, d.Log))
		users.PATCH("/me/social-links", auth, handlers.UpdateSocialLinks(d.Pool, d.Log))
		users.GET("", handlers.ListUsers(d.Pool, d.Log))
		users.GET("/:id", handlers.GetUser(d.Pool, d.Log))
		users.POST("/:id/views", auth, handlers.RecordProfileView(d.Pool, d.Log))
		users.GET("/:id/views", auth, handlers.GetProfileViewers(d.Pool, d.Log))
		users.POST("/:id/ratings", auth, handlers.RateUser(d.Pool, d.Log))
		users.GET("/:id/ratings", handlers.GetUserRatings(d.Pool, d.Log))
		users.GET("/:id/ratings/average", handlers.GetAverageRating(d.Pool, d.Log))
		users.POST("/:id/verifications", auth, handlers.VerifyUser(d.Pool, d.Log))
		users.GET("/:id/verifications", handlers.GetUserVerifications(d.Pool, d.Log))
		users.DELETE("/:id/verifications", auth, handlers.RemoveVerification(d.Pool, d.Log))
		users.POST("/me/activities", auth, handlers.CreateExtraActivity(d.Pool, d.Log))
		users.GET("/:id/activities", handlers.GetExtraActivities(d.Pool, d.Log))
		users.PUT("/me/activities/:activityId", auth, handlers.UpdateExtraActivity(d.Pool, d.Log))
		users.DELETE("/me/activities/:activityId", auth, handlers.DeleteExtraActivity(d.Pool, d.Log))
		users.GET("/me/notifications", auth, handlers.GetNotifications(d.Pool, d.Log))
		users.PUT("/me/notifications/:notifId", auth, handlers.UpdateNotification(d.Pool, d.Log))
		users.DELETE("/me/notifications/:notifId", auth, handlers.DeleteNotification(d.Pool, d.Log))
	}

	// Projects
	projects := v1.Group("/projects")
	{
		projects.POST("", auth, handlers.CreateProject(d.Pool, d.Log))
		projects.GET("", handlers.ListProjects(d.Pool, d.Log))
		projects.GET("/me", auth, handlers.GetMyProjects(d.Pool, d.Log))
		projects.GET("/:id", handlers.GetProject(d.Pool, d.Log))
		projects.PUT("/:id", auth, handlers.UpdateProject(d.Pool, d.Log))
		projects.DELETE("/:id", auth, handlers.DeleteProject(d.Pool, d.Log))
		projects.POST("/:id/like", auth, handlers.ToggleLike(d.Pool, d.Log))
		projects.GET("/:id/stats", handlers.GetProjectStats(d.Pool, d.Log))
		projects.POST("/:id/comments", auth, handlers.CreateComment(d.Pool, d.Log))
		projects.GET("/:id/comments", handlers.GetComments(d.Pool, d.Log))
		projects.POST("/:id/views", auth, handlers.RecordProjectView(d.Pool, d.Log))
		projects.GET("/:id/views/count", handlers.GetProjectViewCount(d.Pool, d.Log))
		projects.POST("/:id/enrollments", auth, handlers.ApplyToProject(d.Pool, d.Log))
		projects.GET("/:id/enrollments", auth, handlers.GetProjectEnrollments(d.Pool, d.Log))
		projects.POST("/:id/registrations", auth, handlers.RegisterForEvent(d.Pool, d.Log))
		projects.GET("/:id/registrations", auth, handlers.GetEventRegistrations(d.Pool, d.Log))
	}

	comments := v1.Group("/comments")
	{
		comments.GET("/:commentId/replies", handlers.GetReplies(d.Pool, d.Log))
		comments.PUT("/:commentId", auth, handlers.UpdateComment(d.Pool, d.Log))
		comments.DELETE("/:commentId", auth, handlers.DeleteComment(d.Pool, d.Log))
	}

	enrollments := v1.Group("/enrollments", auth)
	{
		enrollments.GET("/me", handlers.GetMyEnrollments(d.Pool, d.Log))
		enrollments.PATCH("/:id/accept", handlers.AcceptEnrollment(d.Pool, d.Log))
		enrollments.PATCH("/:id/reject", handlers.RejectEnrollment(d.Pool, d.Log))
	}

	eventRegs := v1.Group("/registrations", auth)
	{
		eventRegs.GET("/me", handlers.GetMyRegistrations(d.Pool, d.Log))
		eventRegs.PATCH("/:id/cancel", handlers.CancelRegistration(d.Pool, d.Log))
	}

	chats := v1.Group("/chats", auth)
	{
		chats.POST("", handlers.CreateOrGetChat(d.Pool, d.Log))
		chats.GET("", handlers.GetMyChats(d.Pool, d.Log))
		chats.GET("/:id", handlers.GetChatDetail(d.Pool, d.Log))
		chats.GET("/:id/messages", handlers.GetMessages(d.Pool, d.Log))
		chats.POST("/:id/messages", handlers.SendMessage(d.Pool, d.Log, nil, nil))
	}

	v1.GET("/roles", handlers.ListAvailableRoles(d.Pool, d.Log))
	v1.POST("/roles", auth, admin, handlers.CreateRole(d.Pool, d.Log))
	v1.POST("/role-requests", auth, handlers.SubmitRoleRequest(d.Pool, d.Log))
	v1.GET("/role-requests", auth, admin, handlers.ListRoleRequests(d.Pool, d.Log))

	v1.POST("/placement-reviews", auth, handlers.CreatePlacementReview(d.Pool, d.Log))
	v1.GET("/placement-reviews", handlers.ListPlacementReviews(d.Pool, d.Log))
	v1.GET("/placement-reviews/:id", handlers.GetPlacementReview(d.Pool, d.Log))

	v1.GET("/metadata/:key", handlers.GetMetadata(d.Pool, d.Log))
	v1.PUT("/metadata/:key", auth, admin, handlers.SetMetadata(d.Pool, d.Log))

	v1.GET("/topics", handlers.ListTopics(d.Pool, d.Log))
	v1.POST("/topics", auth, admin, handlers.CreateTopic(d.Pool, d.Log))
}
