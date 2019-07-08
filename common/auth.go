package common

import (
	"fmt"
	"git.stuhome.com/Sunmxt/wing/model"
	"github.com/jinzhu/gorm"
	zxcvbn "github.com/nbutton23/zxcvbn-go"
	ldap "gopkg.in/ldap.v3"
	"regexp"
)

var ReMail *regexp.Regexp

func (ctx *OperationContext) NewLDAPSearchRequest(username string) (*ldap.SearchRequest, error) {
	if ctx.Runtime.Config == nil {
		return nil, ErrConfigMissing
	}
	config := ctx.Runtime.Config
	searchRequest := ldap.NewSearchRequest(
		config.Auth.LDAP.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(config.Auth.LDAP.SearchPattern, username),
		[]string{config.Auth.LDAP.NameAttribute},
		nil,
	)
	return searchRequest, nil
}

func (ctx *OperationContext) LDAPByName(username string) (*ldap.SearchResult, error) {
	result, request := (*ldap.SearchResult)(nil), (*ldap.SearchRequest)(nil)
	conn, err := ctx.LDAPRootConnection()
	if err != nil {
		return nil, err
	}
	if request, err = ctx.NewLDAPSearchRequest(username); err != nil {
		return nil, err
	}
	if result, err = conn.Search(request); err != nil {
		return nil, NewInternalError(err)
	}
	return result, err
}

func (ctx *OperationContext) AddLDAPAccount(username, password, commonName string) error {
	if !ReMail.Match([]byte(username)) {
		return ErrUsernameNotMail
	}
	score := zxcvbn.PasswordStrength(password, []string{username})
	if score.Score < 2 || len(password) < 6 {
		return ErrWeakPassword
	}

	ctx.Log.Infof("add LDAP user \"%v\".", username)
	if ctx.Runtime.Config == nil {
		return ErrConfigMissing
	}
	config := ctx.Runtime.Config
	conn, err := ctx.LDAPRootConnection()

	hasher := model.NewMD5Hasher()
	if password, err = hasher.HashString(password); err != nil {
		return NewInternalError(err)
	}
	req := ldap.NewAddRequest(fmt.Sprintf(config.Auth.LDAP.RegisterRDN, username)+","+config.Auth.LDAP.BaseDN, nil)
	req.Attribute(config.Auth.LDAP.NameAttribute, []string{commonName})
	req.Attribute("userPassword", []string{"{MD5}" + password})
	req.Attribute("objectClass", config.Auth.LDAP.RegisterObjectClasses)
	for attr, pattern := range config.Auth.LDAP.RegisterAttributes {
		req.Attribute(attr, []string{fmt.Sprintf(pattern, username, commonName)})
	}

	if err = conn.Add(req); err != nil {
		return NewInternalError(err)
	}
	conn.Close()

	return nil
}

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
	if account, err = ctx.LegacyAccountByName(username, true, ""); err != nil {
		return nil, err
	}
	return account, nil
}

func (ctx *OperationContext) LegacyAccountByName(username string, create bool, passwordHash string) (*model.Account, error) {
	if !ReMail.Match([]byte(username)) {
		return nil, ErrUsernameNotMail
	}

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

	hasher := model.NewMD5Hasher()

	var toVerify string
	if toVerify, err = hasher.HashString(password); err != nil {
		return nil, NewInternalError(err)
	}

	if account, err = ctx.LegacyAccountByName(username, false, ""); err != nil {
		return nil, err
	}
	if toVerify != account.Credentials {
		return nil, ErrInvalidAccount
	}
	return account, nil
}

func (ctx *OperationContext) AddLegacyAccount(username, password string) error {
	ctx.Log.Infof("try to add legacy account \"%v\".", username)
	score := zxcvbn.PasswordStrength(password, []string{username})
	if score.Score < 2 || len(password) < 6 {
		return ErrWeakPassword
	}

	account, err := ctx.LegacyAccountByName(username, false, "")
	if err != nil && err != ErrInvalidUsername {
		return err
	}
	if account != nil {
		return ErrAccountExists
	}
	hasher := model.NewMD5Hasher()
	if password, err = hasher.HashString(password); err != nil {
		return err
	}
	if _, err = ctx.LegacyAccountByName(username, true, password); err != nil {
		return err
	}
	return nil
}
