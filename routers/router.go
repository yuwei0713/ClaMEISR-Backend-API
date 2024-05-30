package routers

import (
	"ginapi/controllers"

	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func InitRouters() *gin.Engine {
	Router = gin.New()
	DashBoardRoute := Router.Group("/")
	{
		DashBoardRoute.GET("/", func(c *gin.Context) {
			c.String(200, "test")
		})
		QuestionManageRoute := Router.Group("/QuestionManage")
		{
			QuestionManageRoute.GET("/", controllers.ShowQuestionnaire())
			QuestionManageRoute.POST("/QuestionDetail", controllers.ShowDetailQuestionnaire())
		}
		SearchRouter := Router.Group("/Search")
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
			}
			SearchFillStatusRoute := SearchRouter.Group("/FillStatus")
			{
				SearchFillStatusRoute.GET("/", controllers.ShowSearch("FillStatus", "0"))
				SearchFillStatusRoute.POST("/DetailData", controllers.ShowDetail("FillResult", "0"))
			}
			// 	DataExportRoute := Router.Group("/DataExport")
			// 	{
			// 		DataExportRoute.POST("/")
			// 		CustomizeExportRoute := Router.Group("/Customize")
			// 		{
			// 			CustomizeExportRoute.GET("/")
			// 			CustomizeExportRoute.POST("/")
			// 		}
			// 	}
		}
		ManageRouter := Router.Group("/Manage")
		{
			UserManageRoute := ManageRouter.Group("/Users")
			{
				UserManageRoute.GET("/", controllers.UserManage("Search", "0"))
				// UserManageRoute.POST("/UserInsert")
				// UserManageRoute.POST("/UserUpdate")
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
				// SchoolManageRoute.POST("/ClassDelete")
			}
		}
	}
	return Router
}
