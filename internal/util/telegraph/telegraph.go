package telegraph

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"gitlab.com/toby3d/telegraph"
)

type Telegraph struct {
	conf *Config
}

type Config struct {
	ShortName  string
	AuthorName string
}

type Option func(*Config)

func WithShortName(name string) Option {
	return func(c *Config) {
		c.ShortName = name
	}
}

func WithAuthorName(name string) Option {
	return func(c *Config) {
		c.AuthorName = name
	}
}

const (
	defaultShortName  = "VerseMaster"
	defaultAuthorName = "Yoan"
)

var cache *expirable.LRU[string, telegraph.Account]

var cacheOnce sync.Once

func newAccountCache() *expirable.LRU[string, telegraph.Account] {

	cacheOnce.Do(func() {
		cache = expirable.NewLRU[string, telegraph.Account](5, nil, time.Hour)
	})

	return cache
}

func New(opts ...Option) *Telegraph {
	conf := &Config{
		ShortName:  defaultShortName,
		AuthorName: defaultAuthorName,
	}
	for _, opt := range opts {
		opt(conf)
	}
	return &Telegraph{
		conf: conf,
	}
}

func (t *Telegraph) CreatePage(title, content, imgUrl string) (*telegraph.Page, error) {
	account, err := t.getAccount()
	if err != nil {
		return nil, err
	}
	nodes, err := telegraph.ContentFormat(content)
	if err != nil {
		return nil, err
	}
	rs, err := account.CreatePage(telegraph.Page{
		AuthorName: t.conf.AuthorName,
		Title:      title,
		Content:    nodes,
		ImageURL:   imgUrl,
	}, false)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func (t *Telegraph) getAccount() (*telegraph.Account, error) {

	cache := newAccountCache()
	if account, ok := cache.Get(t.conf.ShortName); ok {
		return &account, nil
	}

	account, err := telegraph.CreateAccount(telegraph.Account{
		ShortName:  t.conf.ShortName,
		AuthorName: t.conf.AuthorName,
	})
	if err != nil {
		return nil, err
	}

	cache.Add(t.conf.ShortName, *account)
	return account, nil
}
