package routers

import (
	"github.com/astaxie/beego"
	"github.com/toophy/doors/controllers"
)

func init() {
	beego.AutoRouter(&controllers.LocalController{})
}
