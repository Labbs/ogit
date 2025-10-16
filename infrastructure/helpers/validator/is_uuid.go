package validator

import v "github.com/go-playground/validator/v10"

func IsValidUUID(uuid string) bool {
	validate := v.New()

	type tempStruct struct {
		Value string `validate:"uuid4"`
	}

	temp := tempStruct{Value: uuid}
	err := validate.Struct(temp)
	return err == nil
}
