package accountModuleDto

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	errorHelpers "go-gin-test-job/src/common/error-helpers"
	errorMessages "go-gin-test-job/src/common/error-messages"
	"go-gin-test-job/src/common/validations"
	"go-gin-test-job/src/database/entities"
	"strings"
)

type PostCreateAccountRequestDto struct {
	Address string                 `json:"address" validate:"AccountAddressValidation" example:"1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a"`
	Name    string                 `json:"name" validate:"required,max=255" example:"Main Account"`
	Rank    int8                   `json:"rank" validate:"required,min=0,max=100" example:"50"`
	Memo    *string                `json:"memo" validate:"omitempty,max=65535" example:"Important account for transactions"`
	Status  entities.AccountStatus `json:"status" validate:"AccountStatusValidation" enums:"On,Off" example:"On"`
}

var postCreateAccountRequestDtoValidator *validator.Validate

func init() {
	postCreateAccountRequestDtoValidator = validator.New()
	_ = postCreateAccountRequestDtoValidator.RegisterValidation("AccountAddressValidation", validations.AccountAddressValidation)
	_ = postCreateAccountRequestDtoValidator.RegisterValidation("AccountStatusValidation", validations.AccountStatusValidation)
}

func validatePostCreateAccountRequestDto(dto *PostCreateAccountRequestDto) error {
	return postCreateAccountRequestDtoValidator.Struct(dto)
}

// CreatePostCreateAccountRequestDto is the Gin version for handling the request
func CreatePostCreateAccountRequestDto(c *gin.Context) (PostCreateAccountRequestDto, error) {
	var dto PostCreateAccountRequestDto
	// Parse body params into DTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		errorMessage := PostCreateAccountRequestDtoQueryParseErrorMessage(err)
		return dto, errorHelpers.RespondBadRequestError(c, errorMessage)
	}
	// Validate the DTO
	if err := validatePostCreateAccountRequestDto(&dto); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errorMessage := PostCreateAccountRequestDtoValidateErrorMessage(err)
			return dto, errorHelpers.RespondBadRequestError(c, errorMessage)
		}
	}
	return dto, nil
}

func PostCreateAccountRequestDtoQueryParseErrorMessage(err error) string {
	return errorMessages.DefaultQueryParseErrorMessage()
}

func PostCreateAccountRequestDtoValidateErrorMessage(err validator.FieldError) string {
	var errorMessage string
	if err.Field() == "Address" && err.Tag() == "AccountAddressValidation" {
		errorMessage = fmt.Sprintf("%s format is wrong", err.Field())
	} else if err.Field() == "Status" && err.Tag() == "AccountStatusValidation" {
		errorMessage = fmt.Sprintf("%s must be one of the next values: %s", err.Field(), strings.Join(entities.AccountStatusList, ","))
	} else if err.Field() == "Name" && err.Tag() == "required" {
		errorMessage = fmt.Sprintf("%s is required", err.Field())
	} else if err.Field() == "Name" && err.Tag() == "max" {
		errorMessage = fmt.Sprintf("%s must be shorter than or equal to %s characters", err.Field(), err.Param())
	} else if err.Field() == "Rank" && err.Tag() == "required" {
		errorMessage = fmt.Sprintf("%s is required", err.Field())
	} else if err.Field() == "Rank" && (err.Tag() == "min" || err.Tag() == "max") {
		errorMessage = fmt.Sprintf("%s must be between 0 and 100", err.Field())
	} else if err.Field() == "Memo" && err.Tag() == "max" {
		errorMessage = fmt.Sprintf("%s must be shorter than or equal to %s characters", err.Field(), err.Param())
	} else {
		errorMessage = errorMessages.DefaultFieldErrorMessage(err.Field())
	}
	return errorMessage
}
