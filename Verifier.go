package gin_endpoint_verifier

import (
	"github.com/gin-gonic/gin"
	"regexp"
	"fmt"
	"errors"
	"strings"
)


type fieldRequirements struct  {
	name string
	format *regexp.Regexp
}

func verifyInput(c * gin.Context, verify []fieldRequirements) (map[string]string, error) {
	var errorStrings []string
	finalValues := make(map[string]string)

	for _,value := range verify {
		field, exists := c.GetPostForm(value.name)

		if !exists {
			error := fmt.Sprintf("\"%s\" is needed for this request",value.name)
			errorStrings = append(errorStrings, error)
			continue
		}

		if value.format != nil && !value.format.MatchString(field){
			error := fmt.Sprintf("\"%s\" is invalid for \"%s\", should match \"%s\" ",field,value.name,value.format.String())
			errorStrings = append(errorStrings, error)
			continue
		}
		finalValues[value.name] = field
	}

	if len(errorStrings) > 0 {
		return finalValues , errors.New(strings.Join(errorStrings," "))
	} else {
		return finalValues , nil
	}
}