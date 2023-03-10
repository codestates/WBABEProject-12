package router

//router.go : api 전체 인입에 대한 관리 및 구성을 담당하는 파일
import (
	"fmt"
	ctl "go-ordering/controller"
	logger "go-ordering/logger"

	"github.com/gin-gonic/gin"
	swgFiles "github.com/swaggo/files"
	ginSwg "github.com/swaggo/gin-swagger"

	"go-ordering/docs" //swagger에 의해 자동 생성된 package
)

type Router struct {
	ct *ctl.Controller
}

func NewRouter(ctl *ctl.Controller) (*Router, error) {
	fmt.Println("router.NewRouter ctl : ", ctl)

	r := &Router{ct: ctl} //controller 포인터를 ct로 복사, 할당

	return r, nil
}

// cross domain을 위해 사용
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("router.CORS c : ", c)

		//~ 생략
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		//허용할 header 타입에 대해 열거
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, X-Forwarded-For, Authorization, accept, origin, Cache-Control, X-Requested-With")
		//허용할 method에 대해 열거
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// 임의 인증을 위한 함수
func liteAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("router.liteAuth c : ", c)
		//~ 생략
		if c == nil {
			c.Abort() // 미들웨어에서 사용, 이후 요청 중지
			return
		}
		//http 헤더내 "Authorization" 폼의 데이터를 조회
		auth := c.GetHeader("Authorization")
		//실제 인증기능이 올수있다. 단순히 출력기능만 처리 현재는 출력예시
		fmt.Println("Authorization-word ", auth)

		c.Next()
	}
}

// 실제 라우팅
func (p *Router) Idx() *gin.Engine {

	// 컨피그나 상황에 맞게 gin 모드 설정
	gin.SetMode(gin.ReleaseMode)
	// gin.SetMode(gin.DebugMode)

	r := gin.Default() //gin 선언
	//gin.Default()와 동일
	fmt.Println("router.Idx c : ", r)

	// 기존의 logger, recovery 대신 logger에서 선언한 미들웨어 사용
	//r.Use(gin.Logger())   //gin 내부 log, logger 미들웨어 사용 선언
	//r.Use(gin.Recovery()) //gin 내부 recover, recovery 미들웨어 사용 - 패닉복구
	r.Use(logger.GinLogger())
	r.Use(logger.GinRecovery(true))

	r.Use(CORS()) //crossdomain 미들웨어 사용 등록

	logger.Info("start server")

	r.GET("/swagger/:any", ginSwg.WrapHandler(swgFiles.Handler))
	docs.SwaggerInfo.Host = "localhost" //swagger 정보 등록

	//피주문자 그룹
	seller := r.Group("oos/seller", liteAuth())
	{
		seller.POST("/createMenu", p.ct.CreateMenu)
		seller.POST("/updateMenu", p.ct.UpdateMenu)
		seller.POST("/deleteMenu", p.ct.DeleteMenu)

		seller.POST("/searchMenu", p.ct.SearchMenu)
		seller.POST("/orderStates", p.ct.OrderStates)

		seller.POST("/viewMenu", p.ct.ViewMenu)

	}

	//주문자 그룹
	order := r.Group("oos/order", liteAuth())
	{

		order.POST("/searchMenu", p.ct.SearchMenu)

		order.POST("/createReview", p.ct.CreateReview)
		order.POST("/newOrder", p.ct.NewOrder)
		order.POST("/changeOrder", p.ct.ChangeOrder)
		order.POST("/searchOrder", p.ct.SearchOrder)

	}

	return r
}
