package common

import (
	"fmt"
	"git.stuhome.com/Sunmxt/wing/model"
	"github.com/jinzhu/gorm"
	ldap "gopkg.in/ldap.v3"
)

func (ctx *OperationContext) AuthAsLDAPUser(username, password string) (*model.Account, error) {
	ctx.Log.Infof("try to auth \"%v\" via LDAP.", username)

	if ctx.Runtime.Config == nil {
		return nil, ErrConfigMissing
	}
	config := ctx.Runtime.Config
	conn, err := ctx.LDAPRootConnection()
	if err != nil {
		return nil, err
	}
	searchRequest, result := ldap.NewSearchRequest(
		config.Auth.LDAP.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(config.Auth.LDAP.SearchPattern, username),
		[]string{config.Auth.LDAP.NameAttribute},
		nil,
	), (*ldap.SearchResult)(nil)
	if result, err = conn.Search(searchRequest); err != nil {
		return nil, NewInternalError(err)
	}
	if len(result.Entries) < 1 {
		return nil, ErrInvalidAccount
	}
	if len(result.Entries) > 1 {
		ctx.Log.Infof("more then 1 LDAP Entries found for user \"%v\".", err.Error())
		return nil, ErrInvalidAccount
	}
	userDN := result.Entries[0].DN
	if err = conn.Bind(userDN, password); err != nil {
		ctx.Log.Infof("bind user DN \"%v\" error: "+err.Error(), userDN)
		return nil, NewInternalError(err)
	}
	conn.Close()

	ctx.Log.Infof("valid LDAP user \"%v\"", username)
	var account *model.Account
	if account, err = ctx.legacyAccountByName(username, true, ""); err != nil {
		return nil, err
	}
	return account, nil
}

func (ctx *OperationContext) legacyAccountByName(username string, create bool, passwordHash string) (*model.Account, error) {
	ctx.Log.Infof("load legacy user \"%v\"", username)

	db, err := ctx.Database()
	if err != nil {
		return nil, err
	}
	account := &model.Account{}
	if err = db.Where(&model.Account{Name: username}).First(account).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			if !create {
				return nil, ErrInvalidUsername
			}
			account.Credentials = passwordHash
			account.Name = username
			if err = db.Save(account).Error; err != nil {
				return nil, NewInternalError(err)
			}
			ctx.Log.Infof("new legacy user created: " + username)
		}
	}
	return account, nil
}

func (ctx *OperationContext) AuthAsLegacyUser(username, password string) (account *model.Account, err error) {
	ctx.Log.Infof("try to auth \"%v\" as legacy user.", username)

	if ctx.Runtime.Config == nil {
		return nil, ErrConfigMissing
	}
	config := ctx.Runtime.Config

	var hasher *model.SecretHasher
	if hasher, err = model.NewSecretHasher(config.SessionToken); err != nil {
		return nil, ErrInvalidSessionToken
	}
	var toVerify string
	toVerify, err = hasher.HashString(password)
	if account, err = ctx.legacyAccountByName(username, false, ""); err != nil {
		return nil, err
	}
	if toVerify != account.Credentials {
		return nil, ErrInvalidAccount
	}
	return account, nil
}
