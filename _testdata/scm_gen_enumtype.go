package scm

import (
	"fmt"

	"github.com/peter-edge/go-stringhelper"
)

type CheckoutOptionsType uint

var CheckoutOptionsTypeGit CheckoutOptionsType = 0
var CheckoutOptionsTypeGithub CheckoutOptionsType = 1
var CheckoutOptionsTypeHg CheckoutOptionsType = 2
var CheckoutOptionsTypeBitbucket CheckoutOptionsType = 3

var checkoutOptionsTypeToString = map[CheckoutOptionsType]string{
	CheckoutOptionsTypeGit:       "git",
	CheckoutOptionsTypeGithub:    "github",
	CheckoutOptionsTypeHg:        "hg",
	CheckoutOptionsTypeBitbucket: "bitbucket",
}

var stringToCheckoutOptionsType = map[string]CheckoutOptionsType{
	"git":       CheckoutOptionsTypeGit,
	"github":    CheckoutOptionsTypeGithub,
	"hg":        CheckoutOptionsTypeHg,
	"bitbucket": CheckoutOptionsTypeBitbucket,
}

func AllCheckoutOptionsTypes() []CheckoutOptionsType {
	return []CheckoutOptionsType{
		CheckoutOptionsTypeGit,
		CheckoutOptionsTypeGithub,
		CheckoutOptionsTypeHg,
		CheckoutOptionsTypeBitbucket,
	}
}

func CheckoutOptionsTypeOf(s string) (CheckoutOptionsType, error) {
	checkoutOptionsType, ok := stringToCheckoutOptionsType[s]
	if !ok {
		return 0, newErrorUnknownCheckoutOptionsType(s)
	}
	return checkoutOptionsType, nil
}

func (this CheckoutOptionsType) String() string {
	if int(this) < len(checkoutOptionsTypeToString) {
		return checkoutOptionsTypeToString[this]
	}
	panic(newErrorUnknownCheckoutOptionsType(this).Error())
}

type CheckoutOptions interface {
	fmt.Stringer
	Type() CheckoutOptionsType
}

func (this *GitCheckoutOptions) Type() CheckoutOptionsType {
	return CheckoutOptionsTypeGit
}

func (this *GithubCheckoutOptions) Type() CheckoutOptionsType {
	return CheckoutOptionsTypeGithub
}

func (this *HgCheckoutOptions) Type() CheckoutOptionsType {
	return CheckoutOptionsTypeHg
}

func (this *BitbucketCheckoutOptions) Type() CheckoutOptionsType {
	return CheckoutOptionsTypeBitbucket
}

func (this *GitCheckoutOptions) String() string {
	return stringhelper.String(this)
}

func (this *GithubCheckoutOptions) String() string {
	return stringhelper.String(this)
}

func (this *HgCheckoutOptions) String() string {
	return stringhelper.String(this)
}

func (this *BitbucketCheckoutOptions) String() string {
	return stringhelper.String(this)
}

func CheckoutOptionsSwitch(
	checkoutOptions CheckoutOptions,
	gitCheckoutOptionsFunc func(gitCheckoutOptions *GitCheckoutOptions) error,
	githubCheckoutOptionsFunc func(githubCheckoutOptions *GithubCheckoutOptions) error,
	hgCheckoutOptionsFunc func(hgCheckoutOptions *HgCheckoutOptions) error,
	bitbucketCheckoutOptionsFunc func(bitbucketCheckoutOptions *BitbucketCheckoutOptions) error,
) error {
	switch checkoutOptions.Type() {
	case CheckoutOptionsTypeGit:
		return gitCheckoutOptionsFunc(checkoutOptions.(*GitCheckoutOptions))
	case CheckoutOptionsTypeGithub:
		return githubCheckoutOptionsFunc(checkoutOptions.(*GithubCheckoutOptions))
	case CheckoutOptionsTypeHg:
		return hgCheckoutOptionsFunc(checkoutOptions.(*HgCheckoutOptions))
	case CheckoutOptionsTypeBitbucket:
		return bitbucketCheckoutOptionsFunc(checkoutOptions.(*BitbucketCheckoutOptions))
	default:
		return newErrorUnknownCheckoutOptionsType(checkoutOptions.Type())
	}
}

func (this CheckoutOptionsType) NewCheckoutOptions(
	gitCheckoutOptionsFunc func() (*GitCheckoutOptions, error),
	githubCheckoutOptionsFunc func() (*GithubCheckoutOptions, error),
	hgCheckoutOptionsFunc func() (*HgCheckoutOptions, error),
	bitbucketCheckoutOptionsFunc func() (*BitbucketCheckoutOptions, error),
) (CheckoutOptions, error) {
	switch this {
	case CheckoutOptionsTypeGit:
		return gitCheckoutOptionsFunc()
	case CheckoutOptionsTypeGithub:
		return githubCheckoutOptionsFunc()
	case CheckoutOptionsTypeHg:
		return hgCheckoutOptionsFunc()
	case CheckoutOptionsTypeBitbucket:
		return bitbucketCheckoutOptionsFunc()
	default:
		return nil, newErrorUnknownCheckoutOptionsType(this)
	}
}

func (this CheckoutOptionsType) Produce(
	checkoutOptionsTypeGitFunc func() (interface{}, error),
	checkoutOptionsTypeGithubFunc func() (interface{}, error),
	checkoutOptionsTypeHgFunc func() (interface{}, error),
	checkoutOptionsTypeBitbucketFunc func() (interface{}, error),
) (interface{}, error) {
	switch this {
	case CheckoutOptionsTypeGit:
		return checkoutOptionsTypeGitFunc()
	case CheckoutOptionsTypeGithub:
		return checkoutOptionsTypeGithubFunc()
	case CheckoutOptionsTypeHg:
		return checkoutOptionsTypeHgFunc()
	case CheckoutOptionsTypeBitbucket:
		return checkoutOptionsTypeBitbucketFunc()
	default:
		return nil, newErrorUnknownCheckoutOptionsType(this)
	}
}

func (this CheckoutOptionsType) Handle(
	checkoutOptionsTypeGitFunc func() error,
	checkoutOptionsTypeGithubFunc func() error,
	checkoutOptionsTypeHgFunc func() error,
	checkoutOptionsTypeBitbucketFunc func() error,
) error {
	switch this {
	case CheckoutOptionsTypeGit:
		return checkoutOptionsTypeGitFunc()
	case CheckoutOptionsTypeGithub:
		return checkoutOptionsTypeGithubFunc()
	case CheckoutOptionsTypeHg:
		return checkoutOptionsTypeHgFunc()
	case CheckoutOptionsTypeBitbucket:
		return checkoutOptionsTypeBitbucketFunc()
	default:
		return newErrorUnknownCheckoutOptionsType(this)
	}
}

func newErrorUnknownCheckoutOptionsType(value interface{}) error {
	return fmt.Errorf("scm: UnknownCheckoutOptionsType: %v", value)
}
