package common

import (
	"errors"
	mlog "git.stuhome.com/Sunmxt/wing/log"
	"git.stuhome.com/Sunmxt/wing/model"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var ErrConfigMissing error = errors.New("Configuration is missing in context.")

type OperationContext struct {
	Runtime     *WingRuntime
	Log         *log.Entry
	DB          *gorm.DB
	RBACContext *model.RBACContext
	Account     model.Account
	Client      *kubernetes.Clientset
}

func (ctx *OperationContext) LogContext() {
}

func (ctx *OperationContext) KubeClient() (*kubernetes.Clientset, error) {
	if ctx.Client != nil {
		return ctx.Client, nil
	}
	clientset, err := kubernetes.NewForConfig(ctx.Runtime.ClusterConfig)
	if err != nil {
		return nil, err
	}
	ctx.Client = clientset
	return ctx.Client, nil
}

func (ctx *OperationContext) Database() (db *gorm.DB, err error) {
	if ctx.DB != nil {
		return ctx.DB, nil
	}
	if ctx.Runtime.Config == nil {
		return nil, ErrConfigMissing
	}
	if ctx.DB, err = gorm.Open(ctx.Runtime.Config.DB.SQLEngine, ctx.Runtime.Config.DB.SQLDsn); err != nil {
		return nil, errors.New("Failed to open database: " + err.Error())
	}
	if ctx.Runtime.Config.Debug {
		ctx.DB.LogMode(true)
	}
	ctx.DB.SetLogger(mlog.GormLogger{
		Log: ctx.Log.WithFields(log.Fields{
			"type": "gorm",
		}),
	})
	return ctx.DB, nil
}

func (ctx *OperationContext) GetAccount() *model.Account {
	if ctx.Account.ID > 0 {
		return &ctx.Account
	}
	db, err := ctx.Database()
	if err != nil {
		return nil
	}
	if err = db.Where(model.Account{Name: ctx.Account.Name}).First(&ctx.Account).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			ctx.Log.Info("[Role] Anonymous.")
		} else {
			ctx.Log.Error("[Role] Failed to load account: " + err.Error())
		}
		return nil
	}
	return &ctx.Account
}

func (ctx *OperationContext) RBAC() *model.RBACContext {
	if ctx.RBACContext != nil {
		return ctx.RBACContext
	}
	if ctx.Account.Name == "" {
		ctx.Log.Info("[RBAC] Anonymous.")
		return nil
	}
	ctx.RBACContext = model.NewRBACContext(ctx.Account.Name)
	db, err := ctx.Database()
	if err != nil {
		ctx.Log.Warnf("[RBAC] Cannot load RBAC rule for user \"%v\": %v", ctx.Account.Name, err.Error())
		return nil
	}
	if err = ctx.RBACContext.Load(db); err != nil {
		ctx.Log.Warn("[RBAC] RBAC rules not loaded: " + err.Error())
		return nil
	}
	return ctx.RBACContext
}

func (ctx *OperationContext) Permitted(resource string, verbs int64) bool {
	rbac := ctx.RBAC()
	if rbac == nil || !rbac.Permitted(resource, verbs) {
		return false
	}
	return true
}
