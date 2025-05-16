package accountModuleDto

import (
	"go-gin-test-job/src/database/entities"
)

type AccountDto struct {
	Id        int64   `json:"id" example:"1"`
	Address   string  `json:"address" example:"1JzfdUygUFk2M6KS3ngFMGRsy5vsH4N37a"`
	Name      string  `json:"name" example:"Main Account"`
	Rank      int8    `json:"rank" example:"50"`
	Memo      *string `json:"memo" example:"Important account for transactions"`
	Balance   string  `json:"balance" example:"12.1234"`
	Status    string  `json:"status" example:"On"`
	CreatedAt int64   `json:"created_at" example:"1600000000000"`
	UpdatedAt int64   `json:"updated_at" example:"1600000000000"`
}

func CreateAccountDto(account *entities.Account) AccountDto {
	return AccountDto{
		Id:        account.Id,
		Address:   account.Address,
		Name:      account.Name,
		Rank:      account.Rank,
		Memo:      account.Memo,
		Balance:   account.Balance.String(),
		Status:    string(account.Status),
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}
}
