package pool

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/os/grpool"
)

type Job func(context.Context) interface{}

func New(ctx context.Context, jobs ...Job) <-chan interface{} {
	var wg sync.WaitGroup
	rs := make(chan interface{})

	wg.Add(len(jobs))

	for _, job := range jobs {
		job := job
		grpool.Add(ctx, func(ctx context.Context) {
			defer wg.Done()
			rs <- job(ctx)
		})
	}

	grpool.Add(ctx, func(ctx context.Context) {
		wg.Wait()
		close(rs)
	})

	return rs
}
