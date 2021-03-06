package gin_endpoint_verifier

import (
	"github.com/gin-gonic/gin"
	"regexp"
	"fmt"
	"errors"
	"strings"
	"net/url"
	"net/http"
	"bytes"
	"github.com/dgrijalva/jwt-go"
)


type FieldRequirements struct  {
	Name   string
	Format *regexp.Regexp
}

const EmailRegexString = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
const AllAcceptingRegexString = "^*"

const FormEncodedHeader = "application/x-www-form-urlencoded; charset=utf-8"

//todo fix this horrible code duplication and header auth
func ObtainVerifyPostInput(c * gin.Context, verify []FieldRequirements) (map[string]interface{}, error) {

	var errorStrings []string
	finalValues := make(map[string]interface{})

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

func ParseAndHandleFormResponse(response *http.Response) (url.Values,int,error){
	buf := bytes.NewBuffer(make([]byte, 0, response.ContentLength))
	_, _ = buf.ReadFrom(response.Body)

	values,_:=url.ParseQuery(buf.String())
	response.Body.Close()

	if response.StatusCode != 200 {
		return values,response.StatusCode,errors.New(values.Get("error"))
	} else {
		return values,response.StatusCode,nil
	}
}

//gets the byte stream however parses it as an error if the status is failing
func GetRespByteStream(response *http.Response) ([]byte, int, error){
	buf := bytes.NewBuffer(make([]byte, 0, response.ContentLength))
	_, _ = buf.ReadFrom(response.Body)

	if response.StatusCode != 200 {
		values,_:=url.ParseQuery(buf.String())
		response.Body.Close()
		return buf.Bytes(),response.StatusCode,errors.New(values.Get("error"))
	} else {
		return buf.Bytes(),200,nil
	}
}

//takes an error and serialises it to the body as a form value
func HandleRequestErrors(c *gin.Context, code int,err error) {
	if err != nil {
		v := url.Values{}
		v.Add("error",err.Error())
		c.String(code,v.Encode())
	} else {
		c.String(code,"Encountered unknown error")
	}
}

func GetEmailRegex() * regexp.Regexp {
	return regexp.MustCompile(EmailRegexString)
}

func GetGeneralRegex() * regexp.Regexp {
	return regexp.MustCompile(AllAcceptingRegexString)
}


func EncodeJWT(toEncode string, key string, rootKey string) (string,error){
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		rootKey: toEncode})
	secretKey :=[]byte (key)
	return token.SignedString(secretKey)
}

func DecodeJWT(originToken string,key string, rootKey string)(url.Values,error){
	secretKey :=[]byte (key)
	token, err := jwt.Parse(originToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil,err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		return nil,errors.New("invalid token")
	}

	tokenString := claims[rootKey].(string)
	parsedValues , parseError := url.ParseQuery(tokenString)

	if parseError != nil {
		return nil, parseError
	}

	return parsedValues,nil
}

func verifyHeader(c *gin.Context) (error) {
	//add future supported headers here?
	if c.GetHeader("Content-Type") != FormEncodedHeader {
		return errors.New("Content Type invalid, must be Form encoded")
	} else {
		return nil
	}
}