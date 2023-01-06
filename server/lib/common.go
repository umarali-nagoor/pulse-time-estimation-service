package lib

import (
	// "github.com/gin-gonic/gin"
	// "github.com/gin-gonic/gin/binding"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func parsePayload(input map[string]interface{}) {
	fmt.Printf("***** Inside parsePayload %v:", input)
	fmt.Printf("***** format", input["format_version"])
	fmt.Printf("***** terraform", input["terraform_version"])
}

// BindJSON is a shortcut for c.BindWith(obj, binding.JSON)
func BindJSON(obj interface{}, c *gin.Context) error {
	return BindWith(obj, binding.JSON, c)
}

// BindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func BindWith(obj interface{}, b binding.Binding, c *gin.Context) error {
	if err := b.Bind(c.Request, obj); err != nil {
		return err
	}
	return nil
}
