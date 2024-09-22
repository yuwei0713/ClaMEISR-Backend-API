package routers

import (
	"encoding/gob"
	"ginapi/controllers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func InitRouters() *gin.Engine {
	Router = gin.New()

	// using gob.Register create map for cookie store own data
	gob.Register(map[string]interface{}{})

	// create new cookie store，set secret key
	store := cookie.NewStore([]byte("secret"))
	Router.Use(sessions.Sessions("auth-session", store))

	AllRouter := Router.Group("/api")
	{
		AllRouter.POST("/Login", controllers.Login)
		AllRouter.POST("/Register", controllers.Register)
		// Router.POST("/Logout")

		DashBoardRoute := AllRouter.Group("/")
		{
			DashBoardRoute.GET("/", func(c *gin.Context) {
				c.String(200, "test")
			})
			QuestionManageRoute := AllRouter.Group("/QuestionManage")
			{
				QuestionManageRoute.GET("/", controllers.ShowQuestionnaire())
				QuestionManageRoute.POST("/QuestionDetail", controllers.ShowDetailQuestionnaire())
			}
			SearchRouter := AllRouter.Group("/Search")
			{
				SearchTeacherRoute := SearchRouter.Group("/Teacher")
				{
					SearchTeacherRoute.GET("/", controllers.ShowSearch("Teacher", "0"))
					SearchTeacherRoute.POST("/user", controllers.ShowDetail("Teacher", "0"))
				}
				SearchChildrenRoute := SearchRouter.Group("/Children")
				{
					SearchChildrenRoute.GET("/", controllers.ShowSearch("Children", "0"))
					SearchChildrenRoute.POST("/Child", controllers.ShowDetail("Children", "0"))
					SearchChildrenRoute.POST("/Update", controllers.ChildManage("Update", "0"))
				}
				SearchFillStatusRoute := SearchRouter.Group("/FillStatus")
				{
					SearchFillStatusRoute.GET("/", controllers.ShowSearch("FillStatus", "0"))
					SearchFillStatusRoute.POST("/DetailData", controllers.ShowDetail("FillResult", "0"))
				}
				DataExportRoute := AllRouter.Group("/DataExport")
				{
					DataExportRoute.GET("", controllers.ExportToExcel())
					// CustomizeExportRoute := Router.Group("/Customize")
					// {
					// 	CustomizeExportRoute.GET("/")
					// 	CustomizeExportRoute.POST("/")
					// }
				}
			}
			ManageRouter := AllRouter.Group("/Manage")
			{
				UserManageRoute := ManageRouter.Group("/Users")
				{
					UserManageRoute.GET("/", controllers.UserManage("Search", "0"))
					UserManageRoute.POST("/UserInsert", controllers.ClaMEISR_Register)
					UserManageRoute.POST("/UserUpdate", controllers.UserManage("Update", "0"))
					// UserManageRoute.POST("/UserDelete")
				}
				SchoolManageRoute := ManageRouter.Group("/Schools")
				{
					SchoolManageRoute.GET("/", controllers.SchoolManage("Search", "0"))
					SchoolManageRoute.POST("/SchoolInsert", controllers.SchoolManage("Insert", "0"))
					// SchoolManageRoute.POST("/SchoolUpdate")
					// SchoolManageRoute.POST("/SchoolDelete")
					SchoolManageRoute.POST("/ClassInsert", controllers.SchoolManage("Insert", "0"))
					SchoolManageRoute.POST("/ClassUpdate", controllers.SchoolManage("Update", "0"))
					SchoolManageRoute.POST("/ClassDelete", controllers.SchoolManage("Delete", "0"))
				}
			}
		}
	}

	return Router
}
