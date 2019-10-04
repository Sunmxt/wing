package controller

import (
	"fmt"
	"git.stuhome.com/Sunmxt/wing/common"
	ccommon "git.stuhome.com/Sunmxt/wing/controller/common"
	"git.stuhome.com/Sunmxt/wing/model/account"
	"github.com/jinzhu/gorm"
	zxcvbn "github.com/nbutton23/zxcvbn-go"
	ldap "gopkg.in/ldap.v3"
)

func NewLDAPSearchRequest(ctx *ccommon.OperationContext, username string) (*ldap.SearchRequest, error) {
	if ctx.Runtime.Config == nil {
		return nil, common.ErrConfigMissing
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

func LDAPByName(ctx *ccommon.OperationContext, username string) (*ldap.SearchResult, error) {
	result, request := (*ldap.SearchResult)(nil), (*ldap.SearchRequest)(nil)
	conn, err := ctx.LDAPRootConnection()
	if err != nil {
		return nil, err
	}
	if request, err = NewLDAPSearchRequest(ctx, username); err != nil {
		return nil, err
	}
	if result, err = conn.Search(request); err != nil {
		return nil, common.NewInternalError(err)
	}
	return result, err
}

func AddLDAPAccount(ctx *ccommon.OperationContext, username, password, commonName string) error {
	if !common.ReMail.Match([]byte(username)) {
		return common.ErrUsernameNotMail
	}
	score := zxcvbn.PasswordStrength(password, []string{username})
	if score.Score < 2 || len(password) < 6 {
		return common.ErrWeakPassword
	}

	ctx.Log.Infof("add LDAP user \"%v\".", username)
	if ctx.Runtime.Config == nil {
		return common.ErrConfigMissing
	}
	config := ctx.Runtime.Config
	conn, err := ctx.LDAPRootConnection()

	hasher := account.NewMD5Hasher()
	if password, err = hasher.HashString(password); err != nil {
		return common.NewInternalError(err)
	}
	req := ldap.NewAddRequest(fmt.Sprintf(config.Auth.LDAP.RegisterRDN, username)+","+config.Auth.LDAP.BaseDN, nil)
	req.Attribute(config.Auth.LDAP.NameAttribute, []string{commonName})
	req.Attribute("userPassword", []string{"{MD5}" + password})
	req.Attribute("objectClass", config.Auth.LDAP.RegisterObjectClasses)
	for attr, pattern := range config.Auth.LDAP.RegisterAttributes {
		req.Attribute(attr, []string{fmt.Sprintf(pattern, username, commonName)})
	}

	if err = conn.Add(req); err != nil {
		return common.NewInternalError(err)
	}
	conn.Close()

	return nil
}

func AuthAsLDAPUser(ctx *ccommon.OperationContext, username, password string) (*account.Account, error) {
	ctx.Log.Infof("try to auth \"%v\" via LDAP.", username)

	if ctx.Runtime.Config == nil {
		return nil, common.ErrConfigMissing
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
		return nil, common.NewInternalError(err)
	}
	if len(result.Entries) < 1 {
		return nil, common.ErrInvalidAccount
	}
	if len(result.Entries) > 1 {
		ctx.Log.Infof("more then 1 LDAP Entries found for user \"%v\".", err.Error())
		return nil, common.ErrInvalidAccount
	}
	userDN := result.Entries[0].DN
	if err = conn.Bind(userDN, password); err != nil {
		ctx.Log.Infof("bind user DN \"%v\" error: "+err.Error(), userDN)
		return nil, common.NewInternalError(err)
	}
	conn.Close()

	ctx.Log.Infof("valid LDAP user \"%v\"", username)
	var account *account.Account
	if account, err = LegacyAccountByName(ctx, username, true, ""); err != nil {
		return nil, err
	}
	return account, nil
}

func LegacyAccountByName(ctx *ccommon.OperationContext, username string, create bool, passwordHash string) (*account.Account, error) {
	if !common.ReMail.Match([]byte(username)) {
		return nil, common.ErrUsernameNotMail
	}

	ctx.Log.Infof("load legacy user \"%v\"", username)

	db, err := ctx.Database()
	if err != nil {
		return nil, err
	}
	user := &account.Account{}
	if err = db.Where(&account.Account{Name: username}).First(user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			if !create {
				return nil, common.ErrInvalidUsername
			}
			user.Credentials = passwordHash
			user.Name = username
			if err = db.Save(user).Error; err != nil {
				return nil, common.NewInternalError(err)
			}
			ctx.Log.Infof("new legacy user created: " + username)
		}
	}
	return user, nil
}

func AuthAsLegacyUser(ctx *ccommon.OperationContext, username, password string) (user *account.Account, err error) {
	ctx.Log.Infof("try to auth \"%v\" as legacy user.", username)

	hasher := account.NewMD5Hasher()

	var toVerify string
	if toVerify, err = hasher.HashString(password); err != nil {
		return nil, common.NewInternalError(err)
	}

	if user, err = LegacyAccountByName(ctx, username, false, ""); err != nil {
		return nil, err
	}
	if toVerify != user.Credentials {
		return nil, common.ErrInvalidAccount
	}
	return user, nil
}

func AddLegacyAccount(ctx *ccommon.OperationContext, username, password string) error {
	ctx.Log.Infof("try to add legacy account \"%v\".", username)
	score := zxcvbn.PasswordStrength(password, []string{username})
	if score.Score < 2 || len(password) < 6 {
		return common.ErrWeakPassword
	}

	user, err := LegacyAccountByName(ctx, username, false, "")
	if err != nil && err != common.ErrInvalidUsername {
		return err
	}
	if user != nil {
		return common.ErrAccountExists
	}
	hasher := account.NewMD5Hasher()
	if password, err = hasher.HashString(password); err != nil {
		return err
	}
	if _, err = LegacyAccountByName(ctx, username, true, password); err != nil {
		return err
	}
	return nil
}
