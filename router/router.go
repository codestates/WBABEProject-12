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
	/* [코드리뷰]
	 * Group을 사용하여 API 성격에 따라 request를 관리하는 코드는 매우 좋은 코드입니다.
	 * 또한 주문자와 피주문자를 잘 나누어주셨습니다.
     * 일반적으로 현업에서도 이와 같은 코드를 자주 사용합니다. 훌륭합니다.
	 *
	 * 코드의 확장성을 고려하였을때, endpoint 관리를 함께 고려한 코드를 개발하는 것도 추천드립니다.
	 * 예를들어 /order/status 를 호출하는 클라이언트(Web, App, etc..)들이 실시간으로 들어오고 있을 때,
	 * controller의 GetOrderStatus function을 변경해야 하는 상황이 발생한다면,
	 * /order/status2 로 받아주는 경우가 있을 것이고(/order/status는 그대로 받아주면서)
	 * 처음부터 /order/v1/status 로 관리되며, /order/v2/status 리뉴얼 버전에 따라 version up을 시켜
	 * v01 방식의 클라이언트와, v02 방식의 클라이언트를 모두 받아줄 수 있는 확장성 있는 코드를 구현해보시는 것을 추천드립니다.
	 */

	return r
}
