package ado

type ADOContext struct {
	ProjectName     string
	OrganizationUrl string
	accessToken     string
}

func NewADOContext(projectName, organizationUrl, accessToken string) ADOContext {
	return ADOContext{
		ProjectName:     projectName,
		OrganizationUrl: organizationUrl,
		accessToken:     accessToken,
	}
}

func (context *ADOContext) GetAccessToken() string {
	return context.accessToken
}
