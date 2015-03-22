package testdata

//go:generate gen-enumtype

// @gen-enumtype CheckoutOptions git 0
type GitCheckoutOptions struct {
	User       string
	Repository string
	Branch     string
	CommitId   string
}

// @gen-enumtype CheckoutOptions bitbuckethg 1
type BitbucketHgCheckoutOptions struct {
	User        string
	Repository  string
	ChangesetId string
}
