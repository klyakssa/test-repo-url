package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RouterGroup struct {
	*gin.RouterGroup
}

func NewGroup(r *gin.RouterGroup) *RouterGroup {
	return &RouterGroup{r}
}

func (r *RouterGroup) Group(grp string) *RouterGroup {
	return &RouterGroup{RouterGroup: r.RouterGroup.Group(grp)}
}

func (r *RouterGroup) SGET(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	r.RouterGroup.GET(pattern, func(c *gin.Context) {
		handler(c.Writer, c.Request)
	})
}

func (r *RouterGroup) SPOST(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	r.RouterGroup.POST(pattern, func(c *gin.Context) {
		handler(c.Writer, c.Request)
	})
}

func (r *RouterGroup) GET(pattern string, handler gin.HandlerFunc) {
	r.RouterGroup.GET(pattern, handler)
}

func (r *RouterGroup) Middleware(middleware ...gin.HandlerFunc) {
	r.RouterGroup.Use(middleware...)
}
