package gin_endpoint_verifier

import (
	"github.com/gin-gonic/gin"
	"regexp"
	"fmt"
	"errors"
	"strings"
	"net/url"
)


type FieldRequirements struct  {
	Name   string
	Format *regexp.Regexp
}

const EmailRegexString = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
const AllAcceptingRegexString = "^*"

const FormEncodedHeader = "application/x-www-form-urlencoded; charset=utf-8"

//todo fix this horrible code duplication
func ObtainVerifyPostInput(c * gin.Context, verify []FieldRequirements) (map[string]string, error) {
	headerErr := verifyHeader(c)
	if headerErr != nil {
		return nil,headerErr
	}

	var errorStrings []string
	finalValues := make(map[string]string)

	for _,value := range verify {
		field, exists := c.GetPostForm(value.Name) //todo add multiple header type support

		if !exists {
			error := fmt.Sprintf("\"%s\" is needed for this request",value.Name)
			errorStrings = append(errorStrings, error)
			continue
		}

		if value.Format != nil && !value.Format.MatchString(field){
			error := fmt.Sprintf("\"%s\" is invalid for \"%s\", should match \"%s\" ",field,value.Name,value.Format.String())
			errorStrings = append(errorStrings, error)
			continue
		}
		finalValues[value.Name] = field
	}

	if len(errorStrings) > 0 {
		return finalValues , errors.New(strings.Join(errorStrings," "))
	} else {
		return finalValues , nil
	}
}

func ObtainVerifyGetInput(c * gin.Context, verify []FieldRequirements)(map[string]interface{},error){
	headerErr := verifyHeader(c)
	if headerErr != nil {
		return nil,headerErr
	}

	var errorStrings []string
	finalValues := make(map[string]interface{})


	for _,value := range verify {
		field := c.Query(value.Name) //todo add multiple header type support

		if field == "" {
			error := fmt.Sprintf("\"%s\" is needed for this request",value.Name)
			errorStrings = append(errorStrings, error)
			continue
		}

		if value.Format != nil && !value.Format.MatchString(field){
			error := fmt.Sprintf("\"%s\" is invalid for \"%s\", should match \"%s\" ",field,value.Name,value.Format.String())
			errorStrings = append(errorStrings, error)
			continue
		}
		finalValues[value.Name] = field
	}

	if len(errorStrings) > 0 {
		return finalValues , errors.New(strings.Join(errorStrings," "))
	} else {
		return finalValues , nil
	}
}

func GetEmailRegex() * regexp.Regexp {
	return regexp.MustCompile(EmailRegexString)
}

func GetGeneralRegex() * regexp.Regexp {
	return regexp.MustCompile(AllAcceptingRegexString)
}

func HandleRequestErrors(c *gin.Context, code int,err error) {
	v := url.Values{}
	v.Add("error",err.Error())
	c.String(code,v.Encode())
}

func verifyHeader(c *gin.Context) (error) {
	//add future supported headers here?
	if c.GetHeader("Content-Type") != FormEncodedHeader {
		return errors.New("Content Type invalid, must be Form encoded")
	} else {
		return nil
	}
}