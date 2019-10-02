package gitlab

import (
	"git.stuhome.com/Sunmxt/wing/common"
)

type UserIdentity struct {
	Provider  string `json:"provider"`
	ExternUID string `json:"extern_uid"`
}

type User struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
	CreateAt  string `json:"create_at"`
	IsAdmin   bool   `json:"is_admin"`
	// bio: null
	// location
	PublicEmail      string         `json:"public_email"`
	Skype            string         `json:"skype"`
	Linkedin         string         `json:"linkedin"`
	Twitter          string         `json:"twitter"`
	WebsiteURL       string         `json:"website_url"`
	Organization     string         `json:"organization"`
	LastSignInAt     string         `json:"last_sign_in_at"`
	ConfirmedAt      string         `json:"confirmed_at"`
	ThemeID          uint           `json:"theme_id"`
	LastActivityOn   string         `json:"last_activity_on"`
	ColorSchemeID    uint           `json:"color_scheme_id"`
	ProjectsLimit    uint           `json:"projects_limit"`
	CurrentSignInAt  string         `json:"current_sign_in_at"`
	Identities       []UserIdentity `json:"identities"`
	CanCreateGroup   bool           `json:"can_create_group"`
	CanCreateProject bool           `json:"can_create_project"`
	TwoFactorEnabled bool           `json:"two_factor_enabled"`
	External         bool           `json:"external"`
	PrivateProfile   bool           `json:"private_profile"`
}

type UserContext struct {
	Cursor GitlabPagination
	Client *GitlabClient
	Error  error
}

func NewUserContext(client *GitlabClient) *UserContext {
	return &UserContext{
		Client: client,
	}
}

//func (c *UserContext) parseUserDetail(resp *http.Response, user *User) *User {
//	body := make([]byte, resp.ContentLength)
//	if _, err := io.ReadFull(resp.Body, body); err != nil {
//		c.Client.Err("[Gitlab Client] read user detail response body failure: " + err.Error())
//		c.Error = err
//		return nil
//	}
//	if user == nil {
//		user = &User{}
//	}
//	if err := json.Unmarshal(body, user); err != nil {
//		c.Client.Err("[Gitlab Client] unmarshal user detail failure: " + err.Error())
//		c.Error = err
//		return nil
//	}
//	return user
//}

func (c *UserContext) Current() *User {
	if c.Client == nil || c.Client.Endpoint == nil {
		c.Error = common.ErrEndpointMissing
		return nil
	}
	req, err := c.Client.NewRequest("GET", "api/v4/user", nil)
	if err != nil {
		c.Error = err
		return nil
	}
	user := &User{}
	if _, err = c.Client.Do(req, user); err != nil {
		c.Error = err
		return nil
	}
	return user
}
