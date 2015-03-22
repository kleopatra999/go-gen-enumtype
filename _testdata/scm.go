package scm

//go:generate gen-enumtype

// @gen-enumtype CheckoutOptions git 0
type GitCheckoutOptions struct {
	User     string
	Host     string
	Path     string
	Branch   string
	CommitId string
}

// @gen-enumtype CheckoutOptions github 1
type GithubCheckoutOptions struct {
	User       string
	Repository string
	Branch     string
	CommitId   string
}

// @gen-enumtype CheckoutOptions hg 2
type HgCheckoutOptions struct {
	User        string
	Host        string
	Path        string
	ChangesetId string
}

// @gen-enumtype CheckoutOptions bitbucket 3
type BitbucketCheckoutOptions struct {
	User        string
	Repository  string
	Branch      string // only set if BitbucketType == BitbucketTypeGit
	CommitId    string // only set if BitbucketType == BitbucketTypeGit
	ChangesetId string // only set if BitbucketType == BitbucketTypeHg
}
