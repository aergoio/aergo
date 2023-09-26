package types

type validator func(tx *TxBody) error

var govValidators map[string]validator

func InitGovernance(consensus string, isPublic bool) {
	sysValidator := ValidateSystemTx
	if consensus != "dpos" {
		sysValidator = func(tx *TxBody) error {
			return ErrTxInvalidType
		}
	}

	govValidators = map[string]validator{
		AergoSystem: sysValidator,
		AergoName:   validateNameTx,
		AergoEnterprise: func(tx *TxBody) error {
			if isPublic {
				return ErrTxOnlySupportedInPriv
			}
			return nil
		},
	}
}
