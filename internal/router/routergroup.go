package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RouterGroup struct {
	r *gin.RouterGroup
}

func NewGroup(r *gin.RouterGroup) *RouterGroup {
	return &RouterGroup{r: r}
}

func (r *RouterGroup) Group(grp string) *RouterGroup {
	return &RouterGroup{r: r.r.Group(grp)}
}

func (r *RouterGroup) SGET(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	r.r.GET(pattern, func(c *gin.Context) {
		handler(c.Writer, c.Request)
	})
}

func (r *RouterGroup) SPOST(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	r.r.POST(pattern, func(c *gin.Context) {
		handler(c.Writer, c.Request)
	})
}
