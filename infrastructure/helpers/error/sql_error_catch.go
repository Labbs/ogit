package error

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	// Map des patterns génériques vers les erreurs GORM
	errorMap = map[string]error{
		"repository.name":                 gorm.ErrDuplicatedKey,
		"repository_name":                 gorm.ErrDuplicatedKey,
		"user.email":                      gorm.ErrDuplicatedKey,
		"user_email":                      gorm.ErrDuplicatedKey,
		"unique constraint failed":        gorm.ErrDuplicatedKey,
		"duplicate key":                   gorm.ErrDuplicatedKey,
		"duplicate entry":                 gorm.ErrDuplicatedKey,
		"violates unique constraint":      gorm.ErrDuplicatedKey,
		"foreign key constraint failed":   gorm.ErrForeignKeyViolated,
		"violates foreign key constraint": gorm.ErrForeignKeyViolated,
		"foreign key constraint fails":    gorm.ErrForeignKeyViolated,
		"check constraint failed":         gorm.ErrCheckConstraintViolated,
		"violates check constraint":       gorm.ErrCheckConstraintViolated,
		"check constraint":                gorm.ErrCheckConstraintViolated,
	}
)

func Catch(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return gorm.ErrRecordNotFound
	}

	errorMsg := strings.ToLower(err.Error())

	for pattern, gormErr := range errorMap {
		if strings.Contains(errorMsg, pattern) {
			return gormErr
		}
	}

	return err
}
