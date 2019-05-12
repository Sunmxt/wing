package uac

import (
	"fmt"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// High-level role rule interface
type ContextRoleRule struct {
	Resource string
	Verbs    int64

	dirty   bool
	roleCtx *ContextRole
}

func (r *ContextRoleRule) Save(db *gorm.DB) (err error) {
	record := &RoleRecord{}
	if !r.roleCtx.Loaded() {
		err = r.roleCtx.Load(db)
		if err != nil {
			return err
		}
	}

	if err = db.Where(&RoleRecord{RoleID: r.roleCtx.ID}).First(record).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.Errorf("[RBAC] Cannot fetch record for resource \"%v\" of role \"%v\": %v", r.Resource, r.roleCtx.Name, err.Error())
			return err
		}
	}
	record.ResourceName = r.Resource
	record.Verbs = r.Verbs
	record.RoleID = r.roleCtx.ID

	err = db.Save(record).Error
	if err != nil {
		log.Errorf("[RBAC] Cannot save record for resource \"%v\" of role \"%v\": %v", r.Resource, r.roleCtx.Name, err.Error())
		return err
	}
	return nil
}

func NewContextRoleRule(resource string, role *ContextRole) *ContextRoleRule {
	return &ContextRoleRule{
		Resource: resource,
		roleCtx:  role,
		dirty:    false,
	}
}

// High-level Role interface.
type ContextRole struct {
	ID    int
	Name  string
	Rules map[string]*ContextRoleRule
}

func Role(name string) *ContextRole {
	return &ContextRole{
		Name:  name,
		Rules: make(map[string]*ContextRoleRule, 0),
	}
}

func (r *ContextRole) modifyHelper(resource string, modify func(*ContextRoleRule)) {
	rule, ok := r.Rules[resource]
	if !ok {
		rule = NewContextRoleRule(resource, r)
		r.Rules[resource] = rule
	}
	modify(rule)
	rule.dirty = true
}
func (r *ContextRole) Grant(resource string, verbs int64) *ContextRole {
	r.modifyHelper(resource, func(rule *ContextRoleRule) {
		rule.Verbs |= verbs
	})
	return r
}

func (r *ContextRole) Revoke(resource string, verbs int64) *ContextRole {
	r.modifyHelper(resource, func(rule *ContextRoleRule) {
		rule.Verbs &= verbs ^ VerbAll
	})
	return r
}

func (r *ContextRole) Assign(resource string, verbs int64) *ContextRole {
	r.modifyHelper(resource, func(rule *ContextRoleRule) {
		rule.Verbs = verbs
	})
	return r
}

func (r *ContextRole) Update(db *gorm.DB) (errs []error) {
	errs = make([]error, 0)

	for _, rule := range r.Rules {
		if rule.dirty {
			if err := rule.Save(db); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) < 1 {
		return nil
	}
	return errs
}

func (r *ContextRole) Loaded() bool {
	return r.ID > 0
}

func (r *ContextRole) Load(db *gorm.DB) (err error) {
	role := &RoleModel{}
	err = db.Where(&RoleModel{Name: r.Name}).First(role).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			role.Name = r.Name
			role.ID = 0
			err = db.Save(role).Error
			if role.ID > 0 {
				err = nil
			}
		}
	}
	if err != nil {
		log.Errorf("[RBAC] Failed to load role \"%v\": %v", r.Name, err.Error())
		return err
	}
	r.ID = role.ID

	var records []RoleRecord
	if err = db.Where(&RoleRecord{RoleID: r.ID}).Find(&records).Error; err != nil {
		log.Errorf("[RBAC] Failed to load rules for role \"%v\": %v", r.Name, err.Error())
		return err
	}
	for _, record := range records {
		rule, ok := r.Rules[record.ResourceName]
		if !ok {
			rule = NewContextRoleRule(record.ResourceName, r)
			r.Rules[record.ResourceName] = rule
		}
		rule.Verbs = record.Verbs
	}

	return nil
}

// High-level RBAC interface.
type RBACContext struct {
	User     string
	UserID   int
	Roles    map[string]*ContextRole
	toGrant  []string
	toRevoke []string
}

func NewRBACContext(user string) *RBACContext {
	return &RBACContext{
		User:  user,
		Roles: make(map[string]*ContextRole),
	}
}

func (c *RBACContext) Loaded() bool {
	return c.UserID > 0
}

func (c *RBACContext) Load(db *gorm.DB) (err error) {
	var account Account
	if !c.Loaded() {
		if err = db.Where(&Account{Name: c.User}).First(&account).Error; err != nil { // Not found or error.
			return err
		}
		c.UserID = account.ID
	}

	var bindings []RoleBinding
	if err = db.Preload("Role").Where("account_id = ?", account.ID).Find(&bindings).Error; err != nil {
		return err
	}
	for _, binding := range bindings {
		role, ok := c.Roles[binding.Role.Name]
		if !ok {
			role = Role(binding.Role.Name)
			role.ID = binding.Role.ID
			c.Roles[binding.Role.Name] = role
		}
		if err = role.Load(db); err != nil {
			return err
		}
	}

	return nil
}

func (c *RBACContext) Grant(role *ContextRole) *RBACContext {
	if role == nil {
		return c
	}
	if c.toGrant == nil {
		c.toGrant = make([]string, 0)
	}
	c.Roles[role.Name] = role
	c.toGrant = append(c.toGrant, role.Name)
	return c
}

func (c *RBACContext) Revoke(role *ContextRole) *RBACContext {
	if role == nil {
		return c
	}
	if c.toRevoke == nil {
		c.toRevoke = make([]string, 0)
	}
	delete(c.Roles, role.Name)
	c.toRevoke = append(c.toRevoke, role.Name)
	return c
}

func (c *RBACContext) Update(db *gorm.DB) (err error) {
	if !c.Loaded() {
		if err = c.Load(db); err != nil {
			return err
		}
	}

	if c.toRevoke != nil || c.toGrant != nil {
		tx := db.Begin()
		if c.toGrant != nil {
			binding := &RoleBinding{}
			for _, roleName := range c.toGrant {
				binding.AccountID = c.UserID
				role, ok := c.Roles[roleName]
				if !ok {
					err = fmt.Errorf("[RBAC] Failed to grant role \"%v\" for the missing of role object.", roleName)
					log.Error(err.Error())
					tx.Rollback()
					return err
				}
				binding.RoleID = role.ID
				if err = tx.Save(binding).Error; err != nil {
					log.Errorf("[RBAC] Failed to grant role \"%v\": %v", roleName, err.Error())
				}
			}
		}

		if c.toRevoke != nil {
			if err = tx.Where("name in ?", c.toRevoke).Delete(RoleBinding{}).Error; err != nil {
				log.Errorf("[RBAC] Failed to revoke role \"%v\": %v", c.toRevoke, err.Error())
				tx.Rollback()
				return err
			}
		}
		tx.Commit()
	}

	return c.Load(db)
}

func (c *RBACContext) Permitted(verbs int64, resource string) bool {
	return true
}
